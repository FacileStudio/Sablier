package agents

import (
	"context"
	stderrors "errors"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/agentwork"
	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service backs the agents module: booking agent sessions and reporting the
// time an owner's agents spent, alone and alongside the owner.
type Service struct {
	orm        *gorm.DB
	resolver   attribution.Resolver
	controller *Controller
}

// NewService wires the agents service; resolver attributes each session.
func NewService(orm *gorm.DB, resolver attribution.Resolver) *Service {
	service := &Service{orm: orm, resolver: resolver}
	service.controller = newController(service)
	return service
}

func (service *Service) record(ctx context.Context, session agentwork.Session) (*agentwork.Result, error) {
	result, err := agentwork.Record(ctx, service.orm.WithContext(ctx), service.resolver, session)
	switch {
	case err == nil:
		return result, nil
	case stderrors.Is(err, agentwork.ErrProjectNotVisible):
		return nil, errors.Invalid("project_id or task_id is not one you can see")
	case stderrors.Is(err, agentwork.ErrForeignFacileID):
		return nil, errors.Conflict("facile_id belongs to another user's entry")
	default:
		return nil, errors.Internal("failed to record the agent session", err)
	}
}

func (service *Service) listAgents(ctx context.Context, ownerID int64) ([]schemas.User, error) {
	var agents []schemas.User
	err := service.orm.WithContext(ctx).
		Where("kind = ? AND owner_user_id = ?", schemas.UserKindAgent, ownerID).
		Order("name asc, id asc").
		Find(&agents).Error
	if err != nil {
		return nil, errors.Internal("failed to list agents", err)
	}
	return agents, nil
}

func (service *Service) entriesIn(ctx context.Context, userIDs []int64, from, to time.Time) ([]schemas.TimeEntry, error) {
	var entries []schemas.TimeEntry
	err := service.orm.WithContext(ctx).
		Where("user_id IN ? AND started_at < ? AND (stopped_at IS NULL OR stopped_at > ?)", userIDs, to, from).
		Find(&entries).Error
	if err != nil {
		return nil, errors.Internal("failed to read time entries", err)
	}
	return entries, nil
}

func (service *Service) paired(ctx context.Context, ownerID int64, from, to time.Time) ([]PairedResponse, error) {
	agents, err := service.listAgents(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	ids := []int64{ownerID}
	for _, agent := range agents {
		ids = append(ids, agent.ID)
	}
	entries, err := service.entriesIn(ctx, ids, from, to)
	if err != nil {
		return nil, err
	}
	return summarize(ownerID, agents, entries, from, to, time.Now().UTC()), nil
}

// summarize adds up each agent's time inside [from, to) and the part of it that
// overlaps the owner's own entries.
func summarize(ownerID int64, agents []schemas.User, entries []schemas.TimeEntry, from, to, now time.Time) []PairedResponse {
	byUser := map[int64][]schemas.TimeEntry{}
	for _, entry := range entries {
		byUser[entry.UserID] = append(byUser[entry.UserID], entry)
	}
	ownerSpans := spansOf(byUser[ownerID], from, to, now)
	out := make([]PairedResponse, 0, len(agents))
	for _, agent := range agents {
		row := PairedResponse{AgentID: agent.ID}
		var agentSeconds, coworkSeconds float64
		for _, entry := range byUser[agent.ID] {
			clipped, ok := clip(entry, from, to, now)
			if !ok {
				continue
			}
			seconds := activeSeconds(entry, clipped, now)
			agentSeconds += seconds
			if entry.Cowork {
				coworkSeconds += seconds
			}
		}
		row.AgentSeconds = int64(agentSeconds)
		row.CoworkSeconds = int64(coworkSeconds)
		row.PairedSeconds = int64(overlapSeconds(spansOf(byUser[agent.ID], from, to, now), ownerSpans))
		out = append(out, row)
	}
	return out
}

func spansOf(entries []schemas.TimeEntry, from, to, now time.Time) []span {
	spans := make([]span, 0, len(entries))
	for _, entry := range entries {
		if clipped, ok := clip(entry, from, to, now); ok {
			spans = append(spans, clipped)
		}
	}
	return spans
}
