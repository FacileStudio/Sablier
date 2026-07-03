package timeentries

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"api/internal/errors"
	"api/internal/webhook"
	"api/modules/nookpool"
	"api/schemas"

	enveloppe "github.com/FacileStudio/enveloppe/go"
	"gorm.io/gorm"
)

type Service struct {
	orm         *gorm.DB
	controller  *Controller
	poolService *nookpool.Service
}

func NewService(orm *gorm.DB) *Service {
	service := &Service{orm: orm}
	service.controller = newController(service)
	return service
}

func (service *Service) SetPoolService(ps *nookpool.Service) {
	service.poolService = ps
}

func (service *Service) startTimer(ctx context.Context, userID string, projectID int64, taskID int64, spaceID *string) (*schemas.TimeEntry, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}

	task, err := service.getTask(ctx, projectID, taskID)
	if err != nil {
		return nil, "", err
	}

	var running schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&running).Error
	if err == nil {
		return nil, "", errors.Failed("a timer is already running, stop it first")
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.Internal("failed to check running timer", err)
	}

	var resolvedSpaceID *string
	if spaceID != nil && *spaceID != "" {
		resolvedSpaceID = spaceID
	}

	record := &schemas.TimeEntry{
		ProjectID:          projectID,
		TaskID:             task.ID,
		UserID:             uid,
		SpaceID:            resolvedSpaceID,
		StartedAt:          time.Now().UTC(),
		LastNotificationAt: nil,
	}

	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, "", errors.Internal("failed to start timer", err)
	}

	needsTaskUpdate := false
	if task.Status != schemas.StatusInProgress {
		task.Status = schemas.StatusInProgress
		needsTaskUpdate = true
	}
	if task.ActorID == nil || *task.ActorID != uid {
		task.ActorID = &uid
		needsTaskUpdate = true
	}
	if needsTaskUpdate {
		if err := service.orm.WithContext(ctx).Save(task).Error; err == nil {
			if service.poolService != nil {
				var project schemas.Project
				if err := service.orm.WithContext(ctx).Where("id = ?", projectID).First(&project).Error; err == nil {
					go service.poolService.EmitTaskEvent(enveloppe.ActionUpdated, task, &project)
				}
			}
		}
	}

	service.fireWebhook(ctx, uid, "timer_started", record)
	return record, task.Name, nil
}

func (service *Service) pauseTimer(ctx context.Context, userID string) (*schemas.TimeEntry, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}

	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.NotFound("no running timer")
	}
	if err != nil {
		return nil, "", errors.Internal("failed to find running timer", err)
	}
	if record.PausedAt != nil {
		return nil, "", errors.Failed("timer is already paused")
	}

	now := time.Now().UTC()
	record.PausedAt = &now
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, "", errors.Internal("failed to pause timer", err)
	}
	taskName, err := service.taskName(ctx, record.TaskID)
	if err != nil {
		return nil, "", err
	}
	return &record, taskName, nil
}

func (service *Service) resumeTimer(ctx context.Context, userID string) (*schemas.TimeEntry, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}

	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.NotFound("no running timer")
	}
	if err != nil {
		return nil, "", errors.Internal("failed to find running timer", err)
	}
	if record.PausedAt == nil {
		return nil, "", errors.Failed("timer is not paused")
	}

	now := time.Now().UTC()
	record.PausedDurationMs += now.Sub(*record.PausedAt).Milliseconds()
	record.PausedAt = nil
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, "", errors.Internal("failed to resume timer", err)
	}
	taskName, err := service.taskName(ctx, record.TaskID)
	if err != nil {
		return nil, "", err
	}
	return &record, taskName, nil
}

func (service *Service) stopTimer(ctx context.Context, userID string) (*schemas.TimeEntry, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}

	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.NotFound("no running timer")
	}
	if err != nil {
		return nil, "", errors.Internal("failed to find running timer", err)
	}

	now := time.Now().UTC()
	if record.PausedAt != nil {
		record.PausedDurationMs += now.Sub(*record.PausedAt).Milliseconds()
		record.PausedAt = nil
	}
	record.StoppedAt = &now
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, "", errors.Internal("failed to stop timer", err)
	}
	service.fireWebhook(ctx, uid, "timer_stopped", &record)
	taskName, err := service.taskName(ctx, record.TaskID)
	if err != nil {
		return nil, "", err
	}
	return &record, taskName, nil
}

type timeEntryRow struct {
	schemas.TimeEntry
	UserEmail     string
	UserName      string
	UserColor     string
	UserAvatarURL string
	TaskName      string
}

func (service *Service) listEntries(ctx context.Context, projectID int64, userID int64, spaceID *string) ([]timeEntryRow, error) {
	query := service.orm.WithContext(ctx).
		Model(&schemas.TimeEntry{}).
		Select("time_entries.*, users.email as user_email, users.name as user_name, users.color as user_color, users.avatar_url as user_avatar_url, tasks.name as task_name").
		Joins("JOIN users ON users.id = time_entries.user_id").
		Joins("LEFT JOIN tasks ON tasks.id = time_entries.task_id")
	if projectID > 0 {
		query = query.Where("time_entries.project_id = ?", projectID)
	}
	if userID > 0 {
		query = query.Where("time_entries.user_id = ?", userID)
	}
	if spaceID != nil {
		query = query.Where("time_entries.space_id = ?", *spaceID)
	} else {
		query = query.Where("time_entries.space_id IS NULL")
	}
	var records []timeEntryRow
	if err := query.Order("time_entries.started_at desc").Limit(100).Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list entries", err)
	}
	return records, nil
}

