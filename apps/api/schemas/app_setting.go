package schemas

type AppSetting struct {
	ID                  int    `gorm:"primaryKey"`
	WebhookURL          string `gorm:"not null;default:''"`
	WebhookSecretHeader string `gorm:"not null;default:''"`
	WebhookSecretValue  string `gorm:"not null;default:''"`
	NookPoolURL         string `gorm:"column:nook_pool_url;not null;default:''" json:"nook_pool_url"`
	NookPoolSecret      string `gorm:"column:nook_pool_secret;not null;default:''" json:"nook_pool_secret"`
	NookPoolEnabled     bool   `gorm:"column:nook_pool_enabled;not null;default:false" json:"nook_pool_enabled"`
}
