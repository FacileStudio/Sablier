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
			QueryParams:  []documentation.Field{{Name: "space_id", Type: "string", Description: "Absent: personal rows only. all: personal rows plus every space you belong to. A space id: that space; 404 when you are not a member."}},
			ResponseBody: ListProjectsResponse{},
			Errors:       []documentation.Error{{Status: 404, Code: "not_found", Description: "space_id names a space you are not in."}},
		},
		{
			Method:  "GET",
			Path:    "/tasks/search",
			Summary: "Search tasks",
			Description: "Searches the tasks you can see by task or project name. Ranks the tasks you last worked on first, " +
				"then open Opus tasks, then task-name matches, then project-name matches. Choosing a result chooses its " +
				"project and space too.",
			Auth: "bearer token required",
			QueryParams: []documentation.Field{
				{Name: "q", Type: "string", Description: "Case-insensitive substring, at most 200 characters; empty lists everything"},
				{Name: "limit", Type: "integer", Description: "1 to 50, default 20"},
			},
			ResponseBody: TaskSearchResponse{},
		},
		{
			Method:       "POST",
			Path:         "/projects",
			Summary:      "Create project",
			Description:  "A space_id must name a space you belong to, otherwise the answer is 404.",
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
