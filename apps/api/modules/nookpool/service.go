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
			Emit:   []string{"project.created", "project.updated", "project.deleted", "task.created", "task.updated", "task.deleted", "time_entry.created", "time_entry.updated"},
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

// shouldEmit reports whether events should be produced at all: the pool is
// enabled in settings, or a client exists (env-var configs run with
// Enabled=false and no settings row, so a client — even one mid-reconnect —
// is the signal). Events are queued in the outbox, not sent directly, so a
// temporarily disconnected pool must not drop them.
func (s *Service) shouldEmit() bool {
	s.mu.RLock()
	hasClient := s.client != nil
	s.mu.RUnlock()
	if hasClient {
		return true
	}
	var record schemas.AppSetting
	if err := s.orm.Where("id = ?", appSettingID).First(&record).Error; err != nil {
		return false
	}
	return record.NookPoolEnabled
}

func (s *Service) enqueueOutbox(channel string, evt any) {
	payload, err := json.Marshal(evt)
	if err != nil {
		s.logger.Error("pool: failed to serialize outbox event", slog.Any("error", err), slog.String("channel", channel))
		return
	}
	row := schemas.PoolOutbox{Channel: channel, Payload: string(payload)}
	if err := s.orm.Create(&row).Error; err != nil {
		s.logger.Error("pool: failed to enqueue outbox event", slog.Any("error", err), slog.String("channel", channel))
	}
}

const outboxBatchSize = 100

