package settings

import (
	"net/http"

	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes wires the /settings endpoints. The webhook is app-wide, so
// both routes are open to the SABLIER_ADMINS allow-list only.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service, admins middleware.Admins) {
	router.Route("/settings", func(r chi.Router) {
		r.With(middleware.RequireAdmin(authService, admins)).Get("/", func(w http.ResponseWriter, req *http.Request) {
			resp, err := service.controller.getSettings(req.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.With(middleware.RequireAdmin(authService, admins)).Put("/", func(w http.ResponseWriter, req *http.Request) {
			var body UpdateRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.updateSettings(req.Context(), &body)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}
