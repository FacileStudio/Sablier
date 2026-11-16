package schemas

import (
	"time"

	"gorm.io/gorm"
)

// Checkpoint is the last moment the person confirmed entry: its start, or a later resume,
// edit or "keep going" answer.
func Checkpoint(entry TimeEntry) time.Time {
	if entry.CheckedAt != nil && entry.CheckedAt.After(entry.StartedAt) {
		return *entry.CheckedAt
	}
	return entry.StartedAt
}

// LastActivity is the last sign that the person was still at work on entry: its checkpoint,
// pushed forward by each of their heartbeats up to until that follows the previous sign
// within gap. The first longer gap ends the chain, so a heartbeat days later does not keep
// a forgotten timer alive.
func LastActivity(db *gorm.DB, entry TimeEntry, gap time.Duration, until time.Time) (time.Time, error) {
	last := Checkpoint(entry)
	var beats []time.Time
	err := db.Model(&Heartbeat{}).
		Where("user_id = ? AND at > ? AND at <= ?", entry.UserID, last, until).
		Order("at asc").
		Pluck("at", &beats).Error
	if err != nil {
		return last, err
	}
	for _, at := range beats {
		if at.Sub(last) > gap {
			break
		}
		last = at
	}
	return last, nil
}
