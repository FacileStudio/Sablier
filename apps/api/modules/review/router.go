// Package review serves the Review inbox: the entries whose attribution was not sure
// enough to book silently, and the drafts still waiting to become entries.
package review

import (
	"net/http"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/Sablier/apps/api/modules/heartbeats"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// InboxResponse is GET /review: the flagged entries of the caller and the caller's agents,
// and the caller's drafts over the last 48 hours.
type InboxResponse struct {
	Entries []timeentries.TimeEntryResponse `json:"entries"`
	Drafts  []heartbeats.Draft              `json:"drafts"`
}

// ResolveRequest is the body of POST /review/{entry_id}. A positive TaskID moves the
// entry to that task; an empty body or {} accepts it where it is.
type ResolveRequest struct {
	TaskID int64 `json:"task_id"`
}

// Handlers serves the review routes from the entries and drafts services.
type Handlers struct {
	Entries *timeentries.Service
	Drafts  *heartbeats.Service
}

// RegisterRoutes wires the authenticated /review endpoints.
func RegisterRoutes(router chi.Router, handlers Handlers, authService *auth.Service) {
	router.Route("/review", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		router.Get("/", handlers.inbox)
		router.Post("/{entry_id}", handlers.resolve)
	})
}

func callerID(request *http.Request) (int64, error) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	id, err := strconv.ParseInt(identity.UserID, 10, 64)
	if err != nil {
		return 0, errors.Invalid("invalid user id")
	}
	return id, nil
}

func (h Handlers) inbox(w http.ResponseWriter, request *http.Request) {
	userID, err := callerID(request)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	entries, err := h.Entries.ReviewEntries(request.Context(), userID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	from, to := h.Drafts.DefaultWindow()
	drafts, err := h.Drafts.ListDrafts(request.Context(), userID, from, to)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, InboxResponse{Entries: entries, Drafts: drafts})
}

func (h Handlers) resolve(w http.ResponseWriter, request *http.Request) {
	userID, err := callerID(request)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	entryID, err := strconv.ParseInt(chi.URLParam(request, "entry_id"), 10, 64)
	if err != nil {
		httpjson.WriteError(w, errors.Invalid("invalid entry id"))
		return
	}
	var req ResolveRequest
	if request.ContentLength != 0 {
		if err := httpjson.DecodeJSON(w, request, &req); err != nil {
			httpjson.WriteError(w, err)
			return
		}
	}
	resp, err := h.Entries.ReviewEntry(request.Context(), userID, entryID, req.TaskID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, http.StatusOK, resp)
}
