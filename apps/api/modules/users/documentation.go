package users

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "users",
	Description: "User listing plus current-user retrieval and update routes.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/users",
			Summary:      "List users",
			Description:  "Returns the accounts the caller may see: itself, its own agents, and everyone who shares a space with it.",
			Auth:         "bearer token required",
			ResponseBody: ListResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:      "GET",
			Path:        "/users/{id}",
			Summary:     "Get a user",
			Description: "Returns one account the caller may see (itself, its agents, a space peer). Any other id answers 404.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "id", Type: "string", Description: "User identifier."},
			},
			ResponseBody: PeerResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 404, Code: "not_found", Description: "No user with that identifier that the caller may see."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "GET",
			Path:         "/users/me",
			Summary:      "Return the current user",
			Description:  "Returns the authenticated user with profile metadata. is_admin is true when the account is on SABLIER_ADMINS.",
			Auth:         "bearer token required",
			ResponseBody: MeResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "PATCH",
			Path:         "/users/me",
			Summary:      "Update the current user",
			Description:  "Updates the authenticated user's name, pastel color, rate and workday length. The email and the password are Registre's: there are none to change here.",
			Auth:         "bearer token required",
			RequestBody:  UpdateRequest{},
			ResponseBody: MeResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body or invalid update input."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 404, Code: "not_found", Description: "The authenticated user no longer exists."},
				{Status: 409, Code: "already_exists", Description: "A user with the same email already exists."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "GET",
			Path:         "/users/me/api-token",
			Summary:      "Report API token status",
			Description:  "Says whether the authenticated user has an API token, without revealing it.",
			Auth:         "bearer token required",
			ResponseBody: ApiTokenStatusResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "POST",
			Path:         "/users/me/api-token",
			Summary:      "Issue an API token",
			Description:  "Creates a personal API token, returned once and never shown again.",
			Auth:         "bearer token required",
			RequestBody:  CreateApiTokenRequest{},
			ResponseBody: ApiTokenResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/users/me/api-token",
			Summary:      "Revoke the API token",
			Description:  "Deletes the authenticated user's API token.",
			Auth:         "bearer token required",
			ResponseBody: DeleteTokenResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
	},
}
