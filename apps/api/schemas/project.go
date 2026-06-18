package schemas

import "time"

type Project struct {
	ID          int64     `gorm:"column:id;primaryKey"`
	Name        string    `gorm:"column:name"`
	Description string    `gorm:"column:description"`
	Icon        *string   `gorm:"column:icon;default:'Layout'" json:"icon,omitempty"`
	OwnerID     int64     `gorm:"column:owner_id;index"`
	SpaceID     *string   `gorm:"column:space_id;type:uuid;index" json:"space_id,omitempty"`
	FacileID    *string   `gorm:"column:facile_id;uniqueIndex" json:"facile_id,omitempty"`
	CreatedAt   time.Time `gorm:"column:created_at;autoCreateTime"`
	UpdatedAt   time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (Project) TableName() string { return "projects" }
