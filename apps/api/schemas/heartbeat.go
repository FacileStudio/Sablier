package schemas

import "time"

// Heartbeat is one activity signal from a tool, the raw material drafts are computed from.
// The unique index makes a resent heartbeat a no-op instead of a duplicate row.
type Heartbeat struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	UserID      int64     `gorm:"column:user_id;not null;index;index:idx_heartbeats_user_at,priority:1;uniqueIndex:idx_heartbeats_dedupe,priority:1"`
	At          time.Time `gorm:"column:at;not null;index;index:idx_heartbeats_user_at,priority:2;uniqueIndex:idx_heartbeats_dedupe,priority:4"`
	Source      string    `gorm:"column:source;not null;uniqueIndex:idx_heartbeats_dedupe,priority:2"`
	Entity      string    `gorm:"column:entity;not null;default:'';uniqueIndex:idx_heartbeats_dedupe,priority:3"`
	ProjectHint string    `gorm:"column:project_hint;not null;default:''"`
	ProjectID   *int64    `gorm:"column:project_id"`
	Write       bool      `gorm:"column:write;not null;default:false"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (Heartbeat) TableName() string { return "heartbeats" }

// DraftDismissal is a range the user rejected; no draft is suggested inside it again.
type DraftDismissal struct {
	ID        int64     `gorm:"column:id;primaryKey"`
	UserID    int64     `gorm:"column:user_id;not null;index"`
	StartedAt time.Time `gorm:"column:started_at;not null"`
	StoppedAt time.Time `gorm:"column:stopped_at;not null"`
	CreatedAt time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (DraftDismissal) TableName() string { return "draft_dismissals" }
