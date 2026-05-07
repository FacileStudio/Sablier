package schemas

import "time"

type TimeEntry struct {
	ID                int64      `gorm:"column:id;primaryKey"`
	ProjectID         int64      `gorm:"column:project_id;index"`
	TaskID            int64      `gorm:"column:task_id;index"`
	UserID            int64      `gorm:"column:user_id;index"`
	LegacyDescription string     `gorm:"column:description"`
	StartedAt         time.Time  `gorm:"column:started_at"`
	StoppedAt         *time.Time `gorm:"column:stopped_at"`
	PausedAt          *time.Time `gorm:"column:paused_at"`
	PausedDurationMs  int64      `gorm:"column:paused_duration_ms"`
	CreatedAt         time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt         time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (TimeEntry) TableName() string { return "time_entries" }
