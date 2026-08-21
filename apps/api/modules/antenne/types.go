package antenne

// PoolSettings holds the stored Antenne Pool connection configuration.
type PoolSettings struct {
	URL     string `json:"antenne_url"`
	Secret  string `json:"antenne_secret"`
	Enabled bool   `json:"antenne_enabled"`
}

// UpdatePoolRequest is the request body for updating pool settings.
type UpdatePoolRequest struct {
	URL     string `json:"antenne_url"`
	Secret  string `json:"antenne_secret"`
	Enabled bool   `json:"antenne_enabled"`
}

// PoolSettingsResponse describes the current pool settings and connection state.
type PoolSettingsResponse struct {
	Settings     PoolSettings `json:"pool_settings"`
	Connected    bool         `json:"connected"`
	ConnectError string       `json:"connect_error,omitempty"`
	FromEnv      bool         `json:"from_env,omitempty"`
}

// SyncResult reports how many projects and tasks were synced.
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

// PoolEventToggle represents one pool event's enabled state.
type PoolEventToggle struct {
	Event   string `json:"event"`
	Enabled bool   `json:"enabled"`
}

// PoolEventsResponse lists the current pool event toggles.
type PoolEventsResponse struct {
	Events []PoolEventToggle `json:"events"`
}

// UpdatePoolEventsRequest is the request body for updating pool event toggles.
type UpdatePoolEventsRequest struct {
	Events []PoolEventToggle `json:"events"`
}
