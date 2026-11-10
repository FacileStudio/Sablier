package mcp

import (
	"context"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

// updateEntry is sablier_update_entry: the fields given change, the others stay.
func (t *toolset) updateEntry(ctx context.Context, req *sdk.CallToolRequest, in updateInput) (*sdk.CallToolResult, entryOutput, error) {
	userID, c, err := t.callerClock(ctx, req, in.Timezone)
	if err != nil {
		return nil, entryOutput{}, err
	}
	current, err := t.Entries.Own(ctx, userID, in.EntryID)
	if err != nil {
		return nil, entryOutput{}, mcpkit.Failure(err, hintEntry)
	}
	change, err := t.changes(ctx, userID, in, c, current)
	if err != nil {
		return nil, entryOutput{}, err
	}
	if _, err := t.Entries.Update(ctx, userID, in.EntryID, change); err != nil {
		return nil, entryOutput{}, mcpkit.Failure(err, hintEntry)
	}
	out, err := t.entryView(ctx, userID, in.EntryID, c)
	return nil, entryOutput{Status: statusUpdated, Entry: out}, err
}

// changes merges the given fields over the entry as it is, into the full body PUT
// /time-entries/{id} takes.
func (t *toolset) changes(
	ctx context.Context, userID int64, in updateInput, c clock, current *timeentries.TimeEntryResponse,
) (*timeentries.UpdateEntryRequest, error) {
	change := &timeentries.UpdateEntryRequest{
		ProjectID: current.ProjectID, TaskID: current.TaskID, StartedAt: current.StartedAt, StoppedAt: current.StoppedAt,
	}
	retask := in.TaskID > 0 || in.Task != ""
	if in.ProjectID > 0 && in.ProjectID != current.ProjectID {
		if !retask {
			return nil, errMove
		}
		change.ProjectID = in.ProjectID
	}
	if in.ProjectID == 0 && !retask && in.StartedAt == "" && in.StoppedAt == "" {
		return nil, errNothing
	}
	if retask {
		var err error
		if change.TaskID, err = t.taskFor(ctx, userID, change.ProjectID, in.TaskID, in.Task); err != nil {
			return nil, err
		}
	}
	return change, c.retime(change, in.StartedAt, in.StoppedAt)
}

// retime applies a new start or end to an entry's body.
func (c clock) retime(change *timeentries.UpdateEntryRequest, start, stop string) error {
	var err error
	if start != "" {
		if change.StartedAt, _, err = c.parse("started_at", start); err != nil {
			return err
		}
	}
	if stop != "" {
		stopped, _, err := c.parse("stopped_at", stop)
		change.StoppedAt = &stopped
		return err
	}
	return nil
}
