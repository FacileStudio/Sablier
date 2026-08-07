package antenne

import (
	"net/http"

	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/antenne", func(r chi.Router) {
		r.Use(middleware.RequireAuth(authService))

		r.Get("/", func(w http.ResponseWriter, req *http.Request) {
			resp, err := service.controller.getSettings(req.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.Post("/sync", func(w http.ResponseWriter, req *http.Request) {
			result, err := service.controller.triggerSync(req.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, result)
		})

		r.Put("/", func(w http.ResponseWriter, req *http.Request) {
			var body UpdatePoolRequest
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

		r.Get("/events", func(w http.ResponseWriter, req *http.Request) {
			resp, err := service.controller.getPoolEvents(req.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.Put("/events", func(w http.ResponseWriter, req *http.Request) {
			var body UpdatePoolEventsRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.updatePoolEvents(req.Context(), &body)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}