func (service *Service) listRunningEntries(ctx context.Context, spaceID *string) ([]timeEntryRow, error) {
	query := service.orm.WithContext(ctx).
		Model(&schemas.TimeEntry{}).
		Select("time_entries.*, users.email as user_email, users.name as user_name, users.color as user_color, users.avatar_url as user_avatar_url, tasks.name as task_name").
		Joins("JOIN users ON users.id = time_entries.user_id").
		Joins("LEFT JOIN tasks ON tasks.id = time_entries.task_id").
		Where("time_entries.stopped_at IS NULL")
	if spaceID != nil {
		query = query.Where("time_entries.space_id = ?", *spaceID)
	} else {
		query = query.Where("time_entries.space_id IS NULL")
	}
	var records []timeEntryRow
	if err := query.Order("time_entries.started_at asc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list running entries", err)
	}
	return records, nil
}

func (service *Service) createEntry(ctx context.Context, userID string, projectID int64, taskID int64, startedAt, stoppedAt time.Time, spaceID *string) (*schemas.TimeEntry, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}
	task, err := service.getTask(ctx, projectID, taskID)
	if err != nil {
		return nil, "", err
	}
	var resolvedSpaceID *string
	if spaceID != nil && *spaceID != "" {
		resolvedSpaceID = spaceID
	}
	stopped := stoppedAt
	record := &schemas.TimeEntry{
		ProjectID: projectID,
		TaskID:    task.ID,
		UserID:    uid,
		SpaceID:   resolvedSpaceID,
		StartedAt: startedAt.UTC(),
		StoppedAt: &stopped,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, "", errors.Internal("failed to create entry", err)
	}
	return record, task.Name, nil
}

func (service *Service) updateEntry(ctx context.Context, userID string, entryID int64, req *UpdateEntryRequest) (*schemas.TimeEntry, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}

	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("id = ? AND user_id = ?", entryID, uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", errors.NotFound("time entry not found")
	}
	if err != nil {
		return nil, "", errors.Internal("failed to get entry", err)
	}
	if record.StoppedAt != nil && req.StoppedAt == nil {
		return nil, "", errors.Invalid("only the currently running session can remain running after edit")
	}

	task, err := service.getTask(ctx, req.ProjectID, req.TaskID)
	if err != nil {
		return nil, "", err
	}

	record.ProjectID = req.ProjectID
	record.TaskID = task.ID
	record.StartedAt = req.StartedAt.UTC()
	record.StoppedAt = req.StoppedAt
	if err := service.orm.WithContext(ctx).Save(&record).Error; err != nil {
		return nil, "", errors.Internal("failed to update entry", err)
	}
	return &record, task.Name, nil
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
	TaskID      int64      `json:"task_id"`
	TaskName    string     `json:"task_name"`
	UserID      int64      `json:"user_id"`
	UserEmail   string     `json:"user_email"`
	StartedAt   time.Time  `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
}

func (service *Service) fireWebhook(ctx context.Context, userID int64, event string, entry *schemas.TimeEntry) {
	var user schemas.User
	service.orm.WithContext(ctx).Where("id = ?", userID).First(&user)

	var project schemas.Project
	service.orm.WithContext(ctx).Where("id = ?", entry.ProjectID).First(&project)

	var task schemas.Task
	service.orm.WithContext(ctx).Where("id = ?", entry.TaskID).First(&task)

	data := &webhookTimeEntry{
		ID:          entry.ID,
		ProjectID:   entry.ProjectID,
		ProjectName: project.Name,
		TaskID:      entry.TaskID,
		TaskName:    task.Name,
		UserID:      entry.UserID,
		UserEmail:   user.Email,
		StartedAt:   entry.StartedAt,
		StoppedAt:   entry.StoppedAt,
	}

	var setting schemas.AppSetting
	if err := service.orm.WithContext(ctx).Where("id = 1").First(&setting).Error; err == nil && setting.WebhookURL != "" {
		webhook.Fire(setting.WebhookURL, setting.WebhookSecretHeader, setting.WebhookSecretValue, webhook.Payload{
			Event: event,
			Data:  data,
		})
	}

	if service.poolService != nil {
		action := enveloppe.ActionCreated
		if event == "timer_stopped" {
			action = enveloppe.ActionUpdated
		}
		go service.poolService.EmitTimeEntryEvent(action, entry)
	}
}

func (service *Service) getRunningTimer(ctx context.Context, userID string) (*schemas.TimeEntry, string, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, "", errors.Invalid("invalid user id")
	}
	var record schemas.TimeEntry
	err = service.orm.WithContext(ctx).Where("user_id = ? AND stopped_at IS NULL", uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, "", nil
	}
	if err != nil {
		return nil, "", errors.Internal("failed to get running timer", err)
	}
	taskName, err := service.taskName(ctx, record.TaskID)
	if err != nil {
		return nil, "", err
	}
	return &record, taskName, nil
}

func (service *Service) getTask(ctx context.Context, projectID int64, taskID int64) (*schemas.Task, error) {
	var task schemas.Task
	err := service.orm.WithContext(ctx).Where("id = ? AND project_id = ?", taskID, projectID).First(&task).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Invalid("task_id is invalid for project")
	}
	if err != nil {
		return nil, errors.Internal("failed to get task", err)
	}
	return &task, nil
}

func (service *Service) taskName(ctx context.Context, taskID int64) (string, error) {
	var task schemas.Task
	err := service.orm.WithContext(ctx).Where("id = ?", taskID).First(&task).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return "", errors.NotFound("task not found")
	}
	if err != nil {
		return "", errors.Internal("failed to get task", err)
	}
	return task.Name, nil
}
