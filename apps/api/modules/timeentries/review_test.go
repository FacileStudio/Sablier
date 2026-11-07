package timeentries

import (
	"context"
	"net/http"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

func seedAll(t *testing.T, orm *gorm.DB, values ...any) {
	t.Helper()
	for _, value := range values {
		if err := orm.Create(value).Error; err != nil {
			t.Fatalf("seed %T: %v", value, err)
		}
	}
}

func TestSpaceScopesAndReview(t *testing.T) {
	orm := testdb.Open(t)
	service := NewService(orm)
	ctx := context.Background()
	owner, ownerID := int64(1), int64(1)
	agent := schemas.User{ID: 2, Email: "bulle@agents.test", Kind: schemas.UserKindAgent, OwnerUserID: &owner}
	mine, foreign := schemas.Space{Name: "Mine"}, schemas.Space{Name: "Foreign"}
	seedAll(t, orm, &schemas.User{ID: ownerID, Email: "a@example.test"}, &agent, &mine, &foreign)
	seedAll(t, orm, &schemas.SpaceMember{SpaceID: mine.ID, UserID: ownerID})
	personal, team, secret := schemas.Project{Name: "P"}, schemas.Project{Name: "T", SpaceID: &mine.ID},
		schemas.Project{Name: "S", SpaceID: &foreign.ID}
	seedAll(t, orm, &personal, &team, &secret)
	teamTask := schemas.Task{ProjectID: team.ID, Name: "Work"}
	seedAll(t, orm, &teamTask)
	at := time.Date(2026, 11, 7, 9, 0, 0, 0, time.UTC)
	ai, unsure := attribution.SourceAI, 0.4
	flagged := schemas.TimeEntry{UserID: agent.ID, ProjectID: personal.ID, StartedAt: at, StoppedAt: &at,
		Attribution: schemas.Attribution{AttributionSource: &ai, AttributionConfidence: &unsure, NeedsReview: true}}
	seedAll(t, orm, &flagged,
		&schemas.TimeEntry{UserID: ownerID, ProjectID: team.ID, SpaceID: &mine.ID, StartedAt: at},
		&schemas.TimeEntry{UserID: 3, ProjectID: secret.ID, SpaceID: &foreign.ID, StartedAt: at})

	rows, err := service.listEntries(ctx, ownerID, listFilter{Space: schemas.AllSpaces})
	if err != nil || len(rows) != 2 {
		t.Fatalf("all spaces: %d rows, %v", len(rows), err)
	}
	if rows, err = service.listEntries(ctx, ownerID, listFilter{Space: schemas.AllSpaces, Running: true}); err != nil || len(rows) != 1 {
		t.Fatalf("running in all spaces: %d rows, %v", len(rows), err)
	}
	if _, err := service.listEntries(ctx, ownerID, listFilter{Space: foreign.ID}); errors.Status(err) != http.StatusNotFound {
		t.Fatalf("a foreign space must be a 404, got %v", err)
	}

	inbox, err := service.ReviewEntries(ctx, ownerID)
	if err != nil || len(inbox) != 1 || inbox[0].ID != flagged.ID || !inbox[0].NeedsReview {
		t.Fatalf("review inbox: %+v, %v", inbox, err)
	}
	reviewed, err := service.ReviewEntry(ctx, ownerID, flagged.ID, teamTask.ID)
	if err != nil {
		t.Fatal(err)
	}
	if reviewed.NeedsReview || reviewed.TaskID != teamTask.ID || reviewed.SpaceID == nil || *reviewed.SpaceID != mine.ID ||
		*reviewed.Source != attribution.SourceManual || reviewed.ProjectName != "T" {
		t.Fatalf("reviewed entry: %+v", reviewed)
	}
	if _, err := service.ReviewEntry(ctx, 3, flagged.ID, 0); errors.Status(err) != http.StatusNotFound {
		t.Fatalf("someone else's entry must be a 404, got %v", err)
	}
}
