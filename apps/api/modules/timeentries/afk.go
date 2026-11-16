package timeentries

import (
	"context"
	stderrors "errors"
	"strconv"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"github.com/FacileStudio/antenne-client/go/envelope"
	"gorm.io/gorm"
)

const (
	// IdleCap is how long a running timer may go without any sign of activity before the
	// server stops it.
	IdleCap = 10 * time.Hour
	// activityGap is the longest gap between two heartbeats that still counts as one stretch
	// of work, the same gap drafts merge heartbeats with.
	activityGap = 15 * time.Minute
)

// finish ends record at end, folding a pause still open at end into its paused time, then
// saves it and announces the stop.
func (service *Service) finish(ctx context.Context, record *schemas.TimeEntry, end time.Time) error {
	if record.PausedAt != nil {
		if end.After(*record.PausedAt) {
			record.PausedDurationMs += end.Sub(*record.PausedAt).Milliseconds()
		}
		record.PausedAt = nil
	}
	record.StoppedAt = &end
	if err := service.orm.WithContext(ctx).Save(record).Error; err != nil {
		return errors.Internal("failed to stop timer", err)
	}
	service.emitEntry(envelope.ActionUpdated, record)
	service.fireWebhook(ctx, record.UserID, "timer_stopped", record)
	return nil
}

// markChecked records now as the last check of record when it is still running.
func markChecked(record *schemas.TimeEntry) {
	if record.StoppedAt == nil {
		now := time.Now().UTC()
		record.CheckedAt = &now
	}
}

// lastActive is the last sign that the person was still working on the running record.
func (service *Service) lastActive(ctx context.Context, record schemas.TimeEntry) (time.Time, error) {
	last, err := schemas.LastActivity(service.orm.WithContext(ctx), record, activityGap, time.Now().UTC())
	if err != nil {
		return last, errors.Internal("failed to read recent activity", err)
	}
	return last, nil
}

// checkTimer records the person's "keep going" answer on their running timer.
func (service *Service) checkTimer(ctx context.Context, userID string) (*schemas.TimeEntry, string, error) {
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
	record.CheckedAt = &now
	if err := service.orm.WithContext(ctx).Model(&record).Update("checked_at", now).Error; err != nil {
		return nil, "", errors.Internal("failed to keep the timer", err)
	}
	taskName, err := service.taskName(ctx, record.TaskID)
	if err != nil {
		return nil, "", err
	}
	return &record, taskName, nil
}

// StopIdleTimers stops every running, unpaused timer with no sign of activity for IdleCap.
// It ends one at its last activity, or IdleCap after its checkpoint when nothing followed
// that checkpoint, and returns the owners of the timers it stopped.
func (service *Service) StopIdleTimers(ctx context.Context, now time.Time) ([]int64, error) {
	var records []schemas.TimeEntry
	err := service.orm.WithContext(ctx).
		Where("stopped_at IS NULL AND paused_at IS NULL AND COALESCE(checked_at, started_at) <= ?", now.Add(-IdleCap)).
		Order("id asc").
		Find(&records).Error
	if err != nil {
		return nil, errors.Internal("failed to list running timers", err)
	}
	var stopped []int64
	for i := range records {
		end, idle, err := idleEnd(service.orm.WithContext(ctx), records[i], now)
		if err != nil {
			return stopped, err
		}
		if !idle {
			continue
		}
		if err := service.finish(ctx, &records[i], end); err != nil {
			return stopped, err
		}
		stopped = append(stopped, records[i].UserID)
	}
	return stopped, nil
}

func idleEnd(db *gorm.DB, record schemas.TimeEntry, now time.Time) (time.Time, bool, error) {
	last, err := schemas.LastActivity(db, record, activityGap, now)
	if err != nil {
		return time.Time{}, false, errors.Internal("failed to read recent activity", err)
	}
	if now.Sub(last) < IdleCap {
		return time.Time{}, false, nil
	}
	if last.Equal(schemas.Checkpoint(record)) {
		return last.Add(IdleCap), true, nil
	}
	return last, true, nil
}
