package nookpool

type PoolSettings struct {
	URL     string `json:"nook_pool_url"`
	Secret  string `json:"nook_pool_secret"`
	Enabled bool   `json:"nook_pool_enabled"`
}

type UpdatePoolRequest struct {
	URL     string `json:"nook_pool_url"`
	Secret  string `json:"nook_pool_secret"`
	Enabled bool   `json:"nook_pool_enabled"`
}

type PoolSettingsResponse struct {
	Settings  PoolSettings `json:"pool_settings"`
	Connected bool         `json:"connected"`
}
