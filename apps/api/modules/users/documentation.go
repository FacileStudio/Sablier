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
			Description:  "Returns all authenticated users with profile metadata.",
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
			Description: "Returns one user with profile metadata.",
			Auth:        "bearer token required",
			PathParams: []documentation.Field{
				{Name: "id", Type: "string", Description: "User identifier."},
			},
			ResponseBody: MeResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 404, Code: "not_found", Description: "No user with that identifier."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "GET",
			Path:         "/users/me",
			Summary:      "Return the current user",
			Description:  "Returns the authenticated user with profile metadata.",
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
			Description:  "Updates the authenticated user's name, email, password, and/or pastel color. Replacing an existing password requires current_password; without it only an account that has none can be given a first one.",
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
			Method:       "POST",
			Path:         "/users/me/avatar",
			Summary:      "Upload the current user's avatar",
			Description:  "Stores a new avatar file for the authenticated user and returns the updated profile.",
			Auth:         "bearer token required",
			ResponseBody: MeResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Missing file or unsupported image type."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 404, Code: "not_found", Description: "The authenticated user no longer exists."},
				{Status: 413, Code: "resource_exhausted", Description: "Avatar file is too large."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "DELETE",
			Path:         "/users/me/avatar",
			Summary:      "Remove the current user's avatar",
			Description:  "Deletes the stored avatar file and returns the updated profile.",
			Auth:         "bearer token required",
			ResponseBody: MeResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 404, Code: "not_found", Description: "The authenticated user no longer exists."},
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
