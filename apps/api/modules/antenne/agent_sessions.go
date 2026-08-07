package antenne

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"time"

	enveloppe "github.com/FacileStudio/enveloppe/go"
	antenneclient "github.com/FacileStudio/antenne-client/go"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/gorm"
)

const (
	agentSessionsTaskName    = "Agent sessions"
	agentCatchAllProject     = "Agent work"
	agentCatchAllOwnerID     = 1
	agentCatchAllProjectDesc = "Agent sessions from Jardin without a matching Sablier project"
)

// handleAgentSession materializes a Jardin agent_session event as a time
// entry. A project whose name matches the session's project gets the entry
// under its "Agent sessions" task; anything else is parked in the shared
// "Agent work" project under a task named after the source project, so no
// session is ever dropped. Entries upsert by facile_id, making re-emitted
// history idempotent even without the processed-events ledger.
func (s *Service) handleAgentSession(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt enveloppe.Event[enveloppe.AgentSession]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("pool: failed to decode agent session", slog.Any("error", err))
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
		s.logger.Warn("pool: agent session missing facile_id or project")
		return
	}
	started, err := time.Parse(time.RFC3339, p.StartedAt)
	if err != nil {
		s.logger.Warn("pool: agent session has invalid started_at", slog.String("value", p.StartedAt))
		return
	}
	stopped, err := time.Parse(time.RFC3339, p.StoppedAt)
	if err != nil || stopped.Before(started) {
		s.logger.Warn("pool: agent session has invalid stopped_at", slog.String("value", p.StoppedAt))
		return
	}

	userID := s.resolveActorByEmail(p.UserEmail)
	if userID == nil {
		s.logger.Warn("pool: no user for agent session email",
			slog.String("email", p.UserEmail), slog.String("facile_id", p.FacileID))
		return
	}

	project, taskName := s.agentSessionTarget(p.Project)
	if project == nil {
		s.logger.Error("pool: failed to resolve target project for agent session",
			slog.String("project", p.Project))
		return
	}
	task, err := s.ensureAgentTask(project.ID, taskName)
	if err != nil {
		s.logger.Error("pool: failed to ensure agent session task", slog.Any("error", err))
		return
	}

	description := fmt.Sprintf("jardin: %s@%s", p.Agent, p.Machine)
	if p.Branch != "" {
		description += " branch=" + p.Branch
	}

	var existing schemas.TimeEntry
	err = s.orm.Where("facile_id = ?", p.FacileID).First(&existing).Error
	switch {
	case err == nil:
		updates := map[string]interface{}{
			"project_id": project.ID,
			"task_id":    task.ID,
			"user_id":    *userID,
			"started_at": started.UTC(),
			"stopped_at": stopped.UTC(),
		}
		if err := s.orm.Model(&existing).Updates(updates).Error; err != nil {
			s.logger.Error("pool: failed to update agent session entry", slog.Any("error", err))
			return
		}
	case errors.Is(err, gorm.ErrRecordNotFound):
		facileID := p.FacileID
		stoppedAt := stopped.UTC()
		record := schemas.TimeEntry{
			ProjectID:         project.ID,
			TaskID:            task.ID,
			UserID:            *userID,
			FacileID:          &facileID,
			LegacyDescription: description,
			StartedAt:         started.UTC(),
			StoppedAt:         &stoppedAt,
		}
		if err := s.orm.Create(&record).Error; err != nil {
			s.logger.Error("pool: failed to create agent session entry", slog.Any("error", err))
			return
		}
	default:
		s.logger.Error("pool: failed to look up agent session entry", slog.Any("error", err))
		return
	}

	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("pool: synced agent session",
		slog.String("facile_id", p.FacileID),
		slog.String("project", project.Name),
		slog.String("task", task.Name))
}

func (s *Service) agentSessionTarget(name string) (*schemas.Project, string) {
	var project schemas.Project
	err := s.orm.Where("lower(name) = lower(?)", name).First(&project).Error
	if err == nil {
		return &project, agentSessionsTaskName
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ""
	}

	err = s.orm.Where("lower(name) = lower(?)", agentCatchAllProject).First(&project).Error
	if err == nil {
		return &project, name
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, ""
	}

	facileID := GenerateFacileID()
	record := schemas.Project{
		Name:        agentCatchAllProject,
		Description: agentCatchAllProjectDesc,
		OwnerID:     agentCatchAllOwnerID,
		FacileID:    &facileID,
	}
	if err := s.orm.Create(&record).Error; err != nil {
		if err := s.orm.Where("lower(name) = lower(?)", agentCatchAllProject).First(&record).Error; err != nil {
			return nil, ""
		}
	}
	return &record, name
}

func (s *Service) ensureAgentTask(projectID int64, name string) (*schemas.Task, error) {
	var task schemas.Task
	err := s.orm.Where("project_id = ? AND lower(name) = lower(?)", projectID, name).First(&task).Error
	if err == nil {
		return &task, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	facileID := GenerateFacileID()
	task = schemas.Task{ProjectID: projectID, Name: name, FacileID: &facileID}
	if err := s.orm.Create(&task).Error; err != nil {
		if retryErr := s.orm.Where("project_id = ? AND lower(name) = lower(?)", projectID, name).First(&task).Error; retryErr != nil {
			return nil, err
		}
	}
	return &task, nil
}
