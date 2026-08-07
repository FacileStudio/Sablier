package antenne

type PoolSettings struct {
	URL     string `json:"antenne_url"`
	Secret  string `json:"antenne_secret"`
	Enabled bool   `json:"antenne_enabled"`
}

type UpdatePoolRequest struct {
	URL     string `json:"antenne_url"`
	Secret  string `json:"antenne_secret"`
	Enabled bool   `json:"antenne_enabled"`
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
	"agent_session.created",
	"agent_session.updated",
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
