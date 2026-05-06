package schemas

type AppSetting struct {
	ID                  int    `gorm:"primaryKey"`
	WebhookURL          string `gorm:"not null;default:''"`
	WebhookSecretHeader string `gorm:"not null;default:''"`
	WebhookSecretValue  string `gorm:"not null;default:''"`
}
