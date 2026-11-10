package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusDeleted = "deleted"
	waiting       = "Nothing happened yet: your client cannot ask the user to confirm. Give the user confirm_url and " +
		"ask them to confirm there. Do not call this tool again for the same entry; the request stays open for 7 days."
)

// deleteEntry is sablier_delete_entry: the entry goes once the user confirms.
func (t *toolset) deleteEntry(ctx context.Context, req *sdk.CallToolRequest, in deleteInput) (*sdk.CallToolResult, entryOutput, error) {
	if mcpkit.Retrying(req) {
		return t.resumeDelete(ctx, req)
	}
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, entryOutput{}, err
	}
	entry, err := t.Entries.Own(ctx, userID, in.EntryID)
	if err != nil {
		return nil, entryOutput{}, mcpkit.Failure(err, hintEntry)
	}
	result, outcome, err := t.confirm.Request(ctx, req, kindDelete, target(in.EntryID), deleteQuestion(entry))
	if err != nil {
		return nil, entryOutput{}, mcpkit.Failure(err, hintEntry)
	}
	if result != nil {
		return result, entryOutput{}, nil
	}
	return nil, entryOutput{Status: outcome.Status, EntryID: in.EntryID, ConfirmURL: outcome.ConfirmURL, Message: waiting}, nil
}

// resumeDelete handles the retry that carries the person's answer.
func (t *toolset) resumeDelete(ctx context.Context, req *sdk.CallToolRequest) (*sdk.CallToolResult, entryOutput, error) {
	if _, err := t.userOf(ctx, req); err != nil {
		return nil, entryOutput{}, err
	}
	outcome, err := t.confirm.Resume(ctx, req, kindDelete)
	if err != nil {
		return nil, entryOutput{}, mcpkit.Failure(err, "Call sablier_delete_entry again to ask the user afresh.")
	}
	if outcome.Status == mcpkit.StatusDeclined {
		return nil, entryOutput{Status: mcpkit.StatusDeclined, Message: "The user declined. Nothing happened, and the entry is kept."}, nil
	}
	_, entryID, _ := refOf(outcome.Action)
	return nil, entryOutput{Status: statusDeleted, EntryID: entryID, Message: "The user confirmed and the entry is deleted."}, nil
}

// deleteQuestion is what the person is asked before an entry goes.
func deleteQuestion(e *timeentries.TimeEntryResponse) string {
	net, _ := worked(*e, time.Now())
	return fmt.Sprintf("Delete the time entry of %d min on %s / %s started %s UTC? This cannot be undone.",
		minutes(net), e.ProjectName, e.TaskName, e.StartedAt.UTC().Format("2006-01-02 15:04"))
}
