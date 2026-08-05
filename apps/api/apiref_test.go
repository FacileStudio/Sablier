package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/modules/nookpool"
	"github.com/FacileStudio/Sablier/apps/api/modules/notifications"
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
	nookPoolService := nookpool.NewService(nil, appLogger)
	return buildRouter(nil, noopPinger{}, appEnv, appLogger, notificationsService, nookPoolService)
}

// The registry is hand-written, so it rots the moment someone registers a route
// and forgets it. /files is excluded: it is a static file server, not an API.
func TestEveryRouteIsDocumented(t *testing.T) {
	if missing := apiref.Undocumented(testRouter(t), referenceConfig(), "/files"); len(missing) > 0 {
		t.Errorf("routes missing from the API registry: %v", missing)
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
