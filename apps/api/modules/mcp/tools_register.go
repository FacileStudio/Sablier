package mcp

import (
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

// list is the curated tool set. Reports are open world: teammates and synced tools name
// the projects and tasks of shared spaces. Deleting an entry is the sensitive tool.
// The agent work tools come last, so the older tools keep their order.
func (t *toolset) list() []mcpkit.Tool {
	return append(append(append(t.reads(), t.writes()...), t.sensitive()...), t.agentWork()...)
}

// reads are the four read-only tools.
func (t *toolset) reads() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolStatus, Title: "Timer status",
			Description: "Show whether the user's timer runs: its entry, project, task, start and minutes so far, " +
				"or status stopped. Call it before starting a timer.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true},
		}, t.status),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolEntries, Title: "List time entries",
			Description: "List the time entries that overlap a range, newest first, with the total and the time per " +
				"project for the whole range (time outside the range is not counted). Default range is the last 7 days " +
				"and the user's own time across personal projects and every space. Pass space, project_id, everyone or " +
				"user_id to narrow or widen it, and next_cursor back as cursor for the next page.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.listEntries),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolProjects, Title: "List projects and tasks",
			Description: "List the projects the user can track time on, newest first, each with its open tasks (ids, " +
				"names and status), and the spaces the user belongs to. Pass query to find a project or task by name, " +
				"project_id for all of one project's tasks, include_done for finished tasks.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.listProjects),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolSummary, Title: "Weekly summary",
			Description: "Add up one week, Monday to Sunday in timezone: the total, the time per day, per project and " +
				"task, and per person when everyone or user_id is set. Default is this week and the user's own time. " +
				"A running timer counts up to now.",
			Annotations: &mcpkit.Annotations{ReadOnly: true, Idempotent: true, OpenWorld: true},
		}, t.summary),
	}
}

// writes are the tools that start, stop, add or change time without losing any.
func (t *toolset) writes() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolStart, Title: "Start the timer",
			Description: "Start the user's timer on a project and a task. Pass task_id, or task with a short name of " +
				"what the work is: Sablier creates that task in the project when it is new. Only one timer runs at a " +
				"time: if one runs, stop it first with sablier_stop_timer.",
			Annotations: &mcpkit.Annotations{},
		}, t.start),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolStop, Title: "Stop the timer",
			Description: "Stop the user's running timer and return the finished entry with its minutes. With no " +
				"timer running it changes nothing and returns status not_running.",
			Annotations: &mcpkit.Annotations{Idempotent: true},
		}, t.stop),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolAdd, Title: "Add a time entry",
			Description: "Record time already spent on a project and a task: started_at plus stopped_at or " +
				"duration_minutes. Pass task_id, or task with a name, created in the project when new. Times without " +
				"an offset are read in timezone.",
			Annotations: &mcpkit.Annotations{},
		}, t.addEntry),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolUpdate, Title: "Edit a time entry",
			Description: "Change one of the user's own entries: its task (task_id or task), its project (with a task " +
				"of that project), its start or its end. Fields left out stay as they are. A finished entry cannot " +
				"become running again.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.updateEntry),
	}
}

// agentWork are the tools an agent books its own working time with, as POST
// /api/agents/sessions and POST /api/heartbeats do.
func (t *toolset) agentWork() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolSession, Title: "Record an agent session",
			Description: "Book one finished span of your own work as an agent to the user, under an agent user named " +
				"after agent_key and name. Sablier finds the project and task from the Opus tasks, space and title " +
				"given, or leaves the entry for review. facile_id keys the session: booking the same id again " +
				"replaces that entry, so a retry never doubles the time. Returns the entry with its minutes.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.recordAgentSession),
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolHeartbeat, Title: "Record a heartbeat",
			Description: "Record that the user was active at a moment, from source. Sablier joins heartbeats into " +
				"draft entries the user confirms in Review. Send at most one every 2 minutes per entity; an exact " +
				"repeat changes nothing, and one older than 7 days is skipped.",
			Annotations: &mcpkit.Annotations{Idempotent: true},
		}, t.recordHeartbeat),
	}
}

// sensitive are the tools the user confirms.
func (t *toolset) sensitive() []mcpkit.Tool {
	return []mcpkit.Tool{
		mcpkit.NewTool(mcpkit.ToolSpec{
			Name: toolDelete, Title: "Delete a time entry",
			Description: "Delete one of the user's own entries for good. The user always confirms: in your client " +
				"when it can ask, otherwise the result has status pending_confirmation and a confirm_url to give the " +
				"user. Nothing is deleted before they confirm.",
			Annotations: &mcpkit.Annotations{Destructive: true, Idempotent: true},
		}, t.deleteEntry),
	}
}
