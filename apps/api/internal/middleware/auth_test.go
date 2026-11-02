package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
)

type fakeVerifier struct{ scopes []string }

func (f fakeVerifier) VerifyJWT(context.Context, string) (porte.Identity, error) {
	return porte.Identity{UserID: 7, Subject: "user-7", Scopes: f.scopes, Actor: "bulle"}, nil
}

type fakeStore struct{ porte.SessionStore }

type fakeAuth struct{ sessions *session.Manager }

func (f fakeAuth) RequireAuth(next http.Handler) http.Handler {
	return f.sessions.RequireAuth(next)
}

func (fakeAuth) IdentityForUser(context.Context, int64) (string, string, error) {
	return "7", "user@example.test", nil
}

func serve(t *testing.T, scopes []string, method, path string) int {
	t.Helper()
	manager, err := session.New(
		porte.Config{Issuer: "https://sso.example.test", MachineAudience: "sablier-client", ScopePrefix: "sablier"},
		session.Deps{Sessions: fakeStore{}},
	)
	if err != nil {
		t.Fatal(err)
	}
	manager.WithJWT(fakeVerifier{scopes: scopes})
	handler := http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(http.StatusOK) })
	wrapped := RequireAuth(fakeAuth{sessions: manager})(handler)
	request := httptest.NewRequest(method, path, nil)
	request.Header.Set("Authorization", "Bearer aaa.bbb.ccc")
	recorder := httptest.NewRecorder()
	wrapped.ServeHTTP(recorder, request)
	return recorder.Code
}

func TestDelegatedJWTScopeGate(t *testing.T) {
	cases := []struct {
		name   string
		scopes []string
		method string
		path   string
		want   int
	}{
		{"read scope reads the running timer", []string{"sablier.read"}, http.MethodGet, "/api/time-entries/running", http.StatusOK},
		{"read scope cannot start a timer", []string{"sablier.read"}, http.MethodPost, "/api/time-entries/start", http.StatusForbidden},
		{"read scope cannot stop a timer", []string{"sablier.read"}, http.MethodPost, "/api/time-entries/stop", http.StatusForbidden},
		{"another app's scope reads nothing", []string{"courrier.read"}, http.MethodGet, "/api/time-entries/running", http.StatusForbidden},
		{"write scope can start a timer", []string{"sablier.write"}, http.MethodPost, "/api/time-entries/start", http.StatusOK},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := serve(t, tc.scopes, tc.method, tc.path); got != tc.want {
				t.Fatalf("%s %s with %v: got %d, want %d", tc.method, tc.path, tc.scopes, got, tc.want)
			}
		})
	}
}