// RunOutboxWorker drains the outbox to the pool while connected. Failures
// stop the batch (never skip: skipping would reorder the log) and retry on
// the next tick. It also prunes the idempotency ledger daily.
func (s *Service) RunOutboxWorker(ctx context.Context) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()
	lastPrune := time.Time{}

	for {
		select {
		case <-ticker.C:
			s.drainOutbox()
			if time.Since(lastPrune) > 24*time.Hour {
				s.orm.Where("processed_at < ?", time.Now().Add(-35*24*time.Hour)).Delete(&schemas.PoolProcessedEvent{})
				lastPrune = time.Now()
			}
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) drainOutbox() {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()

	if client == nil || !client.IsConnected() {
		return
	}

	var rows []schemas.PoolOutbox
	if err := s.orm.Order("id ASC").Limit(outboxBatchSize).Find(&rows).Error; err != nil {
		s.logger.Error("pool: failed to read outbox", slog.Any("error", err))
		return
	}

	for i := range rows {
		if err := client.EmitNow(rows[i].Channel, json.RawMessage(rows[i].Payload)); err != nil {
			s.logger.Error("pool: outbox emit failed", slog.Any("error", err), slog.String("channel", rows[i].Channel))
			s.orm.Model(&rows[i]).Updates(map[string]interface{}{
				"attempts":   rows[i].Attempts + 1,
				"last_error": err.Error(),
			})
			return
		}
		s.orm.Delete(&rows[i])
	}
}

func (s *Service) EmitTimeEntryEvent(action enveloppe.Action, entry *schemas.TimeEntry) {
	if !s.shouldEmit() {
		return
	}

	channel := fmt.Sprintf("time_entry.%s", action)
	if !s.IsPoolEventEnabled(channel) {
		return
	}

	e := *entry
	if e.FacileID == nil {
		fid := GenerateFacileID()
		e.FacileID = &fid
		s.orm.Model(&schemas.TimeEntry{}).Where("id = ?", e.ID).Update("facile_id", fid)
	}

	var taskFacileID string
	var task schemas.Task
	if err := s.orm.Where("id = ?", e.TaskID).First(&task).Error; err == nil {
		if task.FacileID == nil {
			fid := GenerateFacileID()
			task.FacileID = &fid
			s.orm.Model(&task).Update("facile_id", fid)
		}
		taskFacileID = *task.FacileID
	}

	var userEmail string
	var user schemas.User
	if err := s.orm.Select("email").Where("id = ?", e.UserID).First(&user).Error; err == nil {
		userEmail = user.Email
	}

	var stoppedAt *string
	if e.StoppedAt != nil {
		formatted := e.StoppedAt.UTC().Format(time.RFC3339)
		stoppedAt = &formatted
	}

	evt := enveloppe.Event[enveloppe.TimeEntry]{
		Version:  enveloppe.EventVersion,
		App:      enveloppe.AppSablier,
		Object:   enveloppe.ObjectTimeEntry,
		Action:   action,
		FacileID: *e.FacileID,
		Payload: enveloppe.TimeEntry{
			FacileID:     *e.FacileID,
			TaskFacileID: taskFacileID,
			UserEmail:    userEmail,
			StartedAt:    e.StartedAt.UTC().Format(time.RFC3339),
			StoppedAt:    stoppedAt,
		},
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_time_entry_%s_%s_%d", action, *e.FacileID, time.Now().UnixMilli()),
	}

	s.enqueueOutbox(channel, evt)
}

func (s *Service) EmitProjectEvent(action enveloppe.Action, project *schemas.Project) {
	if !s.shouldEmit() {
		return
	}

	channel := fmt.Sprintf("project.%s", action)
	if !s.IsPoolEventEnabled(channel) {
		return
	}

	p := *project
	if p.FacileID == nil {
		fid := GenerateFacileID()
		p.FacileID = &fid
		s.orm.Model(&schemas.Project{}).Where("id = ?", p.ID).Update("facile_id", fid)
	}

	var desc *string
	if p.Description != "" {
		desc = &p.Description
	}

	evt := enveloppe.Event[enveloppe.Project]{
		Version:  enveloppe.EventVersion,
		App:      enveloppe.AppSablier,
		Object:   enveloppe.ObjectProject,
		Action:   action,
		FacileID: *p.FacileID,
		Payload: enveloppe.Project{
			FacileID:    *p.FacileID,
			Name:        p.Name,
			Description: desc,
			Icon:        normalizeIcon(p.Icon),
		},
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_project_%s_%s_%d", action, *p.FacileID, time.Now().UnixMilli()),
	}

	s.enqueueOutbox(channel, evt)
}

func (s *Service) EmitTaskEvent(action enveloppe.Action, task *schemas.Task, project *schemas.Project) {
	if !s.shouldEmit() {
		return
	}

	if !s.IsPoolEventEnabled(fmt.Sprintf("task.%s", action)) {
		return
	}

	t := *task
	if t.FacileID == nil {
		fid := GenerateFacileID()
		t.FacileID = &fid
		s.orm.Model(&schemas.Task{}).Where("id = ?", t.ID).Update("facile_id", fid)
	}

	projectFacileID := ""
	if project != nil {
		p := *project
		if p.FacileID == nil {
			fid := GenerateFacileID()
			p.FacileID = &fid
			s.orm.Model(&schemas.Project{}).Where("id = ?", p.ID).Update("facile_id", fid)
			s.EmitProjectEvent(enveloppe.ActionCreated, &p)
		}
		projectFacileID = *p.FacileID
	}

	var actorEmail string
	if t.ActorID != nil {
		var user schemas.User
		if err := s.orm.Select("email").Where("id = ?", *t.ActorID).First(&user).Error; err == nil {
			actorEmail = user.Email
		}
	}

	evt := enveloppe.Event[enveloppe.Task]{
		Version:  enveloppe.EventVersion,
		App:      enveloppe.AppSablier,
		Object:   enveloppe.ObjectTask,
		Action:   action,
		FacileID: *t.FacileID,
		Payload: enveloppe.Task{
			FacileID:        *t.FacileID,
			ProjectFacileID: projectFacileID,
			Name:            t.Name,
			Status:          schemas.NormalizeStatus(t.Status),
			ActorEmail:      actorEmail,
		},
		Timestamp:      time.Now().UTC().Format(time.RFC3339),
		IdempotencyKey: fmt.Sprintf("sablier_task_%s_%s_%d", action, *t.FacileID, time.Now().UnixMilli()),
	}

	s.enqueueOutbox(fmt.Sprintf("task.%s", action), evt)
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
