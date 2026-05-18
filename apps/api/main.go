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

	"api/internal/database"
	documentation "api/internal/documentation"
	"api/internal/env"
	"api/internal/httpjson"
	"api/internal/logger"
	"api/internal/middleware"
	"api/internal/worker"
	"api/modules/auth"
	"api/modules/notifications"
	"api/modules/projects"
	"api/modules/settings"
	"api/modules/timeentries"
	"api/modules/users"
	"api/schemas"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"
	"gorm.io/gorm"
)

type sqlPinger interface {
	PingContext(ctx context.Context) error
}

func createApiServer(db *gorm.DB, sqlDB sqlPinger, appEnv *env.Config, appLogger *slog.Logger, notificationsService *notifications.Service) (*http.Server, error) {
	authService := auth.NewService(db)
	projectService := projects.NewService(db)
	timeEntryService := timeentries.NewService(db)
	userService := users.NewService(db, appEnv.StorageDir)
	settingsService := settings.NewService(db)
	docs := documentation.Response{
		Modules: []documentation.Module{
			auth.Documentation,
			projects.Documentation,
			timeentries.Documentation,
			users.Documentation,
			settings.Documentation,
			notifications.Documentation,
		},
	}

	router := chi.NewRouter()
	router.Use(chimiddleware.RequestID)
	router.Use(chimiddleware.RealIP)
	router.Use(middleware.CORS(appEnv.CORSAllowedOrigins))
	router.Use(middleware.RequestLogger(appLogger))
	router.Use(chimiddleware.Recoverer)

	router.Get("/health", func(w http.ResponseWriter, request *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ok"})
	})
	router.Get("/ready", func(w http.ResponseWriter, request *http.Request) {
		readinessContext, cancel := context.WithTimeout(request.Context(), 2*time.Second)
		defer cancel()
		if err := sqlDB.PingContext(readinessContext); err != nil {
			httpjson.WriteJSON(w, http.StatusServiceUnavailable, map[string]string{"status": "not_ready"})
			return
		}
		httpjson.WriteJSON(w, http.StatusOK, map[string]string{"status": "ready"})
	})
	router.Get("/docs", func(w http.ResponseWriter, request *http.Request) {
		httpjson.WriteJSON(w, http.StatusOK, docs)
	})
	router.Handle("/files/*", http.StripPrefix("/files/", http.FileServer(http.Dir(appEnv.StorageDir))))

	auth.RegisterRoutes(router, authService, *appEnv)
	projects.RegisterRoutes(router, projectService, authService)
	timeentries.RegisterRoutes(router, timeEntryService, authService)
	users.RegisterRoutes(router, userService, authService)
	settings.RegisterRoutes(router, settingsService, authService)
	notifications.RegisterRoutes(router, notificationsService, authService)

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
	appEnv, err := env.Load()
	appLogger := logger.New("info")
	if err != nil {
		appLogger.Error("failed to load config", slog.Any("error", err))
		return
	}
	appLogger = logger.New(appEnv.LogLevel)

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

	server, err := createApiServer(db, sqlDB, &appEnv, appLogger, notificationsService)
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

	select {
	case err := <-serverErrCh:
		if !errors.Is(err, http.ErrServerClosed) {
			appLogger.Error("server stopped", slog.Any("error", err))
		}
	case <-shutdownSignal.Done():
		appLogger.Info("server shutting down")
		shutdownContext, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if err := server.Shutdown(shutdownContext); err != nil {
			appLogger.Error("server shutdown failed", slog.Any("error", err))
			return
		}
		appLogger.Info("server stopped")
	}
}
