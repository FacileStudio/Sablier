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

func (service *Service) getSettings(ctx context.Context) (*Settings, error) {
	var record schemas.AppSetting
	err := service.orm.WithContext(ctx).Where("id = ?", appSettingID).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return &Settings{}, nil
	}
	if err != nil {
		return nil, errors.Internal("failed to get settings", err)
	}
	return &Settings{
		WebhookURL:          record.WebhookURL,
		WebhookSecretHeader: record.WebhookSecretHeader,
		WebhookSecretValue:  record.WebhookSecretValue,
	}, nil
}

func (service *Service) updateSettings(ctx context.Context, req *UpdateRequest) (*Settings, error) {
	record := schemas.AppSetting{
		ID:                  appSettingID,
		WebhookURL:          strings.TrimSpace(req.WebhookURL),
		WebhookSecretHeader: strings.TrimSpace(req.WebhookSecretHeader),
		WebhookSecretValue:  strings.TrimSpace(req.WebhookSecretValue),
	}
	if err := service.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "id"}},
		DoUpdates: clause.AssignmentColumns([]string{"webhook_url", "webhook_secret_header", "webhook_secret_value"}),
	}).Create(&record).Error; err != nil {
		return nil, errors.Internal("failed to update settings", err)
	}
	return &Settings{
		WebhookURL:          record.WebhookURL,
		WebhookSecretHeader: record.WebhookSecretHeader,
		WebhookSecretValue:  record.WebhookSecretValue,
	}, nil
}
