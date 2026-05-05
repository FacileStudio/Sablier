package schemas

import "time"

type TimeEntry struct {
	ID          int64      `gorm:"column:id;primaryKey"`
	ProjectID   int64      `gorm:"column:project_id;index"`
	UserID      int64      `gorm:"column:user_id;index"`
	Description string     `gorm:"column:description"`
	StartedAt   time.Time  `gorm:"column:started_at"`
	StoppedAt   *time.Time `gorm:"column:stopped_at"`
	CreatedAt   time.Time  `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time  `gorm:"column:updated_at;autoUpdateTime"`
}

func (TimeEntry) TableName() string { return "time_entries" }
