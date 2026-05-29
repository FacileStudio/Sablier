package nookpool

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
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

func getEnvPoolConfig() (string, string) {
	return os.Getenv("NOOK_POOL_URL"), os.Getenv("NOOK_POOL_SECRET")
}

func (s *Service) AutoConnect(ctx context.Context) {
	settings, fromEnv, err := s.getSettings(ctx)
	if err != nil {
		s.logger.Error("pool: failed to load settings", slog.Any("error", err))
		return
	}
	if settings.Enabled && settings.URL != "" && settings.Secret != "" {
		if err := s.connect(settings.URL, settings.Secret); err != nil {
			s.logger.Error("pool: auto-connect failed", slog.Any("error", err))
		}
		return
	}

	if fromEnv && settings.URL != "" && settings.Secret != "" {
		s.logger.Info("pool: using env vars for auto-connect")
		if err := s.connect(settings.URL, settings.Secret); err != nil {
			s.logger.Error("pool: auto-connect from env failed", slog.Any("error", err))
		}
	}
}

func (s *Service) getSettings(ctx context.Context) (*PoolSettings, bool, error) {
	var record schemas.AppSetting
	err := s.orm.WithContext(ctx).Where("id = ?", appSettingID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		envURL, envSecret := getEnvPoolConfig()
		if envURL != "" && envSecret != "" {
			return &PoolSettings{
				URL:     envURL,
				Secret:  envSecret,
				Enabled: false,
			}, true, nil
		}
		return &PoolSettings{}, false, nil
	}
	if err != nil {
		return nil, false, errors.Internal("failed to get pool settings", err)
	}
	return &PoolSettings{
		URL:     record.NookPoolURL,
		Secret:  record.NookPoolSecret,
		Enabled: record.NookPoolEnabled,
	}, false, nil
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
		App:      "Sablier",
		Instance: instanceURL,
		Secret:   secret,
		Events: pool.EventConfig{
			Emit:   []string{"project.created", "project.updated", "project.deleted", "task.created", "task.updated", "task.deleted", "timer.started", "timer.stopped"},
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

func (s *Service) IsPoolEventEnabled(event string) bool {
	var record schemas.AppSetting
	if err := s.orm.Where("id = ?", appSettingID).First(&record).Error; err != nil {
		return true
	}
	if record.PoolEvents == "" {
		return true
	}
	var toggles map[string]bool
	if err := json.Unmarshal([]byte(record.PoolEvents), &toggles); err != nil {
		return true
	}
	enabled, exists := toggles[event]
	if !exists {
		return true
	}
	return enabled
}

func (s *Service) getPoolEvents(ctx context.Context) (*PoolEventsResponse, error) {
	var record schemas.AppSetting
	s.orm.WithContext(ctx).Where("id = ?", appSettingID).First(&record)

	toggles := make(map[string]bool)
	if record.PoolEvents != "" {
		json.Unmarshal([]byte(record.PoolEvents), &toggles)
	}

	events := make([]PoolEventToggle, len(AllPoolEvents))
	for i, evt := range AllPoolEvents {
		enabled, exists := toggles[evt]
		if !exists {
			enabled = true
		}
		events[i] = PoolEventToggle{Event: evt, Enabled: enabled}
	}
	return &PoolEventsResponse{Events: events}, nil
}

func (s *Service) updatePoolEvents(ctx context.Context, req *UpdatePoolEventsRequest) (*PoolEventsResponse, error) {
	toggles := make(map[string]bool)
	for _, evt := range req.Events {
		toggles[evt.Event] = evt.Enabled
	}
	data, err := json.Marshal(toggles)
	if err != nil {
		return nil, errors.Internal("failed to serialize pool events", err)
	}

	s.orm.WithContext(ctx).Model(&schemas.AppSetting{}).Where("id = ?", appSettingID).Update("pool_events", string(data))

	return s.getPoolEvents(ctx)
}

func (s *Service) EmitTimerEvent(event string, payload *TimerEventPayload) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil || !client.IsConnected() {
		return
	}

	if !s.IsPoolEventEnabled(event) {
		return
	}

	evt := TimerEvent{
		App:            "sablier",
		Event:          event,
		Payload:        *payload,
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_%s_%d_%d", event, payload.ID, time.Now().UnixMilli()),
	}

	if err := client.Emit(event, evt); err != nil {
		s.logger.Error("pool: failed to emit timer event", slog.Any("error", err), slog.String("event", event))
	}
}

func (s *Service) EmitProjectEvent(action enveloppe.Action, project *schemas.Project) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil || !client.IsConnected() {
		return
	}

	channel := fmt.Sprintf("project.%s", action)
	if !s.IsPoolEventEnabled(channel) {
		return
	}

	if project.FacileID == nil {
		fid := GenerateFacileID()
		project.FacileID = &fid
		s.orm.Model(project).Update("facile_id", fid)
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
			Icon:        normalizeIcon(project.Icon),
		},
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_project_%s_%s_%d", action, *project.FacileID, time.Now().UnixMilli()),
	}

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

	if !s.IsPoolEventEnabled(fmt.Sprintf("task.%s", action)) {
		return
	}

	if task.FacileID == nil {
		fid := GenerateFacileID()
		task.FacileID = &fid
		s.orm.Model(task).Update("facile_id", fid)
	}

	if project != nil && project.FacileID == nil {
		fid := GenerateFacileID()
		project.FacileID = &fid
		s.orm.Model(project).Update("facile_id", fid)
		s.EmitProjectEvent(enveloppe.ActionCreated, project)
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
			Status:          normalizeStatus(task.Status),
		},
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_task_%s_%s_%d", action, *task.FacileID, time.Now().UnixMilli()),
	}

	channel := fmt.Sprintf("task.%s", action)
	if err := client.Emit(channel, evt); err != nil {
		s.logger.Error("pool: failed to emit task event", slog.Any("error", err), slog.String("action", string(action)))
	}
}

