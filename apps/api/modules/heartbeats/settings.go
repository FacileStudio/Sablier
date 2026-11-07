package heartbeats

import (
	"context"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

func (service *Service) settings(ctx context.Context, userID int64) (*DraftSettings, error) {
	var user schemas.User
	err := service.orm.WithContext(ctx).Select("id", "auto_confirm_drafts", "review_threshold").First(&user, userID).Error
	if err != nil {
		return nil, errors.Internal("failed to read draft settings", err)
	}
	threshold := attribution.Threshold(user.ReviewThreshold)
	return &DraftSettings{AutoConfirmDrafts: user.AutoConfirm, ReviewThreshold: &threshold}, nil
}

func (service *Service) updateSettings(ctx context.Context, userID int64, req *DraftSettings) (*DraftSettings, error) {
	updates := map[string]any{"auto_confirm_drafts": req.AutoConfirmDrafts}
	if req.ReviewThreshold != nil {
		if *req.ReviewThreshold < attribution.MinThreshold || *req.ReviewThreshold > attribution.MaxThreshold {
			return nil, errors.Invalid("review_threshold must be between 0.5 and 1")
		}
		updates["review_threshold"] = *req.ReviewThreshold
	}
	err := service.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", userID).Updates(updates).Error
	if err != nil {
		return nil, errors.Internal("failed to update draft settings", err)
	}
	return service.settings(ctx, userID)
}
