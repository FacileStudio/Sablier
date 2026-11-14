package mcp

import (
	"context"
	"net/http"
	"strconv"
	"time"

	"github.com/FacileStudio/porte/mcpauth"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const webActor = "sablier-web"

// pendingUser is the signed-in person of a confirm page request. Only the person's own
// browser session passes: a bearer, CLI or delegated token must never confirm what an
// agent asked for.
func pendingUser(r *http.Request) (string, error) {
	identity, err := mcpauth.RequireOwnSession(r)
	if err != nil {
		return "", err
	}
	return strconv.FormatInt(identity.UserID, 10), nil
}

// describe adds the targeted entry, as it is now, to a confirm page item.
func (t *toolset) describe(ctx context.Context, item mcpkit.PendingItem, a mcpkit.Action) PendingItem {
	out := PendingItem{PendingItem: item}
	if owner, entryID, err := refOf(a); err == nil {
		out.Entry = t.pendingEntry(ctx, owner, entryID)
	}
	return out
}

// done is the confirm page's answer once an action ran: the entry is deleted.
func done(context.Context, mcpkit.Outcome) mcpkit.PendingResult {
	return mcpkit.PendingResult{Status: statusDeleted}
}

// pendingEntry is the entry an action targets, or nil once it is gone.
func (t *toolset) pendingEntry(ctx context.Context, userID, entryID int64) *PendingEntry {
	entry, err := t.Entries.Own(ctx, userID, entryID)
	if err != nil {
		return nil
	}
	net, _ := worked(*entry, time.Now())
	return &PendingEntry{
		ID: entry.ID, Project: entry.ProjectName, Task: entry.TaskName,
		StartedAt: entry.StartedAt, StoppedAt: entry.StoppedAt, Minutes: minutes(net),
	}
}
