package timeentries

import (
	"context"
	stderrors "errors"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"github.com/FacileStudio/antenne-client/go/envelope"
	"gorm.io/gorm"
)

const (
	maxReviewEntries = 200
	ownedByOwner     = "(time_entries.user_id = ? OR time_entries.user_id IN " +
		"(SELECT id FROM users WHERE kind = 'agent' AND owner_user_id = ?))"
)

// ReviewEntries lists the entries waiting for review that belong to ownerID or to one of
// ownerID's agents, newest first.
func (service *Service) ReviewEntries(ctx context.Context, ownerID int64) ([]TimeEntryResponse, error) {
	var records []timeEntryRow
	err := service.rows(ctx).
		Where("time_entries.needs_review = ?", true).
		Where(ownedByOwner, ownerID, ownerID).
		Order("time_entries.started_at desc").Limit(maxReviewEntries).
		Find(&records).Error
	if err != nil {
		return nil, errors.Internal("failed to list entries to review", err)
	}
	return rowsToResponse(records).Entries, nil
}

// ReviewEntry settles a flagged entry of ownerID or one of ownerID's agents: it moves the
// entry to taskID when taskID is positive, else keeps it where it is, and either way marks
// it as the owner's choice, so a replayed agent session never moves it again.
func (service *Service) ReviewEntry(ctx context.Context, ownerID, entryID, taskID int64) (*TimeEntryResponse, error) {
	var entry schemas.TimeEntry
	err := service.orm.WithContext(ctx).Where("time_entries.id = ?", entryID).
		Where(ownedByOwner, ownerID, ownerID).First(&entry).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("time entry not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to get entry", err)
	}
	if taskID > 0 {
		if err := service.moveToTask(ctx, ownerID, &entry, taskID); err != nil {
			return nil, err
		}
	}
	manual, sure := attribution.SourceManual, 1.0
	entry.Attribution = schemas.Attribution{AttributionSource: &manual, AttributionConfidence: &sure}
	if err := service.orm.WithContext(ctx).Save(&entry).Error; err != nil {
		return nil, errors.Internal("failed to review entry", err)
	}
	service.emitEntry(envelope.ActionUpdated, &entry)
	var records []timeEntryRow
	if err := service.rows(ctx).Where("time_entries.id = ?", entry.ID).Find(&records).Error; err != nil || len(records) == 0 {
		return nil, errors.Internal("failed to read the reviewed entry", err)
	}
	return &rowsToResponse(records).Entries[0], nil
}

func (service *Service) moveToTask(ctx context.Context, ownerID int64, entry *schemas.TimeEntry, taskID int64) error {
	var task schemas.Task
	query := service.orm.WithContext(ctx).Model(&schemas.Task{}).Select("tasks.*").
		Joins("JOIN projects ON projects.id = tasks.project_id").Where("tasks.id = ?", taskID)
	err := schemas.VisibleSpaces(query, schemas.Projects, ownerID).First(&task).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return errors.Invalid("task_id is not a task you can see")
	}
	if err != nil {
		return errors.Internal("failed to get task", err)
	}
	project, err := service.projectFor(ctx, ownerID, task.ProjectID)
	if err != nil {
		return err
	}
	entry.ProjectID, entry.TaskID, entry.SpaceID = task.ProjectID, task.ID, project.SpaceID
	return nil
}
