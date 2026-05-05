package projects

import (
	"context"
	stderrors "errors"
	"strconv"

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

func (service *Service) createProject(ctx context.Context, userID string, name, description string) (*schemas.Project, error) {
	ownerID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	record := &schemas.Project{
		Name:        name,
		Description: description,
		OwnerID:     ownerID,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create project", err)
	}
	return record, nil
}

func (service *Service) listProjects(ctx context.Context, userID string) ([]schemas.Project, error) {
	ownerID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	var records []schemas.Project
	if err := service.orm.WithContext(ctx).Where("owner_id = ?", ownerID).Order("created_at desc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list projects", err)
	}
	return records, nil
}

func (service *Service) getProject(ctx context.Context, userID string, projectID int64) (*schemas.Project, error) {
	ownerID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	var record schemas.Project
	err = service.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", projectID, ownerID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.NotFound("project not found")
	}
	if err != nil {
		return nil, errors.Internal("failed to get project", err)
	}
	return &record, nil
}

func (service *Service) updateProject(ctx context.Context, userID string, projectID int64, name, description string) (*schemas.Project, error) {
	record, err := service.getProject(ctx, userID, projectID)
	if err != nil {
		return nil, err
	}
	record.Name = name
	record.Description = description
	if err := service.orm.WithContext(ctx).Save(record).Error; err != nil {
		return nil, errors.Internal("failed to update project", err)
	}
	return record, nil
}

func (service *Service) deleteProject(ctx context.Context, userID string, projectID int64) error {
	ownerID, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return errors.Invalid("invalid user id")
	}
	result := service.orm.WithContext(ctx).Where("id = ? AND owner_id = ?", projectID, ownerID).Delete(&schemas.Project{})
	if result.Error != nil {
		return errors.Internal("failed to delete project", result.Error)
	}
	if result.RowsAffected == 0 {
		return errors.NotFound("project not found")
	}
	return nil
}

func (service *Service) listTasks(ctx context.Context, userID string, projectID int64) ([]schemas.Task, error) {
	if _, err := service.getProject(ctx, userID, projectID); err != nil {
		return nil, err
	}
	var records []schemas.Task
	if err := service.orm.WithContext(ctx).Where("project_id = ?", projectID).Order("name asc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list tasks", err)
	}
	return records, nil
}

func (service *Service) createTask(ctx context.Context, userID string, projectID int64, name string) (*schemas.Task, error) {
	if _, err := service.getProject(ctx, userID, projectID); err != nil {
		return nil, err
	}
	var existing schemas.Task
	err := service.orm.WithContext(ctx).Where("project_id = ? AND lower(name) = lower(?)", projectID, name).First(&existing).Error
	if err == nil {
		return &existing, nil
	}
	if !stderrors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.Internal("failed to check task", err)
	}
	record := &schemas.Task{
		ProjectID: projectID,
		Name:      name,
	}
	if err := service.orm.WithContext(ctx).Create(record).Error; err != nil {
		return nil, errors.Internal("failed to create task", err)
	}
	return record, nil
}
