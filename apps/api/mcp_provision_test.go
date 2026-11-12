package main

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/modules/antenne"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/notifications"
	"github.com/FacileStudio/Sablier/apps/api/modules/spaces"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"gorm.io/gorm"
)

const (
	newSubject = "registre-subject-1"
	newEmail   = "new.person@example.test"
)

// registre is a fake Registre: discovery over mcptest's signing key, so tokens minted
// with its URL as iss verify and the OIDC kit boots as in production, and the deletions
// list the receiver polls.
type registre struct {
	URL  string
	keys *mcptest.Issuer
	mu   sync.Mutex
	gone []directory.Deleted
}

func newRegistre(t *testing.T) *registre {
	t.Helper()
	r := &registre{keys: mcptest.NewIssuer(t)}
	srv := httptest.NewServer(http.HandlerFunc(r.serve))
	t.Cleanup(srv.Close)
	r.URL = srv.URL
	return r
}

func (r *registre) serve(w http.ResponseWriter, req *http.Request) {
	if req.URL.Path == "/api/directory/deletions" {
		r.mu.Lock()
		defer r.mu.Unlock()
		writeJSON(w, map[string]any{"deleted": r.gone})
		return
	}
	writeJSON(w, map[string]any{
		"issuer": r.URL, "jwks_uri": r.keys.URL + "/jwks",
		"authorization_endpoint": r.URL + "/authorize", "token_endpoint": r.URL + "/token",
		"scopes_supported": []string{"openid", "profile", "email", "offline_access", "roles"},
	})
}

// erase lists subject as erased, as user-erase does.
func (r *registre) erase(subject string) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.gone = append(r.gone, directory.Deleted{Subject: subject, DeletedAt: time.Now().UTC()})
}

// token is what Registre mints for an MCP resource: a verified email and no roles.
func (r *registre) token(t *testing.T, resource string) string {
	return r.keys.Mint(t, map[string]any{
		"iss": r.URL, "aud": []string{resource}, "sub": newSubject, "scope": "sablier.read sablier.write",
		"email": newEmail, "email_verified": true, "name": "New Person", "locale": "fr",
	})
}

// writeJSON answers value; a test client that hung up needs no answer.
func writeJSON(w http.ResponseWriter, value any) {
	if err := json.NewEncoder(w).Encode(value); err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
	}
}

// productionEnv is Sablier's production OIDC and MCP environment, OIDC_CLAIMS_SCOPE=roles
// included, pointed at issuer and resource.
func productionEnv(t *testing.T, issuer, resource string) env.Config {
	t.Helper()
	for key, value := range map[string]string{
		"OIDC_ISSUER": issuer, "OIDC_CLIENT_ID": "sablier", "OIDC_CLIENT_SECRET": "secret",
		"OIDC_REDIRECT_URL": "http://localhost/api/auth/oidc/callback", "OIDC_SUCCESS_URL": "http://localhost/",
		"OIDC_MACHINE_AUDIENCE": "sablier", "OIDC_SCOPE_PREFIX": "sablier", "OIDC_CLAIMS_SCOPE": "roles",
		"MCP_RESOURCE": resource, "STORAGE_DIR": t.TempDir(), "DATABASE_URL": "postgres://unused",
	} {
		t.Setenv(key, value)
	}
	appEnv, err := env.Load()
	if err != nil {
		t.Fatal(err)
	}
	return appEnv
}

// sablier is the API wired the way run and createApiServer wire it, on a fake Registre.
type sablier struct {
	db        *gorm.DB
	registre  *registre
	resource  string
	deletions *spaces.Deletions
}

func newSablier(t *testing.T) *sablier {
	t.Helper()
	s := &sablier{db: testdb.Migrated(t), registre: newRegistre(t)}
	srv := httptest.NewUnstartedServer(nil)
	s.resource = "http://" + srv.Listener.Addr().String() + "/mcp"
	appEnv := productionEnv(t, s.registre.URL, s.resource)
	srv.Config.Handler = s.router(t, &appEnv)
	srv.Start()
	t.Cleanup(srv.Close)
	return s
}

