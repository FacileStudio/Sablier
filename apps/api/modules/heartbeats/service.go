package heartbeats

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/modules/timeentries"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

const (
	maxBatch       = 500
	maxEntityLen   = 1024
	maxHintLen     = 200
	maxFactLen     = 200
	maxAge         = 7 * 24 * time.Hour
	maxFutureSkew  = 5 * time.Minute
	retention      = 90 * 24 * time.Hour
	autoConfirmAge = 48 * time.Hour
)

const sourceList = "claude-code, editor, desktop, git, bulle, cli"

// Service stores heartbeats and turns them into drafts and time entries.
type Service struct {
	orm        *gorm.DB
	entries    *timeentries.Service
	resolver   attribution.Resolver
	controller *Controller
	now        func() time.Time
}

// NewService wires the heartbeats service; entries books confirmed drafts and resolver
// suggests where each draft belongs.
func NewService(orm *gorm.DB, entries *timeentries.Service, resolver attribution.Resolver) *Service {
	service := &Service{orm: orm, entries: entries, resolver: resolver, now: func() time.Time { return time.Now().UTC() }}
	service.controller = newController(service)
	return service
}

func validSource(source string) bool {
	return slices.Contains(strings.Split(sourceList, ", "), source)
}

func validateHeartbeat(i int, h HeartbeatInput) error {
	field := fmt.Sprintf("heartbeats[%d]", i)
	switch {
	case !validSource(h.Source):
		return errors.Invalid(field + ".source must be one of " + sourceList)
	case h.At.IsZero():
		return errors.Invalid(field + ".at is required")
	case len(h.Entity) > maxEntityLen:
		return errors.Invalid(field + ".entity is too long")
	case len(h.Project) > maxHintLen:
		return errors.Invalid(field + ".project is too long")
	case len(h.RegistreSpaceID) > maxFactLen || len(h.OpusTaskID) > maxFactLen || len(h.Title) > maxFactLen:
		return errors.Invalid(field + ".registre_space_id, opus_task_id and title are at most 200 characters")
	}
	return nil
}

// Ingest validates a batch at the trust boundary and stores it for userID. POST
// /heartbeats and the MCP tool sablier_record_heartbeat both ingest through it.
func (service *Service) Ingest(ctx context.Context, userID int64, items []HeartbeatInput) (int, error) {
	if len(items) > maxBatch {
		return 0, errors.Invalid("at most 500 heartbeats per call")
	}
	for i, h := range items {
		if err := validateHeartbeat(i, h); err != nil {
			return 0, err
		}
	}
	return service.ingest(ctx, userID, items)
}

// ingest stores the heartbeats inside the accepted window, silently skipping the
// ones too old or too far ahead, and ignores exact repeats through the unique index.
func (service *Service) ingest(ctx context.Context, userID int64, items []HeartbeatInput) (int, error) {
	now := service.now()
	rows := make([]schemas.Heartbeat, 0, len(items))
	for _, h := range items {
		at := h.At.UTC().Truncate(time.Second)
		if at.Before(now.Add(-maxAge)) || at.After(now.Add(maxFutureSkew)) {
			continue
		}
		rows = append(rows, schemas.Heartbeat{
			UserID: userID, At: at, Source: h.Source, Entity: h.Entity,
			ProjectHint: strings.TrimSpace(h.Project), ProjectID: h.ProjectID, Write: h.Write,
			HeartbeatFacts: schemas.HeartbeatFacts{
				RegistreSpaceID: strings.TrimSpace(h.RegistreSpaceID),
				OpusTaskID:      strings.TrimSpace(h.OpusTaskID),
				Title:           strings.TrimSpace(h.Title),
			},
		})
	}
	if len(rows) == 0 {
		return 0, nil
	}
	err := service.orm.WithContext(ctx).Clauses(clause.OnConflict{DoNothing: true}).Create(&rows).Error
	if err != nil {
		return 0, errors.Internal("failed to store heartbeats", err)
	}
	return len(rows), nil
}
