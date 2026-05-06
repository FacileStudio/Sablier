package timeentries

import (
	"context"

	"api/internal/errors"
	"api/schemas"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func toResponse(e *schemas.TimeEntry) TimeEntryResponse {
	return TimeEntryResponse{
		ID:        e.ID,
		ProjectID: e.ProjectID,
		TaskID:    e.TaskID,
		UserID:    e.UserID,
		StartedAt: e.StartedAt,
		StoppedAt: e.StoppedAt,
		CreatedAt: e.CreatedAt,
		UpdatedAt: e.UpdatedAt,
	}
}

func (c *Controller) start(ctx context.Context, userID string, req *StartTimerRequest) (*TimeEntryResponse, error) {
	if req.ProjectID <= 0 {
		return nil, errors.Invalid("project_id is required")
	}
	if req.TaskID <= 0 {
		return nil, errors.Invalid("task_id is required")
	}
	record, taskName, err := c.service.startTimer(ctx, userID, req.ProjectID, req.TaskID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(record)
	resp.TaskName = taskName
	return &resp, nil
}

func (c *Controller) stop(ctx context.Context, userID string) (*TimeEntryResponse, error) {
	record, taskName, err := c.service.stopTimer(ctx, userID)
	if err != nil {
		return nil, err
	}
	resp := toResponse(record)
	resp.TaskName = taskName
	return &resp, nil
}

func (c *Controller) list(ctx context.Context, projectID int64) (*ListEntriesResponse, error) {
	records, err := c.service.listEntries(ctx, projectID)
	if err != nil {
		return nil, err
	}
	return rowsToResponse(records), nil
}

func (c *Controller) listRunning(ctx context.Context) (*ListEntriesResponse, error) {
	records, err := c.service.listRunningEntries(ctx)
	if err != nil {
		return nil, err
	}
	return rowsToResponse(records), nil
}

func rowsToResponse(records []timeEntryRow) *ListEntriesResponse {
	items := make([]TimeEntryResponse, len(records))
	for i, r := range records {
		items[i] = toResponse(&r.TimeEntry)
		items[i].UserEmail = r.UserEmail
		items[i].UserName = r.UserName
		items[i].UserColor = r.UserColor
		items[i].TaskName = r.TaskName
	}
	return &ListEntriesResponse{Entries: items}
}

func (c *Controller) create(ctx context.Context, userID string, req *CreateEntryRequest) (*TimeEntryResponse, error) {
	if req.ProjectID <= 0 {
		return nil, errors.Invalid("project_id is required")
	}
	if req.TaskID <= 0 {
		return nil, errors.Invalid("task_id is required")
	}
	if req.StartedAt.IsZero() {
		return nil, errors.Invalid("started_at is required")
	}
	if req.StoppedAt.IsZero() || req.StoppedAt.Before(req.StartedAt) {
		return nil, errors.Invalid("stopped_at must be after started_at")
	}
	record, taskName, err := c.service.createEntry(ctx, userID, req.ProjectID, req.TaskID, req.StartedAt, req.StoppedAt)
	if err != nil {
		return nil, err
	}
	resp := toResponse(record)
	resp.TaskName = taskName
	return &resp, nil
}

func (c *Controller) update(ctx context.Context, userID string, entryID int64, req *UpdateEntryRequest) (*TimeEntryResponse, error) {
	if req.ProjectID <= 0 {
		return nil, errors.Invalid("project_id is required")
	}
	if req.TaskID <= 0 {
		return nil, errors.Invalid("task_id is required")
	}
	record, taskName, err := c.service.updateEntry(ctx, userID, entryID, req)
	if err != nil {
		return nil, err
	}
	resp := toResponse(record)
	resp.TaskName = taskName
	return &resp, nil
}

func (c *Controller) delete(ctx context.Context, userID string, entryID int64) error {
	return c.service.deleteEntry(ctx, userID, entryID)
}

func (c *Controller) running(ctx context.Context, userID string) (*TimeEntryResponse, error) {
	record, taskName, err := c.service.getRunningTimer(ctx, userID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return nil, nil
	}
	resp := toResponse(record)
	resp.TaskName = taskName
	return &resp, nil
}
