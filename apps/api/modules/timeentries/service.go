package timeentries

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"api/internal/errors"
	"api/internal/webhook"
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
	service.fireWebhook(ctx, uid, "timer_started", record)
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
	service.fireWebhook(ctx, uid, "timer_stopped", &record)
	return &record, nil
}

type timeEntryRow struct {
	schemas.TimeEntry
	UserEmail string
}

func (service *Service) listEntries(ctx context.Context, projectID int64) ([]timeEntryRow, error) {
	query := service.orm.WithContext(ctx).
		Model(&schemas.TimeEntry{}).
		Select("time_entries.*, users.email as user_email").
		Joins("JOIN users ON users.id = time_entries.user_id")
	if projectID > 0 {
		query = query.Where("time_entries.project_id = ?", projectID)
	}
	var records []timeEntryRow
	if err := query.Order("time_entries.started_at desc").Limit(100).Find(&records).Error; err != nil {
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

type webhookTimeEntry struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	ProjectName string     `json:"project_name"`
	UserID      int64      `json:"user_id"`
	UserEmail   string     `json:"user_email"`
	Description string     `json:"description"`
	StartedAt   time.Time  `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
}

func (service *Service) fireWebhook(ctx context.Context, userID int64, event string, entry *schemas.TimeEntry) {
	var setting schemas.UserSetting
	if err := service.orm.WithContext(ctx).Where("user_id = ?", userID).First(&setting).Error; err != nil {
		return
	}
	if setting.WebhookURL == "" {
		return
	}

	var user schemas.User
	service.orm.WithContext(ctx).Where("id = ?", userID).First(&user)

	var project schemas.Project
	service.orm.WithContext(ctx).Where("id = ?", entry.ProjectID).First(&project)

	data := &webhookTimeEntry{
		ID:          entry.ID,
		ProjectID:   entry.ProjectID,
		ProjectName: project.Name,
		UserID:      entry.UserID,
		UserEmail:   user.Email,
		Description: entry.Description,
		StartedAt:   entry.StartedAt,
		StoppedAt:   entry.StoppedAt,
	}

	webhook.Fire(setting.WebhookURL, setting.WebhookSecretHeader, setting.WebhookSecretValue, webhook.Payload{
		Event: event,
		Data:  data,
	})
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
