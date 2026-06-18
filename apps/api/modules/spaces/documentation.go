package spaces

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "spaces",
	Description: "Multi-tenant space management with role-based membership.",
	Routes: []documentation.Route{
		{Method: "GET", Path: "/spaces", Summary: "List spaces for current user", Auth: "bearer token required"},
		{Method: "POST", Path: "/spaces", Summary: "Create space", Auth: "bearer token required"},
		{Method: "GET", Path: "/spaces/{spaceId}", Summary: "Get space", Auth: "bearer token required"},
		{Method: "PUT", Path: "/spaces/{spaceId}", Summary: "Update space", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/spaces/{spaceId}", Summary: "Delete space", Auth: "bearer token required"},
		{Method: "POST", Path: "/spaces/{spaceId}/leave", Summary: "Leave space", Auth: "bearer token required"},
		{Method: "GET", Path: "/spaces/{spaceId}/members", Summary: "List space members", Auth: "bearer token required"},
		{Method: "POST", Path: "/spaces/{spaceId}/members", Summary: "Add member to space", Auth: "bearer token required"},
		{Method: "PUT", Path: "/spaces/{spaceId}/members/{memberId}", Summary: "Update member role", Auth: "bearer token required"},
		{Method: "DELETE", Path: "/spaces/{spaceId}/members/{memberId}", Summary: "Remove member from space", Auth: "bearer token required"},
	},
}
