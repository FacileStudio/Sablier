package decide

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
)

type seen struct {
	path, auth, actingFor string
	body                  map[string]any
}

func stub(t *testing.T, status int, got *seen) *httptest.Server {
	t.Helper()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got.path, got.auth, got.actingFor = r.URL.Path, r.Header.Get("Authorization"), r.Header.Get(bulleActingFor)
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

	decider, backend := New(Config{BulleURL: server.URL, BulleServiceKey: "k", JevAPIKey: "j"}, nil)
	answers, err := decider.Decide(context.Background(), "sara@example.test", "state", questions)
	if backend != BackendBulle || err != nil || answers["task"].Choice != "A" {
		t.Fatalf("bulle: backend=%s answers=%v err=%v", backend, answers, err)
	}
	if got.path != bulleDecidePath || got.auth != "Bearer k" || got.actingFor != "sara@example.test" || got.body["model"] != nil {
		t.Fatalf("bulle request: %+v", got)
	}

	decider, backend = New(Config{JevAPIKey: "j", JevBaseURL: server.URL + "/api", JevEndpoint: "decisions"}, nil)
	if _, err := decider.Decide(context.Background(), "sara@example.test", "state", questions); backend != BackendJev || err != nil {
		t.Fatalf("jev: backend=%s err=%v", backend, err)
	}
	if got.path != jevDecisionsPath || got.auth != "Bearer j" || got.actingFor != "" || got.body["model"] != jevDefaultModel {
		t.Fatalf("jev request: %+v", got)
	}

	if decider, backend = New(Config{}, nil); decider != nil || backend != BackendOff {
		t.Fatalf("no configuration must turn the AI step off, got %s", backend)
	}

	for _, status := range []int{http.StatusNotFound, http.StatusPaymentRequired} {
		refused := stub(t, status, &got)
		decider, _ = New(Config{BulleURL: refused.URL}, nil)
		if _, err := decider.Decide(context.Background(), "sara@example.test", "state", questions); err == nil {
			t.Fatalf("a %d from Bulle must be no answer", status)
		}
	}
}

func TestBulleRequestCarriesCriteria(t *testing.T) {
	var got seen
	server := stub(t, http.StatusOK, &got)
	decider, _ := New(Config{BulleURL: server.URL}, nil)
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
