package agentwork

import (
	stderrors "errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func openDB(t *testing.T) *gorm.DB {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	orm, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatal(err)
	}
	err = orm.AutoMigrate(&schemas.User{}, &schemas.Project{}, &schemas.Task{}, &schemas.TimeEntry{},
		&schemas.Space{}, &schemas.SpaceMember{})
	if err != nil {
		t.Fatal(err)
	}
	return orm
}

func create(t *testing.T, orm *gorm.DB, value any) {
	t.Helper()
	if err := orm.Create(value).Error; err != nil {
		t.Fatal(err)
	}
}

var start = time.Date(2026, 11, 2, 10, 0, 0, 0, time.UTC)

func session(ownerID int64, facileID string, minutes int) Session {
	return Session{
		OwnerID:     ownerID,
		Agent:       schemas.AgentFromName("bulle"),
		FacileID:    facileID,
		ProjectHint: "Sablier",
		StartedAt:   start,
		StoppedAt:   start.Add(time.Duration(minutes) * time.Minute),
	}
}

func TestRecordUpsertsByFacileID(t *testing.T) {
	orm := openDB(t)
	owner := schemas.User{Email: "sara@example.com"}
	create(t, orm, &owner)
	space := schemas.Space{Name: "Facile"}
	create(t, orm, &space)
	create(t, orm, &schemas.SpaceMember{SpaceID: space.ID, UserID: owner.ID, Role: schemas.RoleOwner})

	first, err := Record(orm, session(owner.ID, "bulle_run_1", 10))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Record(orm, session(owner.ID, "bulle_run_1", 25))
	if err != nil {
		t.Fatal(err)
	}

	var entries, agents, memberships int64
	orm.Model(&schemas.TimeEntry{}).Count(&entries)
	orm.Model(&schemas.User{}).Where("kind = ?", schemas.UserKindAgent).Count(&agents)
	orm.Model(&schemas.SpaceMember{}).Where("user_id = ?", first.Agent.ID).Count(&memberships)
	if entries != 1 || agents != 1 || memberships != 1 {
		t.Fatalf("replay must update in place: entries=%d agents=%d memberships=%d", entries, agents, memberships)
	}
	if second.Entry.ID != first.Entry.ID || !second.Entry.StoppedAt.Equal(start.Add(25*time.Minute)) {
		t.Fatalf("entry not updated: %+v", second.Entry)
	}
	if first.Agent.Rate != 0 || first.Agent.Avatar() != "https://blobatar.dev/avatar/Bulle?gen=2" {
		t.Fatalf("agent identity: %+v", first.Agent)
	}

	intruder := schemas.User{Email: "eve@example.com"}
	create(t, orm, &intruder)
	if _, err := Record(orm, session(intruder.ID, "bulle_run_1", 5)); !stderrors.Is(err, ErrForeignFacileID) {
		t.Fatalf("another owner must not take over the entry, got %v", err)
	}
}

func TestResolveTargetOrder(t *testing.T) {
	orm := openDB(t)
	owner := schemas.User{Email: "sara@example.com"}
	create(t, orm, &owner)
	explicit := schemas.Project{Name: "Explicit"}
	named := schemas.Project{Name: "Sablier"}
	running := schemas.Project{Name: "Running"}
	hiddenSpace := "00000000-0000-4000-8000-000000000000"
	hidden := schemas.Project{Name: "Hidden", SpaceID: &hiddenSpace}
	for _, project := range []*schemas.Project{&explicit, &named, &running, &hidden} {
		create(t, orm, project)
	}
	stopped := start.Add(time.Hour)
	create(t, orm, &schemas.TimeEntry{UserID: owner.ID, ProjectID: running.ID, StartedAt: start.Add(-time.Hour), StoppedAt: &stopped})

	cases := []struct {
		name      string
		projectID int64
		hint      string
		want      string
		wantTask  string
	}{
		{"project id wins over the name", explicit.ID, "Sablier", "Explicit", SessionsTaskName},
		{"name matches case-insensitively", 0, "sablier", "Sablier", SessionsTaskName},
		{"running entry when the name misses", 0, "Unknown", "Running", SessionsTaskName},
		{"a project in a foreign space is not matched by name", 0, "Hidden", "Running", SessionsTaskName},
	}
	for _, tc := range cases {
		target, err := ResolveTarget(orm, owner.ID, tc.projectID, tc.hint, start)
		if err != nil {
			t.Fatalf("%s: %v", tc.name, err)
		}
		if target.Project.Name != tc.want || target.Task.Name != tc.wantTask {
			t.Fatalf("%s: got %s/%s", tc.name, target.Project.Name, target.Task.Name)
		}
	}

	if _, err := ResolveTarget(orm, owner.ID, hidden.ID, "", start); !stderrors.Is(err, ErrProjectNotVisible) {
		t.Fatalf("a hidden project id must be refused, got %v", err)
	}
	target, err := ResolveTarget(orm, owner.ID, 0, "Unknown", start.Add(3*time.Hour))
	if err != nil {
		t.Fatal(err)
	}
	if target.Project.Name != CatchAllProject || target.Task.Name != "Unknown" {
		t.Fatalf("catch-all: got %s/%s", target.Project.Name, target.Task.Name)
	}
}
