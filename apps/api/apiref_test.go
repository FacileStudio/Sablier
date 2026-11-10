package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/modules/antenne"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/mcp"
	"github.com/FacileStudio/Sablier/apps/api/modules/notifications"
	"github.com/FacileStudio/Sablier/apps/api/modules/spaces"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/porte/oidc"
	portepg "github.com/FacileStudio/porte/pg"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/apiref"
	"github.com/go-chi/chi/v5"
)

type noopPinger struct{}

func (noopPinger) PingContext(context.Context) error { return nil }

func testRouter(t *testing.T) chi.Router {
	t.Helper()
	appEnv := &env.Config{StorageDir: t.TempDir()}
	appLogger := slog.New(slog.NewTextHandler(io.Discard, nil))
	notificationsService := notifications.NewService(nil, "", "", "", appLogger)
	antenneService := antenne.NewService(nil, appLogger)

	store := portepg.New(nil)
	sessions, err := session.New(appEnv.Porte(), session.Deps{Sessions: store.Sessions(), Logger: appLogger})
	if err != nil {
		t.Fatalf("session.New: %v", err)
	}
	kit, err := oidc.New(context.Background(), appEnv.Porte(), oidc.Deps{Sessions: sessions, Logger: appLogger})
	if err != nil {
		t.Fatalf("oidc.New: %v", err)
	}
	deletions := spaces.NewDeletions(nil, directory.Client{}, nil, nil, appEnv.StorageDir, appLogger)
	ideas := &ideasConfig{URL: "https://bulle.example.test", Token: func(context.Context) (string, error) { return "", nil }}
	router, err := buildRouter(nil, noopPinger{}, appEnv, appLogger, sessions, kit, notificationsService, antenneService, &directory.Syncer{}, deletions, attribution.Resolver{}, ideas, nil)
	if err != nil {
		t.Fatalf("buildRouter: %v", err)
	}
	return router
}

// The registry is hand-written, so it rots the moment someone registers a route
// and forgets it. /files (a static file server) and /version are excluded: neither is part of the API.
func TestEveryRouteIsDocumented(t *testing.T) {
	if missing := apiref.Undocumented(testRouter(t), referenceConfig(), "/files", "/version"); len(missing) > 0 {
		t.Errorf("routes missing from the API registry: %v", missing)
	}
}

// Registre nudges anonymously, so the route must answer without a session or a CSRF token.
func TestTheRegistreNudgeNeedsNoCredentials(t *testing.T) {
	recorder := httptest.NewRecorder()
	testRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, directory.SyncPath, nil))
	if recorder.Code != http.StatusAccepted {
		t.Fatalf("POST %s = %d, want 202", directory.SyncPath, recorder.Code)
	}
}

func TestReferenceIsServedAtDocs(t *testing.T) {
	router := testRouter(t)

	page := httptest.NewRecorder()
	router.ServeHTTP(page, httptest.NewRequest(http.MethodGet, "/docs", nil))
	if page.Code != http.StatusOK {
		t.Fatalf("GET /docs = %d, want 200", page.Code)
	}

	spec := httptest.NewRecorder()
	router.ServeHTTP(spec, httptest.NewRequest(http.MethodGet, "/docs/openapi.json", nil))
	if spec.Code != http.StatusOK {
		t.Fatalf("GET /docs/openapi.json = %d, want 200", spec.Code)
	}
	var document struct {
		OpenAPI string         `json:"openapi"`
		Paths   map[string]any `json:"paths"`
	}
	if err := json.Unmarshal(spec.Body.Bytes(), &document); err != nil {
		t.Fatalf("spec is not JSON: %v", err)
	}
	if document.OpenAPI == "" || len(document.Paths) == 0 {
		t.Fatalf("spec is empty: %+v", document)
	}
}

func TestNoPasswordRouteExists(t *testing.T) {
	router := testRouter(t)
	for _, path := range []string{"/api/auth/login", "/api/auth/register"} {
		recorder := httptest.NewRecorder()
		router.ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, path, nil))
		if recorder.Code == http.StatusOK || recorder.Code == http.StatusCreated {
			t.Errorf("POST %s = %d, want it gone", path, recorder.Code)
		}
	}
}

func TestTheDeletionPushNeedsNoSessionButMustBeSigned(t *testing.T) {
	recorder := httptest.NewRecorder()
	testRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodPost, directory.DeletionsPath, nil))
	if recorder.Code != http.StatusBadRequest {
		t.Fatalf("POST %s without a signed notice = %d, want 400", directory.DeletionsPath, recorder.Code)
	}
}

func TestVersionIsServed(t *testing.T) {
	recorder := httptest.NewRecorder()
	testRouter(t).ServeHTTP(recorder, httptest.NewRequest(http.MethodGet, "/version", nil))
	if recorder.Code != http.StatusOK || !strings.Contains(recorder.Body.String(), `"version"`) {
		t.Fatalf("GET /version = %d %q", recorder.Code, recorder.Body.String())
	}
}

func TestConnectSourcesCarryTheJournalOrigin(t *testing.T) {
	got := connectSources("https://journal.example.test/api")
	if got[len(got)-1] != "https://journal.example.test" {
		t.Fatalf("want the Journal origin last, got %v", got)
	}
	if len(connectSources("")) != len(externalConnect) {
		t.Fatalf("an unset Journal URL must add nothing")
	}
}

// The confirm page routes mount only with the MCP server, so the router above lacks them.
func TestMCPPendingRoutesAreDocumented(t *testing.T) {
	router := chi.NewRouter()
	router.Route("/api", func(api chi.Router) {
		mcp.RegisterPendingRoutes(api, nil, auth.NewService(nil, nil, slog.New(slog.NewTextHandler(io.Discard, nil))))
	})
	if missing := apiref.Undocumented(router, referenceConfig()); len(missing) > 0 {
		t.Errorf("MCP routes missing from the API registry: %v", missing)
	}
}
