package settings

// Settings holds the app-level webhook configuration.
type Settings struct {
	WebhookURL          string `json:"webhook_url"`
	WebhookSecretHeader string `json:"webhook_secret_header"`
	WebhookSecretValue  string `json:"webhook_secret_value"`
}

// UpdateRequest is the body for updating settings.
type UpdateRequest struct {
	WebhookURL          string `json:"webhook_url"`
	WebhookSecretHeader string `json:"webhook_secret_header"`
	WebhookSecretValue  string `json:"webhook_secret_value"`
}

// SettingsResponse wraps the current settings.
type SettingsResponse struct {
	Settings Settings `json:"settings"`
}
