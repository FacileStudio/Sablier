package schemas

type AppSetting struct {
	ID                  int    `gorm:"primaryKey"`
	WebhookURL          string `gorm:"not null;default:''"`
	WebhookSecretHeader string `gorm:"not null;default:''"`
	WebhookSecretValue  string `gorm:"not null;default:''"`
	AntenneURL          string `gorm:"column:antenne_url;not null;default:''" json:"antenne_url"`
	AntenneSecret       string `gorm:"column:antenne_secret;not null;default:''" json:"antenne_secret"`
	AntenneEnabled      bool   `gorm:"column:antenne_enabled;not null;default:false" json:"antenne_enabled"`
	PoolEvents          string `gorm:"column:pool_events;not null;default:''" json:"pool_events"`
}
