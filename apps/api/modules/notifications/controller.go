package notifications

import "context"

type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func (c *Controller) getVAPIDPublicKey(_ context.Context) *VAPIDPublicKeyResponse {
	return &VAPIDPublicKeyResponse{PublicKey: c.service.getVAPIDPublicKey()}
}

func (c *Controller) saveSubscription(ctx context.Context, req *SaveSubscriptionRequest) (*SubscriptionResponse, error) {
	if err := c.service.saveSubscription(ctx, req); err != nil {
		return nil, err
	}
	return &SubscriptionResponse{Saved: true}, nil
}

func (c *Controller) deleteSubscription(ctx context.Context) (*DeleteResponse, error) {
	if err := c.service.deleteSubscription(ctx); err != nil {
		return nil, err
	}
	return &DeleteResponse{Deleted: true}, nil
}

func (c *Controller) broadcastNotification(ctx context.Context, req *BroadcastRequest) *BroadcastResponse {
	sent, failed := c.service.broadcastNotification(ctx, req)
	return &BroadcastResponse{Sent: sent, Failed: failed}
}
