package antenne

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "antenne",
	Description: "Antenne connection management for syncing projects and tasks with other Facile apps. App-wide, so every route is open to SABLIER_ADMINS only (403 for anyone else).",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/antenne/",
			Summary:      "Get pool settings",
			Description:  "Returns the Antenne instance URL, whether a shared secret is set, and the connection status. The secret itself is never returned.",
			Auth:         "SABLIER_ADMINS account required",
			ResponseBody: PoolSettingsResponse{},
		},
		{
			Method:       "POST",
			Path:         "/antenne/sync",
			Summary:      "Reconcile with the pool",
			Description:  "Queues every synced project, its tasks and their time entries as created. Safe to run multiple times; it also runs once at boot.",
			Auth:         "SABLIER_ADMINS account required",
			ResponseBody: SyncResult{},
		},
		{
			Method:       "PUT",
			Path:         "/antenne/",
			Summary:      "Update pool settings",
			Description:  "Updates Antenne connection settings. An absent antenne_secret keeps the stored secret. If enabled with a URL and a secret, connects to the Pool.",
			Auth:         "SABLIER_ADMINS account required",
			RequestBody:  UpdatePoolRequest{},
			ResponseBody: PoolSettingsResponse{},
			Errors: []documentation.Error{
				{Status: 500, Code: "internal", Description: "Failed to save pool settings."},
			},
		},
		{
			Method:       "GET",
			Path:         "/antenne/events",
			Summary:      "Get pool event subscriptions",
			Description:  "Returns the event types this app publishes to and consumes from the Pool.",
			Auth:         "SABLIER_ADMINS account required",
			ResponseBody: PoolEventsResponse{},
		},
		{
			Method:       "PUT",
			Path:         "/antenne/events",
			Summary:      "Update pool event subscriptions",
			Description:  "Replaces the set of event types exchanged with the Pool.",
			Auth:         "SABLIER_ADMINS account required",
			RequestBody:  UpdatePoolEventsRequest{},
			ResponseBody: PoolEventsResponse{},
			Errors: []documentation.Error{
				{Status: 500, Code: "internal", Description: "Failed to save pool event subscriptions."},
			},
		},
	},
}
