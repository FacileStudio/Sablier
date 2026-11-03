package heartbeats

import (
	"net/http"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

const maxIngestBytes = 2 << 20

// RegisterRoutes wires the authenticated /heartbeats endpoints: ingest, drafts,
// confirm, dismiss and the auto-confirm setting.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/heartbeats", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))
		router.Post("/", service.handleIngest)
		router.Get("/drafts", service.handleDrafts)
		router.Post("/drafts/confirm", service.handleConfirm)
		router.Post("/drafts/dismiss", service.handleDismiss)
		router.Get("/settings", service.handleSettings)
		router.Put("/settings", service.handleUpdateSettings)
	})
}

func respond(w http.ResponseWriter, status int, value any, err error) {
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	httpjson.WriteJSON(w, status, value)
}

func (service *Service) handleIngest(w http.ResponseWriter, request *http.Request) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	var req IngestRequest
	if err := httpjson.DecodeJSONLimit(w, request, &req, maxIngestBytes); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	resp, err := service.controller.ingest(request.Context(), identity.UserID, &req)
	respond(w, http.StatusAccepted, resp, err)
}

func (service *Service) handleDrafts(w http.ResponseWriter, request *http.Request) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	query := request.URL.Query()
	resp, err := service.controller.drafts(request.Context(), identity.UserID, query.Get("from"), query.Get("to"))
	respond(w, http.StatusOK, resp, err)
}

func (service *Service) handleConfirm(w http.ResponseWriter, request *http.Request) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	var req ConfirmRequest
	if err := httpjson.DecodeJSON(w, request, &req); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	resp, err := service.controller.confirm(request.Context(), identity.UserID, &req)
	respond(w, http.StatusCreated, resp, err)
}

func (service *Service) handleDismiss(w http.ResponseWriter, request *http.Request) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	var req DismissRequest
	if err := httpjson.DecodeJSON(w, request, &req); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	resp, err := service.controller.dismiss(request.Context(), identity.UserID, &req)
	respond(w, http.StatusOK, resp, err)
}

func (service *Service) handleSettings(w http.ResponseWriter, request *http.Request) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	resp, err := service.settings(request.Context(), userID)
	respond(w, http.StatusOK, resp, err)
}

func (service *Service) handleUpdateSettings(w http.ResponseWriter, request *http.Request) {
	identity, _ := authcontext.IdentityFromContext(request.Context())
	userID, err := parseUserID(identity.UserID)
	if err != nil {
		httpjson.WriteError(w, err)
		return
	}
	var req DraftSettings
	if err := httpjson.DecodeJSON(w, request, &req); err != nil {
		httpjson.WriteError(w, err)
		return
	}
	resp, err := service.updateSettings(request.Context(), userID, &req)
	respond(w, http.StatusOK, resp, err)
}
