package settings

import (
	"context"
	stderrors "errors"
	"strconv"

	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
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

func (service *Service) getSettings(ctx context.Context, userID string) (*Settings, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	var record schemas.UserSetting
	err = service.orm.WithContext(ctx).Where("user_id = ?", uid).First(&record).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return &Settings{WebhookURL: ""}, nil
	}
	if err != nil {
		return nil, errors.Internal("failed to get settings", err)
	}
	rateType := record.RateType
	if rateType == "" {
		rateType = "daily"
	}
	return &Settings{
		WebhookURL:          record.WebhookURL,
		WebhookSecretHeader: record.WebhookSecretHeader,
		WebhookSecretValue:  record.WebhookSecretValue,
		Rate:                record.Rate,
		RateType:            rateType,
	}, nil
}

func (service *Service) updateSettings(ctx context.Context, userID string, req *UpdateRequest) (*Settings, error) {
	uid, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Invalid("invalid user id")
	}
	rateType := req.RateType
	if rateType != "daily" && rateType != "hourly" {
		rateType = "daily"
	}
	record := schemas.UserSetting{
		UserID:              uid,
		WebhookURL:          req.WebhookURL,
		WebhookSecretHeader: req.WebhookSecretHeader,
		WebhookSecretValue:  req.WebhookSecretValue,
		Rate:                req.Rate,
		RateType:            rateType,
	}
	if err := service.orm.WithContext(ctx).Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "user_id"}},
		DoUpdates: clause.AssignmentColumns([]string{"webhook_url", "webhook_secret_header", "webhook_secret_value", "rate", "rate_type"}),
	}).Create(&record).Error; err != nil {
		return nil, errors.Internal("failed to update settings", err)
	}
	return &Settings{
		WebhookURL:          record.WebhookURL,
		WebhookSecretHeader: record.WebhookSecretHeader,
		WebhookSecretValue:  record.WebhookSecretValue,
		Rate:                record.Rate,
		RateType:            rateType,
	}, nil
}
