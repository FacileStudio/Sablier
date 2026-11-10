package schemas

import "time"

type ForfaitState struct {
	ID        int       `gorm:"column:id;primaryKey;autoIncrement:false"`
	Document  string    `gorm:"column:document;type:text;not null"`
	UpdatedAt time.Time `gorm:"column:updated_at;autoUpdateTime"`
}

func (ForfaitState) TableName() string { return "forfait_state" }
