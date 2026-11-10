package antenne

import (
	"encoding/json"
	"testing"
	"time"

	"github.com/FacileStudio/antenne-client/go/envelope"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

const (
	smallID = "AAAAAAAAAAAAAAAAAAAAAAAAAA"
	largeID = "fac_ffffffffffffffffffff"
)

type queued struct {
	channel string
	key     string
	payload map[string]any
	raw     []byte
}

func queuedEvents(t *testing.T, s *Service) []queued {
	t.Helper()
	var rows []struct {
		Channel string
		Payload string
	}
	if err := s.orm.Raw(`SELECT channel, payload::text AS payload FROM antenne_outbox WHERE sent_at IS NULL ORDER BY id`).
		Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	out := make([]queued, len(rows))
	for i, row := range rows {
		var evt struct {
			Key     string         `json:"idempotency_key"`
			Payload map[string]any `json:"payload"`
		}
		if err := json.Unmarshal([]byte(row.Payload), &evt); err != nil {
			t.Fatal(err)
		}
		out[i] = queued{channel: row.Channel, key: evt.Key, payload: evt.Payload, raw: []byte(row.Payload)}
	}
	return out
}

func projectEvent(t *testing.T, action envelope.Action, key, id, name string, space map[string]any, updatedAt string) []byte {
	return wireEvent(t, envelope.AppOpus, envelope.ObjectProject, action, key, map[string]any{
		"facile_id": id, "name": name, "space": space, "updated_at": updatedAt,
	})
}

func personal(email string) map[string]any {
	return map[string]any{"registre_space_id": "", "personal": true, "owner_subject": "", "owner_email": email}
}

func enableEmit(s *Service) {
	s.orm.Create(&schemas.AppSetting{ID: appSettingID, AntenneEnabled: true})
}

func TestInboundProjectResolvesRegistreSpaceAndSubject(t *testing.T) {
	s := newSyncService(t)
	alice := seedUser(t, s.orm, "alice@example.com")
	s.orm.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-alice')`, alice.ID, s.issuer)
	reg := "reg-1"
	space := schemas.Space{Name: "Team", RegistreSpaceID: &reg}
	s.orm.Create(&space)

	shared := map[string]any{"registre_space_id": reg, "personal": false, "owner_subject": "sub-alice", "owner_email": "x@y.z"}
	s.handleProjectCreated(projectEvent(t, envelope.ActionCreated, "k1", testProjectID, "Shared", shared, ""), testMeta)
	mine := map[string]any{"registre_space_id": "", "personal": true, "owner_subject": "sub-alice", "owner_email": "nobody@x.z"}
	s.handleProjectCreated(projectEvent(t, envelope.ActionCreated, "k2", testTaskID, "Mine", mine, ""), testMeta)

	var p schemas.Project
	s.orm.Where("facile_id = ?", testProjectID).First(&p)
	if p.SpaceID == nil || *p.SpaceID != space.ID || p.OwnerID != alice.ID {
		t.Fatalf("registre space not resolved: %+v", p)
	}
	var own schemas.Project
	s.orm.Where("facile_id = ?", testTaskID).First(&own)
	if own.SpaceID != nil || own.OwnerID != alice.ID {
		t.Fatalf("personal owner not resolved by subject: %+v", own)
	}
}

func TestInboundProjectWithUnresolvableSpaceStaysReplayable(t *testing.T) {
	s := newSyncService(t)
	missing := map[string]any{"registre_space_id": "reg-unknown", "personal": false}
	s.handleProjectCreated(projectEvent(t, envelope.ActionCreated, "k1", testProjectID, "A", missing, ""), testMeta)
	s.handleProjectCreated(projectEvent(t, envelope.ActionCreated, "k2", testProjectID, "A", personal("ghost@x.z"), ""), testMeta)
	if countOf(s, &schemas.Project{}) != 0 || s.alreadyProcessed("k1") || s.alreadyProcessed("k2") {
		t.Fatal("unresolvable space must skip without consuming the event")
	}
}

func TestInboundProjectLastWriterWins(t *testing.T) {
	s := newSyncService(t)
	now := time.Now().UTC()
	s.handleProjectCreated(projectEvent(t, envelope.ActionCreated, "k1", testProjectID, "A", personal("owner@example.com"), now.Format(time.RFC3339Nano)), testMeta)
	older := now.Add(-time.Minute).Format(time.RFC3339Nano)
	s.handleProjectUpdated(projectEvent(t, envelope.ActionUpdated, "k2", testProjectID, "Old", personal("owner@example.com"), older), testMeta)
	var p schemas.Project
	s.orm.First(&p)
	if p.Name != "A" {
		t.Fatalf("older update applied: %q", p.Name)
	}
	newer := now.Add(time.Minute).Format(time.RFC3339Nano)
	s.handleProjectUpdated(projectEvent(t, envelope.ActionUpdated, "k3", testProjectID, "New", personal("owner@example.com"), newer), testMeta)
	s.orm.First(&p)
	if p.Name != "New" {
		t.Fatalf("newer update not applied: %q", p.Name)
	}
}

func TestProjectTwinMergeConvergesOnSmallerID(t *testing.T) {
	s := newSyncService(t)
	enableEmit(s)
	keep := largeID
	local := schemas.Project{Name: "Alpha", OwnerID: 1, FacileID: &keep}
	s.orm.Create(&local)
	s.orm.Create(&schemas.Task{ProjectID: local.ID, Name: "T"})

	s.handleProjectCreated(projectEvent(t, envelope.ActionCreated, "k1", smallID, "alpha", personal("owner@example.com"), ""), testMeta)
	s.orm.First(&local, local.ID)
	if countOf(s, &schemas.Project{}) != 1 || *local.FacileID != smallID || local.SyncedAt == nil {
		t.Fatalf("smaller incoming id not adopted: %+v", local)
	}
	sent := queuedEvents(t, s)
	if len(sent) != 1 || sent[0].channel != "task.created" || sent[0].payload["project_facile_id"] != smallID {
		t.Fatalf("adopted project must resend its tasks: %+v", sent)
	}

	other := smallID
	twin := schemas.Project{Name: "Beta", OwnerID: 1, FacileID: &other}
	s.orm.Create(&twin)
	s.handleProjectCreated(projectEvent(t, envelope.ActionCreated, "k2", largeID[:len(largeID)-1]+"e", "Beta", personal("owner@example.com"), ""), testMeta)
	s.orm.First(&twin, twin.ID)
	if countOf(s, &schemas.Project{}) != 2 || *twin.FacileID != smallID || !s.alreadyProcessed("k2") {
		t.Fatalf("larger incoming id must be ignored: %+v", twin)
	}
}

func seedSyncedTask(t *testing.T, s *Service) (*schemas.Project, *schemas.Task) {
	t.Helper()
	s.handleProjectCreated(opusProject(t, envelope.ActionCreated, "p1", "Alpha"), testMeta)
	var project schemas.Project
	s.orm.First(&project)
	linked := time.Now()
	local := schemas.Task{ProjectID: project.ID, Name: "Docs"}
	local.SyncedAt = &linked
	s.orm.Create(&local)
	s.handleTaskCreated(opusTask(t, envelope.ActionCreated, "t1", map[string]any{
		"name": "Docs", "status": "archived", "description": "write it", "due_date": "2026-12-01", "priority": "urgent",
	}), testMeta)
	var task schemas.Task
	s.orm.Where("facile_id = ?", testTaskID).First(&task)
	return &project, &task
}

func TestTaskSuffixAndOpusFieldsRoundTrip(t *testing.T) {
	s := newSyncService(t)
	project, task := seedSyncedTask(t, s)
	if task.Name != "Docs (2)" || task.Status != schemas.StatusDone || task.Description != "write it" ||
		task.DueDate == nil || *task.DueDate != "2026-12-01" || task.Priority != "urgent" {
		t.Fatalf("inbound fields not stored: %+v", task)
	}
	enableEmit(s)
	s.EmitTaskEvent(envelope.ActionUpdated, task, project)
	sent := queuedEvents(t, s)
	got := sent[len(sent)-1]
	if got.payload["name"] != "Docs" || got.payload["description"] != "write it" ||
		got.payload["due_date"] != "2026-12-01" || got.payload["priority"] != "urgent" {
		t.Fatalf("echo lost Opus fields or leaked the suffix: %+v", got.payload)
	}
	if got.key[:len("task.updated:"+testTaskID+":")] != "task.updated:"+testTaskID+":" {
		t.Fatalf("idempotency key shape: %q", got.key)
	}
}

func TestInboundTaskUpdateMovesTaskAndEntries(t *testing.T) {
	s := newSyncService(t)
	_, task := seedSyncedTask(t, s)
	other := smallID
	target := schemas.Project{Name: "Beta", OwnerID: 1, FacileID: &other}
	s.orm.Create(&target)
	s.orm.Create(&schemas.TimeEntry{ProjectID: task.ProjectID, TaskID: task.ID, UserID: 1})

	s.handleTaskUpdated(opusTask(t, envelope.ActionUpdated, "t2", map[string]any{
		"name": "Docs", "project_facile_id": smallID, "updated_at": time.Now().Add(time.Minute).UTC().Format(time.RFC3339Nano),
	}), testMeta)
	var moved schemas.Task
	var entry schemas.TimeEntry
	s.orm.First(&moved, task.ID)
	s.orm.First(&entry)
	if moved.ProjectID != target.ID || moved.Name != "Docs" || moved.SyncedName != "" || entry.ProjectID != target.ID {
		t.Fatalf("task not moved: task %+v entry %+v", moved, entry)
	}
}

func TestTimeEntryEventsOnlyForSyncedProjects(t *testing.T) {
	s := newSyncService(t)
	_, task := seedSyncedTask(t, s)
	enableEmit(s)
	stopped := time.Now().UTC()
	entry := schemas.TimeEntry{ProjectID: task.ProjectID, TaskID: task.ID, UserID: 1, StoppedAt: &stopped, PausedDurationMs: 42}
	s.orm.Create(&entry)
	s.EmitTimeEntryEvent(envelope.ActionDeleted, &entry)

	space := schemas.Space{Name: "Local only"}
	s.orm.Create(&space)
	hidden := schemas.Project{Name: "Hidden", OwnerID: 1, SpaceID: &space.ID}
	s.orm.Create(&hidden)
	hiddenTask := schemas.Task{ProjectID: hidden.ID, Name: "H"}
	s.orm.Create(&hiddenTask)
	s.EmitTimeEntryEvent(envelope.ActionCreated, &schemas.TimeEntry{ProjectID: hidden.ID, TaskID: hiddenTask.ID, UserID: 1})
	s.EmitProjectEvent(envelope.ActionUpdated, &hidden)

	sent := queuedEvents(t, s)
	if len(sent) != 1 || sent[0].channel != "time_entry.deleted" {
		t.Fatalf("unexpected outbox: %+v", sent)
	}
	p := sent[0].payload
	if p["facile_id"] == "" || p["task_facile_id"] != testTaskID || p["project_facile_id"] != testProjectID ||
		p["user_email"] != "owner@example.com" || p["paused_duration_ms"] != float64(42) || p["stopped_at"] == nil {
		t.Fatalf("time entry payload: %+v", p)
	}
}

func TestReconcileQueuesProjectsThenTasksThenEntries(t *testing.T) {
	s := newSyncService(t)
	enableEmit(s)
	project := seedProject(t, s.orm, "Local")
	task := schemas.Task{ProjectID: project.ID, Name: "T"}
	s.orm.Create(&task)
	s.orm.Create(&schemas.TimeEntry{ProjectID: project.ID, TaskID: task.ID, UserID: 1})

	result, err := s.Reconcile(t.Context())
	if err != nil || result.ProjectsSynced != 1 || result.TasksSynced != 1 || result.TimeEntriesSynced != 1 {
		t.Fatalf("reconcile: %+v %v", result, err)
	}
	sent := queuedEvents(t, s)
	if len(sent) != 3 || sent[0].channel != "project.created" || sent[1].channel != "task.created" ||
		sent[2].channel != "time_entry.created" || sent[0].payload["description"] != "" {
		t.Fatalf("unexpected order: %+v", sent)
	}
	if space := sent[0].payload["space"].(map[string]any); space["personal"] != true || space["owner_email"] != "owner@example.com" {
		t.Fatalf("space block: %+v", space)
	}
}
