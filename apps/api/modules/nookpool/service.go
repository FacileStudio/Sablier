package nookpool

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"strings"
	"sync"
	"time"

	"api/internal/errors"
	"api/schemas"

	pool "github.com/FacileStudio/pool/go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	stderrors "errors"

	enveloppe "github.com/FacileStudio/enveloppe/go"
)

const appSettingID = 1

type Service struct {
	orm        *gorm.DB
	client     *pool.Client
	controller *Controller
	mu         sync.RWMutex
	logger     *slog.Logger
}

func NewService(orm *gorm.DB, logger *slog.Logger) *Service {
	service := &Service{orm: orm, logger: logger}
	service.controller = newController(service)
	return service
}

func (s *Service) AutoConnect(ctx context.Context) {
	settings, err := s.getSettings(ctx)
	if err != nil {
		s.logger.Error("pool: failed to load settings", slog.Any("error", err))
		return
	}
	if !settings.Enabled || settings.URL == "" || settings.Secret == "" {
		return
	}
	if err := s.connect(settings.URL, settings.Secret); err != nil {
		s.logger.Error("pool: auto-connect failed", slog.Any("error", err))
	}
}

func (s *Service) getSettings(ctx context.Context) (*PoolSettings, error) {
	var record schemas.AppSetting
	err := s.orm.WithContext(ctx).Where("id = ?", appSettingID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return &PoolSettings{}, nil
	}
	if err != nil {
		return nil, errors.Internal("failed to get pool settings", err)
	}
	return &PoolSettings{
		URL:     record.NookPoolURL,
		Secret:  record.NookPoolSecret,
		Enabled: record.NookPoolEnabled,
	}, nil
}

func (s *Service) updateSettings(ctx context.Context, req *UpdatePoolRequest) (*PoolSettings, string, error) {
	record := schemas.AppSetting{
		ID:              appSettingID,
		NookPoolURL:     strings.TrimSpace(req.URL),
		NookPoolSecret:  strings.TrimSpace(req.Secret),
		NookPoolEnabled: req.Enabled,
	}
	if err := s.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"nook_pool_url", "nook_pool_secret", "nook_pool_enabled"}),
	}).Create(&record).Error; err != nil {
		return nil, "", errors.Internal("failed to update pool settings", err)
	}

	var connectErr string
	if req.Enabled && req.URL != "" && req.Secret != "" {
		if err := s.connect(req.URL, req.Secret); err != nil {
			s.logger.Error("pool: connect failed after settings update", slog.Any("error", err))
			connectErr = err.Error()
		}
	} else {
		s.disconnect()
	}

	return &PoolSettings{
		URL:     record.NookPoolURL,
		Secret:  record.NookPoolSecret,
		Enabled: record.NookPoolEnabled,
	}, connectErr, nil
}

func (s *Service) connect(instanceURL, secret string) error {
	s.disconnect()

	cfg := &pool.Config{
		App:      "sablier",
		Instance: instanceURL,
		Secret:   secret,
		Events: pool.EventConfig{
			Emit:   []string{"project.created", "project.updated", "project.deleted", "task.created", "task.updated", "task.deleted"},
			Listen: []string{"project.created", "project.updated", "project.deleted", "task.created", "task.updated", "task.deleted"},
		},
	}

	client := pool.NewClient(cfg,
		pool.WithOnConnect(func() {
			s.logger.Info("pool: connected")
		}),
		pool.WithOnDisconnect(func() {
			s.logger.Info("pool: disconnected")
		}),
		pool.WithOnError(func(err error) {
			s.logger.Error("pool: error", slog.Any("error", err))
		}),
	)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := client.Connect(ctx); err != nil {
		return err
	}

	s.mu.Lock()
	s.client = client
	s.mu.Unlock()

	s.setupListeners()
	s.logger.Info("pool: connected and listeners registered")
	return nil
}

func (s *Service) disconnect() {
	s.mu.Lock()
	client := s.client
	s.client = nil
	s.mu.Unlock()

	if client != nil {
		client.Disconnect()
	}
}

func (s *Service) isConnected() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.client == nil {
		return false
	}
	return s.client.IsConnected()
}

func (s *Service) Shutdown() {
	s.disconnect()
}

func (s *Service) EmitProjectEvent(action enveloppe.Action, project *schemas.Project) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil || !client.IsConnected() {
		return
	}

	if project.FacileID == nil {
		return
	}

	var desc *string
	if project.Description != "" {
		desc = &project.Description
	}

	evt := enveloppe.Event[enveloppe.Project]{
		App:      enveloppe.AppSablier,
		Object:   enveloppe.ObjectProject,
		Action:   action,
		FacileID: *project.FacileID,
		Payload: enveloppe.Project{
			FacileID:    *project.FacileID,
			Name:        project.Name,
			Description: desc,
		},
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_project_%s_%s_%d", action, *project.FacileID, time.Now().UnixMilli()),
	}

	channel := fmt.Sprintf("project.%s", action)
	if err := client.Emit(channel, evt); err != nil {
		s.logger.Error("pool: failed to emit project event", slog.Any("error", err), slog.String("action", string(action)))
	}
}

func (s *Service) EmitTaskEvent(action enveloppe.Action, task *schemas.Task, project *schemas.Project) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil || !client.IsConnected() {
		return
	}

	if task.FacileID == nil {
		return
	}

	projectFacileID := ""
	if project != nil && project.FacileID != nil {
		projectFacileID = *project.FacileID
	}

	evt := enveloppe.Event[enveloppe.Task]{
		App:      enveloppe.AppSablier,
		Object:   enveloppe.ObjectTask,
		Action:   action,
		FacileID: *task.FacileID,
		Payload: enveloppe.Task{
			FacileID:        *task.FacileID,
			ProjectFacileID: projectFacileID,
			Name:            task.Name,
		},
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_task_%s_%s_%d", action, *task.FacileID, time.Now().UnixMilli()),
	}

	channel := fmt.Sprintf("task.%s", action)
	if err := client.Emit(channel, evt); err != nil {
		s.logger.Error("pool: failed to emit task event", slog.Any("error", err), slog.String("action", string(action)))
	}
}

func GenerateFacileID() string {
	b := make([]byte, 10)
	rand.Read(b)
	return "fac_" + hex.EncodeToString(b)
}

func (s *Service) setupListeners() {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil {
		return
	}

	client.Listen("project.created", func(payload json.RawMessage, meta pool.EventMeta) {
		s.handleProjectCreated(payload, meta)
	})
	client.Listen("project.updated", func(payload json.RawMessage, meta pool.EventMeta) {
		s.handleProjectUpdated(payload, meta)
	})
	client.Listen("project.deleted", func(payload json.RawMessage, meta pool.EventMeta) {
		s.handleProjectDeleted(payload, meta)
	})
	client.Listen("task.created", func(payload json.RawMessage, meta pool.EventMeta) {
		s.handleTaskCreated(payload, meta)
	})
	client.Listen("task.updated", func(payload json.RawMessage, meta pool.EventMeta) {
		s.handleTaskUpdated(payload, meta)
	})
	client.Listen("task.deleted", func(payload json.RawMessage, meta pool.EventMeta) {
		s.handleTaskDeleted(payload, meta)
	})
}
