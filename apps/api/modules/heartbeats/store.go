package heartbeats

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"slices"
	"strings"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// computeForUser loads userID's heartbeats in [from, to] and returns the spans left
// once their time entries, a running one included, and dismissed ranges are removed.
func (service *Service) computeForUser(ctx context.Context, userID int64, from, to time.Time) ([]span, error) {
	index, err := service.loadProjects(ctx, userID)
	if err != nil {
		return nil, err
	}
	var rows []schemas.Heartbeat
	err = service.orm.WithContext(ctx).
		Where("user_id = ? AND at >= ? AND at <= ?", userID, from.UTC(), to.UTC()).
		Order("at asc").Find(&rows).Error
	if err != nil {
		return nil, errors.Internal("failed to list heartbeats", err)
	}
	beats := make([]beat, len(rows))
	for i, row := range rows {
		beats[i] = index.resolve(row)
	}
	busy, err := service.busyIntervals(ctx, userID, from, to.Add(mergeGap+loneHeartbeat))
	if err != nil {
		return nil, err
	}
	return computeDrafts(beats, busy), nil
}

// busyIntervals is every range overlapping [from, to] that a draft must not cover.
func (service *Service) busyIntervals(ctx context.Context, userID int64, from, to time.Time) ([]interval, error) {
	var entries []schemas.TimeEntry
	err := service.orm.WithContext(ctx).
		Where("user_id = ? AND started_at < ? AND (stopped_at IS NULL OR stopped_at > ?)", userID, to.UTC(), from.UTC()).
		Find(&entries).Error
	if err != nil {
		return nil, errors.Internal("failed to list time entries", err)
	}
	var dismissals []schemas.DraftDismissal
	err = service.orm.WithContext(ctx).
		Where("user_id = ? AND started_at < ? AND stopped_at > ?", userID, to.UTC(), from.UTC()).
		Find(&dismissals).Error
	if err != nil {
		return nil, errors.Internal("failed to list dismissed drafts", err)
	}
	busy := make([]interval, 0, len(entries)+len(dismissals))
	for _, e := range entries {
		stop := service.now()
		if e.StoppedAt != nil {
			stop = *e.StoppedAt
		}
		busy = append(busy, interval{Start: e.StartedAt.UTC(), Stop: stop.UTC()})
	}
	for _, d := range dismissals {
		busy = append(busy, interval{Start: d.StartedAt.UTC(), Stop: d.StoppedAt.UTC()})
	}
	return busy, nil
}

// overlapsEntry reports whether userID already has an entry, running or not, inside the range.
func (service *Service) overlapsEntry(ctx context.Context, userID int64, from, to time.Time) (bool, error) {
	var count int64
	err := service.orm.WithContext(ctx).Model(&schemas.TimeEntry{}).
		Where("user_id = ? AND started_at < ? AND (stopped_at IS NULL OR stopped_at > ?)", userID, to.UTC(), from.UTC()).
		Count(&count).Error
	if err != nil {
		return false, errors.Internal("failed to check overlapping entries", err)
	}
	return count > 0, nil
}

func toDraft(s span) Draft {
	sources := slices.Clone(s.Sources)
	slices.Sort(sources)
	sum := sha256.Sum256([]byte(s.Start.UTC().Format(time.RFC3339) + "|" + strings.Join(sources, ",")))
	return Draft{
		ID:          hex.EncodeToString(sum[:8]),
		StartedAt:   s.Start,
		StoppedAt:   s.Stop,
		ProjectID:   s.ProjectID,
		ProjectHint: s.Hint,
		Sources:     sources,
		Heartbeats:  len(s.Beats),
	}
}
