package schemas

type UserSetting struct {
	UserID              int64   `gorm:"primaryKey"`
	WebhookURL          string  `gorm:"not null;default:''"`
	WebhookSecretHeader string  `gorm:"not null;default:''"`
	WebhookSecretValue  string  `gorm:"not null;default:''"`
	Rate                float64 `gorm:"not null;default:0"`
	RateType            string  `gorm:"not null;default:'daily'"`
}
