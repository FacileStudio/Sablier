package notifications

// SaveSubscriptionRequest is the body for POST /notifications/subscriptions.
type SaveSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	P256DH   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

// VAPIDPublicKeyResponse carries the public VAPID key for the frontend.
type VAPIDPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}

// SubscriptionResponse reports whether a subscription was saved.
type SubscriptionResponse struct {
	Saved bool `json:"saved"`
}

// DeleteResponse reports whether a subscription was deleted.
type DeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type reminderTarget struct {
	UserID         int64  `gorm:"column:user_id"`
	Locale         string `gorm:"column:locale"`
	SubscriptionID int64  `gorm:"column:subscription_id"`
	Endpoint       string `gorm:"column:endpoint"`
	P256DH         string `gorm:"column:p256dh"`
	Auth           string `gorm:"column:auth"`
}
