package spaces

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "spaces",
	Description: "Multi-tenant space management with role-based membership.",
	Routes: []documentation.Route{
		{Method: "GET", Path: "/api/spaces", Summary: "List spaces for current user", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/spaces", Summary: "Create space", Auth: "bearer token required"},
		{Method: "GET", Path: "/api/spaces/{spaceId}", Summary: "Get space", Auth: "bearer token required"},
		{Method: "PUT", Path: "/api/spaces/{spaceId}", Summary: "Update space", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/api/spaces/{spaceId}", Summary: "Delete space", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/spaces/{spaceId}/leave", Summary: "Leave space", Auth: "bearer token required"},
		{Method: "GET", Path: "/api/spaces/{spaceId}/members", Summary: "List space members", Auth: "bearer token required"},
		{Method: "POST", Path: "/api/spaces/{spaceId}/members", Summary: "Add member to space", Auth: "bearer token required"},
		{Method: "PUT", Path: "/api/spaces/{spaceId}/members/{memberId}", Summary: "Update member role", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/api/spaces/{spaceId}/members/{memberId}", Summary: "Remove member from space", Auth: "bearer token required"},
	},
}
