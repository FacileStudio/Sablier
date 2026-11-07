package review

import (
	documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
)

// Documentation describes the review routes for GET /docs.
var Documentation = documentation.Module{
	Name:        "review",
	Description: "Entries whose attribution was not sure enough to book silently, and drafts waiting to be confirmed.",
	Routes: []documentation.Route{
		{
			Method:  "GET",
			Path:    "/review",
			Summary: "List what waits for review",
			Description: "Returns the entries flagged needs_review that belong to the caller or to the caller's agents, " +
				"newest first, and the caller's drafts over the last 48 hours with their suggestion.",
			Auth:         "bearer token required",
			ResponseBody: InboxResponse{},
		},
		{
			Method:  "POST",
			Path:    "/review/{entry_id}",
			Summary: "Settle a flagged entry",
			Description: "With a task_id, moves the entry to that task, its project and its space. With an empty body " +
				"or {}, accepts it where it is. Either way it clears needs_review and sets attribution_source to " +
				"manual, so a replayed agent session never moves it again.",
			Auth:         "bearer token required",
			RequestBody:  ResolveRequest{},
			ResponseBody: timeentries.TimeEntryResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "task_id is not a task you can see."},
				{Status: 404, Code: "not_found", Description: "No such entry of yours or of your agents."},
			},
		},
	},
}
