package mcp_test

import (
	"context"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/mcp"
	"github.com/FacileStudio/Sablier/apps/api/modules/projects"
	"github.com/FacileStudio/Sablier/apps/api/modules/spaces"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/mcpauth"
	"github.com/FacileStudio/tronc/mcp/mcptest"
	"github.com/go-chi/chi/v5"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
	"gorm.io/gorm"
)

const spaceID = "7a1d1f3e-5b8c-4c2a-9a51-0d6c1e2b3f40"

// identities resolves each test subject to its user.
type identities struct {
	porte.IdentityStore
	users map[string]int64
}

// Find resolves a known subject.
func (s identities) Find(_ context.Context, provider, sub string) (porte.StoredIdentity, error) {
	id, ok := s.users[sub]
	if !ok {
		return porte.StoredIdentity{}, porte.ErrNotFound
	}
	return porte.StoredIdentity{UserID: id, Provider: provider, Subject: sub}, nil
}

// fixture is the MCP endpoint over Sablier's real services: an owner and a teammate
// sharing a space, a stranger outside it, and an agent of the owner.
type fixture struct {
	db       *gorm.DB
	issuer   *mcptest.Issuer
	resource string
	users    map[string]int64
	project  schemas.Project
	shared   schemas.Project
	task     schemas.Task
	router   chi.Router
	front    func(http.Handler) http.Handler
}

// newFixture starts the endpoint and the confirm page routes on any free port.
func newFixture(t *testing.T) *fixture {
	t.Helper()
	return newFixtureOn(t, "")
}

// newFixtureOn starts the endpoint on addr, any free port when empty.
func newFixtureOn(t *testing.T, addr string) *fixture {
	t.Helper()
	db := testdb.Open(t)
	must(t, db.AutoMigrate(&schemas.AppSetting{}))
	f := &fixture{db: db, issuer: mcptest.NewIssuer(t), users: map[string]int64{}, front: func(next http.Handler) http.Handler { return next }}
	f.seed(t)
	var handler http.Handler
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { f.front(handler).ServeHTTP(w, r) }))
	if addr != "" {
		listener, err := net.Listen("tcp", addr)
		must(t, err)
		srv.Listener = listener
	}
	srv.Start()
	t.Cleanup(srv.Close)
	f.resource = srv.URL + "/mcp"
	service := f.service(t)
	f.router = chi.NewRouter()
	mcp.RegisterRoutes(f.router, service)
	f.router.Route("/api", func(r chi.Router) { mcp.RegisterPendingRoutes(r, service, sessions{db: db}) })
	handler = f.router
	return f
}

// service builds the MCP server over the fixture's database and issuer.
func (f *fixture) service(t *testing.T) *mcp.Service {
	t.Helper()
	verifier, err := mcpauth.Verifier(context.Background(), mcpauth.Config{
		Issuer: f.issuer.URL, Resource: f.resource, ScopePrefix: "sablier", Identities: identities{users: f.users},
	})
	must(t, err)
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	service, err := mcp.NewService(mcp.Deps{
		Entries: timeentries.NewService(f.db), Projects: projects.NewService(f.db), Spaces: spaces.NewService(f.db),
		Accounts: auth.NewService(f.db, nil, logger),
	}, mcp.Config{
		Resource: f.resource, Issuer: f.issuer.URL, Version: "test", StateKey: []byte("state-key"),
		Verifier: verifier, Pending: mcptest.NewPending(), PerMinute: 100_000, Logger: logger,
	})
	must(t, err)
	return service
}

