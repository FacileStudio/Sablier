package mcp_test

import (
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/tronc/mcp/mcptest"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// timerOut is the part of a timer result the tests read.
type timerOut struct {
	Status string `json:"status"`
	Entry  struct {
		ID      int64  `json:"id"`
		Project string `json:"project"`
		Task    string `json:"task"`
		TaskID  int64  `json:"task_id"`
		State   string `json:"state"`
	} `json:"entry"`
}

// TestMCPConformance runs tronc's conformance kit, with deleting an entry as the
// sensitive tool.
func TestMCPConformance(t *testing.T) {
	f := newFixture(t)
	mcptest.Run(t, mcptest.Config{
		URL:     f.resource,
		Token:   f.token,
		Confirm: []mcptest.ConfirmCase{{Tool: "sablier_delete_entry", Setup: f.freshEntry}},
	})
}

// freshEntry books an entry to delete and counts how many times it was deleted.
func (f *fixture) freshEntry(tb testing.TB) (map[string]any, func() int) {
	e := f.entry(tb, "owner", time.Now().Add(-2*time.Hour), time.Hour)
	return map[string]any{"entry_id": e.ID}, func() int { return 1 - f.exists(e.ID) }
}

// TestMCPTimerFlow walks a model's usual path: check, start on a new task by name, refuse
// a second timer, stop, stop again.
func TestMCPTimerFlow(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner")
	var out timerOut
	call(t, session, "sablier_timer_status", nil, &out)
	if out.Status != "stopped" {
		t.Fatalf("no timer runs yet, got %+v", out)
	}
	call(t, session, "sablier_start_timer", map[string]any{"project_id": f.shared.ID, "task": "Client call"}, &out)
	if out.Status != "started" || out.Entry.Task != "Client call" || out.Entry.Project != "Acme" || out.Entry.State != "running" {
		t.Fatalf("the timer must start on a new task, got %+v", out)
	}
	again := refused(t, session, "sablier_start_timer", map[string]any{"project_id": f.shared.ID, "task_id": f.task.ID})
	if !strings.Contains(again, "sablier_stop_timer") {
		t.Fatalf("a second timer must point at sablier_stop_timer, got %q", again)
	}
	var stopped, twice timerOut
	call(t, session, "sablier_stop_timer", nil, &stopped)
	if stopped.Status != "stopped" || stopped.Entry.State != "" || stopped.Entry.ID != out.Entry.ID {
		t.Fatalf("stop must finish the entry, got %+v", stopped)
	}
	call(t, session, "sablier_stop_timer", nil, &twice)
	if twice.Status != "not_running" {
		t.Fatalf("a second stop changes nothing, got %+v", twice)
	}
	var reused timerOut
	call(t, session, "sablier_start_timer", map[string]any{"project_id": f.shared.ID, "task": "client CALL"}, &reused)
	if reused.Entry.TaskID != out.Entry.TaskID {
		t.Fatalf("a task name already there must be reused, got %d and %d", reused.Entry.TaskID, out.Entry.TaskID)
	}
}

// TestMCPAddEditList adds an entry in local time, edits it, and finds it in the list
// with the range totals.
func TestMCPAddEditList(t *testing.T) {
	f := newFixture(t)
	session := f.connect(t, "owner")
	var added struct {
		Status string `json:"status"`
		Entry  struct {
			ID        int64  `json:"id"`
			StartedAt string `json:"started_at"`
			Minutes   int    `json:"minutes"`
		} `json:"entry"`
	}
	call(t, session, "sablier_add_entry", map[string]any{
		"project_id": f.shared.ID, "task_id": f.task.ID, "started_at": "2026-11-09T09:00", "duration_minutes": 90,
		"timezone": "Europe/Paris",
	}, &added)
	if added.Status != "created" || added.Entry.Minutes != 90 || added.Entry.StartedAt != "2026-11-09T09:00:00+01:00" {
		t.Fatalf("want 90 minutes from 09:00 Paris time, got %+v", added)
	}
	call(t, session, "sablier_update_entry", map[string]any{"entry_id": added.Entry.ID, "stopped_at": "2026-11-09T10:00:00Z"}, &added)
	if added.Status != "updated" || added.Entry.Minutes != 120 {
		t.Fatalf("the new end must keep the start and make 120 minutes, got %+v", added)
	}
	checkListed(t, session, added.Entry.ID)
}

// checkListed finds the edited entry in its day with the range totals.
func checkListed(t *testing.T, session *sdk.ClientSession, id int64) {
	t.Helper()
	var list struct {
		Total struct {
			Minutes int `json:"minutes"`
			Entries int `json:"entries"`
		} `json:"total"`
		ByProject []struct{ Name string } `json:"by_project"`
		Entries   []struct{ ID int64 }    `json:"entries"`
	}
	call(t, session, "sablier_list_entries", map[string]any{"from": "2026-11-09", "to": "2026-11-09", "timezone": "Europe/Paris"}, &list)
	if list.Total.Minutes != 120 || list.Total.Entries != 1 || list.ByProject[0].Name != "Acme" || list.Entries[0].ID != id {
		t.Fatalf("the edited entry must count 120 minutes, got %+v", list)
	}
	if text := refused(t, session, "sablier_update_entry", map[string]any{"entry_id": id}); !strings.Contains(text, "at least one") {
		t.Fatalf("an empty edit must say what to pass, got %q", text)
	}
}
