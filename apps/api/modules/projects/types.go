package projects

import "time"

// CreateProjectRequest is the body for creating a project.
type CreateProjectRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
	SpaceID     *string `json:"space_id"`
}

// UpdateProjectRequest is the body for updating a project.
type UpdateProjectRequest struct {
	Name        string  `json:"name"`
	Description string  `json:"description"`
	Icon        *string `json:"icon"`
}

// CreateTaskRequest is the body for creating a task.
type CreateTaskRequest struct {
	Name string `json:"name"`
}

// UpdateTaskRequest is the body for updating a task.
type UpdateTaskRequest struct {
	Name   string  `json:"name"`
	Status *string `json:"status"`
}

// ProjectResponse is the serialized shape of a project.
type ProjectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	Icon        *string   `json:"icon"`
	OwnerID     int64     `json:"owner_id"`
	SpaceID     *string   `json:"space_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// ListProjectsResponse wraps a project list.
type ListProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

// TaskResponse is the serialized shape of a task.
type TaskResponse struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	Status    string    `json:"status"`
	ActorID   *int64    `json:"actor_id"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
	TaskSyncResponse
}

// TaskSyncResponse is what Opus owns on a synced task; Sablier shows it read-only.
type TaskSyncResponse struct {
	Description string  `json:"description"`
	DueDate     *string `json:"due_date"`
	Priority    string  `json:"priority"`
}

// ListTasksResponse wraps a task list.
type ListTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}

// DeleteProjectResponse reports whether a project was deleted.
type DeleteProjectResponse struct {
	Deleted bool `json:"deleted"`
}

// DeleteTaskResponse reports whether a task was deleted and unlinked sessions.
type DeleteTaskResponse struct {
	Deleted          bool  `json:"deleted"`
	SessionsUnlinked int64 `json:"sessions_unlinked"`
}

// TaskSearchResponse is GET /tasks/search.
type TaskSearchResponse struct {
	Results []TaskSearchResult `json:"results"`
}

// TaskSearchResult is one task with its project and space. Choosing it chooses all three.
// LastUsedAt is the caller's latest entry on the task, and Opus marks a task Opus knows.
type TaskSearchResult struct {
	Task       TaskRef    `json:"task"`
	Project    ProjectRef `json:"project"`
	Space      *SpaceRef  `json:"space"`
	LastUsedAt *time.Time `json:"last_used_at"`
	Opus       bool       `json:"opus"`
}

// TaskRef names a task in a search result.
type TaskRef struct {
	ID       int64   `json:"id"`
	Name     string  `json:"name"`
	FacileID *string `json:"facile_id"`
}

// ProjectRef names a project in a search result.
type ProjectRef struct {
	ID   int64   `json:"id"`
	Name string  `json:"name"`
	Icon *string `json:"icon"`
}

// SpaceRef names a space in a search result; personal tasks carry none.
type SpaceRef struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}
