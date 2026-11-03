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

// autoConfirmable keeps the drafts that ended before cutoff and resolved to a project.
func autoConfirmable(spans []span, cutoff time.Time) []span {
	var out []span
	for _, s := range spans {
		if s.ProjectID != nil && !s.Stop.After(cutoff) {
			out = append(out, s)
		}
	}
	return out
}

func (service *Service) autoConfirmUser(ctx context.Context, userID int64, now time.Time, logger *slog.Logger) int {
	spans, err := service.computeForUser(ctx, userID, now.Add(-maxAge-autoConfirmAge), now)
	if err != nil {
		logger.Error("heartbeats: failed to compute drafts", slog.Int64("user_id", userID), slog.Any("error", err))
		return 0
	}
	confirmed := 0
	for _, s := range autoConfirmable(spans, now.Add(-autoConfirmAge)) {
		req := &ConfirmRequest{StartedAt: s.Start, StoppedAt: s.Stop, ProjectID: *s.ProjectID}
		if _, err := service.confirm(ctx, userID, req); err != nil {
			logger.Warn("heartbeats: failed to auto-confirm draft", slog.Int64("user_id", userID), slog.Any("error", err))
			continue
		}
		confirmed++
	}
	return confirmed
}
