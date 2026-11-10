package timeentries

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// RangeFilter narrows InRange. Space follows schemas.ScopeSpace; UserID 0 keeps every
// person the scope shows. Limit caps the rows read.
type RangeFilter struct {
	Space     string
	ProjectID int64
	UserID    int64
	From      time.Time
	To        time.Time
	Limit     int
}

// userKey is the decimal user id the controller methods take.
func userKey(userID int64) string {
	return strconv.FormatInt(userID, 10)
}

// Running is GET /time-entries/running: userID's running entry, or nil.
func (service *Service) Running(ctx context.Context, userID int64) (*TimeEntryResponse, error) {
	return service.controller.running(ctx, userKey(userID))
}

// Start is POST /time-entries/start for userID.
func (service *Service) Start(ctx context.Context, userID, projectID, taskID int64) (*TimeEntryResponse, error) {
	return service.controller.start(ctx, userKey(userID), &StartTimerRequest{ProjectID: projectID, TaskID: taskID})
}

// Stop is POST /time-entries/stop for userID.
func (service *Service) Stop(ctx context.Context, userID int64) (*TimeEntryResponse, error) {
	return service.controller.stop(ctx, userKey(userID))
}

// Update is PUT /time-entries/{id} for userID.
func (service *Service) Update(ctx context.Context, userID, entryID int64, req *UpdateEntryRequest) (*TimeEntryResponse, error) {
	return service.controller.update(ctx, userKey(userID), entryID, req)
}

// Delete is DELETE /time-entries/{id} for userID.
func (service *Service) Delete(ctx context.Context, userID, entryID int64) error {
	return service.controller.delete(ctx, userKey(userID), entryID)
}

// Own loads one of userID's own entries with its names, the entries PUT and DELETE
// /time-entries/{id} accept, or a not_found error.
func (service *Service) Own(ctx context.Context, userID, entryID int64) (*TimeEntryResponse, error) {
	var records []timeEntryRow
	err := service.rows(ctx).Where("time_entries.id = ? AND time_entries.user_id = ?", entryID, userID).
		Limit(1).Find(&records).Error
	if err != nil {
		return nil, errors.Internal("failed to get entry", err)
	}
	if len(records) == 0 {
		return nil, errors.NotFound("time entry not found")
	}
	return &rowsToResponse(records).Entries[0], nil
}

// InRange lists the entries the caller can see that overlap [From, To), newest first,
// through the same space scope as GET /time-entries.
func (service *Service) InRange(ctx context.Context, callerID int64, filter RangeFilter) ([]TimeEntryResponse, error) {
	query, err := schemas.ScopeSpace(service.rows(ctx), schemas.Entries, callerID, filter.Space)
	if err != nil {
		return nil, spaceError(err)
	}
	query = query.Where("time_entries.started_at < ? AND (time_entries.stopped_at IS NULL OR time_entries.stopped_at > ?)",
		filter.To.UTC(), filter.From.UTC())
	if filter.ProjectID > 0 {
		query = query.Where("time_entries.project_id = ?", filter.ProjectID)
	}
	if filter.UserID > 0 {
		query = query.Where("time_entries.user_id = ?", filter.UserID)
	}
	var records []timeEntryRow
	err = query.Order("time_entries.started_at desc").Order("time_entries.id desc").Limit(filter.Limit).Find(&records).Error
	if err != nil && !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to list entries", err)
	}
	return rowsToResponse(records).Entries, nil
}
