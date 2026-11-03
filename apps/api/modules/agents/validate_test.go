package agents

import (
	"testing"
	"time"
)

func TestToSessionValidatesAtTheBoundary(t *testing.T) {
	now := time.Date(2026, 11, 3, 12, 0, 0, 0, time.UTC)
	started, stopped := now.Add(-10*time.Minute), now
	active := int64(240)
	valid := func() RecordSessionRequest {
		return RecordSessionRequest{AgentKey: "bulle", Name: "Jarvis", FacileID: "bulle_run_1",
			StartedAt: &started, StoppedAt: &stopped, ActiveSeconds: &active}
	}

	session, err := toSession(1, ptr(valid()), now)
	if err != nil {
		t.Fatal(err)
	}
	if session.PausedMs != 6*60*1000 || session.Agent.AvatarURL != "https://blobatar.dev/avatar/Jarvis?gen=2" {
		t.Fatalf("session: %+v", session)
	}

	future := now.Add(time.Hour)
	bad := map[string]func(*RecordSessionRequest){
		"uppercase key":     func(r *RecordSessionRequest) { r.AgentKey = "Bulle" },
		"http avatar":       func(r *RecordSessionRequest) { r.AvatarURL = "http://example.com/a.png" },
		"javascript avatar": func(r *RecordSessionRequest) { r.AvatarURL = "javascript:alert(1)" },
		"missing facile id": func(r *RecordSessionRequest) { r.FacileID = " " },
		"stopped first":     func(r *RecordSessionRequest) { r.StoppedAt, r.StartedAt = &started, &stopped },
		"future stop":       func(r *RecordSessionRequest) { r.StoppedAt = &future },
		"no times":          func(r *RecordSessionRequest) { r.StartedAt = nil },
	}
	for name, mutate := range bad {
		req := valid()
		mutate(&req)
		if _, err := toSession(1, &req, now); err == nil {
			t.Errorf("%s: expected a validation error", name)
		}
	}
}

func ptr[T any](v T) *T { return &v }
