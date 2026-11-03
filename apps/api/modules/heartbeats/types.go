package heartbeats

import "time"

// HeartbeatInput is one activity signal sent by a tool.
type HeartbeatInput struct {
	At        time.Time `json:"at"`
	Source    string    `json:"source"`
	Entity    string    `json:"entity"`
	Project   string    `json:"project"`
	ProjectID *int64    `json:"project_id"`
	Write     bool      `json:"write"`
}

// IngestRequest is the body for POST /heartbeats.
type IngestRequest struct {
	Heartbeats []HeartbeatInput `json:"heartbeats"`
}

// IngestResponse counts the heartbeats inside the accepted time window.
type IngestResponse struct {
	Accepted int `json:"accepted"`
}

// Draft is a suggested time entry computed from heartbeats.
type Draft struct {
	ID          string    `json:"id"`
	StartedAt   time.Time `json:"started_at"`
	StoppedAt   time.Time `json:"stopped_at"`
	ProjectID   *int64    `json:"project_id"`
	ProjectHint string    `json:"project_hint"`
	Sources     []string  `json:"sources"`
	Heartbeats  int       `json:"heartbeats"`
}

// ConfirmRequest is the body for turning a draft into a time entry.
type ConfirmRequest struct {
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
	ProjectID int64     `json:"project_id"`
	TaskID    int64     `json:"task_id"`
}

// DismissRequest is the body for hiding a range from future drafts.
type DismissRequest struct {
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
}

// DismissResponse reports that the range was stored.
type DismissResponse struct {
	Dismissed bool `json:"dismissed"`
}

// DraftSettings holds the user's draft preferences.
type DraftSettings struct {
	AutoConfirmDrafts bool `json:"auto_confirm_drafts"`
}
