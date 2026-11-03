package agents

import (
	"net/http"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes wires the authenticated /agents endpoints. porte's scope gate
// already holds a delegated bearer to the write scope on the POST.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/agents", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.controller.list(request.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/paired", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			query := request.URL.Query()
			resp, err := service.controller.paired(request.Context(), identity.UserID, query.Get("from"), query.Get("to"))
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/sessions", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req RecordSessionRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.recordSession(request.Context(), identity.UserID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}
