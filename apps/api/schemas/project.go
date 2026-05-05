package schemas

import "time"

type Project struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	OwnerID     int64     `gorm:"column:owner_id;index"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Project) TableName() string { return "projects" }
