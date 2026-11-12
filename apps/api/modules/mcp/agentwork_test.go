package mcp_test

import (
	"encoding/json"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

// sessionOut is the part of a sablier_record_agent_session result the tests read.
type sessionOut struct {
	Status  string `json:"status"`
	EntryID int64  `json:"entry_id"`
	AgentID int64  `json:"agent_id"`
	Agent   string `json:"agent"`
	Minutes int    `json:"minutes"`
}

// TestMCPRecordsAnAgentSessionOnceByFacileID books the session Bulle books after a run,
// twice, and expects one entry on the owner's agent with the active time only.
func TestMCPRecordsAnAgentSessionOnceByFacileID(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner")
	stopped := time.Now().UTC().Add(-time.Minute).Truncate(time.Second)
	args := map[string]any{
		"agent_key": "bulle", "name": "Bulle", "facile_id": "bulle_run_42",
		"started_at": stopped.Add(-30 * time.Minute).Format(time.RFC3339Nano), "stopped_at": stopped.Format(time.RFC3339Nano),
		"active_seconds": 20 * 60, "cowork": true, "description": "Inbox triage", "title": "Inbox triage",
		"registre_space_id": "space-1", "opus_workspace": "studio", "opus_task_ids": []string{"opus_task_1"},
	}
	var first, second sessionOut
	call(t, session, "sablier_record_agent_session", args, &first)
	call(t, session, "sablier_record_agent_session", args, &second)
	if first.Status != "recorded" || first.Agent != "Bulle" || first.Minutes != 20 || second.EntryID != first.EntryID {
		t.Fatalf("one entry of 20 active minutes, booked twice: %+v then %+v", first, second)
	}
	var agent schemas.User
	must(t, f.db.First(&agent, first.AgentID).Error)
	if agent.Kind != schemas.UserKindAgent || agent.OwnerUserID == nil || *agent.OwnerUserID != f.users["owner"] {
		t.Fatalf("the session belongs to an agent of the owner, got %+v", agent)
	}
	var count int64
	must(t, f.db.Model(&schemas.TimeEntry{}).Where("facile_id = ?", "bulle_run_42").Count(&count).Error)
	if count != 1 {
		t.Fatalf("want 1 entry for the facile id, got %d", count)
	}
}

// TestMCPAgentSessionIsValidatedLikeREST refuses what POST /agents/sessions refuses.
func TestMCPAgentSessionIsValidatedLikeREST(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner")
	now := time.Now().UTC()
	base := map[string]any{"agent_key": "Bulle!", "name": "Bulle", "facile_id": "x",
		"started_at": now.Add(-time.Hour).Format(time.RFC3339), "stopped_at": now.Format(time.RFC3339)}
	if text := refused(t, session, "sablier_record_agent_session", base); !strings.Contains(strings.ToLower(text), "agent_key") {
		t.Fatalf("a bad agent_key must be named, got %q", text)
	}
	base["agent_key"], base["started_at"] = "bulle", "yesterday"
	if text := refused(t, session, "sablier_record_agent_session", base); !strings.Contains(strings.ToLower(text), "started_at") {
		t.Fatalf("a bad started_at must be named, got %q", text)
	}
}

// TestMCPRecordsAHeartbeatOnce stores Bulle's presence heartbeat, ignores its exact
// repeat and skips one older than the ingest window.
func TestMCPRecordsAHeartbeatOnce(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner")
	args := map[string]any{"at": time.Now().UTC().Format(time.RFC3339), "source": "bulle", "entity": "bulle:session/7",
		"registre_space_id": "space-1", "opus_task_id": "opus_task_1"}
	var out struct {
		Status   string `json:"status"`
		Accepted int    `json:"accepted"`
	}
	call(t, session, "sablier_record_heartbeat", args, &out)
	call(t, session, "sablier_record_heartbeat", args, &out)
	var count int64
	must(t, f.db.Model(&schemas.Heartbeat{}).Where("user_id = ? AND entity = ?", f.users["owner"], "bulle:session/7").Count(&count).Error)
	if out.Status != "recorded" || count != 1 {
		t.Fatalf("want one stored heartbeat, got %+v and %d rows", out, count)
	}
	args["at"] = time.Now().UTC().Add(-8 * 24 * time.Hour).Format(time.RFC3339)
	call(t, session, "sablier_record_heartbeat", args, &out)
	if out.Status != "skipped" {
		t.Fatalf("a heartbeat older than 7 days is skipped, got %+v", out)
	}
	args["source"] = "browser"
	if text := refused(t, session, "sablier_record_heartbeat", args); !strings.Contains(strings.ToLower(text), "source") {
		t.Fatalf("an unknown source must be named, got %q", text)
	}
}

// TestMCPAgentWorkToolsTakeBullesFields pins the argument names Bulle's agent_time.go
// maps its REST bodies onto.
func TestMCPAgentWorkToolsTakeBullesFields(t *testing.T) {
	f := newFixture(t)
	listed, err := f.connect(t, "owner").ListTools(t.Context(), nil)
	must(t, err)
	want := map[string][]string{
		"sablier_record_agent_session": {"agent_key", "name", "avatar_url", "facile_id", "started_at", "stopped_at",
			"active_seconds", "cowork", "description", "title", "registre_space_id", "opus_workspace", "opus_task_ids"},
		"sablier_record_heartbeat": {"at", "source", "entity", "project", "registre_space_id", "opus_task_id", "title"},
	}
	for _, tool := range listed.Tools {
		names, ok := want[tool.Name]
		if !ok {
			continue
		}
		delete(want, tool.Name)
		schema, _ := json.Marshal(tool.InputSchema)
		for _, name := range names {
			if !strings.Contains(string(schema), `"`+name+`"`) {
				t.Errorf("%s takes no %s: %s", tool.Name, name, schema)
			}
		}
	}
	if len(want) > 0 {
		t.Fatalf("tools not listed: %v", want)
	}
}
