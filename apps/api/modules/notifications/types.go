package notifications

type SaveSubscriptionRequest struct {
	Endpoint string `json:"endpoint"`
	P256DH   string `json:"p256dh"`
	Auth     string `json:"auth"`
}

type VAPIDPublicKeyResponse struct {
	PublicKey string `json:"public_key"`
}

type SubscriptionResponse struct {
	Saved bool `json:"saved"`
}

type DeleteResponse struct {
	Deleted bool `json:"deleted"`
}

type BroadcastRequest struct {
	Title string `json:"title"`
	Body  string `json:"body"`
	Icon  string `json:"icon,omitempty"`
}

type BroadcastResponse struct {
	Sent   int `json:"sent"`
	Failed int `json:"failed"`
}
