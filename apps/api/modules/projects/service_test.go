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

	projects, err := service.listProjects(context.Background())
	if err != nil {
		t.Fatalf("list projects: %v", err)
	}

	if len(projects) != 2 {
		t.Fatalf("expected 2 projects, got %d", len(projects))
	}
}

func TestGetProjectAllowsForeignOwner(t *testing.T) {
	service := newTestService(t)
	foreignProject := seedProject(t, service.orm, 2, "Not mine")

	project, err := service.getProject(context.Background(), foreignProject.ID)
	if err != nil {
		t.Fatalf("get project: %v", err)
	}

	if project.ID != foreignProject.ID {
		t.Fatalf("expected project %d, got %d", foreignProject.ID, project.ID)
	}
}

func TestListTasksAllowsForeignProject(t *testing.T) {
	service := newTestService(t)
	foreignProject := seedProject(t, service.orm, 2, "Not mine")
	task := seedTask(t, service.orm, foreignProject.ID, "Shared task")

	tasks, err := service.listTasks(context.Background(), foreignProject.ID)
	if err != nil {
		t.Fatalf("list tasks: %v", err)
	}

	if len(tasks) != 1 {
		t.Fatalf("expected 1 task, got %d", len(tasks))
	}
	if tasks[0].ID != task.ID {
		t.Fatalf("expected task %d, got %d", task.ID, tasks[0].ID)
	}
}

func TestCreateTaskAllowsForeignProject(t *testing.T) {
	service := newTestService(t)
	foreignProject := seedProject(t, service.orm, 2, "Not mine")

	task, err := service.createTask(context.Background(), foreignProject.ID, "Shared task")
	if err != nil {
		t.Fatalf("create task: %v", err)
	}

	if task.ProjectID != foreignProject.ID {
		t.Fatalf("expected project %d, got %d", foreignProject.ID, task.ProjectID)
	}
}
