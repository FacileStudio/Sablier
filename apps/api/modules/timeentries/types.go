package timeentries

import "time"

type StartTimerRequest struct {
	ProjectID   int64  `json:"project_id"`
	Description string `json:"description"`
}

type CreateEntryRequest struct {
	ProjectID   int64     `json:"project_id"`
	Description string    `json:"description"`
	StartedAt   time.Time `json:"started_at"`
	StoppedAt   time.Time `json:"stopped_at"`
}

type UpdateEntryRequest struct {
	ProjectID   int64      `json:"project_id"`
	Description string     `json:"description"`
	StartedAt   time.Time  `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
}

type TimeEntryResponse struct {
	ID          int64      `json:"id"`
	ProjectID   int64      `json:"project_id"`
	UserID      int64      `json:"user_id"`
	Description string     `json:"description"`
	StartedAt   time.Time  `json:"started_at"`
	StoppedAt   *time.Time `json:"stopped_at"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   time.Time  `json:"updated_at"`
}

type ListEntriesResponse struct {
	Entries []TimeEntryResponse `json:"entries"`
}
