package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"strconv"
	"sync"
	"syscall"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/database"
	documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"
	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/internal/worker"
	"github.com/FacileStudio/Sablier/apps/api/modules/agents"
	"github.com/FacileStudio/Sablier/apps/api/modules/antenne"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/heartbeats"
	"github.com/FacileStudio/Sablier/apps/api/modules/notifications"
	"github.com/FacileStudio/Sablier/apps/api/modules/projects"
	"github.com/FacileStudio/Sablier/apps/api/modules/settings"
	"github.com/FacileStudio/Sablier/apps/api/modules/spaces"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/modules/users"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/FacileStudio/tronc/spa"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type sqlPinger interface {
	PingContext(ctx context.Context) error
}

func referenceConfig() apiref.Config {
	return apiref.Config{
		Title:       "Sablier API",
		Description: "Self-hosted time tracker for small teams.",
		Servers:     []string{"/api"},
		Registry: documentation.Response{
			Modules: []documentation.Module{
				auth.Documentation,
				projects.Documentation,
				timeentries.Documentation,
				users.Documentation,
				settings.Documentation,
				spaces.Documentation,
				antenne.Documentation,
				notifications.Documentation,
				agents.Documentation,
				heartbeats.Documentation,
			},
		},
	}
}

// journalConfigExtra hangs the browser's Journal settings off /auth/config.
//
// The client is adapter-static served by this binary, so it has no environment
// of its own and the key has to arrive over HTTP. /auth/config is the endpoint
// that already exists for exactly this — unauthenticated, rate-limited, read
// once at boot — and reusing it beats minting a second public route that would
// carry two fields. Nothing here is secret: a Journal public key is meant to
// sit in a bundle, and the origin allowlist and daily quota on the key are what
// bound its abuse.
//
// Unset JOURNAL_BROWSER_KEY or JOURNAL_BROWSER_URL omits the block entirely and
// the client reports nothing, which is what every environment except production
// wants. The URL is its own variable and not JOURNAL_URL, which the server SDK
// points at a Docker-internal address no browser can reach.
func journalConfigExtra(appEnv *env.Config) func() map[string]any {
	return func() map[string]any {
		if appEnv.JournalBrowserKey == "" || appEnv.JournalBrowserURL == "" {
			return nil
		}
		return map[string]any{
			"journal": map[string]any{
				"url": appEnv.JournalBrowserURL,
				"key": appEnv.JournalBrowserKey,
			},
		}
	}
}

// buildAuth constructs porte: one session manager and the OIDC kit over the
// identity tables. Sign-in goes through Registre only; there is no local
// password kit. Discovery runs here, so an unreachable or half-configured
// issuer fails at boot rather than on somebody's first login.
func buildAuth(ctx context.Context, db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger) (*session.Manager, *oidc.Kit, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, err
	}
	store := portepg.New(sqlDB)
	users := auth.NewUserStore(db)
	cfg := appEnv.Porte()

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return nil, nil, err
	}
	kit, err := oidc.New(ctx, cfg, oidc.Deps{
		Users:       users,
		Identities:  store.Identities(),
		Sessions:    sessions,
		Codes:       store.LoginCodes(),
		Logger:      appLogger,
		ConfigExtra: journalConfigExtra(appEnv),
	})
	if err != nil {
		return nil, nil, err
	}
	return sessions, kit, nil
}

// uploadSourceMaps ships this build's maps to Journal so a browser stack trace
// resolves to real file names instead of a hashed chunk and a column number.
//
// It runs from the process that serves the build, because the two ship in the
// same image and cannot drift apart that way — and because .git is excluded
// from the Docker context, which rules out reading a commit at build time. The
// release is read from the client's own version.json, which is the exact string
// the browser reports, so the two always agree.
//
// In a goroutine and never fatal: this is a debugging convenience, and an app
// that cannot upload its maps must still serve.
func uploadSourceMaps(appEnv env.Config, appLogger *slog.Logger) {
	dir := os.Getenv("SOURCEMAP_DIR")
	if dir == "" || appEnv.JournalURL == "" || appEnv.JournalToken == "" {
		return
	}

	release, err := clientRelease(os.Getenv("CLIENT_DIR"))
	if err != nil {
		appLogger.Warn("source maps: no release to upload under", slog.Any("error", err))
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	result, err := journal.UploadSourceMaps(ctx, journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken}, dir, release)
	if err != nil {
		appLogger.Warn("source maps: upload failed",
			slog.String("release", release), slog.Any("error", err))
		return
	}
	if result.Uploaded > 0 {
		appLogger.Info("source maps uploaded",
			slog.String("release", release),
			slog.Int("uploaded", result.Uploaded),
			slog.Int("skipped", result.Skipped))
	}
}

