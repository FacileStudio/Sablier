package schemas

import "time"

type User struct {
	ID           int64     `gorm:"column:id;primaryKey"`
	Email        string    `gorm:"column:email;uniqueIndex"`
	Name         string    `gorm:"column:name"`
	AvatarURL    string    `gorm:"column:avatar_url"`
	Color        string    `gorm:"column:color"`
	PasswordHash string    `gorm:"column:password_hash"`
	Rate         float64   `gorm:"column:rate;not null;default:0"`
	RateType     string    `gorm:"column:rate_type;not null;default:'daily'"`
	WorkdayHours float64   `gorm:"column:workday_hours;not null;default:8"`
	CreatedAt    time.Time `gorm:"column:created_at;autoCreateTime"`
}

func (User) TableName() string { return "users" }
