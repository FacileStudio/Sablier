package antenne

import (
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"

	enveloppe "github.com/FacileStudio/enveloppe/go"
	antenneclient "github.com/FacileStudio/antenne-client/go"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	orm, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	err = orm.AutoMigrate(&schemas.User{}, &schemas.Project{}, &schemas.Task{}, &schemas.TimeEntry{},
		&schemas.AppSetting{}, &schemas.PoolOutbox{}, &schemas.PoolProcessedEvent{})
	if err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	return NewService(orm, slog.New(slog.NewTextHandler(io.Discard, nil)))
}

func seedUser(t *testing.T, orm *gorm.DB, email string) *schemas.User {
	t.Helper()
	user := &schemas.User{Email: email, Name: "Sara"}
	if err := orm.Create(user).Error; err != nil {
		t.Fatal(err)
	}
	return user
}

func seedProject(t *testing.T, orm *gorm.DB, name string) *schemas.Project {
	t.Helper()
	project := &schemas.Project{Name: name, OwnerID: 1}
	if err := orm.Create(project).Error; err != nil {
		t.Fatal(err)
	}
	return project
}

func sessionEvent(t *testing.T, facileID, project, email, key string) json.RawMessage {
	t.Helper()
	evt := enveloppe.Event[enveloppe.AgentSession]{
		Version:  enveloppe.EventVersion,
		App:      enveloppe.AppMycelium,
		Object:   enveloppe.ObjectAgentSession,
		Action:   enveloppe.ActionCreated,
		FacileID: facileID,
		Payload: enveloppe.AgentSession{
			FacileID:  facileID,
			Project:   project,
			Machine:   "lucy",
			Agent:     "claude",
			Branch:    "main",
			UserEmail: email,
			StartedAt: "2026-08-01T10:00:00Z",
			StoppedAt: "2026-08-01T10:30:00Z",
			TokensIn:  1000,
			TokensOut: 500,
		},
		Timestamp:      "2026-08-01T11:00:00Z",
		IdempotencyKey: key,
	}
	data, err := json.Marshal(evt)
	if err != nil {
		t.Fatal(err)
	}
	return data
}

var testMeta = antenneclient.EventMeta{Channel: "agent_session.created", Sender: "Mycelium"}

func TestAgentSessionCreatesEntryInMatchingProject(t *testing.T) {
	s := newTestService(t)
	user := seedUser(t, s.orm, "sara@example.com")
	project := seedProject(t, s.orm, "Mycelium")

	s.handleAgentSession(sessionEvent(t, "fac_abc", "Mycelium", "sara@example.com", "key_1"), testMeta)

	var entry schemas.TimeEntry
	if err := s.orm.Where("facile_id = ?", "fac_abc").First(&entry).Error; err != nil {
		t.Fatalf("entry not created: %v", err)
	}
	if entry.ProjectID != project.ID || entry.UserID != user.ID {
		t.Fatalf("bad attribution: %+v", entry)
	}
	if entry.StoppedAt == nil {
		t.Fatal("entry must be stopped, not running")
	}
	var task schemas.Task
	if err := s.orm.First(&task, entry.TaskID).Error; err != nil {
		t.Fatal(err)
	}
	if task.Name != agentSessionsTaskName || task.ProjectID != project.ID {
		t.Fatalf("expected auto-created 'Agent sessions' task, got %+v", task)
	}
	if !strings.Contains(entry.LegacyDescription, "claude@lucy") {
		t.Fatalf("description missing audit trail: %q", entry.LegacyDescription)
	}
}

func TestAgentSessionIdempotentReplay(t *testing.T) {
	s := newTestService(t)
	seedUser(t, s.orm, "sara@example.com")
	seedProject(t, s.orm, "Mycelium")

	payload := sessionEvent(t, "fac_abc", "Mycelium", "sara@example.com", "key_1")
	s.handleAgentSession(payload, testMeta)
	s.handleAgentSession(payload, testMeta)

	var count int64
	s.orm.Model(&schemas.TimeEntry{}).Count(&count)
	if count != 1 {
		t.Fatalf("replay must not duplicate: %d entries", count)
	}
}

func TestAgentSessionUpsertsByFacileID(t *testing.T) {
	s := newTestService(t)
	seedUser(t, s.orm, "sara@example.com")
	seedProject(t, s.orm, "Mycelium")

	s.handleAgentSession(sessionEvent(t, "fac_abc", "Mycelium", "sara@example.com", "key_1"), testMeta)
	s.handleAgentSession(sessionEvent(t, "fac_abc", "Mycelium", "sara@example.com", "key_2"), testMeta)

	var count int64
	s.orm.Model(&schemas.TimeEntry{}).Count(&count)
	if count != 1 {
		t.Fatalf("same facile_id with new key must update, not duplicate: %d entries", count)
	}
}

func TestAgentSessionParksUnmatchedProject(t *testing.T) {
	s := newTestService(t)
	seedUser(t, s.orm, "sara@example.com")

	s.handleAgentSession(sessionEvent(t, "fac_xyz", "SomeRepo", "sara@example.com", "key_1"), testMeta)

	var project schemas.Project
	if err := s.orm.Where("name = ?", agentCatchAllProject).First(&project).Error; err != nil {
		t.Fatalf("catch-all project not created: %v", err)
	}
	var task schemas.Task
	if err := s.orm.Where("project_id = ? AND name = ?", project.ID, "SomeRepo").First(&task).Error; err != nil {
		t.Fatalf("per-source task not created: %v", err)
	}
	var entry schemas.TimeEntry
	if err := s.orm.Where("facile_id = ?", "fac_xyz").First(&entry).Error; err != nil {
		t.Fatalf("parked entry not created: %v", err)
	}
	if entry.TaskID != task.ID {
		t.Fatal("entry not attached to parked task")
	}
}

func TestAgentSessionSkipsUnknownUser(t *testing.T) {
	s := newTestService(t)
	seedProject(t, s.orm, "Mycelium")

	s.handleAgentSession(sessionEvent(t, "fac_abc", "Mycelium", "nobody@example.com", "key_1"), testMeta)

	var count int64
	s.orm.Model(&schemas.TimeEntry{}).Count(&count)
	if count != 0 {
		t.Fatal("unknown user must not create entries")
	}
	if s.alreadyProcessed("key_1") {
		t.Fatal("skipped event must stay replayable")
	}
}

func TestAgentSessionRejectsInvalidTimes(t *testing.T) {
	s := newTestService(t)
	seedUser(t, s.orm, "sara@example.com")
	seedProject(t, s.orm, "Mycelium")

	evt := enveloppe.Event[enveloppe.AgentSession]{
		Payload: enveloppe.AgentSession{
			FacileID: "fac_bad", Project: "Mycelium", UserEmail: "sara@example.com",
			StartedAt: "2026-08-01T10:30:00Z", StoppedAt: "2026-08-01T10:00:00Z",
		},
		IdempotencyKey: "key_bad",
	}
	data, _ := json.Marshal(evt)
	s.handleAgentSession(data, testMeta)

	var count int64
	s.orm.Model(&schemas.TimeEntry{}).Count(&count)
	if count != 0 {
		t.Fatal("stopped-before-started must be rejected")
	}
}
