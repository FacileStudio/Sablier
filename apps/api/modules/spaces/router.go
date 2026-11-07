package spaces

import (
	"net/http"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/modules/auth"
	"github.com/FacileStudio/porte/directory"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
)

// RegisterRoutes wires the authenticated, read-only /spaces endpoints. Creating, renaming,
// inviting, changing a role, removing and leaving all happen in Registre, which pushes the
// result here through the directory sync.
func RegisterRoutes(router chi.Router, service *Service, authService *auth.Service) {
	router.Route("/spaces", func(router chi.Router) {
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

		router.Route("/{spaceId}", func(router chi.Router) {
			router.Get("/", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				spaceID := chi.URLParam(request, "spaceId")
				resp, err := service.controller.get(request.Context(), spaceID, identity.UserID)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})

			router.Get("/members", func(w http.ResponseWriter, request *http.Request) {
				identity, _ := authcontext.IdentityFromContext(request.Context())
				spaceID := chi.URLParam(request, "spaceId")
				resp, err := service.controller.listMembers(request.Context(), spaceID, identity.UserID)
				if err != nil {
					httpjson.WriteError(w, err)
					return
				}
				httpjson.WriteJSON(w, http.StatusOK, resp)
			})
		})
	})
}

// RegisterSync mounts Registre's nudge on the /api group it is given, so it answers at
// exactly directory.SyncPath. It sits outside every auth middleware on purpose: the nudge
// is an anonymous, empty POST that only makes the next authenticated pull happen sooner.
func RegisterSync(router chi.Router, syncer *directory.Syncer) {
	router.Method(http.MethodPost, strings.TrimPrefix(directory.SyncPath, "/api"), syncer.Handler())
}

// RegisterDeletions mounts Registre's signed user.deleted push on the /api group it is
// given, so it answers at exactly directory.DeletionsPath. It sits outside every auth
// middleware: the handler verifies the notice's signature itself.
func RegisterDeletions(router chi.Router, deletions *Deletions) {
	router.Method(http.MethodPost, strings.TrimPrefix(directory.DeletionsPath, "/api"), deletions.Handler())
}
