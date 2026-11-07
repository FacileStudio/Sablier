package heartbeats

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

const draftTaskName = "Auto-tracked"

// confirm books a draft as a regular time entry, refusing a range that an entry
// created since the draft was computed now overlaps.
func (service *Service) confirm(ctx context.Context, userID int64, req *ConfirmRequest) (*timeentries.TimeEntryResponse, error) {
	index, err := service.loadProjects(ctx, userID)
	if err != nil {
		return nil, err
	}
	if !index.visible[req.ProjectID] {
		return nil, errors.Invalid("project_id is not a project you can see")
	}
	overlaps, err := service.overlapsEntry(ctx, userID, req.StartedAt, req.StoppedAt)
	if err != nil {
		return nil, err
	}
	if overlaps {
		return nil, errors.Conflict("the range overlaps an existing time entry")
	}
	taskID := req.TaskID
	if taskID <= 0 {
		task, err := service.ensureTask(ctx, req.ProjectID)
		if err != nil {
			return nil, err
		}
		taskID = task.ID
	}
	return service.entries.CreateEntry(ctx, userID, &timeentries.CreateEntryRequest{
		ProjectID:   req.ProjectID,
		TaskID:      taskID,
		StartedAt:   req.StartedAt.UTC(),
		StoppedAt:   req.StoppedAt.UTC(),
		Attribution: req.attribution,
	})
}

func (service *Service) ensureTask(ctx context.Context, projectID int64) (*schemas.Task, error) {
	var task schemas.Task
	lookup := func() error {
		return service.orm.WithContext(ctx).
			Where("project_id = ? AND lower(name) = lower(?)", projectID, draftTaskName).First(&task).Error
	}
	err := lookup()
	if err == nil {
		return &task, nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to find the draft task", err)
	}
	task = schemas.Task{ProjectID: projectID, Name: draftTaskName}
	if err := service.orm.WithContext(ctx).Create(&task).Error; err != nil {
		if retryErr := lookup(); retryErr != nil {
			return nil, errors.Internal("failed to create the draft task", err)
		}
	}
	return &task, nil
}

func (service *Service) dismiss(ctx context.Context, userID int64, req *DismissRequest) error {
	record := schemas.DraftDismissal{UserID: userID, StartedAt: req.StartedAt.UTC(), StoppedAt: req.StoppedAt.UTC()}
	if err := service.orm.WithContext(ctx).Create(&record).Error; err != nil {
		return errors.Internal("failed to dismiss the draft", err)
	}
	return nil
}
