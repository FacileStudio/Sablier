package mcp

import (
	"context"
	"net/http"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	mcpkit "github.com/FacileStudio/tronc/mcp"
)

const webActor = "sablier-web"

// pendingUser is the signed-in person of a confirm page request. Only the person's own
// session passes: a delegated token must never confirm what an agent asked for.
func pendingUser(r *http.Request) (string, error) {
	authenticated, ok := porte.From(r.Context())
	identity, found := authcontext.IdentityFromContext(r.Context())
	if !ok || !found {
		return "", errors.Unauthorized("missing auth")
	}
	if authenticated.SessionID == 0 || authenticated.Actor != "" {
		return "", mcpkit.ErrDelegated
	}
	return identity.UserID, nil
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
