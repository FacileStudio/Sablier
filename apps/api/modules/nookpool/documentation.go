package nookpool

import documentation "api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "nookpool",
	Description: "Nook Pool connection management for syncing projects and tasks with other Facile apps.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/nook-pool/",
			Summary:      "Get pool settings",
			Description:  "Returns the current Nook Pool connection settings and status.",
			Auth:         "required",
			ResponseBody: "PoolSettingsResponse",
		},
		{
			Method:       "PUT",
			Path:         "/nook-pool/",
			Summary:      "Update pool settings",
			Description:  "Updates Nook Pool connection settings. If enabled with valid URL and secret, connects to the Pool.",
			Auth:         "required",
			RequestBody:  "UpdatePoolRequest",
			ResponseBody: "PoolSettingsResponse",
			Errors: []documentation.Error{
				{Status: 500, Code: "internal", Description: "Failed to save pool settings."},
			},
		},
	},
}
