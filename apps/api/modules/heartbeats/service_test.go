package heartbeats

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

var testNow = time.Date(2026, 3, 10, 12, 0, 0, 0, time.UTC)

func newTestService(t *testing.T) *Service {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	orm, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}
	err = orm.AutoMigrate(&schemas.User{}, &schemas.Project{}, &schemas.Task{}, &schemas.TimeEntry{},
		&schemas.SpaceMember{}, &schemas.Heartbeat{}, &schemas.DraftDismissal{})
	if err != nil {
		t.Fatalf("migrate test database: %v", err)
	}
	service := NewService(orm, timeentries.NewService(orm))
	service.now = func() time.Time { return testNow }
	return service
}

func seed(t *testing.T, orm *gorm.DB, value any) {
	t.Helper()
	if err := orm.Create(value).Error; err != nil {
		t.Fatalf("seed %T: %v", value, err)
	}
}

func seedUser(t *testing.T, orm *gorm.DB, id int64, autoConfirm bool) {
	t.Helper()
	seed(t, orm, &schemas.User{ID: id, Email: fmt.Sprintf("u%d@example.test", id), PasswordHash: "hash"})
	if err := orm.Model(&schemas.User{}).Where("id = ?", id).Update("auto_confirm_drafts", autoConfirm).Error; err != nil {
		t.Fatalf("set auto confirm: %v", err)
	}
}

func TestConfirmRefusesOverlap(t *testing.T) {
	service := newTestService(t)
	seedUser(t, service.orm, 1, true)
	project := schemas.Project{Name: "Sablier", OwnerID: 1}
	seed(t, service.orm, &project)
	stopped := testNow.Add(-time.Hour)
	seed(t, service.orm, &schemas.TimeEntry{ProjectID: project.ID, TaskID: 1, UserID: 1,
		StartedAt: testNow.Add(-2 * time.Hour), StoppedAt: &stopped})

	_, err := service.confirm(context.Background(), 1, &ConfirmRequest{
		StartedAt: testNow.Add(-90 * time.Minute), StoppedAt: testNow.Add(-30 * time.Minute), ProjectID: project.ID,
	})
	var apiErr *errors.Error
	if !stderrors.As(err, &apiErr) || apiErr.Code != "already_exists" {
		t.Fatalf("expected an overlap conflict, got %v", err)
	}

	entry, err := service.confirm(context.Background(), 1, &ConfirmRequest{
		StartedAt: testNow.Add(-30 * time.Minute), StoppedAt: testNow.Add(-10 * time.Minute), ProjectID: project.ID,
	})
	if err != nil || entry.TaskName != draftTaskName {
		t.Fatalf("expected the free range to be booked on %q, got %+v, %v", draftTaskName, entry, err)
	}
}

func TestRunMaintenanceSelectsStaleProjectDrafts(t *testing.T) {
	service := newTestService(t)
	seedUser(t, service.orm, 1, true)
	seedUser(t, service.orm, 2, false)
	project := schemas.Project{Name: "Sablier", OwnerID: 1}
	seed(t, service.orm, &project)
	beat := func(userID int64, age time.Duration, hint string) {
		seed(t, service.orm, &schemas.Heartbeat{UserID: userID, At: testNow.Add(-age), Source: "editor", ProjectHint: hint})
	}
	beat(1, 72*time.Hour, "sablier")
	beat(1, 60*time.Hour, "unknown project")
	beat(1, time.Hour, "sablier")
	beat(2, 72*time.Hour, "sablier")
	beat(1, 100*24*time.Hour, "sablier")

	confirmed, err := service.RunMaintenance(context.Background(), slog.New(slog.NewTextHandler(io.Discard, nil)))
	if err != nil || confirmed != 1 {
		t.Fatalf("expected one auto-confirmed draft, got %d, %v", confirmed, err)
	}
	var entries []schemas.TimeEntry
	service.orm.Find(&entries)
	if len(entries) != 1 || entries[0].UserID != 1 || !entries[0].StartedAt.Equal(testNow.Add(-72*time.Hour)) {
		t.Fatalf("expected user 1's 72h-old draft booked, got %+v", entries)
	}
	var remaining int64
	service.orm.Model(&schemas.Heartbeat{}).Count(&remaining)
	if remaining != 4 {
		t.Fatalf("expected the 100-day-old heartbeat pruned, %d left", remaining)
	}
}
