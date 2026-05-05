package settings

import (
	"context"
	"strings"

	"api/internal/authcontext"
	"api/internal/errors"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) getSettings(ctx context.Context) (*SettingsResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	s, err := c.service.getSettings(ctx, identity.UserID)
	if err != nil {
		return nil, err
	}
	return &SettingsResponse{Settings: *s}, nil
}

func (c *Controller) updateSettings(ctx context.Context, req *UpdateRequest) (*SettingsResponse, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return nil, errors.Unauthorized("missing auth")
	}
	s, err := c.service.updateSettings(ctx, identity.UserID, strings.TrimSpace(req.WebhookURL))
	if err != nil {
		return nil, err
	}
	return &SettingsResponse{Settings: *s}, nil
}
