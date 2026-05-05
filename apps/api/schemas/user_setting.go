package schemas

type UserSetting struct {
	UserID     int64  `gorm:"primaryKey"`
	WebhookURL string `gorm:"not null;default:''"`
}
