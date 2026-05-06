package settings

type Settings struct {
	WebhookURL          string `json:"webhook_url"`
	WebhookSecretHeader string `json:"webhook_secret_header"`
	WebhookSecretValue  string `json:"webhook_secret_value"`
}

type UpdateRequest struct {
	WebhookURL          string `json:"webhook_url"`
	WebhookSecretHeader string `json:"webhook_secret_header"`
	WebhookSecretValue  string `json:"webhook_secret_value"`
}

type SettingsResponse struct {
	Settings Settings `json:"settings"`
}