func (s *sablier) router(t *testing.T, appEnv *env.Config) http.Handler {
	t.Helper()
	sqlDB, err := s.db.DB()
	if err != nil {
		t.Fatal(err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	ctx := context.Background()
	accounts, err := auth.NewAccounts(ctx, sqlDB, logger)
	if err != nil {
		t.Fatal(err)
	}
	s.deletions = newDeletions(s.db, accounts, appEnv, logger)
	parts, err := buildAuth(ctx, sqlDB, people{accounts: accounts, erased: s.deletions}, appEnv, logger)
	if err != nil {
		t.Fatal(err)
	}
	setup, err := newMCPSetup(ctx, appEnv, sqlDB, parts, logger)
	if err != nil {
		t.Fatal(err)
	}
	router, err := buildRouter(s.db, noopPinger{}, appEnv, logger, parts, notifications.NewService(nil, "", "", "", logger),
		antenne.NewService(nil, logger), nil, s.deletions, attribution.Resolver{}, nil, setup)
	if err != nil {
		t.Fatal(err)
	}
	return router
}

// accounts counts the users with email.
func (s *sablier) accounts(t *testing.T, email string) int64 {
	t.Helper()
	var n int64
	if err := s.db.Model(&schemas.User{}).Where("email = ?", email).Count(&n).Error; err != nil {
		t.Fatal(err)
	}
	return n
}

// initialize sends an MCP initialize with token and returns the status.
func (s *sablier) initialize(t *testing.T, token string) int {
	t.Helper()
	body := `{"jsonrpc":"2.0","id":1,"method":"initialize","params":{"protocolVersion":"2025-06-18",` +
		`"capabilities":{},"clientInfo":{"name":"t","version":"1"}}}`
	req, err := http.NewRequest(http.MethodPost, s.resource, strings.NewReader(body))
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("Authorization", "Bearer "+token)
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json, text/event-stream")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	if err := resp.Body.Close(); err != nil {
		t.Fatal(err)
	}
	return resp.StatusCode
}

// TestTheFirstMCPCallProvisionsTheAccount calls a tool with the token Registre mints for
// an MCP resource. A person who never opened Sablier gets their account on that call.
func TestTheFirstMCPCallProvisionsTheAccount(t *testing.T) {
	s := newSablier(t)
	session := mcptest.Connect(t, s.resource, s.registre.token(t, s.resource), mcptest.ClientOptions{})
	if result := mcptest.Call(t, session, "sablier_timer_status", map[string]any{}); result.IsError {
		t.Fatalf("the first call must work: %s", mcptest.Text(result))
	}
	var user schemas.User
	if err := s.db.Where("email = ?", newEmail).First(&user).Error; err != nil {
		t.Fatalf("the first MCP call must create the account: %v", err)
	}
	if user.Kind != schemas.UserKindHuman || user.Name != "New Person" {
		t.Fatalf("provisioned account, got %+v", user)
	}
}

// comeBack re-creates the erased subject's account, as a race with the erasure would.
func (s *sablier) comeBack(t *testing.T) {
	t.Helper()
	back := schemas.User{Email: newEmail, Name: "New Person"}
	if err := s.db.Create(&back).Error; err != nil {
		t.Fatal(err)
	}
	if err := s.db.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, ?)`,
		back.ID, s.registre.URL, newSubject).Error; err != nil {
		t.Fatal(err)
	}
}

// pollErases polls Registre's deletions and expects the subject's account gone.
func (s *sablier) pollErases(t *testing.T, what string) {
	t.Helper()
	if err := s.deletions.Poll(context.Background()); err != nil || s.accounts(t, newEmail) != 0 {
		t.Fatalf("%s: %v, %d accounts left", what, err, s.accounts(t, newEmail))
	}
}

// TestAnErasedSubjectStaysErased walks porte's erased-subject recipe: provision, erase,
// a token minted before the erasure is refused, and an account that comes back for the
// subject is erased again on the next poll, without a restart and touching no one else.
func TestAnErasedSubjectStaysErased(t *testing.T) {
	s := newSablier(t)
	bystander := schemas.User{Email: "bystander@example.test", Name: "Bystander"}
	if err := s.db.Create(&bystander).Error; err != nil {
		t.Fatal(err)
	}
	early := s.registre.token(t, s.resource)
	if code := s.initialize(t, early); code != http.StatusOK {
		t.Fatalf("the first call provisions, got %d", code)
	}
	s.registre.erase(newSubject)
	s.pollErases(t, "the poll must erase the account")
	if code := s.initialize(t, early); code != http.StatusUnauthorized {
		t.Fatalf("an erased subject must not be provisioned again, got %d", code)
	}
	s.comeBack(t)
	s.pollErases(t, "an account back for an erased subject must be erased again")
	if s.accounts(t, bystander.Email) != 1 {
		t.Fatal("the erasure must touch no one else")
	}
}
