package projects

import (
	"context"
	"testing"

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

	err = orm.AutoMigrate(&schemas.Project{}, &schemas.Task{})
	if err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return NewService(orm)
}

func seedProject(t *testing.T, orm *gorm.DB, ownerID int64, name string) schemas.Project {
	t.Helper()

	project := schemas.Project{
		Name:        name,
		Description: name + " description",
		OwnerID:     ownerID,
	}
	if err := orm.Create(&project).Error; err != nil {
		t.Fatalf("create project: %v", err)
	}
	return project
}

func seedTask(t *testing.T, orm *gorm.DB, projectID int64, name string) schemas.Task {
	t.Helper()

	task := schemas.Task{
		ProjectID: projectID,
		Name:      name,
	}
	if err := orm.Create(&task).Error; err != nil {
		t.Fatalf("create task: %v", err)
	}
	return task
}

func TestListProjectsReturnsOnlyOwnedProjects(t *testing.T) {
	service := newTestService(t)
	seedProject(t, service.orm, 1, "Mine")
	seedProject(t, service.orm, 2, "Not mine")

	projects, err := service.listProjects(context.Background(), "1")
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(projects) != 1 {
		t.Fatalf("expected 1 project, got %d", len(projects))
	}
	if projects[0].OwnerID != 1 {
		t.Fatalf("expected owner 1, got %d", projects[0].OwnerID)
	}
	if projects[0].Name != "Mine" {
		t.Fatalf("expected project Mine, got %q", projects[0].Name)
	}
}

func TestListTasksRejectsForeignProject(t *testing.T) {
	service := newTestService(t)
	foreignProject := seedProject(t, service.orm, 2, "Not mine")
	seedTask(t, service.orm, foreignProject.ID, "Secret task")

	_, err := service.listTasks(context.Background(), "1", foreignProject.ID)
	if err == nil || err.Error() != "project not found" {
		t.Fatalf("expected project not found, got %v", err)
	}
}

func TestCreateTaskRejectsForeignProject(t *testing.T) {
	service := newTestService(t)
	foreignProject := seedProject(t, service.orm, 2, "Not mine")

	_, err := service.createTask(context.Background(), "1", foreignProject.ID, "Should fail")
	if err == nil || err.Error() != "project not found" {
		t.Fatalf("expected project not found, got %v", err)
	}
}
