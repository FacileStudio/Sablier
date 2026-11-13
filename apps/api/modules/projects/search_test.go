package projects

import (
	"context"
	stderrors "errors"
	"slices"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

func TestSearchTasksRanksRecentThenOpusThenNames(t *testing.T) {
	orm := testdb.Migrated(t)
	service := NewService(orm)
	mine, foreign := schemas.Space{Name: "Facile"}, schemas.Space{Name: "Elsewhere"}
	seedAll(t, orm, &schemas.User{ID: 1, Email: "a@example.test"}, &mine, &foreign)
	seedAll(t, orm, &schemas.SpaceMember{SpaceID: mine.ID, UserID: 1})
	team := schemas.Project{Name: "Docs", OwnerID: 1, SpaceID: &mine.ID}
	personal := schemas.Project{Name: "Home", OwnerID: 1}
	hidden := schemas.Project{Name: "Docs secret", OwnerID: 2, SpaceID: &foreign.ID}
	seedAll(t, orm, &team, &personal, &hidden)
	synced := time.Date(2026, 11, 1, 0, 0, 0, 0, time.UTC)
	tasks := []*schemas.Task{
		{ProjectID: personal.ID, Name: "Write docs"},
		{ProjectID: team.ID, Name: "Opus open", TaskSync: schemas.TaskSync{SyncedAt: &synced}},
		{ProjectID: team.ID, Name: "Opus done", Status: schemas.StatusDone, TaskSync: schemas.TaskSync{SyncedAt: &synced}},
		{ProjectID: team.ID, Name: "Recent"},
		{ProjectID: team.ID, Name: "Older"},
		{ProjectID: hidden.ID, Name: "Leaked docs"},
	}
	for _, task := range tasks {
		seedAll(t, orm, task)
	}
	stop := synced.Add(time.Hour)
	seedAll(t, orm,
		&schemas.TimeEntry{UserID: 1, ProjectID: team.ID, TaskID: tasks[4].ID, StartedAt: synced, StoppedAt: &stop},
		&schemas.TimeEntry{UserID: 1, ProjectID: team.ID, TaskID: tasks[3].ID, StartedAt: stop, StoppedAt: &stop})

	rows, err := service.searchTasks(context.Background(), 1, "", 10)
	if err != nil {
		t.Fatal(stderrors.Unwrap(err))
	}
	if got := names(rows); !slices.Equal(got, []string{"Recent", "Older", "Opus open", "Opus done", "Write docs"}) {
		t.Fatalf("ranking: %v", got)
	}
	if rows[0].LastUsedAt == nil || !rows[0].LastUsedAt.Equal(stop) || !rows[2].Opus || rows[4].SpaceID != nil {
		t.Fatalf("row fields: %+v", rows)
	}

	rows, err = service.searchTasks(context.Background(), 1, "DOCS", 10)
	if err != nil {
		t.Fatal(err)
	}
	if got := names(rows); !slices.Equal(got, []string{"Recent", "Older", "Opus open", "Write docs", "Opus done"}) {
		t.Fatalf("task name matches rank above project matches, and a foreign space stays hidden: %v", got)
	}
	if rows, err = service.searchTasks(context.Background(), 1, "100%_", 10); err != nil || len(rows) != 0 {
		t.Fatalf("LIKE wildcards in q must be literal: %v %v", names(rows), err)
	}
}

func seedAll(t *testing.T, orm *gorm.DB, values ...any) {
	t.Helper()
	for _, value := range values {
		if err := orm.Create(value).Error; err != nil {
			t.Fatalf("seed %T: %v", value, err)
		}
	}
}

func names(rows []searchRow) []string {
	out := make([]string, len(rows))
	for i, row := range rows {
		out[i] = row.TaskName
	}
	return out
}
