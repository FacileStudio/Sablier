package mcp_test

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"strings"
	"testing"
	"time"
)

// TestMCPRespectsSpaces checks a stranger reaches nothing of a space, and a teammate sees
// the space's time but changes only their own, exactly like the REST API.
func TestMCPRespectsSpaces(t *testing.T) {
	f := newFixture(t)
	mine := f.entry(t, "owner", time.Now().Add(-3*time.Hour), time.Hour)
	stranger := f.connect(t, "stranger")
	refused(t, stranger, "sablier_start_timer", map[string]any{"project_id": f.shared.ID, "task_id": f.task.ID})
	refused(t, stranger, "sablier_list_entries", map[string]any{"space": spaceID})
	refused(t, stranger, "sablier_delete_entry", map[string]any{"entry_id": mine.ID})
	var projects struct {
		Projects []struct{ ID int64 } `json:"projects"`
	}
	call(t, stranger, "sablier_list_projects", nil, &projects)
	if len(projects.Projects) != 0 {
		t.Fatalf("a stranger must see no project, got %+v", projects)
	}
	checkTeammate(t, f, mine.ID)
}

// checkTeammate sees the space's time with who booked it, but cannot edit it, and the
// owner finds the shared project by a task name.
func checkTeammate(t *testing.T, f *fixture, entryID int64) {
	t.Helper()
	teammate := f.connect(t, "teammate")
	var list struct {
		Entries []struct {
			ID     int64  `json:"id"`
			Person string `json:"person"`
		} `json:"entries"`
	}
	call(t, teammate, "sablier_list_entries", map[string]any{"space": spaceID, "everyone": true}, &list)
	if len(list.Entries) != 1 || list.Entries[0].Person != "owner" {
		t.Fatalf("a teammate sees the space's time with who booked it, got %+v", list)
	}
	text := refused(t, teammate, "sablier_update_entry", map[string]any{"entry_id": entryID, "task": "Mine now"})
	if !strings.Contains(text, "sablier_list_entries") {
		t.Fatalf("editing someone else's entry must be refused with a hint, got %q", text)
	}
	var projects struct {
		Projects []struct{ ID int64 } `json:"projects"`
	}
	call(t, f.connect(t, "owner"), "sablier_list_projects", map[string]any{"query": "desi"}, &projects)
	if len(projects.Projects) != 1 || projects.Projects[0].ID != f.shared.ID {
		t.Fatalf("a task name must find its project, got %+v", projects)
	}
}

// TestMCPAgentTokenIsRefused checks a token naming an agent user never acts, as
// RequireAuth refuses it.
func TestMCPAgentTokenIsRefused(t *testing.T) {
	f := newFixture(t)
	refused(t, f.connect(t, "agent"), "sablier_timer_status", nil)
}

// TestMCPConfirmPage runs a delete from a client that cannot ask, then confirms it on
// the page, which refuses a delegated token.
func TestMCPConfirmPage(t *testing.T) {
	f := newFixture(t)
	e := f.entry(t, "owner", time.Now().Add(-2*time.Hour), time.Hour)
	var out struct {
		Status     string `json:"status"`
		ConfirmURL string `json:"confirm_url"`
	}
	call(t, f.connect(t, "owner"), "sablier_delete_entry", map[string]any{"entry_id": e.ID}, &out)
	if out.Status != "pending_confirmation" || f.exists(e.ID) != 1 {
		t.Fatalf("nothing is deleted before the user confirms, got %+v", out)
	}
	id := out.ConfirmURL[strings.LastIndex(out.ConfirmURL, "=")+1:]
	owner := strconv.FormatInt(f.users["owner"], 10)
	if code, body := f.page(http.MethodGet, "/api/mcp/pending", owner, ""); code != 200 || !strings.Contains(body, `"project":"Acme"`) {
		t.Fatalf("the page lists the entry, got %d %s", code, body)
	}
	if code, _ := f.page(http.MethodPost, "/api/mcp/pending/"+id+"/confirm", owner, "bulle"); code != http.StatusForbidden {
		t.Fatalf("a delegated token must not confirm, got %d", code)
	}
	cli := httptest.NewRequest(http.MethodPost, "/api/mcp/pending/"+id+"/confirm", strings.NewReader("{}"))
	cli.Header.Set("X-Test-User", owner)
	cli.Header.Set("X-Test-CLI", "1")
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, cli)
	if recorder.Code != http.StatusForbidden || f.exists(e.ID) != 1 {
		t.Fatalf("a CLI session must not confirm, got %d", recorder.Code)
	}
	if code, _ := f.page(http.MethodPost, "/api/mcp/pending/"+id+"/confirm", strconv.FormatInt(f.users["teammate"], 10), ""); code != 404 {
		t.Fatalf("someone else cannot confirm, got %d", code)
	}
	if code, _ := f.page(http.MethodPost, "/api/mcp/pending/"+id+"/confirm", owner, ""); code != 200 || f.exists(e.ID) != 0 {
		t.Fatalf("the owner's confirmation deletes the entry, got %d", code)
	}
	if code, _ := f.page(http.MethodPost, "/api/mcp/pending/"+id+"/confirm", owner, ""); code != 404 {
		t.Fatalf("a confirmation runs once, got %d", code)
	}
}

// page calls a confirm page route as user, through a delegated token when actor is set.
func (f *fixture) page(method, path, user, actor string) (int, string) {
	req := httptest.NewRequest(method, path, strings.NewReader("{}"))
	req.Header.Set("X-Test-User", user)
	if actor != "" {
		req.Header.Set("X-Test-Actor", actor)
	}
	recorder := httptest.NewRecorder()
	f.router.ServeHTTP(recorder, req)
	return recorder.Code, recorder.Body.String()
}
