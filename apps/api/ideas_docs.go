package main

import (
	documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"
	"github.com/FacileStudio/tronc/bulle"
)

var ideasErrors = []documentation.Error{
	{Status: 401, Code: "unauthenticated", Description: "Authorization header is missing or invalid."},
	{Status: 402, Code: "payment_required", Description: "The plan has no ai.access. The error carries an upgrade_url."},
	{Status: 404, Code: "not_found", Description: "The space is unknown or the user is not a member of it."},
}

var ideasSpaceParam = documentation.Field{Name: "space_id", Type: "string", Description: "Sablier space id. Absent means the personal space."}

var ideasDocumentation = documentation.Module{
	Name:        "ideas",
	Description: "AI suggestions for Sablier, generated and run by Bulle. Mounted only when BULLE_URL is set.",
	Routes: []documentation.Route{
		{
			Method:       "GET",
			Path:         "/ideas",
			Summary:      "List ideas",
			Description:  "Returns ai_access false and no ideas when the plan lacks ai.access, with the upgrade_url.",
			Auth:         "session required",
			QueryParams:  []documentation.Field{ideasSpaceParam},
			ResponseBody: bulle.IdeasPage{},
			Errors:       ideasErrors,
		},
		{
			Method:       "POST",
			Path:         "/ideas/generate",
			Summary:      "Generate ideas",
			Description:  "Starts a batch in Bulle scoped to Sablier.",
			Auth:         "session required",
			QueryParams:  []documentation.Field{ideasSpaceParam},
			ResponseBody: bulle.Batch{},
			Errors:       ideasErrors,
		},
		{
			Method:       "POST",
			Path:         "/ideas/{id}/run",
			Summary:      "Run an idea",
			Description:  "Runs the idea in the background in Bulle and returns the run id to poll.",
			Auth:         "session required",
			QueryParams:  []documentation.Field{ideasSpaceParam},
			ResponseBody: bulle.RunStart{},
			Errors:       ideasErrors,
		},
		{
			Method:  "POST",
			Path:    "/ideas/{id}/dismiss",
			Summary: "Dismiss an idea",
			Auth:    "session required",
			Errors:  ideasErrors,
		},
		{
			Method:       "GET",
			Path:         "/ideas/runs/{run}",
			Summary:      "Read a run",
			Description:  "State, summary, output links, pending approvals and the link that continues the run in Bulle.",
			Auth:         "session required",
			ResponseBody: bulle.Run{},
			Errors:       ideasErrors,
		},
		{
			Method:      "POST",
			Path:        "/ideas/runs/{run}/approvals/{call}",
			Summary:     "Answer an approval",
			Description: "Body {\"approve\": bool}. Approving needs ai.access.",
			Auth:        "session required",
			Errors:      ideasErrors,
		},
	},
}
