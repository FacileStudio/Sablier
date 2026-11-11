package users

import (
	"context"
	"net/http"
	"strings"

	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"
	"github.com/FacileStudio/tronc/errors"
)

// Controller serves the user, profile and api-token routes.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (controller *Controller) list(context context.Context) (*ListResponse, error) {
	identity, ok := authcontext.IdentityFromContext(context)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}

	users, err := controller.service.listUsers(context, identity.UserID)
	if err != nil {
		return nil, err
	}

	return &ListResponse{Users: users}, nil
}

func (controller *Controller) get(context context.Context, userID string) (*PeerResponse, error) {
	identity, ok := authcontext.IdentityFromContext(context)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	peer, err := controller.service.getPeer(context, identity.UserID, userID)
	if err != nil {
		return nil, err
	}
	return &PeerResponse{User: *peer}, nil
}

func (controller *Controller) me(context context.Context) (*MeResponse, error) {
	identity, ok := authcontext.IdentityFromContext(context)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}

	user, err := controller.service.getUser(context, identity.UserID)
	if err != nil {
		return nil, err
	}

	if user.Email == "" {
		user.Email = identity.Email
	}
	user.IsAdmin = controller.service.admins.Allows(identity.Email)

	return &MeResponse{User: *user}, nil
}

// updateMe applies a subset of the profile.
func (controller *Controller) updateMe(r *http.Request, req *UpdateRequest) (*MeResponse, error) {
	context := r.Context()
	identity, ok := authcontext.IdentityFromContext(context)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}

	var name *string
	if req.Name != nil {
		trimmed := strings.TrimSpace(*req.Name)
		if len(trimmed) > 80 {
			return nil, errors.Invalid("name must be at most 80 characters")
		}
		name = &trimmed
	}

	var color *string
	if req.Color != nil {
		normalized, ok := usercolor.Normalize(*req.Color)
		if !ok {
			return nil, errors.Invalid("color must be one of: AD9EF0, F09ED6, EE7E89, EEB47E, A9EE7E, 7EEEDB")
		}
		color = &normalized
	}

	var rate *float64
	if req.Rate != nil {
		v := *req.Rate
		if v < 0 {
			v = 0
		}
		rate = &v
	}

	var rateType *string
	if req.RateType != nil {
		t := *req.RateType
		if t != "daily" && t != "hourly" {
			t = "daily"
		}
		rateType = &t
	}

	var workdayHours *float64
	if req.WorkdayHours != nil {
		h := *req.WorkdayHours
		if h <= 0 {
			h = 8
		}
		workdayHours = &h
	}

	if name == nil && color == nil && rate == nil && rateType == nil && workdayHours == nil {
		return nil, errors.Invalid("at least one field must be provided")
	}

	user, err := controller.service.updateUser(context, identity.UserID, name, color, rate, rateType, workdayHours)
	if err != nil {
		return nil, err
	}
	user.IsAdmin = controller.service.admins.Allows(identity.Email)

	return &MeResponse{User: *user}, nil
}

func (controller *Controller) getApiToken(context context.Context) (*ApiTokenStatusResponse, error) {
	identity, ok := authcontext.IdentityFromContext(context)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	record, err := controller.service.getApiToken(context, identity.UserID)
	if err != nil {
		return nil, err
	}
	if record == nil {
		return &ApiTokenStatusResponse{HasToken: false}, nil
	}
	return &ApiTokenStatusResponse{
		HasToken:  true,
		Name:      record.Label,
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (controller *Controller) createApiToken(context context.Context, req *CreateApiTokenRequest) (*ApiTokenResponse, error) {
	identity, ok := authcontext.IdentityFromContext(context)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	name := strings.TrimSpace(req.Name)
	if name == "" {
		name = "CLI"
	}
	rawToken, record, err := controller.service.createApiToken(context, identity.UserID, name)
	if err != nil {
		return nil, err
	}
	return &ApiTokenResponse{
		Token:     rawToken,
		Name:      record.Label,
		CreatedAt: record.CreatedAt.UTC().Format(time.RFC3339),
	}, nil
}

func (controller *Controller) deleteApiToken(context context.Context) error {
	identity, ok := authcontext.IdentityFromContext(context)
	if !ok {
		return errors.Unauthorized("missing auth")
	}
	return controller.service.deleteApiToken(context, identity.UserID)
}
