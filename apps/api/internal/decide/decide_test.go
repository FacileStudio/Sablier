package decide

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/FacileStudio/tronc/bulle"
)

type seen struct {
	path, auth, actingFor string
	body                  map[string]any
}

func stub(t *testing.T, status int, got *seen) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path, got.auth, got.actingFor = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get(bulle.ActingForHeader)
		if err := json.NewDecoder(r.Body).Decode(&got.body); err != nil {
			t.Errorf("decode: %v", err)
		}
		w.WriteHeader(status)
		_, _ = w.Write([]byte(`{"answers":{"task":{"choice":"A","confidence":0.9,"probabilities":{"A":0.9}}}}`))
	}))
	t.Cleanup(server.Close)
	return server
}

func TestBackends(t *testing.T) {
	questions := map[string]Question{"task": {Type: "choice", Instructions: "pick", Options: []string{"A", "B"}}}
	var got seen
	server := stub(t, http.StatusOK, &got)

	decider, backend, err := New(Config{BulleURL: server.URL, BulleServiceKey: "k"})
	if backend != BackendBulle || err != nil {
		t.Fatalf("bulle: backend=%s err=%v", backend, err)
	}
	answers, err := decider.Decide(context.Background(), "sara@example.test", "state", questions)
	if err != nil || answers["task"].Choice != "A" {
		t.Fatalf("bulle: answers=%v err=%v", answers, err)
	}
	if got.path != "/api/ai/decide" || got.auth != "Bearer k" || got.actingFor != "sara@example.test" || got.body["model"] != nil {
		t.Fatalf("bulle request: %+v", got)
	}

	for _, cfg := range []Config{{}, {BulleURL: server.URL}, {BulleServiceKey: "k"}} {
		if decider, backend, err := New(cfg); decider != nil || backend != BackendOff || err != nil {
			t.Fatalf("%+v must turn the AI step off, got %s %v", cfg, backend, err)
		}
	}
	if _, backend, err := New(Config{BulleURL: "ftp://bulle", BulleServiceKey: "k"}); backend != BackendOff || err == nil {
		t.Fatalf("an invalid BULLE_URL must be refused, got %s %v", backend, err)
	}

	for _, status := range []int{http.StatusNotFound, http.StatusPaymentRequired, http.StatusFound} {
		refused := stub(t, status, &got)
		decider, _, _ = New(Config{BulleURL: refused.URL, BulleServiceKey: "k"})
		if _, err := decider.Decide(context.Background(), "sara@example.test", "state", questions); err == nil {
			t.Fatalf("a %d from Bulle must be no answer", status)
		}
	}
}

func TestBulleRequestCarriesCriteria(t *testing.T) {
	var got seen
	server := stub(t, http.StatusOK, &got)
	decider, _, _ := New(Config{BulleURL: server.URL, BulleServiceKey: "k"})
	questions := map[string]Question{"task": {Type: "choice", Instructions: "pick",
		Criteria: map[string]string{"A": "work on A", "B": "work on B"}, Options: []string{"A", "B"}}}
	if _, err := decider.Decide(context.Background(), "sara@example.test", "state", questions); err != nil {
		t.Fatal(err)
	}
	sent, _ := got.body["questions"].(map[string]any)["task"].(map[string]any)
	if criteria, ok := sent["criteria"].(map[string]any); !ok || criteria["A"] != "work on A" || criteria["B"] != "work on B" {
		t.Fatalf("a choice must reach Bulle with its criteria keyed by option, got %v", sent)
	}
}