// clientRelease reads the build identity SvelteKit writes next to the bundle,
// which is the same value the browser reports as `release`.
func clientRelease(clientDir string) (string, error) {
	if clientDir == "" {
		clientDir = "./client"
	}
	raw, err := os.ReadFile(filepath.Join(clientDir, "_app", "version.json"))
	if err != nil {
		return "", err
	}
	var payload struct {
		Version string `json:"version"`
	}
	if err := json.Unmarshal(raw, &payload); err != nil {
		return "", err
	}
	if payload.Version == "" {
		return "", fmt.Errorf("client version.json carries no version")
	}
	return payload.Version, nil
}

// buildRouter assembles the chi router with all app middleware and module
// routes mounted.
//
// TrustedProxies matters behind Traefik and Cloudflare: RemoteAddr is only
// the real visitor address if both are trusted, since Traefik replaces the
// forwarded chain rather than extending it, so the visitor survives only in
// Cf-Connecting-Ip. TRUSTED_PROXIES=private,cloudflare covers all three hops.
func buildRouter(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger, sessions *session.Manager, kit *oidc.Kit, notificationsService *notifications.Service, antenneService *antenne.Service, syncer *directory.Syncer, deletions *spaces.Deletions) chi.Router {
	authService := auth.NewService(db, sessions, appLogger)
	projectService := projects.NewService(db)
	timeEntryService := timeentries.NewService(db)
	userService := users.NewService(db, authService)
	settingsService := settings.NewService(db)
	spaceService := spaces.NewService(db)
	agentService := agents.NewService(db)
	projectService.SetPoolService(antenneService)
	timeEntryService.SetPoolService(antenneService)
	heartbeatService := heartbeats.NewService(db, timeEntryService)

	clientDir := spa.DirFromEnv()
	router := httpx.NewRouter(httpx.Config{
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		Logger:         appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
		Security: securityConfig(clientDir, appLogger),
	})

	health.Mount(router, sqlDB.PingContext)
	router.Get("/version", versionHandler)
	apiref.Mount(router, referenceConfig())
	router.Handle("/files/*", http.StripPrefix("/files/", http.FileServer(http.Dir(appEnv.StorageDir))))

	router.Route("/api", func(api chi.Router) {
		auth.RegisterRoutes(api, sessions, kit)
		projects.RegisterRoutes(api, projectService, authService)
		timeentries.RegisterRoutes(api, timeEntryService, authService)
		users.RegisterRoutes(api, userService, authService)
		settings.RegisterRoutes(api, settingsService, authService)
		spaces.RegisterRoutes(api, spaceService, authService)
		if syncer != nil {
			spaces.RegisterSync(api, syncer)
		}
		if deletions != nil {
			spaces.RegisterDeletions(api, deletions)
		}
		antenne.RegisterRoutes(api, antenneService, authService)
		notifications.RegisterRoutes(api, notificationsService, authService)
		agents.RegisterRoutes(api, agentService, authService)
		heartbeats.RegisterRoutes(api, heartbeatService, authService)
	})

	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router
}

func createApiServer(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger, notificationsService *notifications.Service, antenneService *antenne.Service, syncer *directory.Syncer, deletions *spaces.Deletions) (*http.Server, error) {
	sessions, kit, err := buildAuth(context.Background(), db, appEnv, appLogger)
	if err != nil {
		return nil, err
	}
	router := buildRouter(db, sqlDB, appEnv, appLogger, sessions, kit, notificationsService, antenneService, syncer, deletions)

	antenneService.AutoConnect(context.Background())

	addr := ":" + strconv.Itoa(appEnv.Port)
	server := &http.Server{
		Addr:              addr,
		Handler:           router,
		ReadHeaderTimeout: 5 * time.Second,
		ReadTimeout:       10 * time.Second,
		WriteTimeout:      15 * time.Second,
		IdleTimeout:       60 * time.Second,
	}

	return server, nil
}

// newDirectory builds the Registre space sync when OIDC is configured, using the app's own
// client credentials, and returns nils otherwise: without Registre there is nothing to sync
// and no nudge route to mount.
//
// ready runs once, in the background, after the first snapshot applied: the import has
// succeeded by then and every imported space carries its Registre id, so the Antenne
// reconcile can tell which projects are shareable.
func newDirectory(db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger, ready func(context.Context)) (*spaces.Registre, *directory.Syncer) {
	if appEnv.OIDC == nil {
		return nil, nil
	}
	registre := spaces.NewRegistre(db, appEnv.OIDC.Issuer, appLogger)
	var once sync.Once
	syncer := &directory.Syncer{
		Client: directory.Client{
			Issuer:       appEnv.OIDC.Issuer,
			ClientID:     appEnv.OIDC.ClientID,
			ClientSecret: appEnv.OIDC.ClientSecret,
		},
		Apply: func(ctx context.Context, snapshot directory.Snapshot) error {
			if err := registre.Apply(ctx, snapshot); err != nil {
				return err
			}
			once.Do(func() { go ready(ctx) })
			return nil
		},
		Logger: appLogger,
	}
	return registre, syncer
}

