package spaces

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name: "spaces",
	Description: "Read-only shared spaces and their members. Registre owns spaces: create, rename, " +
		"invite, role changes, removal and leaving happen at https://sso.facile.studio/spaces.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/spaces",
			Summary:      "List spaces for current user",
			Auth:         "bearer token required",
			ResponseBody: ListSpacesResponse{},
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
			Method:       "GET",
			Path:         "/spaces/{spaceId}/members",
			Summary:      "List space members",
			Auth:         "bearer token required",
			PathParams:   []documentation.Field{{Name: "spaceId", Type: "string", Description: "Space ID"}},
			ResponseBody: ListMembersResponse{},
		},
		{
			Method:  "POST",
			Path:    "/registre/sync",
			Summary: "Registre's nudge: schedules a pull of the space snapshot and answers 202",
			Auth:    "none; the pull itself authenticates with the app's OIDC client credentials",
		},
	},
}
