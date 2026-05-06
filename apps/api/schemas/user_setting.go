package schemas

type UserSetting struct {
	UserID              int64  `gorm:"primaryKey"`
	WebhookURL          string `gorm:"not null;default:''"`
	WebhookSecretHeader string `gorm:"not null;default:''"`
	WebhookSecretValue  string `gorm:"not null;default:''"`
}
