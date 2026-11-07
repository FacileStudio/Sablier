package projects

import (
	stderrors "errors"
	"net/http"
	"strconv"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

func (service *Service) requireVisibleProject(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, request *http.Request) {
		id, err := strconv.ParseInt(chi.URLParam(request, "id"), 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Invalid("invalid project id"))
			return
		}
		identity, _ := authcontext.IdentityFromContext(request.Context())
		userID, err := strconv.ParseInt(identity.UserID, 10, 64)
		if err != nil {
			httpjson.WriteError(w, errors.Unauthorized("not authenticated"))
			return
		}
		_, err = schemas.VisibleProject(service.orm.WithContext(request.Context()), userID, id)
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			httpjson.WriteError(w, errors.NotFound("project not found"))
			return
		}
		if err != nil {
			httpjson.WriteError(w, errors.Internal("failed to load the project", err))
			return
		}
		next.ServeHTTP(w, request)
	})
}
