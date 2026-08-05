package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"path/filepath"
	"syscall"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/database"
	documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"
	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/internal/worker"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/nookpool"
	"github.com/FacileStudio/Sablier/apps/api/modules/notifications"
	"github.com/FacileStudio/Sablier/apps/api/modules/projects"
	"github.com/FacileStudio/Sablier/apps/api/modules/settings"
	"github.com/FacileStudio/Sablier/apps/api/modules/spaces"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/modules/users"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"github.com/FacileStudio/Journal/sdk/journal"
	"github.com/FacileStudio/tronc/health"
	"github.com/FacileStudio/tronc/healthcheck"
	"github.com/FacileStudio/tronc/httpjson"
	"github.com/FacileStudio/tronc/httpx"
	"github.com/FacileStudio/tronc/logger"
	troncmiddleware "github.com/FacileStudio/tronc/middleware"
	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type sqlPinger interface {
	PingContext(ctx context.Context) error
}

func createApiServer(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger, notificationsService *notifications.Service, nookPoolService *nookpool.Service) (*http.Server, error) {
	authService := auth.NewService(db, appEnv.StorageDir, appLogger)
	projectService := projects.NewService(db)
	timeEntryService := timeentries.NewService(db)
	userService := users.NewService(db, appEnv.StorageDir)
	settingsService := settings.NewService(db)
	spaceService := spaces.NewService(db)
	projectService.SetPoolService(nookPoolService)
	timeEntryService.SetPoolService(nookPoolService)
	docs := documentation.Response{
		Modules: []documentation.Module{
			auth.Documentation,
			projects.Documentation,
			timeentries.Documentation,
			users.Documentation,
			settings.Documentation,
			spaces.Documentation,
			nookpool.Documentation,
			notifications.Documentation,
		},
	}
	openapiSpec := documentation.ToOpenAPI(docs)

	router := httpx.NewRouter(httpx.Config{
		Logger: appLogger,
		CORS: troncmiddleware.CORSConfig{
			AllowedOrigins: appEnv.CORSAllowedOrigins,
		},
	})

	health.Mount(router, sqlDB.PingContext)
	router.Get("/docs", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`<!doctype html>
<html>
<head>
  <title>Sablier API</title>
  <meta charset="utf-8" />
</head>
<body>
  <script id="api-reference" data-url="/openapi"></script>
  <script src="https://cdn.jsdelivr.net/npm/@scalar/api-reference"></script>
</body>
</html>`))
	})
	router.Get("/openapi", func(w http.ResponseWriter, _ *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, openapiSpec)
	})
	router.Handle("/files/*", http.StripPrefix("/files/", http.FileServer(http.Dir(appEnv.StorageDir))))

	router.Route("/api", func(api chi.Router) {
		auth.RegisterRoutes(api, authService, *appEnv)
		projects.RegisterRoutes(api, projectService, authService)
		timeentries.RegisterRoutes(api, timeEntryService, authService)
		users.RegisterRoutes(api, userService, authService)
		settings.RegisterRoutes(api, settingsService, authService)
		spaces.RegisterRoutes(api, spaceService, authService)
		nookpool.RegisterRoutes(api, nookPoolService, authService)
		notifications.RegisterRoutes(api, notificationsService, authService)
	})

	nookPoolService.AutoConnect(context.Background())

	addr := ":" + appEnv.Port
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

	appEnv, err := env.Load()
	appLogger := logger.New(logger.Config{})
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return
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

	db, err := database.Open(appEnv.DatabaseURL)
	if err != nil {
		appLogger.Error("failed to open database", slog.Any("error", err))
		return
	}

	if err := schemas.Migrate(db); err != nil {
		appLogger.Error("failed to run migrations", slog.Any("error", err))
		return
	} else {
		appLogger.Info("database migrations applied")
	}

	if err := os.MkdirAll(filepath.Join(appEnv.StorageDir, "avatars"), 0o755); err != nil {
		appLogger.Error("failed to prepare storage", slog.Any("error", err))
		return
	}
	sqlDB, err := db.DB()
	if err != nil {
		appLogger.Error("failed to access database handle", slog.Any("error", err))
		return
	}

	defer func() {
		if err := sqlDB.Close(); err != nil {
			appLogger.Error("failed to close database", slog.Any("error", err))
		}
	}()

	notificationsService := notifications.NewService(db, appEnv.VAPIDPublicKey, appEnv.VAPIDPrivateKey, appEnv.VAPIDSubject, appLogger)
	nookPoolService := nookpool.NewService(db, appLogger)

	server, err := createApiServer(db, sqlDB, &appEnv, appLogger, notificationsService, nookPoolService)
	if err != nil {
		appLogger.Error("failed to create server", slog.Any("error", err))
		return
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

	go nookPoolService.RunOutboxWorker(shutdownSignal)

	appLogger.Info("pool outbox worker started")

	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		nookPoolService.Shutdown()
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return
		}
		appLogger.Info("server stopped")
	}
}
