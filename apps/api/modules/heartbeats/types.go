package heartbeats

import (
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

// HeartbeatInput is one activity signal sent by a tool.
type HeartbeatInput struct {
	At        time.Time `json:"at"`
	Source    string    `json:"source"`
	Entity    string    `json:"entity"`
	Project   string    `json:"project"`
	ProjectID *int64    `json:"project_id"`
	Write     bool      `json:"write"`
	HeartbeatHints
}

// HeartbeatHints are optional attribution facts: the Registre space the tool works in,
// the Opus task (facile id) it works on and a short title.
type HeartbeatHints struct {
	RegistreSpaceID string `json:"registre_space_id"`
	OpusTaskID      string `json:"opus_task_id"`
	Title           string `json:"title"`
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
	ID          string      `json:"id"`
	StartedAt   time.Time   `json:"started_at"`
	StoppedAt   time.Time   `json:"stopped_at"`
	ProjectID   *int64      `json:"project_id"`
	ProjectHint string      `json:"project_hint"`
	Sources     []string    `json:"sources"`
	Heartbeats  int         `json:"heartbeats"`
	Suggestion  *Suggestion `json:"suggestion"`
}

// Suggestion is the resolver's best guess for a draft. TaskID is null when it found a
// project but no task; confirming then books on the project's Auto-tracked task. The
// draft carries no suggestion when nothing matched.
type Suggestion struct {
	TaskID      *int64  `json:"task_id"`
	ProjectID   int64   `json:"project_id"`
	SpaceID     *string `json:"space_id"`
	Source      string  `json:"source"`
	Confidence  float64 `json:"confidence"`
	NeedsReview bool    `json:"needs_review"`
}

// ConfirmRequest is the body for turning a draft into a time entry.
type ConfirmRequest struct {
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
	ProjectID int64     `json:"project_id"`
	TaskID    int64     `json:"task_id"`

	attribution schemas.Attribution
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

// DraftSettings holds the user's draft preferences. ReviewThreshold is the confidence, from
// 0.5 to 1, below which an attribution waits in Review; a PUT that leaves it out keeps it.
type DraftSettings struct {
	AutoConfirmDrafts bool     `json:"auto_confirm_drafts"`
	ReviewThreshold   *float64 `json:"review_threshold"`
}
