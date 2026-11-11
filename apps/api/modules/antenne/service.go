package antenne

import (
	"context"
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	antenneclient "github.com/FacileStudio/antenne-client/go"
	"github.com/FacileStudio/antenne-client/go/catalog"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const appSettingID = 1

// EmitEvents is the channels Sablier publishes. It and ListenEvents are the
// bus registration lists; antenne.yaml documents them and a test keeps all
// three in step.
func EmitEvents() []string {
	return names(catalog.ProjectCreated, catalog.ProjectUpdated, catalog.ProjectDeleted,
		catalog.TaskCreated, catalog.TaskUpdated, catalog.TaskDeleted,
		catalog.TimeEntryCreated, catalog.TimeEntryUpdated, catalog.TimeEntryDeleted)
}

// ListenEvents is the channels Sablier handles, the keys of listeners.
func ListenEvents() []string {
	return names(catalog.ProjectCreated, catalog.ProjectUpdated, catalog.ProjectDeleted,
		catalog.TaskCreated, catalog.TaskUpdated, catalog.TaskDeleted,
		catalog.AgentSessionCreated, catalog.AgentSessionUpdated)
}

func names(channels ...catalog.Channel) []string {
	out := make([]string, len(channels))
	for i, channel := range channels {
		out[i] = string(channel)
	}
	return out
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

// getEnvPoolConfig reads the bus from the environment through
// antenne-client's ConfigFromEnv, which takes ANTENNE_URL or its older synonym
// ANTENNE_INSTANCE. A URL without a secret configures nothing, as before.
func getEnvPoolConfig() (string, string) {
	cfg, err := antenneclient.ConfigFromEnv("Sablier")
	if err != nil {
		return "", ""
	}
	return cfg.Instance, cfg.Secret
}

func (s *Service) AutoConnect(ctx context.Context) {
	settings, fromEnv, err := s.getSettings(ctx)
	if err != nil {
		s.logger.Error("antenne: failed to load settings", slog.Any("error", err))
		return
	}
	if settings.Enabled && settings.URL != "" && settings.Secret != "" {
		if err := s.connect(settings.URL, settings.Secret); err != nil {
			s.logger.Warn("antenne: auto-connect failed, retrying", slog.Any("error", err))
		}
		return
	}

	if fromEnv && settings.URL != "" && settings.Secret != "" {
		s.logger.Info("antenne: using env vars for auto-connect")
		if err := s.connect(settings.URL, settings.Secret); err != nil {
			s.logger.Warn("antenne: auto-connect from env failed, retrying", slog.Any("error", err))
		}
	}
}

// poolConfig is the stored connection, secret included. It stays inside the
// package: the API shows PoolSettings, which carries no secret.
type poolConfig struct {
	URL     string
	Secret  string
	Enabled bool
}

func (c poolConfig) view() PoolSettings {
	return PoolSettings{URL: c.URL, SecretSet: c.Secret != "", Enabled: c.Enabled}
}

func (s *Service) getSettings(ctx context.Context) (*poolConfig, bool, error) {
	var record schemas.AppSetting
	err := s.orm.WithContext(ctx).Where("id = ?", appSettingID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		envURL, envSecret := getEnvPoolConfig()
		if envURL != "" && envSecret != "" {
			return &poolConfig{
				URL:     envURL,
				Secret:  envSecret,
				Enabled: false,
			}, true, nil
		}
		return &poolConfig{}, false, nil
	}
	if err != nil {
		return nil, false, errors.Internal("failed to get pool settings", err)
	}
	return &poolConfig{
		URL:     record.AntenneURL,
		Secret:  record.AntenneSecret,
		Enabled: record.AntenneEnabled,
	}, false, nil
}

func (s *Service) updateSettings(ctx context.Context, req *UpdatePoolRequest) (*poolConfig, string, error) {
	current, _, err := s.getSettings(ctx)
	if err != nil {
		return nil, "", err
	}
	next := poolConfig{URL: strings.TrimSpace(req.URL), Secret: current.Secret, Enabled: req.Enabled}
	if req.Secret != nil {
		next.Secret = strings.TrimSpace(*req.Secret)
	}
	record := schemas.AppSetting{
		ID:             appSettingID,
		AntenneURL:     next.URL,
		AntenneSecret:  next.Secret,
		AntenneEnabled: next.Enabled,
	}
	if err := s.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"antenne_url", "antenne_secret", "antenne_enabled"}),
	}).Create(&record).Error; err != nil {
		return nil, "", errors.Internal("failed to update pool settings", err)
	}

	var connectErr string
	if next.Enabled && next.URL != "" && next.Secret != "" {
		if err := s.connect(next.URL, next.Secret); err != nil {
			s.logger.Error("antenne: connect failed after settings update", slog.Any("error", err))
			connectErr = err.Error()
		}
	} else {
		s.disconnect()
	}

	return &next, connectErr, nil
}

// connect replaces the client with one for the given instance URL and secret
// and registers the listeners. The client keeps dialling until it is
// disconnected (antenne-client's persistent connect), so a bus that is down at
// boot or during a deploy is reached when it returns, and the outbox holds
// what Sablier publishes meanwhile. The first dial's error is still returned,
// so the settings page can show it.
//
// A reconnect the client is going to retry is the mechanism working, not an
// incident: an Antenne restart takes a few seconds and produces three of
// these. Logging them at error painted the dashboard red on every deploy,
// which teaches people to scroll past the colour that is supposed to mean
// something, so transient errors are logged at warn and only a terminal
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

	var dialling atomic.Bool
	firstErr := make(chan error, 1)
	client := antenneclient.NewClient(cfg,
		antenneclient.WithPersistentConnect(),
		antenneclient.WithOnConnect(func() {
			s.logger.Info("antenne: connected")
		}),
		antenneclient.WithOnDisconnect(func() {
			s.logger.Info("antenne: disconnected")
		}),
		antenneclient.WithOnError(func(err error) {
			var transient *antenneclient.TransientError
			if !stderrors.As(err, &transient) {
				s.logger.Error("antenne: error", slog.Any("error", err))
				return
			}
			if dialling.Load() {
				select {
				case firstErr <- transient.Err:
				default:
				}
			}
			s.logger.Warn("antenne: reconnecting",
				slog.Int("attempt", transient.Attempt),
				slog.Any("error", transient.Err))
		}),
	)
	s.setupListeners(client)

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	dialling.Store(true)
	err := client.Connect(ctx)
	dialling.Store(false)
	if err != nil {
		return err
	}

	s.mu.Lock()
	s.client = client
	s.mu.Unlock()

	select {
	case err := <-firstErr:
		return err
	default:
		return nil
	}
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

// listeners maps every channel Sablier listens to onto its handler.
func (s *Service) listeners() map[catalog.Channel]antenneclient.EventHandler {
	return map[catalog.Channel]antenneclient.EventHandler{
		catalog.ProjectCreated:      s.handleProjectCreated,
		catalog.ProjectUpdated:      s.handleProjectUpdated,
		catalog.ProjectDeleted:      s.handleProjectDeleted,
		catalog.TaskCreated:         s.handleTaskCreated,
		catalog.TaskUpdated:         s.handleTaskUpdated,
		catalog.TaskDeleted:         s.handleTaskDeleted,
		catalog.AgentSessionCreated: s.handleAgentSession,
		catalog.AgentSessionUpdated: s.handleAgentSession,
	}
}

func (s *Service) setupListeners(client *antenneclient.Client) {
	for channel, handler := range s.listeners() {
		client.Listen(string(channel), handler)
	}
}
