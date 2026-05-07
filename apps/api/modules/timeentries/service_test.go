package timeentries

import (
	"context"
	"testing"
	"time"

	"api/schemas"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestService(t *testing.T) *Service {
	t.Helper()

	orm, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	err = orm.AutoMigrate(&schemas.User{}, &schemas.Project{}, &schemas.Task{}, &schemas.TimeEntry{})
	if err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return NewService(orm)
}

func seedProject(t *testing.T, orm *gorm.DB, ownerID int64) schemas.Project {
	t.Helper()

	project := schemas.Project{
		Name:        "Shared project",
		Description: "Shared project description",
		OwnerID:     ownerID,
	}
	if err := orm.Create(&project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	return project
}

func seedTask(t *testing.T, orm *gorm.DB, projectID int64) schemas.Task {
	t.Helper()

	task := schemas.Task{
		ProjectID: projectID,
		Name:      "Shared task",
	}
	if err := orm.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	return task
}

func seedEntry(t *testing.T, orm *gorm.DB, projectID int64, taskID int64, userID int64) schemas.TimeEntry {
	t.Helper()

	stoppedAt := time.Now().UTC()
	entry := schemas.TimeEntry{
		ProjectID: projectID,
		TaskID:    taskID,
		UserID:    userID,
		StartedAt: stoppedAt.Add(-time.Hour),
		StoppedAt: &stoppedAt,
	}
	if err := orm.Create(&entry).Error; err != nil {
		t.Fatalf("create entry: %v", err)
	}
	return entry
}

func seedUser(t *testing.T, orm *gorm.DB, id int64, email string, name string, color string) schemas.User {
	t.Helper()

	user := schemas.User{
		ID:           id,
		Email:        email,
		Name:         name,
		Color:        color,
		PasswordHash: "hash",
	}
	if err := orm.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}

func TestUpdateEntryRejectsForeignOwner(t *testing.T) {
	service := newTestService(t)
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	entry := seedEntry(t, service.orm, project.ID, task.ID, 1)

	_, _, err := service.updateEntry(context.Background(), "2", entry.ID, &UpdateEntryRequest{
		ProjectID: project.ID,
		TaskID:    task.ID,
		StartedAt: time.Now().UTC().Add(-30 * time.Minute),
		StoppedAt: ptrTime(time.Now().UTC()),
	})
	if err == nil || err.Error() != "time entry not found" {
		t.Fatalf("expected time entry not found, got %v", err)
	}
}

func TestDeleteEntryRejectsForeignOwner(t *testing.T) {
	service := newTestService(t)
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	entry := seedEntry(t, service.orm, project.ID, task.ID, 1)

	err := service.deleteEntry(context.Background(), "2", entry.ID)
	if err == nil || err.Error() != "time entry not found" {
		t.Fatalf("expected time entry not found, got %v", err)
	}
}

func TestListEntriesIncludesUserColor(t *testing.T) {
	service := newTestService(t)
	user := seedUser(t, service.orm, 1, "jane@example.com", "Jane Doe", "AD9EF0")
	project := seedProject(t, service.orm, user.ID)
	task := seedTask(t, service.orm, project.ID)
	seedEntry(t, service.orm, project.ID, task.ID, user.ID)

	records, err := service.listEntries(context.Background(), project.ID, 0)
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].UserEmail != user.Email {
		t.Fatalf("expected user email %q, got %q", user.Email, records[0].UserEmail)
	}
	if records[0].UserName != user.Name {
		t.Fatalf("expected user name %q, got %q", user.Name, records[0].UserName)
	}
	if records[0].UserColor != user.Color {
		t.Fatalf("expected user color %q, got %q", user.Color, records[0].UserColor)
	}
}

func TestUpdateEntryAllowsEditingRunningSession(t *testing.T) {
	service := newTestService(t)
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	running := schemas.TimeEntry{
		ProjectID: project.ID,
		TaskID:    task.ID,
		UserID:    1,
		StartedAt: time.Now().UTC().Add(-time.Hour),
	}
	if err := service.orm.Create(&running).Error; err != nil {
		t.Fatalf("create running entry: %v", err)
	}

	updatedStart := time.Now().UTC().Add(-90 * time.Minute)
	record, taskName, err := service.updateEntry(context.Background(), "1", running.ID, &UpdateEntryRequest{
		ProjectID: project.ID,
		TaskID:    task.ID,
		StartedAt: updatedStart,
		StoppedAt: nil,
	})
	if err != nil {
		t.Fatalf("update running entry: %v", err)
	}
	if taskName != task.Name {
		t.Fatalf("expected task name %q, got %q", task.Name, taskName)
	}
	if record.StoppedAt != nil {
		t.Fatalf("expected running entry to remain running, got stopped_at=%v", record.StoppedAt)
	}
	if !record.StartedAt.Equal(updatedStart) {
		t.Fatalf("expected started_at %v, got %v", updatedStart, record.StartedAt)
	}
}

func TestUpdateEntryRejectsTurningStoppedEntryBackIntoRunning(t *testing.T) {
	service := newTestService(t)
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	entry := seedEntry(t, service.orm, project.ID, task.ID, 1)

	_, _, err := service.updateEntry(context.Background(), "1", entry.ID, &UpdateEntryRequest{
		ProjectID: project.ID,
		TaskID:    task.ID,
		StartedAt: entry.StartedAt.Add(-15 * time.Minute),
		StoppedAt: nil,
	})
	if err == nil || err.Error() != "only the currently running session can remain running after edit" {
		t.Fatalf("expected running-session-only error, got %v", err)
	}
}

func TestPauseAndResumeTimerAdjustPausedDuration(t *testing.T) {
	service := newTestService(t)
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	record, _, err := service.startTimer(context.Background(), "1", project.ID, task.ID)
	if err != nil {
		t.Fatalf("start timer: %v", err)
	}

	paused, taskName, err := service.pauseTimer(context.Background(), "1")
	if err != nil {
		t.Fatalf("pause timer: %v", err)
	}
	if taskName != task.Name {
		t.Fatalf("expected task name %q, got %q", task.Name, taskName)
	}
	if paused.PausedAt == nil {
		t.Fatal("expected paused_at to be set")
	}

	time.Sleep(15 * time.Millisecond)

	resumed, _, err := service.resumeTimer(context.Background(), "1")
	if err != nil {
		t.Fatalf("resume timer: %v", err)
	}
	if resumed.PausedAt != nil {
		t.Fatalf("expected paused_at to be cleared, got %v", resumed.PausedAt)
	}
	if resumed.PausedDurationMs <= 0 {
		t.Fatalf("expected paused duration to increase, got %d", resumed.PausedDurationMs)
	}

	stored, _, err := service.getRunningTimer(context.Background(), "1")
	if err != nil {
		t.Fatalf("get running timer: %v", err)
	}
	if stored == nil || stored.ID != record.ID {
		t.Fatalf("expected running timer %d, got %#v", record.ID, stored)
	}
	if stored.PausedAt != nil {
		t.Fatalf("expected stored timer to be resumed, got paused_at=%v", stored.PausedAt)
	}
	if stored.PausedDurationMs <= 0 {
		t.Fatalf("expected stored paused duration to persist, got %d", stored.PausedDurationMs)
	}
}

func TestStopTimerWhilePausedFinalizesPausedDuration(t *testing.T) {
	service := newTestService(t)
	project := seedProject(t, service.orm, 1)
	task := seedTask(t, service.orm, project.ID)
	if _, _, err := service.startTimer(context.Background(), "1", project.ID, task.ID); err != nil {
		t.Fatalf("start timer: %v", err)
	}
	if _, _, err := service.pauseTimer(context.Background(), "1"); err != nil {
		t.Fatalf("pause timer: %v", err)
	}

	time.Sleep(15 * time.Millisecond)

	stopped, _, err := service.stopTimer(context.Background(), "1")
	if err != nil {
		t.Fatalf("stop timer: %v", err)
	}
	if stopped.StoppedAt == nil {
		t.Fatal("expected stopped_at to be set")
	}
	if stopped.PausedAt != nil {
		t.Fatalf("expected paused_at to be cleared on stop, got %v", stopped.PausedAt)
	}
	if stopped.PausedDurationMs <= 0 {
		t.Fatalf("expected paused duration to be retained on stop, got %d", stopped.PausedDurationMs)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
