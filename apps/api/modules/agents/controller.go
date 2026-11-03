package agents

import (
	"context"
	"strconv"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/agentwork"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/tronc/errors"
)

// Controller serves the agents routes.
type Controller struct {
	service *Service
}

func newController(service *Service) *Controller {
	return &Controller{service: service}
}

func parseOwner(userID string) (int64, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return 0, errors.Invalid("invalid user id")
	}
	return id, nil
}

func (c *Controller) recordSession(ctx context.Context, userID string, req *RecordSessionRequest) (*RecordSessionResponse, error) {
	ownerID, err := parseOwner(userID)
	if err != nil {
		return nil, err
	}
	session, err := toSession(ownerID, req, time.Now())
	if err != nil {
		return nil, err
	}
	result, err := c.service.record(ctx, session)
	if err != nil {
		return nil, err
	}
	return &RecordSessionResponse{Entry: toEntryResponse(result), Agent: toAgentResponse(result.Agent)}, nil
}

func (c *Controller) list(ctx context.Context, userID string) ([]AgentResponse, error) {
	ownerID, err := parseOwner(userID)
	if err != nil {
		return nil, err
	}
	agents, err := c.service.listAgents(ctx, ownerID)
	if err != nil {
		return nil, err
	}
	out := make([]AgentResponse, 0, len(agents))
	for _, agent := range agents {
		out = append(out, toAgentResponse(agent))
	}
	return out, nil
}

func (c *Controller) paired(ctx context.Context, userID, rawFrom, rawTo string) ([]PairedResponse, error) {
	ownerID, err := parseOwner(userID)
	if err != nil {
		return nil, err
	}
	from, fromErr := time.Parse(time.RFC3339, rawFrom)
	to, toErr := time.Parse(time.RFC3339, rawTo)
	if fromErr != nil || toErr != nil || !to.After(from) {
		return nil, errors.Invalid("from and to must be RFC3339 times with to after from")
	}
	if to.Sub(from) > 366*24*time.Hour {
		return nil, errors.Invalid("the window is at most one year")
	}
	return c.service.paired(ctx, ownerID, from.UTC(), to.UTC())
}

func toAgentResponse(agent schemas.User) AgentResponse {
	response := AgentResponse{ID: agent.ID, Name: agent.Name, AvatarURL: agent.Avatar(), Color: agent.Color}
	if agent.AgentKey != nil {
		response.AgentKey = *agent.AgentKey
	}
	if agent.OwnerUserID != nil {
		response.OwnerUserID = *agent.OwnerUserID
	}
	return response
}

func toEntryResponse(result *agentwork.Result) SessionEntryResponse {
	entry := result.Entry
	response := SessionEntryResponse{
		ID:             entry.ID,
		ProjectID:      entry.ProjectID,
		TaskID:         entry.TaskID,
		TaskName:       result.Target.Task.Name,
		UserID:         entry.UserID,
		Description:    entry.LegacyDescription,
		StartedAt:      entry.StartedAt,
		StoppedAt:      entry.StoppedAt,
		PausedDuration: entry.PausedDurationMs,
		Cowork:         entry.Cowork,
	}
	if entry.FacileID != nil {
		response.FacileID = *entry.FacileID
	}
	return response
}
