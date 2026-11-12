package mcp

// statusInput is sablier_timer_status's input.
type statusInput struct {
	Timezone string `json:"timezone,omitempty" jsonschema:"IANA timezone the times come back in, such as Europe/Paris. Default UTC"`
}

// taskRef names the task time goes to, by id or by name.
type taskRef struct {
	ProjectID int64  `json:"project_id" jsonschema:"the project, from sablier_list_projects"`
	TaskID    int64  `json:"task_id,omitempty" jsonschema:"a task of that project, from sablier_list_projects"`
	Task      string `json:"task,omitempty" jsonschema:"what the time is for, as a task name; used when task_id is not given and created in the project when new"`
}

// startInput is sablier_start_timer's input.
type startInput struct {
	taskRef
	Timezone string `json:"timezone,omitempty" jsonschema:"IANA timezone the times come back in. Default UTC"`
}

// addInput is sablier_add_entry's input.
type addInput struct {
	taskRef
	StartedAt       string `json:"started_at" jsonschema:"when the work started: 2026-11-10T09:30 in timezone, or an RFC 3339 time"`
	StoppedAt       string `json:"stopped_at,omitempty" jsonschema:"when it stopped, same forms; or pass duration_minutes"`
	DurationMinutes int    `json:"duration_minutes,omitempty" jsonschema:"how long it lasted, used when stopped_at is not given"`
	Timezone        string `json:"timezone,omitempty" jsonschema:"IANA timezone for times without an offset. Default UTC"`
}

// updateInput is sablier_update_entry's input: only the fields given change.
type updateInput struct {
	EntryID   int64  `json:"entry_id" jsonschema:"one of your own entries, from sablier_list_entries"`
	ProjectID int64  `json:"project_id,omitempty" jsonschema:"move the entry to this project; then pass task_id or task too"`
	TaskID    int64  `json:"task_id,omitempty" jsonschema:"book the entry on this task"`
	Task      string `json:"task,omitempty" jsonschema:"book the entry on the task of this name, created when new"`
	StartedAt string `json:"started_at,omitempty" jsonschema:"new start: 2026-11-10T09:30 in timezone, or an RFC 3339 time"`
	StoppedAt string `json:"stopped_at,omitempty" jsonschema:"new end, same forms"`
	Timezone  string `json:"timezone,omitempty" jsonschema:"IANA timezone for times without an offset. Default UTC"`
}

// deleteInput is sablier_delete_entry's input.
type deleteInput struct {
	EntryID int64 `json:"entry_id" jsonschema:"one of your own entries, from sablier_list_entries"`
}

// scopeInput selects whose time a report covers and where.
type scopeInput struct {
	Space     string `json:"space,omitempty" jsonschema:"all (default) for personal plus every space you belong to, personal, or a space id"`
	ProjectID int64  `json:"project_id,omitempty" jsonschema:"only this project"`
	Everyone  bool   `json:"everyone,omitempty" jsonschema:"include teammates' and agents' time in shared spaces; default is your own time only"`
	UserID    int64  `json:"user_id,omitempty" jsonschema:"only this person's time, a user id from an entry's detail"`
	Timezone  string `json:"timezone,omitempty" jsonschema:"IANA timezone for dates and the times returned, such as Europe/Paris. Default UTC"`
}

// entriesInput is sablier_list_entries's input.
type entriesInput struct {
	scopeInput
	From           string `json:"from,omitempty" jsonschema:"start of the range: a date (2026-11-03), a local time or an RFC 3339 time. Default 7 days before to"`
	To             string `json:"to,omitempty" jsonschema:"end of the range, same forms; a date includes that whole day. Default now"`
	Cursor         string `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page"`
	ResponseFormat string `json:"response_format,omitempty" jsonschema:"concise (default) or detailed, which adds space, attribution and user ids"`
}

// summaryInput is sablier_weekly_summary's input.
type summaryInput struct {
	scopeInput
	WeekOf string `json:"week_of,omitempty" jsonschema:"any date in the week, such as 2026-11-12; weeks start on Monday. Default this week"`
}

// projectsInput is sablier_list_projects's input.
type projectsInput struct {
	Space       string `json:"space,omitempty" jsonschema:"all (default), personal, or a space id"`
	Query       string `json:"query,omitempty" jsonschema:"keep projects or tasks whose name contains this, case ignored"`
	ProjectID   int64  `json:"project_id,omitempty" jsonschema:"only this project, with up to 500 of its tasks"`
	IncludeDone bool   `json:"include_done,omitempty" jsonschema:"also list tasks whose status is done"`
	Cursor      string `json:"cursor,omitempty" jsonschema:"next_cursor from the previous page"`
}

// agentSessionInput is sablier_record_agent_session's input, the body of POST
// /api/agents/sessions.
type agentSessionInput struct {
	AgentKey      string `json:"agent_key" jsonschema:"the agent's stable key, 1 to 40 of a-z, 0-9 and -, such as bulle; Sablier keeps one agent user per key for the user"`
	Name          string `json:"name" jsonschema:"the agent's display name, at most 80 characters"`
	AvatarURL     string `json:"avatar_url,omitempty" jsonschema:"an https avatar URL; default is a blob drawn from the name"`
	FacileID      string `json:"facile_id" jsonschema:"the session's own id, at most 200 characters; booking the same id again replaces that entry"`
	StartedAt     string `json:"started_at" jsonschema:"when the session started, an RFC 3339 time"`
	StoppedAt     string `json:"stopped_at" jsonschema:"when it stopped, an RFC 3339 time at most 24 hours after started_at and not in the future"`
	ActiveSeconds *int64 `json:"active_seconds,omitempty" jsonschema:"seconds the agent actually worked; the rest of the span is paused time. Default the whole span"`
	Cowork        bool   `json:"cowork,omitempty" jsonschema:"true when the user worked alongside the agent"`
	Description   string `json:"description,omitempty" jsonschema:"what the session did, at most 500 characters"`
	sessionHints
}

// sessionHints are the optional facts that tell Sablier where an agent session belongs.
type sessionHints struct {
	Title           string   `json:"title,omitempty" jsonschema:"a short title of the work, used to find its task"`
	RegistreSpaceID string   `json:"registre_space_id,omitempty" jsonschema:"the Registre space the work belongs to, searched first"`
	OpusWorkspace   string   `json:"opus_workspace,omitempty" jsonschema:"the Opus workspace the work belongs to"`
	OpusTaskIDs     []string `json:"opus_task_ids,omitempty" jsonschema:"facile ids of the Opus tasks the session touched, most recent last, at most 50"`
}

// heartbeatInput is sablier_record_heartbeat's input, one heartbeat of POST
// /api/heartbeats.
type heartbeatInput struct {
	At              string `json:"at" jsonschema:"when the user was active, an RFC 3339 time within the last 7 days"`
	Source          string `json:"source" jsonschema:"what saw the activity: claude-code, editor, desktop, git, bulle or cli"`
	Entity          string `json:"entity,omitempty" jsonschema:"what the user worked on, such as a file path or bulle:session/<id>, at most 1024 characters"`
	Project         string `json:"project,omitempty" jsonschema:"a project name hint, at most 200 characters"`
	RegistreSpaceID string `json:"registre_space_id,omitempty" jsonschema:"the Registre space the work belongs to"`
	OpusTaskID      string `json:"opus_task_id,omitempty" jsonschema:"facile id of the Opus task being worked on"`
	Title           string `json:"title,omitempty" jsonschema:"a short title of the work"`
}
