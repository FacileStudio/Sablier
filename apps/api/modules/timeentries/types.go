package timeentries

import "time"

type StartTimerRequest struct {
	ProjectID int64 `json:"project_id"`
	TaskID    int64 `json:"task_id"`
}

type CreateEntryRequest struct {
	ProjectID int64     `json:"project_id"`
	TaskID    int64     `json:"task_id"`
	StartedAt time.Time `json:"started_at"`
	StoppedAt time.Time `json:"stopped_at"`
}

type UpdateEntryRequest struct {
	ProjectID int64      `json:"project_id"`
	TaskID    int64      `json:"task_id"`
	StartedAt time.Time  `json:"started_at"`
	StoppedAt *time.Time `json:"stopped_at"`
}

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
	StartedAt      time.Time  `json:"started_at"`
	StoppedAt      *time.Time `json:"stopped_at"`
	CreatedAt      time.Time  `json:"created_at"`
	UpdatedAt      time.Time  `json:"updated_at"`
}

type ListEntriesResponse struct {
	Entries []TimeEntryResponse `json:"entries"`
}
