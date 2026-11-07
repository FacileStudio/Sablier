package agents

import documentation "github.com/FacileStudio/Sablier/apps/api/internal/documentation"

var Documentation = documentation.Module{
	Name:        "agents",
	Description: "AI agents tracked as their own users, owned by the human they work for.",
	Routes: []documentation.Route{
		{
			Method:  "POST",
			Path:    "/agents/sessions",
			Summary: "Record an agent session",
			Description: "Upserts one finished span of agent work by facile_id, booked to the caller's agent " +
				"named by agent_key, which is created on first use. The attribution resolver tries task_id or project_id, " +
				"then opus_task_ids (opus_updated_task_id first) or opus_project_id, then the project name inside the " +
				"space named by registre_space_id or opus_workspace, then the caller's entry running at started_at, " +
				"then an AI pick among that space's tasks, then the shared Agent work project. The entry carries " +
				"attribution_source, attribution_confidence and needs_review; an entry reassigned in Review keeps " +
				"its task when the session is replayed.",
			Auth:         "bearer token required (write scope for delegated tokens)",
			RequestBody:  RecordSessionRequest{},
			ResponseBody: RecordSessionResponse{},
			Errors: []documentation.Error{
				{Status: 400, Code: "invalid_argument", Description: "A field is missing or malformed."},
				{Status: 409, Code: "already_exists", Description: "facile_id belongs to another user's entry."},
			},
		},
		{
			Method:       "GET",
			Path:         "/agents",
			Summary:      "List my agents",
			Auth:         "bearer token required",
			ResponseBody: []AgentResponse{},
		},
		{
			Method:  "GET",
			Path:    "/agents/paired",
			Summary: "Agent, cowork and paired time",
			Auth:    "bearer token required",
			QueryParams: []documentation.Field{
				{Name: "from", Type: "string", Description: "Window start, RFC3339"},
				{Name: "to", Type: "string", Description: "Window end, RFC3339"},
			},
			ResponseBody: []PairedResponse{},
		},
	},
}
