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
	Settings     PoolSettings `json:"pool_settings"`
	Connected    bool         `json:"connected"`
	ConnectError string       `json:"connect_error,omitempty"`
	FromEnv      bool         `json:"from_env,omitempty"`
}

type SyncResult struct {
	ProjectsSynced int `json:"projects_synced"`
	TasksSynced    int `json:"tasks_synced"`
}
