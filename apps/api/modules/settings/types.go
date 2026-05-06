package settings

type Settings struct {
	WebhookURL          string  `json:"webhook_url"`
	WebhookSecretHeader string  `json:"webhook_secret_header"`
	WebhookSecretValue  string  `json:"webhook_secret_value"`
	Rate                float64 `json:"rate"`
	RateType            string  `json:"rate_type"`
}

type UpdateRequest struct {
	WebhookURL          string  `json:"webhook_url"`
	WebhookSecretHeader string  `json:"webhook_secret_header"`
	WebhookSecretValue  string  `json:"webhook_secret_value"`
	Rate                float64 `json:"rate"`
	RateType            string  `json:"rate_type"`
}

type SettingsResponse struct {
	Settings Settings `json:"settings"`
}
