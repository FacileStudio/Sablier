package nookpool

import "time"

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
	"timer.started",
	"timer.stopped",
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

type TimerEventPayload struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	ProjectName string     `json:"project_name"`
	TaskID      int64      `json:"task_id"`
	TaskName    string     `json:"task_name"`
	UserID      int64      `json:"user_id"`
	UserEmail   string     `json:"user_email"`
	StartedAt   time.Time  `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at,omitempty"`
}

type TimerEvent struct {
	App            string            `json:"app"`
	Event          string            `json:"event"`
	Payload        TimerEventPayload `json:"payload"`
	Timestamp      string            `json:"timestamp"`
	IdempotencyKey string            `json:"idempotency_key"`
}
