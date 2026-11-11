package settings

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "settings",
	Description: "The app-wide webhook configuration, readable and writable by SABLIER_ADMINS only.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/settings",
			Summary:      "Return the webhook settings",
			Description:  "Returns the app-wide webhook URL and header name, and whether a secret is set. The secret itself is never returned.",
			Auth:         "SABLIER_ADMINS account required",
			ResponseBody: SettingsResponse{},
			Errors: []documentation.Error{
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 403, Code: "permission_denied", Description: "The account is not on SABLIER_ADMINS."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
		{
			Method:       "PUT",
			Path:         "/settings",
			Summary:      "Update the webhook settings",
			Description:  "Updates the app-wide webhook. An absent webhook_secret_value keeps the stored secret; an empty one clears it.",
			Auth:         "SABLIER_ADMINS account required",
			RequestBody:  UpdateRequest{},
			ResponseBody: SettingsResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "Invalid JSON body."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 403, Code: "permission_denied", Description: "The account is not on SABLIER_ADMINS."},
				{Status: 500, Code: "internal", Description: "Unexpected server error."},
			},
		},
	},
}
