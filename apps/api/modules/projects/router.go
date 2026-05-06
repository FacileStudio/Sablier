package projects

import (
	"net/http"
	"strconv"

	"api/internal/authcontext"
	"api/internal/errors"
	"api/internal/httpjson"
	"api/internal/middleware"
	"api/modules/auth"

	"github.com/go-chi/chi/v5"
)

func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/projects", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			resp, err := service.controller.list(request.Context())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req CreateProjectRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.create(request.Context(), identity.UserID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Get("/{id}", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid project id"))
				return
			}
			resp, err := service.controller.get(request.Context(), id)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid project id"))
				return
			}
			var req UpdateProjectRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.update(request.Context(), id, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid project id"))
				return
			}
			if err := service.controller.delete(request.Context(), id); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]bool{"deleted": true})
		})

		router.Get("/{id}/tasks", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid project id"))
				return
			}
			resp, err := service.controller.listTasks(request.Context(), id)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/{id}/tasks", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid project id"))
				return
			}
			var req CreateTaskRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.createTask(request.Context(), id, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Put("/{id}/tasks/{taskId}", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid project id"))
				return
			}
			taskID, err := strconv.ParseInt(chi.URLParam(request, "taskId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid task id"))
				return
			}
			var req UpdateTaskRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.updateTask(request.Context(), id, taskID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{id}/tasks/{taskId}", func(w http.ResponseWriter, request *http.Request) {
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid project id"))
				return
			}
			taskID, err := strconv.ParseInt(chi.URLParam(request, "taskId"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid task id"))
				return
			}
			affected, err := service.controller.deleteTask(request.Context(), id, taskID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{"deleted": true, "sessions_unlinked": affected})
		})
	})
}
