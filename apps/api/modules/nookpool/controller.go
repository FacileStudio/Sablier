package nookpool

import (
	"context"
)

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) getSettings(ctx context.Context) (*PoolSettingsResponse, error) {
	s, fromEnv, err := c.service.getSettings(ctx)
	if err != nil {
		return nil, err
	}
	return &PoolSettingsResponse{
		Settings:  *s,
		Connected: c.service.isConnected(),
		FromEnv:   fromEnv,
	}, nil
}

func (c *Controller) updateSettings(ctx context.Context, req *UpdatePoolRequest) (*PoolSettingsResponse, error) {
	s, connectErr, err := c.service.updateSettings(ctx, req)
	if err != nil {
		return nil, err
	}
	resp := &PoolSettingsResponse{
		Settings:  *s,
		Connected: c.service.isConnected(),
	}
	if connectErr != "" {
		resp.ConnectError = connectErr
	}
	return resp, nil
}
