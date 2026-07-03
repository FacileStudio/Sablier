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

var AllPoolEvents = []string{
	"time_entry.created",
	"time_entry.updated",
	"project.created",
	"project.updated",
	"project.deleted",
	"task.created",
	"task.updated",
	"task.deleted",
}

type PoolEventToggle struct {
	Event   string `json:"event"`
	Enabled bool   `json:"enabled"`
}

type PoolEventsResponse struct {
	Events []PoolEventToggle `json:"events"`
}

type UpdatePoolEventsRequest struct {
	Events []PoolEventToggle `json:"events"`
}
