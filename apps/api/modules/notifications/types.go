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

// BroadcastRequest is the body for the test broadcast endpoint.
type BroadcastRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
}

// BroadcastResponse reports how many pushes were sent and failed.
type BroadcastResponse struct {
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
}
