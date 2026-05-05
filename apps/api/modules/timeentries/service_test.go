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

func seedUser(t *testing.T, orm *gorm.DB, id int64, email string, color string) schemas.User {
	t.Helper()

	user := schemas.User{
		ID:           id,
		Email:        email,
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
	user := seedUser(t, service.orm, 1, "jane@example.com", "AD9EF0")
	project := seedProject(t, service.orm, user.ID)
	task := seedTask(t, service.orm, project.ID)
	seedEntry(t, service.orm, project.ID, task.ID, user.ID)

	records, err := service.listEntries(context.Background(), project.ID)
	if err != nil {
		t.Fatalf("list entries: %v", err)
	}
	if len(records) != 1 {
		t.Fatalf("expected 1 record, got %d", len(records))
	}
	if records[0].UserEmail != user.Email {
		t.Fatalf("expected user email %q, got %q", user.Email, records[0].UserEmail)
	}
	if records[0].UserColor != user.Color {
		t.Fatalf("expected user color %q, got %q", user.Color, records[0].UserColor)
	}
}

func ptrTime(value time.Time) *time.Time {
	return &value
}
