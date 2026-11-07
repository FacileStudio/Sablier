package timeentries

import (
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

// StartTimerRequest is the body for starting the timer. SpaceID is accepted for older
// clients and ignored: the entry takes its project's space.
type StartTimerRequest struct {
	ProjectID int64   `json:"project_id"`
	TaskID    int64   `json:"task_id"`
	SpaceID   *string `json:"space_id"`
}

// CreateEntryRequest is the body for creating a manual time entry. SpaceID is ignored like
// StartTimerRequest's. Attribution is set only by the modules that book on a user's behalf.
type CreateEntryRequest struct {
	ProjectID   int64               `json:"project_id"`
	TaskID      int64               `json:"task_id"`
	SpaceID     *string             `json:"space_id"`
	StartedAt   time.Time           `json:"started_at"`
	StoppedAt   time.Time           `json:"stopped_at"`
	Attribution schemas.Attribution `json:"-"`
}

// UpdateEntryRequest is the body for updating a time entry.
type UpdateEntryRequest struct {
	ProjectID int64      `json:"project_id"`
	TaskID    int64      `json:"task_id"`
	StartedAt time.Time  `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at"`
}

// TimeEntryResponse is the serialized shape of a time entry.
type TimeEntryResponse struct {
	ID             int64      `json:"id"`
	ProjectID      int64      `json:"project_id"`
	TaskID         int64      `json:"task_id"`
	TaskName       string     `json:"task_name"`
	UserID         int64      `json:"user_id"`
	UserEmail      string     `json:"user_email,omitempty"`
	UserName       string     `json:"user_name,omitempty"`
	UserColor      string     `json:"user_color,omitempty"`
	UserAvatarURL  string     `json:"user_avatar_url,omitempty"`
	UserKind       string     `json:"user_kind,omitempty"`
	ProjectName    string     `json:"project_name,omitempty"`
	SpaceID        *string    `json:"space_id"`
	StartedAt      time.Time  `json:"started_at"`
	StoppedAt      *time.Time `json:"stopped_at"`
	PausedAt       *time.Time `json:"paused_at"`
	PausedDuration int64      `json:"paused_duration_ms"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
	AttributionResponse
}

// AttributionResponse says why an entry landed on its task. Source is one of explicit,
// opus, name, running, ai, fallback or manual, and null on an entry a person booked.
type AttributionResponse struct {
	Source      *string  `json:"attribution_source"`
	Confidence  *float64 `json:"attribution_confidence"`
	NeedsReview bool     `json:"needs_review"`
}

// ListEntriesResponse wraps a time-entry list.
type ListEntriesResponse struct {
	Entries []TimeEntryResponse `json:"entries"`
}

// RunningEntryResponse wraps the currently running time entry if any.
type RunningEntryResponse struct {
	Entry *TimeEntryResponse `json:"entry"`
}

// DeleteEntryResponse reports whether a time entry was deleted.
type DeleteEntryResponse struct {
	Deleted bool `json:"deleted"`
}
