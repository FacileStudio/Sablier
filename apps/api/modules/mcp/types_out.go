package mcp

// entryOut is one time entry. State is running, paused, or empty once stopped.
type entryOut struct {
	ID          int64        `json:"id"`
	ProjectID   int64        `json:"project_id"`
	Project     string       `json:"project"`
	TaskID      int64        `json:"task_id"`
	Task        string       `json:"task"`
	StartedAt   string       `json:"started_at"`
	StoppedAt   string       `json:"stopped_at,omitempty"`
	Minutes     int          `json:"minutes"`
	State       string       `json:"state,omitempty"`
	Person      string       `json:"person,omitempty"`
	NeedsReview bool         `json:"needs_review,omitempty"`
	Detail      *entryDetail `json:"detail,omitempty"`
}

// entryDetail is what response_format detailed adds to an entry.
type entryDetail struct {
	UserID        int64    `json:"user_id"`
	SpaceID       *string  `json:"space_id"`
	PausedMinutes int      `json:"paused_minutes"`
	Attribution   *string  `json:"attribution_source"`
	Confidence    *float64 `json:"attribution_confidence"`
}

// timerOutput answers the timer tools.
type timerOutput struct {
	Status  string    `json:"status"`
	Entry   *entryOut `json:"entry,omitempty"`
	Message string    `json:"message,omitempty"`
}

// entryOutput answers the tools that write one entry.
type entryOutput struct {
	Status     string    `json:"status"`
	Entry      *entryOut `json:"entry,omitempty"`
	EntryID    int64     `json:"entry_id,omitempty"`
	ConfirmURL string    `json:"confirm_url,omitempty"`
	Message    string    `json:"message,omitempty"`
}

// total is an amount of tracked time.
type total struct {
	Minutes int     `json:"minutes"`
	Hours   float64 `json:"hours"`
	Entries int     `json:"entries"`
}

// line is the time of one task, person or day.
type line struct {
	ID      int64   `json:"id,omitempty"`
	Name    string  `json:"name"`
	Weekday string  `json:"weekday,omitempty"`
	Minutes int     `json:"minutes"`
	Hours   float64 `json:"hours"`
}

// projectLine is the time of one project, split by task in a summary.
type projectLine struct {
	ID      int64   `json:"id"`
	Name    string  `json:"name"`
	Minutes int     `json:"minutes"`
	Hours   float64 `json:"hours"`
	Tasks   []line  `json:"tasks,omitempty"`
}

// entriesOutput answers sablier_list_entries. Totals cover the whole range, not the page.
type entriesOutput struct {
	From       string        `json:"from"`
	To         string        `json:"to"`
	Total      total         `json:"total"`
	ByProject  []projectLine `json:"by_project"`
	Entries    []entryOut    `json:"entries"`
	NextCursor string        `json:"next_cursor,omitempty"`
}

// summaryOutput answers sablier_weekly_summary.
type summaryOutput struct {
	WeekStart   string        `json:"week_start"`
	WeekEnd     string        `json:"week_end"`
	Timezone    string        `json:"timezone"`
	Total       total         `json:"total"`
	Days        []line        `json:"days"`
	ByProject   []projectLine `json:"by_project"`
	ByPerson    []line        `json:"by_person,omitempty"`
	NeedsReview int           `json:"needs_review,omitempty"`
	Running     bool          `json:"includes_running_timer,omitempty"`
	Truncated   bool          `json:"truncated,omitempty"`
}

// projectOut is one project and its tasks.
type projectOut struct {
	ID        int64     `json:"id"`
	Name      string    `json:"name"`
	SpaceID   string    `json:"space_id,omitempty"`
	Space     string    `json:"space,omitempty"`
	Tasks     []taskOut `json:"tasks"`
	MoreTasks int       `json:"more_tasks,omitempty"`
}

// taskOut is one task.
type taskOut struct {
	ID     int64  `json:"id"`
	Name   string `json:"name"`
	Status string `json:"status"`
}

// spaceOut is a space the person belongs to.
type spaceOut struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

// projectsOutput answers sablier_list_projects.
type projectsOutput struct {
	Spaces     []spaceOut   `json:"spaces"`
	Projects   []projectOut `json:"projects"`
	NextCursor string       `json:"next_cursor,omitempty"`
}

// agentSessionOutput answers sablier_record_agent_session with the entry the session
// was booked as.
type agentSessionOutput struct {
	Status      string  `json:"status"`
	EntryID     int64   `json:"entry_id"`
	AgentID     int64   `json:"agent_id"`
	Agent       string  `json:"agent"`
	ProjectID   int64   `json:"project_id"`
	TaskID      int64   `json:"task_id"`
	Task        string  `json:"task"`
	SpaceID     *string `json:"space_id"`
	Minutes     int     `json:"minutes"`
	NeedsReview bool    `json:"needs_review"`
}

// heartbeatOutput answers sablier_record_heartbeat.
type heartbeatOutput struct {
	Status   string `json:"status"`
	Accepted int    `json:"accepted"`
	Message  string `json:"message,omitempty"`
}