// seed creates the people, the space, a personal and a shared project.
func (f *fixture) seed(t *testing.T) {
	t.Helper()
	for _, name := range []string{"owner", "teammate", "stranger"} {
		user := schemas.User{Email: name + "@example.test", Name: name, Kind: schemas.UserKindHuman}
		must(t, f.db.Create(&user).Error)
		f.users[name] = user.ID
	}
	owner := f.users["owner"]
	agent := schemas.User{Email: "agent@example.test", Name: "Claude", Kind: schemas.UserKindAgent, OwnerUserID: &owner}
	must(t, f.db.Create(&agent).Error)
	f.users["agent"] = agent.ID
	must(t, f.db.Create(&schemas.Space{ID: spaceID, Name: "Studio"}).Error)
	for _, name := range []string{"owner", "teammate"} {
		must(t, f.db.Create(&schemas.SpaceMember{SpaceID: spaceID, UserID: f.users[name], Role: schemas.RoleMember}).Error)
	}
	space := spaceID
	f.project = schemas.Project{Name: "Personal site", OwnerID: owner}
	f.shared = schemas.Project{Name: "Acme", OwnerID: owner, SpaceID: &space}
	must(t, f.db.Create(&f.project).Error)
	must(t, f.db.Create(&f.shared).Error)
	f.task = schemas.Task{ProjectID: f.shared.ID, Name: "Design"}
	must(t, f.db.Create(&f.task).Error)
}

// token is a token for the owner with scopes, or both when none is given.
func (f *fixture) token(t testing.TB, scopes ...string) string {
	if len(scopes) == 0 {
		scopes = []string{"sablier.read", "sablier.write"}
	}
	return f.issuer.Token(t, f.resource, "owner", scopes...)
}

// connect opens a session for subject with both scopes.
func (f *fixture) connect(t *testing.T, subject string) *sdk.ClientSession {
	t.Helper()
	return mcptest.Connect(t, f.resource, f.issuer.Token(t, f.resource, subject, "sablier.read", "sablier.write"), mcptest.ClientOptions{})
}

// entry books a finished entry for user on the shared project's task.
func (f *fixture) entry(t testing.TB, user string, start time.Time, length time.Duration) schemas.TimeEntry {
	t.Helper()
	stop, space := start.Add(length), spaceID
	e := schemas.TimeEntry{ProjectID: f.shared.ID, TaskID: f.task.ID, UserID: f.users[user], SpaceID: &space, StartedAt: start, StoppedAt: &stop}
	must(t, f.db.Create(&e).Error)
	return e
}

// exists reports 1 while entry id is stored, 0 once it is gone.
func (f *fixture) exists(id int64) int {
	var count int64
	f.db.Model(&schemas.TimeEntry{}).Where("id = ?", id).Count(&count)
	return int(count)
}

// call runs a tool and decodes its result, failing on a tool error.
func call(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any, out any) {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if result.IsError {
		t.Fatalf("%s: %s", tool, mcptest.Text(result))
	}
	mcptest.Decode(t, result, out)
}

// refused runs a tool and returns its error text, failing when it worked.
func refused(t *testing.T, session *sdk.ClientSession, tool string, args map[string]any) string {
	t.Helper()
	result := mcptest.Call(t, session, tool, args)
	if !result.IsError {
		t.Fatalf("%s must be refused, got %s", tool, mcptest.Text(result))
	}
	return mcptest.Text(result)
}

// must fails the test on err.
func must(t testing.TB, err error) {
	t.Helper()
	if err != nil {
		t.Fatal(err)
	}
}

// sessions stands in for porte on the confirm page: X-Test-User is the signed-in user,
// and X-Test-Actor makes the request a delegated token.
type sessions struct {
	db *gorm.DB
}

// RequireAuth attaches the test identity.
func (s sessions) RequireAuth(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		id, err := strconv.ParseInt(r.Header.Get("X-Test-User"), 10, 64)
		if err != nil {
			http.Error(w, "unauthenticated", http.StatusUnauthorized)
			return
		}
		identity := porte.Identity{UserID: id, SessionID: 1, Actor: r.Header.Get("X-Test-Actor")}
		if identity.Actor != "" {
			identity.SessionID = 0
		}
		next.ServeHTTP(w, r.WithContext(porte.WithIdentity(r.Context(), identity)))
	})
}

// IdentityForUser is the auth service's lookup.
func (s sessions) IdentityForUser(ctx context.Context, userID int64) (string, string, error) {
	return auth.NewService(s.db, nil, slog.Default()).IdentityForUser(ctx, userID)
}
