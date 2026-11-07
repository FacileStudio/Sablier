package antenne

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	antenneclient "github.com/FacileStudio/antenne-client/go"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const appSettingID = 1

// EmitEvents and ListenEvents are the pool registration lists; antenne.yaml
// documents them and a test keeps both in step.
func EmitEvents() []string {
	return []string{"project.created", "project.updated", "project.deleted", "task.created", "task.updated", "task.deleted", "time_entry.created", "time_entry.updated", "time_entry.deleted"}
}

func ListenEvents() []string {
	return []string{"project.created", "project.updated", "project.deleted", "task.created", "task.updated", "task.deleted", "agent_session.created", "agent_session.updated"}
}

// Service manages the Antenne pool connection, its settings, and the
// event outbox.
type Service struct {
	orm        *gorm.DB
	client     *antenneclient.Client
	controller *Controller
	mu         sync.RWMutex
	logger     *slog.Logger
	issuer     string
	resolver   attribution.Resolver
}

// NewService constructs a Service bound to orm and logger.
func NewService(orm *gorm.DB, logger *slog.Logger) *Service {
	service := &Service{orm: orm, logger: logger}
	service.controller = newController(service)
	return service
}

// SetResolver sets the attribution resolver Mycelium agent sessions are booked through.
func (s *Service) SetResolver(resolver attribution.Resolver) {
	s.resolver = resolver
}

// SetIssuer names the OIDC issuer whose porte subjects identify people on the wire.
func (s *Service) SetIssuer(issuer string) {
	s.issuer = issuer
}

func getEnvPoolConfig() (string, string) {
	return os.Getenv("ANTENNE_URL"), os.Getenv("ANTENNE_SECRET")
}

func (s *Service) AutoConnect(ctx context.Context) {
	settings, fromEnv, err := s.getSettings(ctx)
	if err != nil {
		s.logger.Error("antenne: failed to load settings", slog.Any("error", err))
		return
	}
	if settings.Enabled && settings.URL != "" && settings.Secret != "" {
		if err := s.connect(settings.URL, settings.Secret); err != nil {
			s.logger.Error("antenne: auto-connect failed", slog.Any("error", err))
		}
		return
	}

	if fromEnv && settings.URL != "" && settings.Secret != "" {
		s.logger.Info("antenne: using env vars for auto-connect")
		if err := s.connect(settings.URL, settings.Secret); err != nil {
			s.logger.Error("antenne: auto-connect from env failed", slog.Any("error", err))
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
		URL:     record.AntenneURL,
		Secret:  record.AntenneSecret,
		Enabled: record.AntenneEnabled,
	}, false, nil
}

func (s *Service) updateSettings(ctx context.Context, req *UpdatePoolRequest) (*PoolSettings, string, error) {
	record := schemas.AppSetting{
		ID:             appSettingID,
		AntenneURL:     strings.TrimSpace(req.URL),
		AntenneSecret:  strings.TrimSpace(req.Secret),
		AntenneEnabled: req.Enabled,
	}
	if err := s.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"antenne_url", "antenne_secret", "antenne_enabled"}),
	}).Create(&record).Error; err != nil {
		return nil, "", errors.Internal("failed to update pool settings", err)
	}

	var connectErr string
	if req.Enabled && req.URL != "" && req.Secret != "" {
		if err := s.connect(req.URL, req.Secret); err != nil {
			s.logger.Error("antenne: connect failed after settings update", slog.Any("error", err))
			connectErr = err.Error()
		}
	} else {
		s.disconnect()
	}

	return &PoolSettings{
		URL:     record.AntenneURL,
		Secret:  record.AntenneSecret,
		Enabled: record.AntenneEnabled,
	}, connectErr, nil
}

// connect establishes (or re-establishes) the Antenne client connection
// using the given instance URL and secret, wiring up event listeners on
// success.
//
// A reconnect the client is going to retry is the mechanism working, not an
// incident: an Antenne restart takes a few seconds and produces three of
// these. Logging them at error painted the dashboard red on every deploy,
// which teaches people to scroll past the colour that is supposed to mean
// something, so transient errors are logged at warn and only the terminal
// failure lands at error.
func (s *Service) connect(instanceURL, secret string) error {
	s.disconnect()

	cfg := &antenneclient.Config{
		App:      "Sablier",
		Instance: instanceURL,
		Secret:   secret,
		Events: antenneclient.EventConfig{
			Emit:   EmitEvents(),
			Listen: ListenEvents(),
		},
	}

	client := antenneclient.NewClient(cfg,
		antenneclient.WithOnConnect(func() {
			s.logger.Info("antenne: connected")
		}),
		antenneclient.WithOnDisconnect(func() {
			s.logger.Info("antenne: disconnected")
		}),
		antenneclient.WithOnError(func(err error) {
			var transient *antenneclient.TransientError
			if stderrors.As(err, &transient) {
				s.logger.Warn("antenne: reconnecting",
					slog.Int("attempt", transient.Attempt),
					slog.Any("error", transient.Err))
				return
			}
			s.logger.Error("antenne: error", slog.Any("error", err))
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
	s.logger.Info("antenne: connected and listeners registered")
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
	return record.AntenneEnabled
}

func (s *Service) enqueueOutbox(channel string, evt any) {
	payload, err := json.Marshal(evt)
	if err != nil {
		s.logger.Error("antenne: failed to serialize outbox event", slog.Any("error", err), slog.String("channel", channel))
		return
	}
	row := schemas.AntenneOutbox{Channel: channel, Payload: string(payload)}
	if err := s.orm.Create(&row).Error; err != nil {
		s.logger.Error("antenne: failed to enqueue outbox event", slog.Any("error", err), slog.String("channel", channel))
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
				s.orm.Where("processed_at < ?", time.Now().Add(-ledgerRetention)).Delete(&schemas.AntenneProcessedEvent{})
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

	var rows []schemas.AntenneOutbox
	if err := s.orm.Order("id ASC").Limit(outboxBatchSize).Find(&rows).Error; err != nil {
		s.logger.Error("antenne: failed to read outbox", slog.Any("error", err))
		return
	}

	for i := range rows {
		if err := client.EmitNow(rows[i].Channel, json.RawMessage(rows[i].Payload)); err != nil {
			s.logger.Error("antenne: outbox emit failed", slog.Any("error", err), slog.String("channel", rows[i].Channel))
			s.orm.Model(&rows[i]).Updates(map[string]interface{}{
				"attempts":   rows[i].Attempts + 1,
				"last_error": err.Error(),
			})
			return
		}
		s.orm.Delete(&rows[i])
	}
}

// GenerateFacileID returns a random "fac_"-prefixed hex identifier.
func GenerateFacileID() string {
	return schemas.NewFacileID()
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

	client.Listen("project.created", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleProjectCreated(payload, meta)
	})
	client.Listen("project.updated", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleProjectUpdated(payload, meta)
	})
	client.Listen("project.deleted", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleProjectDeleted(payload, meta)
	})
	client.Listen("task.created", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleTaskCreated(payload, meta)
	})
	client.Listen("task.updated", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleTaskUpdated(payload, meta)
	})
	client.Listen("task.deleted", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleTaskDeleted(payload, meta)
	})
	client.Listen("agent_session.created", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleAgentSession(payload, meta)
	})
	client.Listen("agent_session.updated", func(payload json.RawMessage, meta antenneclient.EventMeta) {
		s.handleAgentSession(payload, meta)
	})
}
