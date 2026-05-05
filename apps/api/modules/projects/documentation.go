package projects

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "projects",
	Description: "Project management routes.",
	Routes: []documentation.Route{
		{Method: "GET", Path: "/projects", Summary: "List projects", Auth: "bearer token required"},
		{Method: "POST", Path: "/projects", Summary: "Create project", Auth: "bearer token required"},
		{Method: "GET", Path: "/projects/{id}", Summary: "Get project", Auth: "bearer token required"},
		{Method: "PUT", Path: "/projects/{id}", Summary: "Update project", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/projects/{id}", Summary: "Delete project", Auth: "bearer token required"},
	},
}
