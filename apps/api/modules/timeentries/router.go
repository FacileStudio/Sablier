package timeentries

import (
	"net/http"
	"net/url"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

func parseListFilter(query url.Values) (listFilter, error) {
	filter := listFilter{Space: query.Get("space_id"), Running: query.Get("running") == "true"}
	for name, target := range map[string]*int64{"project_id": &filter.ProjectID, "user_id": &filter.UserID} {
		raw := query.Get(name)
		if raw == "" {
			continue
		}
		id, err := strconv.ParseInt(raw, 10, 64)
		if err != nil {
			return listFilter{}, errors.Invalid("invalid " + name)
		}
		*target = id
	}
	return filter, nil
}

// RegisterRoutes wires the authenticated /time-entries endpoints, including the
// timer subroutes for start, stop, pause and resume.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/time-entries", func(router chi.Router) {
		router.Use(middleware.RequireAuth(authService))

		router.Get("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			filter, err := parseListFilter(request.URL.Query())
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.list(request.Context(), identity.UserID, filter)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Get("/running", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.controller.running(request.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			if resp == nil {
				httpjson.WriteJSON(w, http.StatusOK, map[string]any{"entry": nil})
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]any{"entry": resp})
		})

		router.Post("/start", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req StartTimerRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.start(request.Context(), identity.UserID, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusCreated, resp)
		})

		router.Post("/stop", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.controller.stop(request.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/pause", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.controller.pause(request.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/resume", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			resp, err := service.controller.resume(request.Context(), identity.UserID)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Post("/", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			var req CreateEntryRequest
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

		router.Put("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid entry id"))
				return
			}
			var req UpdateEntryRequest
			if err := httpjson.DecodeJSON(w, request, &req); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			resp, err := service.controller.update(request.Context(), identity.UserID, id, &req)
			if err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, resp)
		})

		router.Delete("/{id}", func(w http.ResponseWriter, request *http.Request) {
			identity, _ := authcontext.IdentityFromContext(request.Context())
			id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
			if err != nil {
				httpjson.WriteError(w, errors.Invalid("invalid entry id"))
				return
			}
			if err := service.controller.delete(request.Context(), identity.UserID, id); err != nil {
				httpjson.WriteError(w, err)
				return
			}
			httpjson.WriteJSON(w, http.StatusOK, map[string]bool{"deleted": true})
		})
	})
}
