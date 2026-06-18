package projects

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "projects",
	Description: "Project management routes.",
	Routes: []documentation.Route{
		{Method: "GET", Path: "/api/projects", Summary: "List projects", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/projects", Summary: "Create project", Auth: "bearer token required"},
		{Method: "GET", Path: "/api/projects/{id}", Summary: "Get project", Auth: "bearer token required"},
		{Method: "PUT", Path: "/api/projects/{id}", Summary: "Update project", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/api/projects/{id}", Summary: "Delete project", Auth: "bearer token required"},
		{Method: "GET", Path: "/api/projects/{id}/tasks", Summary: "List project tasks", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/projects/{id}/tasks", Summary: "Create project task", Auth: "bearer token required"},
	},
}
