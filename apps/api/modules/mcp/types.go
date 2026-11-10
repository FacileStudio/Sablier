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
