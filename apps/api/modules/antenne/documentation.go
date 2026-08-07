package antenne

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "antenne",
	Description: "Antenne connection management for syncing projects and tasks with other Facile apps.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/antenne/",
			Summary:      "Get pool settings",
			Description:  "Returns the current Antenne connection settings and status.",
			Auth:         "required",
			ResponseBody: "PoolSettingsResponse",
		},
		{
			Method:       "POST",
			Path:         "/antenne/sync",
			Summary:      "Trigger initial sync",
			Description:  "Syncs all existing projects and tasks to the Pool. Safe to run multiple times.",
			Auth:         "required",
			ResponseBody: "SyncResult",
		},
		{
			Method:       "PUT",
			Path:         "/antenne/",
			Summary:      "Update pool settings",
			Description:  "Updates Antenne connection settings. If enabled with valid URL and secret, connects to the Pool.",
			Auth:         "required",
			RequestBody:  "UpdatePoolRequest",
			ResponseBody: "PoolSettingsResponse",
			Errors: []documentation.Error{
				{Status: 500, Code: "internal", Description: "Failed to save pool settings."},
			},
		},
		{
			Method:       "GET",
			Path:         "/antenne/events",
			Summary:      "Get pool event subscriptions",
			Description:  "Returns the event types this app publishes to and consumes from the Pool.",
			Auth:         "required",
			ResponseBody: "PoolEventsResponse",
		},
		{
			Method:       "PUT",
			Path:         "/antenne/events",
			Summary:      "Update pool event subscriptions",
			Description:  "Replaces the set of event types exchanged with the Pool.",
			Auth:         "required",
			RequestBody:  "UpdatePoolEventsRequest",
			ResponseBody: "PoolEventsResponse",
			Errors: []documentation.Error{
				{Status: 500, Code: "internal", Description: "Failed to save pool event subscriptions."},
			},
		},
	},
}