func (s *Service) InitialSync(ctx context.Context) (*SyncResult, error) {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil || !client.IsConnected() {
		return nil, errors.Failed("pool is not connected")
	}

	var unflaggedProjects []schemas.Project
	s.orm.WithContext(ctx).Where("facile_id IS NULL").Find(&unflaggedProjects)
	for i := range unflaggedProjects {
		fid := GenerateFacileID()
		unflaggedProjects[i].FacileID = &fid
		s.orm.WithContext(ctx).Save(&unflaggedProjects[i])
	}

	var allProjects []schemas.Project
	s.orm.WithContext(ctx).Find(&allProjects)
	projectCount := 0
	for i := range allProjects {
		if allProjects[i].FacileID == nil {
			continue
		}
		s.EmitProjectEvent(enveloppe.ActionCreated, &allProjects[i])
		projectCount++
		if projectCount%50 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	if projectCount > 0 {
		time.Sleep(2 * time.Second)
	}

	var unflaggedTasks []schemas.Task
	s.orm.WithContext(ctx).Where("facile_id IS NULL").Find(&unflaggedTasks)
	for i := range unflaggedTasks {
		fid := GenerateFacileID()
		unflaggedTasks[i].FacileID = &fid
		s.orm.WithContext(ctx).Save(&unflaggedTasks[i])
	}

	var allTasks []schemas.Task
	s.orm.WithContext(ctx).Order("project_id ASC").Find(&allTasks)
	taskCount := 0
	for i := range allTasks {
		if allTasks[i].FacileID == nil {
			continue
		}
		var project schemas.Project
		if err := s.orm.WithContext(ctx).Where("id = ? AND facile_id IS NOT NULL", allTasks[i].ProjectID).First(&project).Error; err != nil {
			continue
		}
		s.EmitTaskEvent(enveloppe.ActionCreated, &allTasks[i], &project)
		taskCount++
		if taskCount%50 == 0 {
			time.Sleep(100 * time.Millisecond)
		}
	}

	return &SyncResult{
		ProjectsSynced: projectCount,
		TasksSynced:    taskCount,
	}, nil
}

var validStatuses = map[string]bool{
	"to-do":       true,
	"in-progress": true,
	"in-review":   true,
	"done":        true,
}

func normalizeStatus(status string) string {
	if validStatuses[status] {
		return status
	}
	return "to-do"
}

func GenerateFacileID() string {
	b := make([]byte, 10)
	rand.Read(b)
	return "fac_" + hex.EncodeToString(b)
}

var iconTypoMap = map[string]string{
	"Pallete2":                    "Palette2",
	"Siderbar":                    "Sidebar",
	"Magnifer":                    "Magnifier",
	"MagniferBug":                 "MagnifierBug",
	"MagniferZoomIn":              "MagnifierZoomIn",
	"MagniferZoomOut":             "MagnifierZoomOut",
	"MinimalisticMagnifer":        "MinimalisticMagnifier",
	"MinimalisticMagniferBug":     "MinimalisticMagnifierBug",
	"MinimalisticMagniferZoomIn":  "MinimalisticMagnifierZoomIn",
	"MinimalisticMagniferZoomOut": "MinimalisticMagnifierZoomOut",
	"RoundedMagnifer":             "RoundedMagnifier",
	"RoundedMagniferBug":          "RoundedMagnifierBug",
	"RoundedMagniferZoomIn":       "RoundedMagnifierZoomIn",
	"RoundedMagniferZoomOut":      "RoundedMagnifierZoomOut",
	"Condicioner":                 "Conditioner",
	"Condicioner2":                "Conditioner2",
	"ColourTuneing":               "ColourTuning",
	"MaskHapply":                  "MaskHappy",
	"SpedometerLow":               "SpeedometerLow",
	"SpedometerMax":               "SpeedometerMax",
	"SpedometerMiddle":            "SpeedometerMiddle",
	"CardRecive":                  "CardReceive",
	"ReciveSquare":                "ReceiveSquare",
	"ReciveTwiceSquare":           "ReceiveTwiceSquare",
	"PlaaylistMinimalistic":       "PlaylistMinimalistic",
}

func normalizeIcon(icon *string) *string {
	if icon == nil {
		return nil
	}
	if corrected, ok := iconTypoMap[*icon]; ok {
		return &corrected
	}
	return icon
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
