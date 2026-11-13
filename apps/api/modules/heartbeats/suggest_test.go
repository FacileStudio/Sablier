package heartbeats

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

func TestDraftsCarryTheSuggestionAndThreshold(t *testing.T) {
	orm := testdb.Migrated(t)
	service := NewService(orm, timeentries.NewService(orm), attribution.Resolver{})
	service.now = func() time.Time { return testNow }
	registre := "reg-1"
	space := schemas.Space{Name: "Facile", RegistreSpaceID: &registre}
	seed(t, orm, &schemas.User{ID: 1, Email: "a@example.test"})
	seed(t, orm, &space)
	seed(t, orm, &schemas.SpaceMember{SpaceID: space.ID, UserID: 1})
	personal, team := schemas.Project{Name: "Sablier"}, schemas.Project{Name: "Sablier", SpaceID: &space.ID}
	seed(t, orm, &personal)
	seed(t, orm, &team)
	for _, minutes := range []int{50, 40} {
		seed(t, orm, &schemas.Heartbeat{UserID: 1, At: testNow.Add(-time.Duration(minutes) * time.Minute), Source: "editor",
			ProjectHint: "sablier", HeartbeatFacts: schemas.HeartbeatFacts{RegistreSpaceID: registre}})
	}

	drafts, err := service.listDrafts(context.Background(), 1, testNow.Add(-time.Hour), testNow)
	if err != nil || len(drafts) != 1 {
		t.Fatalf("drafts: %+v, %v", drafts, err)
	}
	got := drafts[0].Suggestion
	if got == nil || got.ProjectID != team.ID || got.Source != attribution.SourceName || got.NeedsReview || *got.SpaceID != space.ID {
		t.Fatalf("suggestion: %+v", got)
	}

	strict := 0.95
	settings, err := service.updateSettings(context.Background(), 1, &DraftSettings{AutoConfirmDrafts: true, ReviewThreshold: &strict})
	if err != nil || *settings.ReviewThreshold != strict {
		t.Fatalf("settings: %+v, %v", settings, err)
	}
	if drafts, err = service.listDrafts(context.Background(), 1, testNow.Add(-time.Hour), testNow); err != nil || !drafts[0].Suggestion.NeedsReview {
		t.Fatalf("a 0.9 name match must need review under a 0.95 threshold: %+v, %v", drafts, err)
	}
}
