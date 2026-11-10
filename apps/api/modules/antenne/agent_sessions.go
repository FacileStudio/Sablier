package antenne

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"time"

	antenneclient "github.com/FacileStudio/antenne-client/go"
	"github.com/FacileStudio/antenne-client/go/envelope"

	"github.com/FacileStudio/Sablier/apps/api/internal/agentwork"
	"github.com/FacileStudio/Sablier/apps/api/internal/attribution"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

// handleAgentSession materializes a Mycelium agent_session event as a time
// entry booked to the agent user of the human named by user_email, keyed on
// the agent's name. agentwork.Record picks the project, parking unmatched
// sessions in the shared "Agent work" project so none is ever dropped.
// Entries upsert by facile_id, making re-emitted history idempotent even
// without the processed-events ledger.
func (s *Service) handleAgentSession(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt envelope.Event[envelope.AgentSession]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode agent session", slog.Any("error", err))
		return
	}
	if !s.IsPoolEventEnabled(meta.Channel) {
		return
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return
	}

	p := evt.Payload
	if p.FacileID == "" || p.Project == "" {
		s.logger.Warn("antenne: agent session missing facile_id or project")
		return
	}
	started, err := time.Parse(time.RFC3339, p.StartedAt)
	if err != nil {
		s.logger.Warn("antenne: agent session has invalid started_at", slog.String("value", p.StartedAt))
		return
	}
	stopped, err := time.Parse(time.RFC3339, p.StoppedAt)
	if err != nil || stopped.Before(started) {
		s.logger.Warn("antenne: agent session has invalid stopped_at", slog.String("value", p.StoppedAt))
		return
	}

	userID := s.resolveUser(s.orm, "", p.UserEmail)
	if userID == nil {
		s.logger.Warn("antenne: no user for agent session email",
			slog.String("email", p.UserEmail), slog.String("facile_id", p.FacileID))
		return
	}

	description := fmt.Sprintf("%s%s@%s", schemas.MyceliumDescriptionPrefix, p.Agent, p.Machine)
	if p.Branch != "" {
		description += " branch=" + p.Branch
	}

	result, err := agentwork.Record(context.Background(), s.orm, s.resolver, agentwork.Session{
		OwnerID:     *userID,
		Agent:       schemas.AgentFromName(p.Agent),
		FacileID:    p.FacileID,
		Facts:       attribution.Facts{ProjectHint: p.Project},
		StartedAt:   started,
		StoppedAt:   stopped,
		Description: description,
	})
	if err != nil {
		s.logger.Error("antenne: failed to record agent session",
			slog.String("facile_id", p.FacileID), slog.Any("error", err))
		return
	}

	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced agent session",
		slog.String("facile_id", p.FacileID),
		slog.String("agent", result.Agent.Name),
		slog.String("project", result.Target.Project.Name),
		slog.String("task", result.Target.Task.Name))
}
