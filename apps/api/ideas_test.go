package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/plans"
	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

const ideasIssuer = "https://sso.example.test"

type ideasHarness struct {
	router chi.Router
	bulle  *httptest.Server
	spaces chan string
	calls  atomic.Int32
}

func stubAuthenticate(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		user := r.Header.Get("X-Test-User")
		if user == "" {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		next.ServeHTTP(w, r.WithContext(authcontext.WithIdentity(r.Context(), authcontext.Identity{UserID: user})))
	})
}

func newIdeasHarness(t *testing.T, db *gorm.DB, gate *plans.Gate) *ideasHarness {
	t.Helper()
	h := &ideasHarness{spaces: make(chan string, 4)}
	h.bulle = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		h.calls.Add(1)
		h.spaces <- r.URL.Query().Get("space_id")
		w.Header().Set("Content-Type", "application/json")
		_, _ = io.WriteString(w, `{"ideas":[],"generating":false,"last_batch_at":null}`)
	}))
	t.Cleanup(h.bulle.Close)
	cfg := &ideasConfig{
		URL:        h.bulle.URL,
		UpgradeURL: ideasIssuer + "/spaces",
		Token:      func(context.Context) (string, error) { return "token", nil },
		Gate:       gate,
	}
	h.router = chi.NewRouter()
	h.router.Route("/api", func(api chi.Router) {
		if err := mountIdeas(api, db, cfg, stubAuthenticate); err != nil {
			t.Fatal(err)
		}
	})
	return h
}

func (h *ideasHarness) do(method, path, user string) *httptest.ResponseRecorder {
	request := httptest.NewRequest(method, path, strings.NewReader("{}"))
	if user != "" {
		request.Header.Set("X-Test-User", user)
	}
	recorder := httptest.NewRecorder()
	h.router.ServeHTTP(recorder, request)
	return recorder
}

func ideasDB(t *testing.T) *gorm.DB {
	t.Helper()
	db := testdb.Open(t)
	registre := "spc_registre"
	if err := db.Create(&schemas.Space{ID: "00000000-0000-4000-8000-000000000001", Name: "Team", RegistreSpaceID: &registre}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.Create(&schemas.SpaceMember{SpaceID: "00000000-0000-4000-8000-000000000001", UserID: 1, Role: schemas.RoleMember}).Error; err != nil {
		t.Fatal(err)
	}
	if err := db.AutoMigrate(&schemas.ForfaitState{}); err != nil {
		t.Fatal(err)
	}
	if err := db.Exec(`CREATE TABLE IF NOT EXISTS porte_identities (user_id integer, provider text, subject text)`).Error; err != nil {
		t.Fatal(err)
	}
	return db
}

func TestIdeasAreBehindAuth(t *testing.T) {
	h := newIdeasHarness(t, ideasDB(t), nil)
	for _, path := range []string{"/api/ideas", "/api/ideas/generate", "/api/ideas/runs/r1"} {
		method := http.MethodGet
		if path == "/api/ideas/generate" {
			method = http.MethodPost
		}
		if code := h.do(method, path, "").Code; code != http.StatusUnauthorized {
			t.Errorf("%s %s without a session = %d, want 401", method, path, code)
		}
	}
	if h.calls.Load() != 0 {
		t.Fatalf("Bulle was called %d times without a session", h.calls.Load())
	}
}

func TestIdeasPassTheRegistreSpace(t *testing.T) {
	h := newIdeasHarness(t, ideasDB(t), nil)

	if code := h.do(http.MethodGet, "/api/ideas?space_id=00000000-0000-4000-8000-000000000001", "1").Code; code != http.StatusOK {
		t.Fatalf("member space = %d, want 200", code)
	}
	if got := <-h.spaces; got != "spc_registre" {
		t.Errorf("Bulle got space %q, want spc_registre", got)
	}

	if code := h.do(http.MethodGet, "/api/ideas", "1").Code; code != http.StatusOK {
		t.Fatalf("personal = %d, want 200", code)
	}
	if got := <-h.spaces; got != "" {
		t.Errorf("personal sent space %q, want none", got)
	}

	if code := h.do(http.MethodGet, "/api/ideas?space_id=00000000-0000-4000-8000-000000000001", "2").Code; code != http.StatusNotFound {
		t.Errorf("non-member = %d, want 404", code)
	}
	if h.calls.Load() != 2 {
		t.Errorf("Bulle calls = %d, want 2", h.calls.Load())
	}
}

func TestAFreePlanGetsNoAIAndNoBulleCall(t *testing.T) {
	db := ideasDB(t)
	if err := db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (1, ?, 'sub-free'), (3, ?, 'sub-paid')`, ideasIssuer, ideasIssuer).Error; err != nil {
		t.Fatal(err)
	}
	gate := plans.New(db, ideasIssuer, slog.New(slog.NewTextHandler(io.Discard, nil)))
	snapshot := directory.Snapshot{Version: 1, EntitlementsRevision: 1, People: []directory.Person{
		{Subject: "sub-paid", Entitlements: map[string]json.RawMessage{plans.AIAccess: json.RawMessage("true")}},
	}}
	if err := gate.Apply(context.Background(), snapshot); err != nil {
		t.Fatal(err)
	}
	h := newIdeasHarness(t, db, gate)

	list := h.do(http.MethodGet, "/api/ideas", "1")
	var page struct {
		AIAccess   bool   `json:"ai_access"`
		UpgradeURL string `json:"upgrade_url"`
	}
	if err := json.Unmarshal(list.Body.Bytes(), &page); err != nil {
		t.Fatal(err)
	}
	if list.Code != http.StatusOK || page.AIAccess || page.UpgradeURL != ideasIssuer+"/spaces" {
		t.Errorf("free list = %d %+v, want 200, ai_access false and the upgrade url", list.Code, page)
	}
	if code := h.do(http.MethodPost, "/api/ideas/generate", "1").Code; code != http.StatusPaymentRequired {
		t.Errorf("free generate = %d, want 402", code)
	}
	if h.calls.Load() != 0 {
		t.Fatalf("Bulle was called %d times for a free plan", h.calls.Load())
	}

	if code := h.do(http.MethodPost, "/api/ideas/generate", "3").Code; code != http.StatusAccepted && code != http.StatusBadGateway {
		t.Errorf("paid generate = %d, want it to reach Bulle", code)
	}
	if h.calls.Load() != 1 {
		t.Errorf("Bulle calls for a paid plan = %d, want 1", h.calls.Load())
	}
}

func TestEntitlementsSurviveARestart(t *testing.T) {
	db := ideasDB(t)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	first := plans.New(db, ideasIssuer, logger)
	if err := first.Apply(context.Background(), directory.Snapshot{Version: 1, EntitlementsRevision: 4}); err != nil {
		t.Fatal(err)
	}
	second := plans.New(db, ideasIssuer, logger)
	if second.Active() {
		t.Fatal("a fresh gate must start inactive")
	}
	if err := second.Load(context.Background()); err != nil {
		t.Fatal(err)
	}
	if !second.Active() {
		t.Fatal("the saved state did not end the fail-open period")
	}
	allowed, err := second.HasAI(context.Background(), 9)
	if err != nil || allowed {
		t.Fatalf("a user with no entitlement = %v, %v, want false after a restart", allowed, err)
	}
}