// newDeletions builds the GDPR deletion receiver when OIDC is configured, polling Registre
// with the same client credentials as the snapshot sync. Sablier runs one replica, so this
// is the one poller. Without Registre there is nobody to erase and it returns nil.
func newDeletions(db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger) *spaces.Deletions {
	if appEnv.OIDC == nil {
		return nil
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("deletion receiver not started", slog.Any("error", err))
		return nil
	}
	client := directory.Client{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
	}
	return spaces.NewDeletions(db, client, portepg.New(sqlDB).EraseIdentity, appEnv.StorageDir, appLogger)
}

// externalConnect lists the origins the built client calls besides itself: the icon API
// behind iconify-icon.
var externalConnect = []string{"https://api.iconify.design"}

// securityConfig extends tronc's default security headers with what the built client
// needs: the inline scripts adapter-static writes into its pages, which script-src 'self'
// would otherwise block, and the external origins above. A client directory that cannot be
// hashed is logged and leaves the default script policy, which only reports.
func securityConfig(clientDir string, appLogger *slog.Logger) troncmiddleware.SecurityConfig {
	sources := map[string][]string{"connect-src": externalConnect}
	if spa.Available(clientDir) {
		hashes, err := spa.InlineScriptHashes(clientDir)
		if err != nil {
			appLogger.Warn("failed to hash the client's inline scripts", slog.Any("error", err))
		} else {
			sources["script-src"] = hashes
		}
	}
	return troncmiddleware.SecurityConfig{CSPSources: sources}
}

// Stamped at Docker build time with git describe (leading v stripped),
// matching goreleaser's {{.Version}} so `/api version` reads the same shape
// as the suite's CLI binaries.
var version = "dev"

func main() {
	if len(os.Args) > 1 && os.Args[1] == "version" {
		fmt.Println(version)
		return
	}
	if healthcheck.Handle(os.Args) {
		return
	}
	if err := run(); err != nil {
		os.Exit(1)
	}
}

func run() error {
	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return err
	}
	var journalClient *journal.Client
	appLogger = logger.New(logger.Config{
		Level: appEnv.LogLevel,
		Wrap: func(handler slog.Handler) slog.Handler {
			if appEnv.JournalURL == "" || appEnv.JournalToken == "" {
				return handler
			}
			journalClient = journal.New(journal.Config{URL: appEnv.JournalURL, Token: appEnv.JournalToken})
			return journal.NewHandler(journalClient, handler)
		},
	})
	if journalClient != nil {
		defer journalClient.Close()
	}
	go uploadSourceMaps(appEnv, appLogger)

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return err
	}

	if err := schemas.Migrate(db, appEnv.Porte().Issuer); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return err
	} else {
		appLogger.Info("database migrations applied")
	}

	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return err
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	notificationsService := notifications.NewService(db, appEnv.VAPIDPublicKey, appEnv.VAPIDPrivateKey, appEnv.VAPIDSubject, appLogger)
	antenneService := antenne.NewService(db, appLogger)

	if appEnv.OIDC != nil {
		antenneService.SetIssuer(appEnv.OIDC.Issuer)
	}
	registre, syncer := newDirectory(db, &appEnv, appLogger, antenneService.ReconcileAtBoot)
	deletions := newDeletions(db, &appEnv, appLogger)
	server, err := createApiServer(db, sqlDB, &appEnv, appLogger, notificationsService, antenneService, syncer, deletions)
	if err != nil {
		appLogger.Error("failed to create server", slog.Any("error", err))
		return err
	}
	serverErrCh := make(chan error, 1)

	go func() {
		serverErrCh <- server.ListenAndServe()
	}()

	appLogger.Info("API server started", slog.String("address", server.Addr))

	shutdownSignal, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		worker.RunNotificationWorker(shutdownSignal, notificationsService, appLogger)
	}()

	appLogger.Info("notification worker started")

	go antenneService.RunOutboxWorker(shutdownSignal)

	if deletions != nil {
		go deletions.Run(shutdownSignal)
	}

	if registre != nil {
		go registre.Run(shutdownSignal, syncer)
	} else {
		go antenneService.ReconcileAtBoot(shutdownSignal)
	}

	appLogger.Info("antenne: outbox worker started")

	go worker.RunHeartbeatWorker(shutdownSignal, heartbeats.NewService(db, timeentries.NewService(db)), appLogger)

	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
			return err
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		antenneService.Shutdown()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return err
		}
		appLogger.Info("server stopped")
	}

	return nil
}
