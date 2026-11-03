package agents

import "time"

// RecordSessionRequest is the body for booking one span of agent work. The
// authenticated caller is the agent's owner.
type RecordSessionRequest struct {
	AgentKey      string     `json:"agent_key"`
	Name          string     `json:"name"`
	AvatarURL     string     `json:"avatar_url"`
	FacileID      string     `json:"facile_id"`
	Project       string     `json:"project"`
	ProjectID     int64      `json:"project_id"`
	StartedAt     *time.Time `json:"started_at"`
	StoppedAt     *time.Time `json:"stopped_at"`
	ActiveSeconds *int64     `json:"active_seconds"`
	TokensIn      int64      `json:"tokens_in"`
	TokensOut     int64      `json:"tokens_out"`
	Cowork        bool       `json:"cowork"`
	Description   string     `json:"description"`
}

// AgentResponse is the serialized shape of an agent user.
type AgentResponse struct {
	ID          int64  `json:"id"`
	Name        string `json:"name"`
	AvatarURL   string `json:"avatar_url"`
	Color       string `json:"color"`
	AgentKey    string `json:"agent_key"`
	OwnerUserID int64  `json:"owner_user_id"`
}

// SessionEntryResponse is the time entry an agent session was booked as.
type SessionEntryResponse struct {
	ID             int64      `json:"id"`
	ProjectID      int64      `json:"project_id"`
	TaskID         int64      `json:"task_id"`
	TaskName       string     `json:"task_name"`
	UserID         int64      `json:"user_id"`
	FacileID       string     `json:"facile_id"`
	Description    string     `json:"description"`
	StartedAt      time.Time  `json:"started_at"`
	StoppedAt      *time.Time `json:"stopped_at"`
	PausedDuration int64      `json:"paused_duration_ms"`
	Cowork         bool       `json:"cowork"`
}

// RecordSessionResponse carries the upserted entry and the agent it belongs to.
type RecordSessionResponse struct {
	Entry SessionEntryResponse `json:"entry"`
	Agent AgentResponse        `json:"agent"`
}

// PairedResponse is one agent's time over the requested window, in seconds.
// Paired is the overlap of the owner's own entries with the agent's entries.
type PairedResponse struct {
	AgentID       int64 `json:"agent_id"`
	AgentSeconds  int64 `json:"agent_seconds"`
	CoworkSeconds int64 `json:"cowork_seconds"`
	PairedSeconds int64 `json:"paired_seconds"`
}
