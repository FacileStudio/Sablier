package heartbeats

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/tronc/errors"
)

const (
	defaultDraftWindow = 48 * time.Hour
	maxDraftWindow     = 31 * 24 * time.Hour
	maxActionRange     = 24 * time.Hour
)

// Controller validates heartbeat and draft requests before the service runs them.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func parseUserID(raw string) (int64, error) {
	id, err := strconv.ParseInt(raw, 10, 64)
	if err != nil {
		return 0, errors.Invalid("invalid user id")
	}
	return id, nil
}

func validRange(from, to time.Time, limit time.Duration) error {
	if from.IsZero() || to.IsZero() {
		return errors.Invalid("started_at and stopped_at are required")
	}
	if !to.After(from) {
		return errors.Invalid("stopped_at must be after started_at")
	}
	if to.Sub(from) > limit {
		return errors.Invalid("the range is too long")
	}
	return nil
}

func (c *Controller) ingest(ctx context.Context, rawUserID string, req *IngestRequest) (*IngestResponse, error) {
	userID, err := parseUserID(rawUserID)
	if err != nil {
		return nil, err
	}
	if len(req.Heartbeats) > maxBatch {
		return nil, errors.Invalid("at most 500 heartbeats per call")
	}
	for i, h := range req.Heartbeats {
		if err := validateHeartbeat(i, h); err != nil {
			return nil, err
		}
	}
	accepted, err := c.service.ingest(ctx, userID, req.Heartbeats)
	if err != nil {
		return nil, err
	}
	return &IngestResponse{Accepted: accepted}, nil
}

func (c *Controller) drafts(ctx context.Context, rawUserID, rawFrom, rawTo string) ([]Draft, error) {
	userID, err := parseUserID(rawUserID)
	if err != nil {
		return nil, err
	}
	to, from := c.service.now(), c.service.now().Add(-defaultDraftWindow)
	if rawTo != "" {
		if to, err = time.Parse(time.RFC3339, rawTo); err != nil {
			return nil, errors.Invalid("to must be RFC3339")
		}
	}
	if rawFrom != "" {
		if from, err = time.Parse(time.RFC3339, rawFrom); err != nil {
			return nil, errors.Invalid("from must be RFC3339")
		}
	}
	if err := validRange(from, to, maxDraftWindow); err != nil {
		return nil, err
	}
	return c.service.listDrafts(ctx, userID, from, to)
}

func (c *Controller) confirm(ctx context.Context, rawUserID string, req *ConfirmRequest) (*timeentries.TimeEntryResponse, error) {
	userID, err := parseUserID(rawUserID)
	if err != nil {
		return nil, err
	}
	if err := validRange(req.StartedAt, req.StoppedAt, maxActionRange); err != nil {
		return nil, err
	}
	if req.ProjectID <= 0 {
		return nil, errors.Invalid("project_id is required")
	}
	return c.service.confirm(ctx, userID, req)
}

func (c *Controller) dismiss(ctx context.Context, rawUserID string, req *DismissRequest) (*DismissResponse, error) {
	userID, err := parseUserID(rawUserID)
	if err != nil {
		return nil, err
	}
	if err := validRange(req.StartedAt, req.StoppedAt, maxActionRange); err != nil {
		return nil, err
	}
	if err := c.service.dismiss(ctx, userID, req); err != nil {
		return nil, err
	}
	return &DismissResponse{Dismissed: true}, nil
}
