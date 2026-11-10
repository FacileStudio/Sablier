package antenne

import (
	"context"
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"regexp"
	"testing"
	"time"

	"github.com/FacileStudio/antenne-client/go/envelope"
	"github.com/FacileStudio/antenne-client/go/outbox"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

var updateGolden = flag.Bool("update", false, "rewrite testdata/wire from the current output")

var (
	volatileTimestamp = regexp.MustCompile(`"timestamp":"[^"]*"`)
	volatileKey       = regexp.MustCompile(`"idempotency_key":"([^"]*):[0-9]+"`)
)

func stable(raw []byte) []byte {
	raw = volatileTimestamp.ReplaceAll(raw, []byte(`"timestamp":"T"`))
	return volatileKey.ReplaceAll(raw, []byte(`"idempotency_key":"$1:N"`))
}

func at(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		panic(err)
	}
	return parsed
}

func str(value string) *string { return &value }

type wireFixture struct {
	shared, personal *schemas.Project
	task             *schemas.Task
	entry            *schemas.TimeEntry
}

func seedWire(t *testing.T, s *Service) wireFixture {
	t.Helper()
	enableEmit(s)
	alice := seedUser(t, s.orm, "Alice@Example.com")
	s.orm.Exec(`INSERT INTO porte_identities (user_id, provider, subject) VALUES (?, ?, 'sub-alice')`, alice.ID, s.issuer)
	space := schemas.Space{Name: "Team", RegistreSpaceID: str("reg-1")}
	if err := s.orm.Create(&space).Error; err != nil {
		t.Fatal(err)
	}
	updated := at("2026-11-10T12:34:56.123456Z")
	shared := &schemas.Project{Name: `Café <Équipe> & "co"`, Description: "line\nbreak", Icon: str("Pallete2"),
		OwnerID: alice.ID, SpaceID: &space.ID, FacileID: str("fac_project"), UpdatedAt: updated}
	personal := &schemas.Project{Name: "Perso", OwnerID: alice.ID, FacileID: str("fac_personal"), UpdatedAt: updated}
	for _, project := range []*schemas.Project{shared, personal} {
		if err := s.orm.Create(project).Error; err != nil {
			t.Fatal(err)
		}
		project.UpdatedAt = updated
	}
	task := &schemas.Task{ProjectID: shared.ID, Name: "Design", Status: "in_progress", ActorID: &alice.ID,
		FacileID: str("fac_task"), UpdatedAt: updated,
		TaskSync: schemas.TaskSync{Description: "desc", DueDate: str("2026-11-12")}}
	if err := s.orm.Create(task).Error; err != nil {
		t.Fatal(err)
	}
	task.UpdatedAt = updated
	stopped := at("2026-11-10T11:00:00Z")
	entry := &schemas.TimeEntry{ProjectID: shared.ID, TaskID: task.ID, UserID: alice.ID, FacileID: str("fac_entry"),
		LegacyDescription: "notes", StartedAt: at("2026-11-10T09:30:00Z"), StoppedAt: &stopped,
		PausedDurationMs: 1500, UpdatedAt: updated}
	if err := s.orm.Create(entry).Error; err != nil {
		t.Fatal(err)
	}
	entry.UpdatedAt = updated
	return wireFixture{shared: shared, personal: personal, task: task, entry: entry}
}

func publishEveryChannel(s *Service, f wireFixture) []string {
	var names []string
	for _, action := range []envelope.Action{envelope.ActionCreated, envelope.ActionUpdated, envelope.ActionDeleted} {
		s.EmitProjectEvent(action, f.shared)
		s.EmitTaskEvent(action, f.task, f.shared)
		s.EmitTimeEntryEvent(action, f.entry)
		names = append(names, "project."+string(action), "task."+string(action), "time_entry."+string(action))
	}
	s.EmitProjectEvent(envelope.ActionCreated, f.personal)
	return append(names, "project.created.personal")
}

func TestEveryPublishedEventKeepsItsBytes(t *testing.T) {
	s := newSyncService(t)
	names := publishEveryChannel(s, seedWire(t, s))
	sent := sentEvents(t, s)
	if len(sent) != len(names) {
		t.Fatalf("queued %d events, want %d", len(sent), len(names))
	}
	for i, name := range names {
		path := filepath.Join("testdata", "wire", name+".json")
		got := append(stable(sent[i]), '\n')
		if *updateGolden {
			if err := os.WriteFile(path, got, 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		if string(got) != string(want) {
			t.Errorf("%s changed on the wire:\n got %s\nwant %s", name, got, want)
		}
	}
}

// sentEvents drains the outbox through antenne-client's drainer and returns
// what Sablier hands the client for each row, the bytes that go on the wire.
func sentEvents(t *testing.T, s *Service) [][]byte {
	t.Helper()
	db, err := s.orm.DB()
	if err != nil {
		t.Fatal(err)
	}
	var out [][]byte
	drainer := outbox.NewDrainer(db, func(channel string, payload any) error {
		event, err := canonical(channel, payload.(json.RawMessage))
		out = append(out, event)
		return err
	})
	if _, err := drainer.Drain(context.Background()); err != nil {
		t.Fatal(err)
	}
	return out
}
