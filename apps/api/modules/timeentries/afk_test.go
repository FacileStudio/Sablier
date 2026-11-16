package timeentries

import (
	"context"
	"testing"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

func TestStopIdleTimersStopsOnlyForgottenRunningTimers(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	now := time.Date(2026, 11, 16, 12, 0, 0, 0, time.UTC)
	rows := seedIdleTimers(t, service, now)
	beat := schemas.Heartbeat{UserID: 2, At: now.Add(-11*time.Hour - 50*time.Minute), Source: "desktop"}
	if err := service.orm.Create(&beat).Error; err != nil {
		t.Fatal(err)
	}

	stopped, err := service.StopIdleTimers(ctx, now)
	if err != nil {
		t.Fatal(err)
	}
	if len(stopped) != 2 || stopped[0] != 1 || stopped[1] != 2 {
		t.Fatalf("stopped users %v, want [1 2]", stopped)
	}
	assertStoppedAt(t, service, 1, ptrTime(rows[0].StartedAt.Add(IdleCap)))
	assertStoppedAt(t, service, 2, ptrTime(beat.At))
	assertStoppedAt(t, service, 3, nil)
	assertStoppedAt(t, service, 4, nil)
}

func seedIdleTimers(t *testing.T, service *Service, now time.Time) []schemas.TimeEntry {
	t.Helper()
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	paused := now.Add(-20 * time.Hour)
	rows := []schemas.TimeEntry{
		{UserID: 1, StartedAt: now.Add(-30 * time.Hour)},
		{UserID: 2, StartedAt: now.Add(-12 * time.Hour)},
		{UserID: 3, StartedAt: now.Add(-3 * time.Hour)},
		{UserID: 4, StartedAt: now.Add(-30 * time.Hour), PausedAt: &paused},
	}
	for i := range rows {
		rows[i].ProjectID, rows[i].TaskID = project.ID, task.ID
		if err := service.orm.Create(&rows[i]).Error; err != nil {
			t.Fatal(err)
		}
	}
	return rows
}

func assertStoppedAt(t *testing.T, service *Service, userID int64, want *time.Time) {
	t.Helper()
	var entry schemas.TimeEntry
	if err := service.orm.Where("user_id = ?", userID).First(&entry).Error; err != nil {
		t.Fatal(err)
	}
	switch {
	case want == nil && entry.StoppedAt != nil:
		t.Errorf("user %d: stopped at %v, want running", userID, entry.StoppedAt)
	case want != nil && (entry.StoppedAt == nil || !entry.StoppedAt.Equal(*want)):
		t.Errorf("user %d: stopped at %v, want %v", userID, entry.StoppedAt, *want)
	}
}

func TestStopAtMustFollowTheLastCheck(t *testing.T) {
	service := newTestService(t)
	ctx := context.Background()
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	started, _, err := service.startTimer(ctx, "1", project.ID, task.ID)
	if err != nil {
		t.Fatal(err)
	}
	for _, bad := range []time.Time{started.StartedAt.Add(-time.Minute), time.Now().Add(time.Hour)} {
		if _, _, err := service.stopTimer(ctx, "1", &bad); err == nil {
			t.Fatalf("stop at %v: want an error", bad)
		}
	}
	checked, _, err := service.checkTimer(ctx, "1")
	if err != nil || checked.CheckedAt == nil {
		t.Fatalf("check: %v, %v", checked, err)
	}
	if _, _, err := service.checkTimer(ctx, "2"); err == nil {
		t.Fatal("check without a running timer: want an error")
	}
	end := *checked.CheckedAt
	stopped, _, err := service.stopTimer(ctx, "1", &end)
	if err != nil {
		t.Fatal(err)
	}
	if !stopped.StoppedAt.Equal(end) {
		t.Fatalf("stopped at %v, want %v", stopped.StoppedAt, end)
	}
}
