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
}

// ListTasksResponse wraps a task list.
type ListTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}
