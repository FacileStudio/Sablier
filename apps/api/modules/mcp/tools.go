package mcp

import (
	"context"
	stderrors "errors"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	toolStatus    = "sablier_timer_status"
	toolStart     = "sablier_start_timer"
	toolStop      = "sablier_stop_timer"
	toolEntries   = "sablier_list_entries"
	toolAdd       = "sablier_add_entry"
	toolUpdate    = "sablier_update_entry"
	toolDelete    = "sablier_delete_entry"
	toolProjects  = "sablier_list_projects"
	toolSummary   = "sablier_weekly_summary"
	spacePersonal = "personal"
	hintEntry     = "Call sablier_list_entries for the ids of your entries; only your own entries can be changed."
	hintProject   = "Call sablier_list_projects for the project and task ids you can use."
	hintSpace     = "Call sablier_list_projects for the spaces you belong to."
)

var (
	errNoCaller = stderrors.New("this request carries no valid access token. Reconnect Sablier")
	errTask     = stderrors.New("pass task_id, or task with the task's name, to say what the time is for")
)

// toolset holds what the tool handlers share.
type toolset struct {
	Deps
	confirm *mcpkit.Confirmations
	appURL  string
}

// userOf reads the Sablier user id from the verified token, refused the way RequireAuth
// refuses it: an agent user or a deleted account never acts.
func (t *toolset) userOf(ctx context.Context, req *sdk.CallToolRequest) (int64, error) {
	identity, ok := mcpauth.Identity(req.Extra.TokenInfo)
	if !ok {
		return 0, errNoCaller
	}
	if _, _, err := t.Accounts.IdentityForUser(ctx, identity.UserID); err != nil {
		return 0, errNoCaller
	}
	return identity.UserID, nil
}

// scopeOf maps the space argument to the REST filter: all by default, personal for the
// caller's private rows, or one space id.
func scopeOf(space string) string {
	switch space {
	case "", schemas.AllSpaces:
		return schemas.AllSpaces
	case spacePersonal:
		return ""
	}
	return space
}

// taskFor resolves the task a timer or an entry books on: task_id as given, or the task
// named task in the project, created when it is new, as the web app does.
func (t *toolset) taskFor(ctx context.Context, userID, projectID, taskID int64, name string) (int64, error) {
	if taskID > 0 {
		return taskID, nil
	}
	if name == "" {
		return 0, errTask
	}
	task, err := t.Projects.EnsureTask(ctx, userID, projectID, name)
	if err != nil {
		return 0, mcpkit.Failure(err, hintProject)
	}
	return task.ID, nil
}

// target is the pending-action reference of an entry.
func target(entryID int64) string {
	return strconv.FormatInt(entryID, 10)
}
