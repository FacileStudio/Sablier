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
	"syscall"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/database"
	documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"
	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/internal/worker"
	"github.com/FacileStudio/Sablier/apps/api/modules/antenne"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/notifications"
	"github.com/FacileStudio/Sablier/apps/api/modules/projects"
	"github.com/FacileStudio/Sablier/apps/api/modules/settings"
	"github.com/FacileStudio/Sablier/apps/api/modules/spaces"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/modules/users"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/local"
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
			},
		},
	}
}

// buildAuth constructs porte: one session manager, shared by the OIDC kit and
// the local login, over the identity tables.
//
// One manager and not two: they would each keep their own idea of the clock
// and of whether the cookie is Secure, and porte refuses a kit whose config
// disagrees with its manager's for exactly that reason. Discovery runs here,
// so an unreachable or half-configured issuer fails at boot rather than on
// somebody's first login — which is a change from what this app did, where a
// discovery failure at route-registration time silently left SSO 404ing until
// the next restart.
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

func buildAuth(ctx context.Context, db *gorm.DB, appEnv *env.Config, appLogger *slog.Logger) (*session.Manager, *local.Kit, *oidc.Kit, error) {
	sqlDB, err := db.DB()
	if err != nil {
		return nil, nil, nil, err
	}
	store := portepg.New(sqlDB)
	users := auth.NewUserStore(db)
	cfg := appEnv.Porte()

	sessions, err := session.New(cfg, session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		return nil, nil, nil, err
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
		return nil, nil, nil, err
	}
	// Sablier's floor has always been eight characters. porte defaults to
	// twelve; raising it here would reject a password this app accepted
	// yesterday, which is a product decision and not a migration.
	passwords, err := local.New(local.Config{AllowRegistration: !appEnv.SSOOnly, MinPasswordLength: 8}, local.Deps{
		Users:      users,
		Identities: store.Identities(),
		Sessions:   sessions,
		Logger:     appLogger,
		Count:      users.CountUsers,
	})
	if err != nil {
		return nil, nil, nil, err
	}
	return sessions, passwords, kit, nil
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

func buildRouter(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger, sessions *session.Manager, passwords *local.Kit, kit *oidc.Kit, notificationsService *notifications.Service, antenneService *antenne.Service) chi.Router {
	authService := auth.NewService(db, sessions, passwords, appLogger)
	projectService := projects.NewService(db)
	timeEntryService := timeentries.NewService(db)
	userService := users.NewService(db, appEnv.StorageDir, authService)
	settingsService := settings.NewService(db)
	spaceService := spaces.NewService(db)
	projectService.SetPoolService(antenneService)
	timeEntryService.SetPoolService(antenneService)

	router := httpx.NewRouter(httpx.Config{
		// Behind Traefik and Cloudflare, RemoteAddr is only the
		// visitor if both are trusted: Traefik replaces the forwarded
		// chain rather than extending it, so the visitor survives in
		// Cf-Connecting-Ip alone. TRUSTED_PROXIES=private,cloudflare
		// fills all three.
		TrustedProxies: appEnv.TrustedProxies,
		CDNProxies:     appEnv.CDNProxies,
		CDNHeader:      appEnv.CDNHeader,
		Logger:         appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
	})

	health.Mount(router, sqlDB.PingContext)
	apiref.Mount(router, referenceConfig())
	router.Handle("/files/*", http.StripPrefix("/files/", http.FileServer(http.Dir(appEnv.StorageDir))))

	router.Route("/api", func(api chi.Router) {
		sessions.Mount(api)
		kit.Mount(api)
		auth.RegisterRoutes(api, authService, *appEnv)
		projects.RegisterRoutes(api, projectService, authService)
		timeentries.RegisterRoutes(api, timeEntryService, authService)
		users.RegisterRoutes(api, userService, authService)
		settings.RegisterRoutes(api, settingsService, authService)
		spaces.RegisterRoutes(api, spaceService, authService)
		antenne.RegisterRoutes(api, antenneService, authService)
		notifications.RegisterRoutes(api, notificationsService, authService)
	})

	clientDir := spa.DirFromEnv()
	if spa.Available(clientDir) {
		router.Handle("/*", spa.Handler(spa.Config{Dir: clientDir}))
		appLogger.Info("serving client", slog.String("dir", clientDir))
	}

	return router
}

func createApiServer(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger, notificationsService *notifications.Service, antenneService *antenne.Service) (*http.Server, error) {
	sessions, passwords, kit, err := buildAuth(context.Background(), db, appEnv, appLogger)
	if err != nil {
		return nil, err
	}
	router := buildRouter(db, sqlDB, appEnv, appLogger, sessions, passwords, kit, notificationsService, antenneService)

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

func main() {
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

	if err := os.MkdirAll(filepath.Join(appEnv.StorageDir, "avatars"), 0o755); err != nil {
		appLogger.Error("failed to prepare storage", slog.Any("error", err))
		return err
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

	server, err := createApiServer(db, sqlDB, &appEnv, appLogger, notificationsService, antenneService)
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

	appLogger.Info("antenne: outbox worker started")

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
