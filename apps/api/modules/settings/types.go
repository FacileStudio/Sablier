package settings

type Settings struct {
	WebhookURL string `json:"webhook_url"`
}

type UpdateRequest struct {
	WebhookURL string `json:"webhook_url"`
}

type SettingsResponse struct {
	Settings Settings `json:"settings"`
}
