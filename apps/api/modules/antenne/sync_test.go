package antenne

import (
	"encoding/json"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	enveloppe "github.com/FacileStudio/enveloppe/go"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

func wireEvent(t *testing.T, app enveloppe.App, object enveloppe.ObjectType, action enveloppe.Action, key string, payload map[string]any) json.RawMessage {
	t.Helper()
	data, err := json.Marshal(map[string]any{
		"version":         1,
		"app":             app,
		"object":          object,
		"action":          action,
		"facile_id":       payload["facile_id"],
		"payload":         payload,
		"timestamp":       "2026-10-23T10:00:00Z",
		"idempotency_key": key,
	})
	if err != nil {
		t.Fatal(err)
	}
	return data
}

func opusProject(t *testing.T, action enveloppe.Action, key, name string) json.RawMessage {
	return wireEvent(t, enveloppe.AppOpus, enveloppe.ObjectProject, action, key, map[string]any{
		"id": 7, "facile_id": "fac_p1", "workspace_id": 3, "name": name, "slug": "alpha",
	})
}

func opusTask(t *testing.T, action enveloppe.Action, key string, extra map[string]any) json.RawMessage {
	payload := map[string]any{
		"id": 9, "facile_id": "fac_t1", "project_id": 7, "project_facile_id": "fac_p1",
		"workspace_id": 3, "priority": "high", "due_date": "2026-11-01T00:00:00Z", "actor_id": 4,
	}
	for k, v := range extra {
		payload[k] = v
	}
	return wireEvent(t, enveloppe.AppOpus, enveloppe.ObjectTask, action, key, payload)
}

func countOf(s *Service, model any) int64 {
	var n int64
	s.orm.Model(model).Count(&n)
	return n
}

func TestInboundProjectLifecycle(t *testing.T) {
	s := newTestService(t)

	s.handleProjectCreated(opusProject(t, enveloppe.ActionCreated, "k1", "Alpha"), testMeta)
	var p schemas.Project
	if err := s.orm.Where("facile_id = ?", "fac_p1").First(&p).Error; err != nil || p.Name != "Alpha" {
		t.Fatalf("project not created: %v %+v", err, p)
	}

	s.handleProjectUpdated(opusProject(t, enveloppe.ActionUpdated, "k2", "Alpha 2"), testMeta)
	s.orm.First(&p, p.ID)
	if p.Name != "Alpha 2" {
		t.Fatalf("project not renamed: %q", p.Name)
	}

	s.handleProjectDeleted(opusProject(t, enveloppe.ActionDeleted, "k3", "Alpha 2"), testMeta)
	if countOf(s, &schemas.Project{}) != 0 {
		t.Fatal("project not deleted")
	}
}

func TestInboundProjectUpdateCreatesMissing(t *testing.T) {
	s := newTestService(t)
	s.handleProjectUpdated(opusProject(t, enveloppe.ActionUpdated, "k1", "Alpha"), testMeta)
	if countOf(s, &schemas.Project{}) != 1 {
		t.Fatal("update of unknown project must create it")
	}
}

func TestInboundTaskLifecycleWithOpusExtras(t *testing.T) {
	s := newTestService(t)
	s.handleProjectCreated(opusProject(t, enveloppe.ActionCreated, "k1", "Alpha"), testMeta)

	s.handleTaskCreated(opusTask(t, enveloppe.ActionCreated, "k2", map[string]any{"title": "Write docs", "name": "", "status": "open"}), testMeta)
	var task schemas.Task
	if err := s.orm.Where("facile_id = ?", "fac_t1").First(&task).Error; err != nil {
		t.Fatalf("task not created: %v", err)
	}
	if task.Name != "Write docs" || task.Status != schemas.StatusTodo {
		t.Fatalf("title/status not applied: %+v", task)
	}

	s.handleTaskUpdated(opusTask(t, enveloppe.ActionUpdated, "k3", map[string]any{"name": "Write docs", "status": "in progress"}), testMeta)
	s.orm.First(&task, task.ID)
	if task.Status != schemas.StatusInProgress {
		t.Fatalf("status not mapped: %q", task.Status)
	}

	s.handleTaskUpdated(opusTask(t, enveloppe.ActionUpdated, "k4", map[string]any{"name": "Write docs", "status": "weird column"}), testMeta)
	s.orm.First(&task, task.ID)
	if task.Status != schemas.StatusInProgress {
		t.Fatalf("unknown status must be kept on update, got %q", task.Status)
	}

	s.handleTaskDeleted(opusTask(t, enveloppe.ActionDeleted, "k5", map[string]any{"name": "Write docs"}), testMeta)
	if countOf(s, &schemas.Task{}) != 0 {
		t.Fatal("task not deleted")
	}
}

func TestInboundTaskUnknownStatusOnCreateDefaultsToTodo(t *testing.T) {
	s := newTestService(t)
	s.handleProjectCreated(opusProject(t, enveloppe.ActionCreated, "k1", "Alpha"), testMeta)
	s.handleTaskCreated(opusTask(t, enveloppe.ActionCreated, "k2", map[string]any{"name": "X", "status": "icebox"}), testMeta)
	var task schemas.Task
	s.orm.Where("facile_id = ?", "fac_t1").First(&task)
	if task.Status != schemas.StatusTodo {
		t.Fatalf("got %q", task.Status)
	}
}

func TestInboundTaskWithUnknownProjectIsDropped(t *testing.T) {
	s := newTestService(t)
	s.handleTaskCreated(opusTask(t, enveloppe.ActionCreated, "k1", map[string]any{"name": "Orphan", "status": "open"}), testMeta)
	if countOf(s, &schemas.Task{}) != 0 || countOf(s, &schemas.Project{}) != 0 {
		t.Fatal("no orphan task and no phantom project allowed")
	}
	if s.alreadyProcessed("k1") {
		t.Fatal("dropped event must stay replayable")
	}
}

func TestInboundIdempotentReplay(t *testing.T) {
	s := newTestService(t)
	p := opusProject(t, enveloppe.ActionCreated, "k1", "Alpha")
	s.handleProjectCreated(p, testMeta)
	s.handleProjectCreated(p, testMeta)
	if countOf(s, &schemas.Project{}) != 1 {
		t.Fatal("replay duplicated project")
	}

	s.handleProjectUpdated(opusProject(t, enveloppe.ActionUpdated, "k2", "Renamed"), testMeta)
	s.handleProjectUpdated(opusProject(t, enveloppe.ActionUpdated, "k2", "Other"), testMeta)
	var row schemas.Project
	s.orm.First(&row)
	if row.Name != "Renamed" {
		t.Fatalf("replayed key must be ignored, got %q", row.Name)
	}
}

func TestInboundIgnoresOwnEvents(t *testing.T) {
	s := newTestService(t)
	data := wireEvent(t, enveloppe.AppSablier, enveloppe.ObjectProject, enveloppe.ActionCreated, "k1", map[string]any{"facile_id": "fac_self", "name": "Mine"})
	s.handleProjectCreated(data, testMeta)
	if countOf(s, &schemas.Project{}) != 0 {
		t.Fatal("own event applied")
	}
}

func TestInboundDoesNotReEmit(t *testing.T) {
	s := newTestService(t)
	s.orm.Create(&schemas.AppSetting{ID: appSettingID, AntenneEnabled: true})
	s.handleProjectCreated(opusProject(t, enveloppe.ActionCreated, "k1", "Alpha"), testMeta)
	s.handleTaskCreated(opusTask(t, enveloppe.ActionCreated, "k2", map[string]any{"name": "T", "status": "done"}), testMeta)
	s.handleTaskDeleted(opusTask(t, enveloppe.ActionDeleted, "k3", nil), testMeta)
	s.handleProjectDeleted(opusProject(t, enveloppe.ActionDeleted, "k4", "Alpha"), testMeta)
	if n := countOf(s, &schemas.AntenneOutbox{}); n != 0 {
		t.Fatalf("inbound changes queued %d outbound events", n)
	}
}

func TestProjectDeleteRemovesItsTasks(t *testing.T) {
	s := newTestService(t)
	s.handleProjectCreated(opusProject(t, enveloppe.ActionCreated, "k1", "Alpha"), testMeta)
	s.handleTaskCreated(opusTask(t, enveloppe.ActionCreated, "k2", map[string]any{"name": "T", "status": "open"}), testMeta)
	s.handleProjectDeleted(opusProject(t, enveloppe.ActionDeleted, "k3", "Alpha"), testMeta)
	if countOf(s, &schemas.Task{}) != 0 {
		t.Fatal("orphan task left behind")
	}
}

func TestMapRemoteStatus(t *testing.T) {
	cases := map[string]string{
		"open": schemas.StatusTodo, "To Do": schemas.StatusTodo, "In Progress": schemas.StatusInProgress,
		"in_progress": schemas.StatusInProgress, "In Review": schemas.StatusInReview, "DONE": schemas.StatusDone,
		"closed": schemas.StatusDone,
	}
	for in, want := range cases {
		if got, ok := mapRemoteStatus(in); !ok || got != want {
			t.Errorf("%q: got %q %v want %q", in, got, ok, want)
		}
	}
	if _, ok := mapRemoteStatus("icebox"); ok {
		t.Error("unknown status must not map")
	}
}

func TestLocalEmitCarriesFacileIDs(t *testing.T) {
	s := newTestService(t)
	s.orm.Create(&schemas.AppSetting{ID: appSettingID, AntenneEnabled: true})
	project := seedProject(t, s.orm, "Local")
	task := &schemas.Task{ProjectID: project.ID, Name: "T"}
	s.orm.Create(task)

	s.EmitTaskEvent(enveloppe.ActionCreated, task, project)

	var rows []schemas.AntenneOutbox
	s.orm.Order("id").Find(&rows)
	if len(rows) != 2 || rows[0].Channel != "project.created" || rows[1].Channel != "task.created" {
		t.Fatalf("unexpected outbox: %+v", rows)
	}
	var evt enveloppe.Event[enveloppe.Task]
	json.Unmarshal([]byte(rows[1].Payload), &evt)
	if evt.App != enveloppe.AppSablier || evt.Payload.FacileID == "" || evt.Payload.ProjectFacileID == "" {
		t.Fatalf("event missing identity: %+v", evt)
	}
}

func yamlList(t *testing.T, doc, section string) []string {
	t.Helper()
	block := regexp.MustCompile(`(?m)^  ` + section + `:\n((?:    - .*\n?)+)`).FindStringSubmatch(doc)
	if block == nil {
		t.Fatalf("section %s not found", section)
	}
	var out []string
	for _, line := range strings.Split(strings.TrimSpace(block[1]), "\n") {
		out = append(out, strings.TrimSpace(strings.TrimPrefix(strings.TrimSpace(line), "- ")))
	}
	sort.Strings(out)
	return out
}

func TestRegistrationMatchesAntenneYaml(t *testing.T) {
	raw, err := os.ReadFile("../../../../antenne.yaml")
	if err != nil {
		t.Fatal(err)
	}
	doc := string(raw)
	for section, code := range map[string][]string{"emit": EmitEvents(), "listen": ListenEvents()} {
		got := append([]string(nil), code...)
		sort.Strings(got)
		if strings.Join(got, ",") != strings.Join(yamlList(t, doc, section), ",") {
			t.Errorf("%s drift: code %v yaml %v", section, got, yamlList(t, doc, section))
		}
	}
}
