package mcp

import (
	"context"
	"net/http"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	"github.com/go-chi/chi/v5"
)

const webActor = "sablier-web"

var errDelegated = errors.Forbidden("only the person's own session can confirm actions")

// pendingUser is the signed-in person of a confirm page request. Only the person's own
// session passes: a delegated token must never confirm what an agent asked for.
func pendingUser(r *http.Request) (mcpkit.Caller, error) {
	authenticated, ok := porte.From(r.Context())
	identity, found := authcontext.IdentityFromContext(r.Context())
	if !ok || !found {
		return mcpkit.Caller{}, errors.Unauthorized("missing auth")
	}
	if authenticated.SessionID == 0 || authenticated.Actor != "" {
		return mcpkit.Caller{}, errDelegated
	}
	return mcpkit.Caller{UserID: identity.UserID, Actor: webActor}, nil
}

// handleList answers the person's open actions with their entries.
func (s *Service) handleList(w http.ResponseWriter, r *http.Request) {
	who, err := pendingUser(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	list, err := s.tools.pendingList(r.Context(), who.UserID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, list)
}

// handleConfirm runs one action.
func (s *Service) handleConfirm(w http.ResponseWriter, r *http.Request) {
	who, err := pendingUser(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if _, err := s.tools.confirm.Confirm(r.Context(), who, chi.URLParam(r, "id")); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, PendingResult{Status: statusDeleted})
}

// handleDismiss closes one action without running it.
func (s *Service) handleDismiss(w http.ResponseWriter, r *http.Request) {
	who, err := pendingUser(r)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	if err := s.tools.confirm.Dismiss(r.Context(), who, chi.URLParam(r, "id")); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, PendingResult{Status: mcpkit.OutcomeDismissed})
}

// pendingList renders the person's open actions with their entries as they are now.
func (t *toolset) pendingList(ctx context.Context, userID string) (PendingList, error) {
	actions, err := t.confirm.Open(ctx, userID)
	if err != nil {
		return PendingList{}, err
	}
	list := PendingList{Actions: make([]PendingItem, 0, len(actions))}
	for _, action := range actions {
		item := PendingItem{
			ID: action.ID, Kind: action.Kind, ClientID: action.ClientID, Summary: action.Summary,
			CreatedAt: action.CreatedAt, ExpiresAt: action.ExpiresAt, State: t.confirm.State(ctx, action),
		}
		if owner, entryID, err := refOf(action); err == nil {
			item.Entry = t.pendingEntry(ctx, owner, entryID)
		}
		list.Actions = append(list.Actions, item)
	}
	return list, nil
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
