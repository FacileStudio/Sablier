package auth

import (
	"context"
	"log/slog"
	"net/http"

	"api/internal/env"
	"api/internal/httpjson"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, appEnv env.Config) {
	router.Route("/auth", func(router chi.Router) {
		router.Post("/register", func(w http.ResponseWriter, request *http.Request) {
			var req RegisterRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.register(request.Context(), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Post("/login", func(w http.ResponseWriter, request *http.Request) {
			var req LoginRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}

			resp, err := service.controller.login(request.Context(), &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		if appEnv.OIDC != nil {
			oidc, err := newOIDCHandler(context.Background(), appEnv.OIDC, service)
			if err != nil {
				slog.Error("failed to initialize OIDC provider", slog.Any("error", err))
			} else {
				router.Get("/oidc", oidc.login)
				router.Get("/oidc/callback", oidc.callback)
			}
		}
	})
}
