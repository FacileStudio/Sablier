package projects

import (
	"context"
	"strings"

	"api/internal/errors"
	"api/schemas"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func toResponse(p *schemas.Project) ProjectResponse {
	return ProjectResponse{
		ID:          p.ID,
		Name:        p.Name,
		Description: p.Description,
		OwnerID:     p.OwnerID,
		CreatedAt:   p.CreatedAt,
		UpdatedAt:   p.UpdatedAt,
	}
}

func (c *Controller) create(ctx context.Context, userID string, req *CreateProjectRequest) (*ProjectResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("project name is required")
	}
	record, err := c.service.createProject(ctx, userID, name, strings.TrimSpace(req.Description))
	if err != nil {
		return nil, err
	}
	resp := toResponse(record)
	return &resp, nil
}

func (c *Controller) list(ctx context.Context) (*ListProjectsResponse, error) {
	records, err := c.service.listProjects(ctx)
	if err != nil {
		return nil, err
	}
	items := make([]ProjectResponse, len(records))
	for i, r := range records {
		items[i] = toResponse(&r)
	}
	return &ListProjectsResponse{Projects: items}, nil
}

func (c *Controller) get(ctx context.Context, userID string, projectID int64) (*ProjectResponse, error) {
	record, err := c.service.getProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(record)
	return &resp, nil
}

func (c *Controller) update(ctx context.Context, userID string, projectID int64, req *UpdateProjectRequest) (*ProjectResponse, error) {
	name := strings.TrimSpace(req.Name)
	if name == "" {
		return nil, errors.Invalid("project name is required")
	}
	record, err := c.service.updateProject(ctx, userID, projectID, name, strings.TrimSpace(req.Description))
	if err != nil {
		return nil, err
	}
	resp := toResponse(record)
	return &resp, nil
}

func (c *Controller) delete(ctx context.Context, userID string, projectID int64) error {
	return c.service.deleteProject(ctx, userID, projectID)
}
