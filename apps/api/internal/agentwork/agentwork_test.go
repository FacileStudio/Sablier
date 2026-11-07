package agentwork

import (
	"context"
	stderrors "errors"
	"fmt"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
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
		OwnerID:   ownerID,
		Agent:     schemas.AgentFromName("bulle"),
		FacileID:  facileID,
		Facts:     attribution.Facts{ProjectHint: "Sablier"},
		StartedAt: start,
		StoppedAt: start.Add(time.Duration(minutes) * time.Minute),
	}
}

func TestRecordUpsertsByFacileID(t *testing.T) {
	orm := openDB(t)
	owner := schemas.User{Email: "sara@example.com"}
	create(t, orm, &owner)
	space := schemas.Space{Name: "Facile"}
	create(t, orm, &space)
	create(t, orm, &schemas.SpaceMember{SpaceID: space.ID, UserID: owner.ID, Role: schemas.RoleOwner})

	first, err := Record(context.Background(), orm, attribution.Resolver{}, session(owner.ID, "bulle_run_1", 10))
	if err != nil {
		t.Fatal(err)
	}
	second, err := Record(context.Background(), orm, attribution.Resolver{}, session(owner.ID, "bulle_run_1", 25))
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
	if _, err := Record(context.Background(), orm, attribution.Resolver{}, session(intruder.ID, "bulle_run_1", 5)); !stderrors.Is(err, ErrForeignFacileID) {
		t.Fatalf("another owner must not take over the entry, got %v", err)
	}
}

func TestRecordFlagsTheCatchAllAndKeepsAReviewedEntry(t *testing.T) {
	orm := openDB(t)
	owner := schemas.User{Email: "sara@example.com"}
	create(t, orm, &owner)
	unknown := session(owner.ID, "bulle_run_2", 10)
	unknown.Facts = attribution.Facts{ProjectHint: "Unknown"}

	first, err := Record(context.Background(), orm, attribution.Resolver{}, unknown)
	if err != nil {
		t.Fatal(err)
	}
	if first.Target.Project.Name != CatchAllProject || first.Target.Task.Name != "Unknown" || !first.Entry.NeedsReview ||
		*first.Entry.AttributionSource != attribution.SourceFallback {
		t.Fatalf("catch-all: %s/%s %+v", first.Target.Project.Name, first.Target.Task.Name, first.Entry.Attribution)
	}

	chosen := schemas.Project{Name: "Chosen"}
	create(t, orm, &chosen)
	manual := attribution.SourceManual
	err = orm.Model(&schemas.TimeEntry{}).Where("id = ?", first.Entry.ID).
		Updates(map[string]any{"project_id": chosen.ID, "attribution_source": manual, "needs_review": false}).Error
	if err != nil {
		t.Fatal(err)
	}
	replay, err := Record(context.Background(), orm, attribution.Resolver{}, unknown)
	if err != nil {
		t.Fatal(err)
	}
	if replay.Entry.ProjectID != chosen.ID || replay.Entry.NeedsReview {
		t.Fatalf("a replay must keep the owner's reassignment, got project %d", replay.Entry.ProjectID)
	}
}
