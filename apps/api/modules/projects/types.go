package projects

import "time"

type CreateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type UpdateProjectRequest struct {
	Name        string `json:"name"`
	Description string `json:"description"`
}

type CreateTaskRequest struct {
	Name string `json:"name"`
}

type UpdateTaskRequest struct {
	Name string `json:"name"`
}

type ProjectResponse struct {
	ID          int64     `json:"id"`
	Name        string    `json:"name"`
	Description string    `json:"description"`
	OwnerID     int64     `json:"owner_id"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

type ListProjectsResponse struct {
	Projects []ProjectResponse `json:"projects"`
}

type TaskResponse struct {
	ID        int64     `json:"id"`
	ProjectID int64     `json:"project_id"`
	Name      string    `json:"name"`
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

type ListTasksResponse struct {
	Tasks []TaskResponse `json:"tasks"`
}
