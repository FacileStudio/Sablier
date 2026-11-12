package mcp

import (
	"context"
	"fmt"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/modules/agents"
	"github.com/FacileStudio/Sablier/apps/api/modules/heartbeats"
	mcpkit "github.com/FacileStudio/tronc/mcp"
	sdk "github.com/modelcontextprotocol/go-sdk/mcp"
)

const (
	statusRecorded = "recorded"
	statusSkipped  = "skipped"
)

// recordAgentSession is sablier_record_agent_session: the same upsert as POST
// /api/agents/sessions, booked to the caller's agent user.
func (t *toolset) recordAgentSession(ctx context.Context, req *sdk.CallToolRequest, in agentSessionInput) (*sdk.CallToolResult, agentSessionOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, agentSessionOutput{}, err
	}
	started, err := rfc3339("started_at", in.StartedAt)
	if err != nil {
		return nil, agentSessionOutput{}, err
	}
	stopped, err := rfc3339("stopped_at", in.StoppedAt)
	if err != nil {
		return nil, agentSessionOutput{}, err
	}
	resp, err := t.Agents.RecordSession(ctx, userID, &agents.RecordSessionRequest{
		AgentKey: in.AgentKey, Name: in.Name, AvatarURL: in.AvatarURL, FacileID: in.FacileID,
		StartedAt: &started, StoppedAt: &stopped, ActiveSeconds: in.ActiveSeconds, Cowork: in.Cowork,
		Description: in.Description,
		SessionFacts: agents.SessionFacts{OpusTaskIDs: in.OpusTaskIDs, RegistreSpaceID: in.RegistreSpaceID,
			OpusWorkspace: in.OpusWorkspace, Title: in.Title},
	})
	if err != nil {
		return nil, agentSessionOutput{}, mcpkit.Failure(err, "")
	}
	return nil, sessionOut(resp), nil
}

// recordHeartbeat is sablier_record_heartbeat: one heartbeat through the same ingest as
// POST /api/heartbeats, where an exact repeat changes nothing.
func (t *toolset) recordHeartbeat(ctx context.Context, req *sdk.CallToolRequest, in heartbeatInput) (*sdk.CallToolResult, heartbeatOutput, error) {
	userID, err := t.userOf(ctx, req)
	if err != nil {
		return nil, heartbeatOutput{}, err
	}
	at, err := rfc3339("at", in.At)
	if err != nil {
		return nil, heartbeatOutput{}, err
	}
	accepted, err := t.Heartbeats.Ingest(ctx, userID, []heartbeats.HeartbeatInput{{
		At: at, Source: in.Source, Entity: in.Entity, Project: in.Project,
		HeartbeatHints: heartbeats.HeartbeatHints{RegistreSpaceID: in.RegistreSpaceID, OpusTaskID: in.OpusTaskID, Title: in.Title},
	}})
	if err != nil {
		return nil, heartbeatOutput{}, mcpkit.Failure(err, "")
	}
	if accepted == 0 {
		return nil, heartbeatOutput{Status: statusSkipped, Message: "at is older than 7 days or more than 5 minutes ahead, so nothing was stored."}, nil
	}
	return nil, heartbeatOutput{Status: statusRecorded, Accepted: accepted}, nil
}

// rfc3339 reads a required RFC 3339 time.
func rfc3339(field, value string) (time.Time, error) {
	at, err := time.Parse(time.RFC3339, value)
	if err != nil {
		return time.Time{}, fmt.Errorf("%s must be an RFC 3339 time such as 2026-11-12T09:30:00Z", field)
	}
	return at, nil
}

// sessionOut is the booked entry in the tool's shape.
func sessionOut(resp *agents.RecordSessionResponse) agentSessionOutput {
	entry := resp.Entry
	var worked time.Duration
	if entry.StoppedAt != nil {
		worked = entry.StoppedAt.Sub(entry.StartedAt) - time.Duration(entry.PausedDuration)*time.Millisecond
	}
	return agentSessionOutput{
		Status: statusRecorded, EntryID: entry.ID, AgentID: resp.Agent.ID, Agent: resp.Agent.Name,
		ProjectID: entry.ProjectID, TaskID: entry.TaskID, Task: entry.TaskName, SpaceID: entry.SpaceID,
		Minutes: int(worked / time.Minute), NeedsReview: entry.NeedsReview,
	}
}
