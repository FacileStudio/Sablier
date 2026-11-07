package heartbeats

import (
	documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
)

// Documentation registers the heartbeat and draft routes in the API reference.
var Documentation = documentation.Module{
	Name:        "heartbeats",
	Description: "Activity heartbeats from tools, and the draft time entries computed from them.",
	Routes: []documentation.Route{
		{
			Method:  "POST",
			Path:    "/heartbeats",
			Summary: "Record heartbeats",
			Description: "Stores up to 500 heartbeats. Heartbeats older than 7 days or more than 5 minutes ahead are skipped; exact repeats are ignored. " +
				"registre_space_id, opus_task_id and title are optional attribution hints.",
			Auth:         "bearer token required",
			RequestBody:  IngestRequest{},
			ResponseBody: IngestResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "The body or query failed validation."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
			},
		},
		{
			Method:  "GET",
			Path:    "/heartbeats/drafts",
			Summary: "List drafts",
			Description: "Merges heartbeats closer than 15 minutes, removes time already covered by entries and dismissed ranges, " +
				"and resolves the project. Each draft carries a suggestion {task_id, project_id, space_id, source, " +
				"confidence, needs_review} from the attribution resolver, or null when nothing matched. Defaults to the last 48 hours.",
			Auth: "bearer token required",
			QueryParams: []documentation.Field{
				{Name: "from", Type: "string", Description: "RFC3339 start of the window"},
				{Name: "to", Type: "string", Description: "RFC3339 end of the window"},
			},
			ResponseBody: []Draft{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "The body or query failed validation."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
			},
		},
		{
			Method:       "POST",
			Path:         "/heartbeats/drafts/confirm",
			Summary:      "Confirm a draft",
			Description:  "Creates a time entry for the range. Without task_id the entry goes to the project's Auto-tracked task.",
			Auth:         "bearer token required",
			RequestBody:  ConfirmRequest{},
			ResponseBody: timeentries.TimeEntryResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "The body or query failed validation."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
				{Status: 409, Code: "already_exists", Description: "The range overlaps an existing time entry."},
			},
		},
		{
			Method:       "POST",
			Path:         "/heartbeats/drafts/dismiss",
			Summary:      "Dismiss a draft",
			Description:  "Stores the range so no draft is suggested inside it again.",
			Auth:         "bearer token required",
			RequestBody:  DismissRequest{},
			ResponseBody: DismissResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "The body or query failed validation."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
			},
		},
		{
			Method:       "GET",
			Path:         "/heartbeats/settings",
			Summary:      "Read draft settings",
			Auth:         "bearer token required",
			ResponseBody: DraftSettings{},
			Errors:       []documentation.Error{{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."}},
		},
		{
			Method:  "PUT",
			Path:    "/heartbeats/settings",
			Summary: "Update draft settings",
			Description: "auto_confirm_drafts lets the hourly worker confirm drafts older than 48 hours whose suggestion " +
				"does not need review. review_threshold (0.5 to 1, default 0.8) is the confidence below which an " +
				"attribution waits in Review; leaving it out keeps the stored value.",
			Auth:         "bearer token required",
			RequestBody:  DraftSettings{},
			ResponseBody: DraftSettings{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "The body or query failed validation."},
				{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
			},
		},
	},
}
