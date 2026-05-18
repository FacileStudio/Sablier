package schemas

type PushSubscription struct {
	ID       int64  `gorm:"column:id;primaryKey"`
	UserID   int64  `gorm:"column:user_id;uniqueIndex;not null"`
	Endpoint string `gorm:"column:endpoint;not null"`
	P256DH   string `gorm:"column:p256dh;not null"`
	Auth     string `gorm:"column:auth;not null"`
}

func (PushSubscription) TableName() string { return "push_subscriptions" }
