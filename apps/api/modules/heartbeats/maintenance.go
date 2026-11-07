package heartbeats

import (
	"context"
	"log/slog"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// RunMaintenance confirms the drafts of opted-in users that ended more than
// autoConfirmAge ago and resolved to a project, then deletes heartbeats past retention.
// It returns how many drafts became time entries.
func (service *Service) RunMaintenance(ctx context.Context, logger *slog.Logger) (int, error) {
	now := service.now()
	userIDs, err := service.autoConfirmUsers(ctx, now.Add(-maxAge-autoConfirmAge))
	if err != nil {
		return 0, err
	}
	confirmed := 0
	for _, userID := range userIDs {
		confirmed += service.autoConfirmUser(ctx, userID, now, logger)
	}
	cutoff := now.Add(-retention)
	if err := service.orm.WithContext(ctx).Where("at < ?", cutoff).Delete(&schemas.Heartbeat{}).Error; err != nil {
		return confirmed, errors.Internal("failed to prune heartbeats", err)
	}
	if err := service.orm.WithContext(ctx).Where("stopped_at < ?", cutoff).Delete(&schemas.DraftDismissal{}).Error; err != nil {
		return confirmed, errors.Internal("failed to prune dismissed drafts", err)
	}
	return confirmed, nil
}

func (service *Service) autoConfirmUsers(ctx context.Context, since time.Time) ([]int64, error) {
	var ids []int64
	err := service.orm.WithContext(ctx).Model(&schemas.Heartbeat{}).
		Distinct("heartbeats.user_id").
		Joins("JOIN users ON users.id = heartbeats.user_id").
		Where("users.auto_confirm_drafts = ? AND heartbeats.at >= ?", true, since).
		Pluck("heartbeats.user_id", &ids).Error
	if err != nil {
		return nil, errors.Internal("failed to list users for auto-confirm", err)
	}
	return ids, nil
}

// autoConfirmable keeps the drafts that ended before cutoff and whose suggestion is sure
// enough to book without review.
func autoConfirmable(drafts []Draft, cutoff time.Time) []Draft {
	var out []Draft
	for _, d := range drafts {
		if d.Suggestion != nil && !d.Suggestion.NeedsReview && !d.StoppedAt.After(cutoff) {
			out = append(out, d)
		}
	}
	return out
}

func confirmRequest(d Draft) *ConfirmRequest {
	source, confidence := d.Suggestion.Source, d.Suggestion.Confidence
	req := &ConfirmRequest{StartedAt: d.StartedAt, StoppedAt: d.StoppedAt, ProjectID: d.Suggestion.ProjectID,
		attribution: schemas.Attribution{AttributionSource: &source, AttributionConfidence: &confidence}}
	if d.Suggestion.TaskID != nil {
		req.TaskID = *d.Suggestion.TaskID
	}
	return req
}

func (service *Service) autoConfirmUser(ctx context.Context, userID int64, now time.Time, logger *slog.Logger) int {
	drafts, err := service.listDrafts(ctx, userID, now.Add(-maxAge-autoConfirmAge), now)
	if err != nil {
		logger.Error("heartbeats: failed to compute drafts", slog.Int64("user_id", userID), slog.Any("error", err))
		return 0
	}
	confirmed := 0
	for _, d := range autoConfirmable(drafts, now.Add(-autoConfirmAge)) {
		if _, err := service.confirm(ctx, userID, confirmRequest(d)); err != nil {
			logger.Warn("heartbeats: failed to auto-confirm draft", slog.Int64("user_id", userID), slog.Any("error", err))
			continue
		}
		confirmed++
	}
	return confirmed
}
