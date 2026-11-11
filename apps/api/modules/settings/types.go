package settings

// Settings is the app-level webhook configuration as the API shows it. The
// secret value never leaves the server: WebhookSecretSet says whether one is stored.
type Settings struct {
	WebhookURL          string `json:"webhook_url"`
	WebhookSecretHeader string `json:"webhook_secret_header"`
	WebhookSecretSet    bool   `json:"webhook_secret_set"`
}

// UpdateRequest is the body for updating settings. A missing or null
// webhook_secret_value keeps the stored secret, and an empty string clears it.
type UpdateRequest struct {
	WebhookURL          string  `json:"webhook_url"`
	WebhookSecretHeader string  `json:"webhook_secret_header"`
	WebhookSecretValue  *string `json:"webhook_secret_value"`
}

// SettingsResponse wraps the current settings.
type SettingsResponse struct {
	Settings Settings `json:"settings"`
}
