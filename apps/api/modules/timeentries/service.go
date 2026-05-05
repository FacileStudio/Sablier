package timeentries

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	controller *Controller
}

func NewService(orm *gorm.DB) *Service {
	service := &Service{orm: orm}
	service.controller = newController(service)
	return service
}

func (service *Service) startTimer(ctx context.Context, userID string, projectID int64, description string) (*schemas.TimeEntry, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}

	var running schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&running).Error
	if err == nil {
		return nil, errors.Failed("a timer is already running, stop it first")
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to check running timer", err)
	}

	record := &schemas.TimeEntry{
		ProjectID:   projectID,
		UserID:      uid,
		Description: description,
		StartedAt:   time.Now().UTC(),
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to start timer", err)
	}
	return record, nil
}

func (service *Service) stopTimer(ctx context.Context, userID string) (*schemas.TimeEntry, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}

	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("no running timer")
	}
	if err != nil {
		return nil, errors.Internal("failed to find running timer", err)
	}

	now := time.Now().UTC()
	record.StoppedAt = &now
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to stop timer", err)
	}
	return &record, nil
}

func (service *Service) listEntries(ctx context.Context, userID string, projectID int64) ([]schemas.TimeEntry, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}

	query := service.orm.WithContext(ctx).Where("user_id = ?", uid)
	if projectID > 0 {
		query = query.Where("project_id = ?", projectID)
	}
	var records []schemas.TimeEntry
	if err := query.Order("started_at desc").Limit(100).Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list entries", err)
	}
	return records, nil
}

func (service *Service) createEntry(ctx context.Context, userID string, projectID int64, description string, startedAt, stoppedAt time.Time) (*schemas.TimeEntry, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	stopped := stoppedAt
	record := &schemas.TimeEntry{
		ProjectID:   projectID,
		UserID:      uid,
		Description: description,
		StartedAt:   startedAt.UTC(),
		StoppedAt:   &stopped,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create entry", err)
	}
	return record, nil
}

func (service *Service) updateEntry(ctx context.Context, userID string, entryID int64, req *UpdateEntryRequest) (*schemas.TimeEntry, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}

	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("id = ? AND user_id = ?", entryID, uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("time entry not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to get entry", err)
	}

	record.ProjectID = req.ProjectID
	record.Description = req.Description
	record.StartedAt = req.StartedAt.UTC()
	record.StoppedAt = req.StoppedAt
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to update entry", err)
	}
	return &record, nil
}

func (service *Service) deleteEntry(ctx context.Context, userID string, entryID int64) error {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return errors.Invalid("invalid user id")
	}
	result := service.orm.WithContext(ctx).Where("id = ? AND user_id = ?", entryID, uid).Delete(&schemas.TimeEntry{})
	if result.Error != nil {
		return errors.Internal("failed to delete entry", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("time entry not found")
	}
	return nil
}

func (service *Service) getRunningTimer(ctx context.Context, userID string) (*schemas.TimeEntry, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Internal("failed to get running timer", err)
	}
	return &record, nil
}
