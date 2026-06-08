package notifications

import (
	"net/http"

	"api/internal/httpjson"
	"api/internal/middleware"
	"api/modules/auth"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/notifications", func(r chi.Router) {
		r.Get("/vapid-public-key", func(w http.ResponseWriter, req *http.Request) {
			httpjson.WriteJSON(w, http.StatusOK, service.controller.getVAPIDPublicKey(req.Context()))
		})

		r.With(middleware.RequireAuth(authService)).Post("/subscriptions", func(w http.ResponseWriter, req *http.Request) {
			var body SaveSubscriptionRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.saveSubscription(req.Context(), &body)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		r.With(middleware.RequireAuth(authService)).Delete("/subscriptions", func(w http.ResponseWriter, req *http.Request) {
			resp, err := service.controller.deleteSubscription(req.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		// TEST ONLY — remove before production
		r.Post("/test-broadcast", func(w http.ResponseWriter, req *http.Request) {
			var body BroadcastRequest
			if err := httpjson.DecodeJSON(w, req, &body); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp := service.controller.broadcastNotification(req.Context(), &body)
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})
	})
}
