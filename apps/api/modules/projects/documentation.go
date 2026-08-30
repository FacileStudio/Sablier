package projects

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "projects",
	Description: "Project management routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/projects",
			Summary:      "List projects",
			Auth:         "bearer token required",
			QueryParams:  []documentation.Field{{Name: "space_id", Type: "string", Description: "Filter projects by space ID"}},
			ResponseBody: ListProjectsResponse{},
		},
		{
			Method:       "POST",
			Path:         "/projects",
			Summary:      "Create project",
			Auth:         "bearer token required",
			RequestBody:  CreateProjectRequest{},
			ResponseBody: ProjectResponse{},
		},
		{
			Method:       "GET",
			Path:         "/projects/{id}",
			Summary:      "Get project",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "id", Type: "integer", Description: "Project ID"}},
			ResponseBody: ProjectResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/projects/{id}",
			Summary:      "Update project",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "id", Type: "integer", Description: "Project ID"}},
			RequestBody:  UpdateProjectRequest{},
			ResponseBody: ProjectResponse{},
		},
		{
			Method:       "DELETE",
			Path:         "/projects/{id}",
			Summary:      "Delete project",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "id", Type: "integer", Description: "Project ID"}},
			ResponseBody: DeleteProjectResponse{},
		},
		{
			Method:       "GET",
			Path:         "/projects/{id}/tasks",
			Summary:      "List project tasks",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "id", Type: "integer", Description: "Project ID"}},
			ResponseBody: ListTasksResponse{},
		},
		{
			Method:       "POST",
			Path:         "/projects/{id}/tasks",
			Summary:      "Create project task",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "id", Type: "integer", Description: "Project ID"}},
			RequestBody:  CreateTaskRequest{},
			ResponseBody: TaskResponse{},
		},
		{
			Method:  "PUT",
			Path:    "/projects/{id}/tasks/{taskId}",
			Summary: "Update project task",
			Auth:    "bearer token required",
			PathParams: []documentation.Field{
				{Name: "id", Type: "integer", Description: "Project ID"},
				{Name: "taskId", Type: "integer", Description: "Task ID"},
			},
			RequestBody:  UpdateTaskRequest{},
			ResponseBody: TaskResponse{},
		},
		{
			Method:  "DELETE",
			Path:    "/projects/{id}/tasks/{taskId}",
			Summary: "Delete project task",
			Auth:    "bearer token required",
			PathParams: []documentation.Field{
				{Name: "id", Type: "integer", Description: "Project ID"},
				{Name: "taskId", Type: "integer", Description: "Task ID"},
			},
			ResponseBody: DeleteTaskResponse{},
		},
	},
}
