package mcp_test

import (
	"net/http"
	"os"
	"os/signal"
	"testing"
	"time"
)

// TestServe serves the endpoint on MCP_SERVE for the official conformance suite, with
// the owner's token added to every request that carries none, until interrupted.
func TestServe(t *testing.T) {
	addr := os.Getenv("MCP_SERVE")
	if addr == "" {
		t.Skip("set MCP_SERVE=127.0.0.1:<port> to serve the endpoint")
	}
	f := newFixtureOn(t, addr)
	token := f.token(t)
	f.front = func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Header.Get("Authorization") == "" {
				r.Header.Set("Authorization", "Bearer "+token)
			}
			next.ServeHTTP(w, r)
		})
	}
	f.entry(t, "owner", time.Now().Add(-2*time.Hour), time.Hour)
	t.Logf("serving %s", f.resource)
	stop := make(chan os.Signal, 1)
	signal.Notify(stop, os.Interrupt)
	<-stop
}
