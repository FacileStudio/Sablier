package settings

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const appSettingID = 1

// Service reads and writes the app-level webhook settings.
type Service struct {
	orm        *gorm.DB
	controller *Controller
}

// NewService wires the settings service.
func NewService(orm *gorm.DB) *Service {
	service := &Service{orm: orm}
	service.controller = newController(service)
	return service
}

func (service *Service) record(ctx context.Context) (schemas.AppSetting, error) {
	var record schemas.AppSetting
	err := service.orm.WithContext(ctx).Where("id = ?", appSettingID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return schemas.AppSetting{ID: appSettingID}, nil
	}
	if err != nil {
		return record, errors.Internal("failed to get settings", err)
	}
	return record, nil
}

func view(record schemas.AppSetting) *Settings {
	return &Settings{
		WebhookURL:          record.WebhookURL,
		WebhookSecretHeader: record.WebhookSecretHeader,
		WebhookSecretSet:    record.WebhookSecretValue != "",
	}
}

func (service *Service) getSettings(ctx context.Context) (*Settings, error) {
	record, err := service.record(ctx)
	if err != nil {
		return nil, err
	}
	return view(record), nil
}

func (service *Service) updateSettings(ctx context.Context, req *UpdateRequest) (*Settings, error) {
	current, err := service.record(ctx)
	if err != nil {
		return nil, err
	}
	record := schemas.AppSetting{
		ID:                  appSettingID,
		WebhookURL:          strings.TrimSpace(req.WebhookURL),
		WebhookSecretHeader: strings.TrimSpace(req.WebhookSecretHeader),
		WebhookSecretValue:  current.WebhookSecretValue,
	}
	if req.WebhookSecretValue != nil {
		record.WebhookSecretValue = strings.TrimSpace(*req.WebhookSecretValue)
	}
	if err := service.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"webhook_url", "webhook_secret_header", "webhook_secret_value"}),
	}).Create(&record).Error; err != nil {
		return nil, errors.Internal("failed to update settings", err)
	}
	return view(record), nil
}
