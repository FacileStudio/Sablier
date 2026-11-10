package projects

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Catalog is GET /projects for userID and spaceID, plus the tasks of every project it
// returns, by name.
func (service *Service) Catalog(ctx context.Context, userID int64, spaceID string) ([]schemas.Project, []schemas.Task, error) {
	projects, err := service.listProjects(ctx, userID, spaceID)
	if err != nil || len(projects) == 0 {
		return projects, nil, err
	}
	ids := make([]int64, len(projects))
	for i, project := range projects {
		ids[i] = project.ID
	}
	var tasks []schemas.Task
	err = service.orm.WithContext(ctx).Where("project_id IN ?", ids).Order("name asc").Order("id asc").Find(&tasks).Error
	if err != nil {
		return nil, nil, errors.Internal("failed to list tasks", err)
	}
	return projects, tasks, nil
}

// EnsureTask is POST /projects/{id}/tasks for userID: the task of that name in a project
// userID can see, created when there is none.
func (service *Service) EnsureTask(ctx context.Context, userID, projectID int64, name string) (*TaskResponse, error) {
	_, err := schemas.VisibleProject(service.orm.WithContext(ctx), userID, projectID)
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("project not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to load the project", err)
	}
	return service.controller.createTask(ctx, projectID, &CreateTaskRequest{Name: name})
}
