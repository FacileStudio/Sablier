package spaces

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "spaces",
	Description: "Multi-tenant space management with role-based membership.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/spaces",
			Summary:      "List spaces for current user",
			Auth:         "bearer token required",
			ResponseBody: ListSpacesResponse{},
		},
		{
			Method:       "POST",
			Path:         "/spaces",
			Summary:      "Create space",
			Auth:         "bearer token required",
			RequestBody:  CreateSpaceRequest{},
			ResponseBody: SpaceResponse{},
		},
		{
			Method:       "GET",
			Path:         "/spaces/{spaceId}",
			Summary:      "Get space",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}},
			ResponseBody: SpaceResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/spaces/{spaceId}",
			Summary:      "Update space",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}},
			RequestBody:  UpdateSpaceRequest{},
			ResponseBody: SpaceResponse{},
		},
		{
			Method:       "DELETE",
			Path:         "/spaces/{spaceId}",
			Summary:      "Delete space",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}},
			ResponseBody: DeleteSpaceResponse{},
		},
		{
			Method:       "POST",
			Path:         "/spaces/{spaceId}/leave",
			Summary:      "Leave space",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}},
			ResponseBody: LeaveSpaceResponse{},
		},
		{
			Method:       "GET",
			Path:         "/spaces/{spaceId}/members",
			Summary:      "List space members",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}},
			ResponseBody: ListMembersResponse{},
		},
		{
			Method:       "POST",
			Path:         "/spaces/{spaceId}/members",
			Summary:      "Add member to space",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}},
			RequestBody:  AddMemberRequest{},
			ResponseBody: MemberResponse{},
		},
		{
			Method:  "PUT",
			Path:    "/spaces/{spaceId}/members/{memberId}",
			Summary: "Update member role",
			Auth:    "bearer token required",
			PathParams: []documentation.Field{
				{Name: "spaceId", Type: "string", Description: "Space ID"},
				{Name: "memberId", Type: "string", Description: "Member ID"},
			},
			RequestBody:  UpdateMemberRoleRequest{},
			ResponseBody: MemberResponse{},
		},
		{
			Method:  "DELETE",
			Path:    "/spaces/{spaceId}/members/{memberId}",
			Summary: "Remove member from space",
			Auth:    "bearer token required",
			PathParams: []documentation.Field{
				{Name: "spaceId", Type: "string", Description: "Space ID"},
				{Name: "memberId", Type: "string", Description: "Member ID"},
			},
			ResponseBody: RemoveMemberResponse{},
		},
	},
}
