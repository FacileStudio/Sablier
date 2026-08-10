package antenne

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	antenneclient "github.com/FacileStudio/antenne-client/go"
	enveloppe "github.com/FacileStudio/enveloppe/go"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

func (s *Service) alreadyProcessed(key string) bool {
	if key == "" {
		return false
	}
	var count int64
	s.orm.Model(&schemas.AntenneProcessedEvent{}).Where("idempotency_key = ?", key).Count(&count)
	return count > 0
}

func (s *Service) markProcessed(key string) {
	if key == "" {
		return
	}
	s.orm.Clauses(clause.OnConflict{DoNothing: true}).Create(&schemas.AntenneProcessedEvent{IdempotencyKey: key})
}

func (s *Service) handleProjectCreated(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt enveloppe.Event[enveloppe.Project]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode project.created", slog.Any("error", err))
		return
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return
	}

	if s.createSyncedProject(&evt) {
		s.markProcessed(evt.IdempotencyKey)
	}
}

func (s *Service) createSyncedProject(evt *enveloppe.Event[enveloppe.Project]) bool {
	var existing schemas.Project
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&existing).Error; err == nil {
		return true
	}

	desc := ""
	if evt.Payload.Description != nil {
		desc = *evt.Payload.Description
	}

	facileID := evt.Payload.FacileID
	record := schemas.Project{
		Name:        evt.Payload.Name,
		Description: desc,
		Icon:        normalizeIcon(evt.Payload.Icon),
		OwnerID:     1,
		FacileID:    &facileID,
	}
	if err := s.orm.Create(&record).Error; err != nil {
		s.logger.Error("antenne: failed to create synced project", slog.Any("error", err))
		return false
	}
	s.logger.Info("antenne: synced project created", slog.String("facile_id", facileID))
	return true
}

func (s *Service) handleProjectUpdated(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt enveloppe.Event[enveloppe.Project]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode project.updated", slog.Any("error", err))
		return
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return
	}

	var record schemas.Project
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Info("antenne: project not found for update, creating it", slog.String("facile_id", evt.Payload.FacileID))
			if s.createSyncedProject(&evt) {
				s.markProcessed(evt.IdempotencyKey)
			}
			return
		}
		s.logger.Error("antenne: failed to find project for update", slog.Any("error", err))
		return
	}

	record.Name = evt.Payload.Name
	if evt.Payload.Description != nil {
		record.Description = *evt.Payload.Description
	}
	if evt.Payload.Icon != nil {
		record.Icon = normalizeIcon(evt.Payload.Icon)
	}
	if err := s.orm.Save(&record).Error; err != nil {
		s.logger.Error("antenne: failed to update synced project", slog.Any("error", err))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced project updated", slog.String("facile_id", evt.Payload.FacileID))
}

func (s *Service) handleProjectDeleted(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt enveloppe.Event[enveloppe.Project]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode project.deleted", slog.Any("error", err))
		return
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return
	}

	result := s.orm.Where("facile_id = ?", evt.Payload.FacileID).Delete(&schemas.Project{})
	if result.Error != nil {
		s.logger.Error("antenne: failed to delete synced project", slog.Any("error", result.Error))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced project deleted", slog.String("facile_id", evt.Payload.FacileID))
}

func (s *Service) handleTaskCreated(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt enveloppe.Event[enveloppe.Task]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode task.created", slog.Any("error", err))
		return
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return
	}

	s.upsertSyncedTask(&evt, 0)
}

const maxTaskParentRetries = 5

// upsertSyncedTask creates or updates a task from a pool event. When the
// parent project has not been synced yet, it reschedules itself with
// time.AfterFunc instead of sleeping: handlers run on the pool client's read
// loop, so a blocking wait here would also block the project.created event
// it is waiting for.
func (s *Service) upsertSyncedTask(evt *enveloppe.Event[enveloppe.Task], attempt int) {
	var existing schemas.Task
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&existing).Error; err == nil {
		updates := map[string]interface{}{}
		if evt.Payload.Name != "" && evt.Payload.Name != existing.Name {
			updates["name"] = evt.Payload.Name
		}
		if evt.Payload.Status != "" {
			normalized := schemas.NormalizeStatus(evt.Payload.Status)
			if normalized != existing.Status {
				updates["status"] = normalized
			}
		}
		if actorID := s.resolveActorByEmail(evt.Payload.ActorEmail); actorID != nil {
			if existing.ActorID == nil || *existing.ActorID != *actorID {
				updates["actor_id"] = *actorID
			}
		}
		if len(updates) > 0 {
			if err := s.orm.Model(&existing).Updates(updates).Error; err != nil {
				s.logger.Error("antenne: failed to upsert synced task", slog.Any("error", err))
				return
			}
			s.logger.Info("antenne: synced task updated", slog.String("facile_id", evt.Payload.FacileID))
		}
		s.markProcessed(evt.IdempotencyKey)
		return
	}

	var project schemas.Project
	if err := s.orm.Where("facile_id = ?", evt.Payload.ProjectFacileID).First(&project).Error; err != nil {
		if attempt < maxTaskParentRetries {
			time.AfterFunc(1*time.Second, func() {
				s.upsertSyncedTask(evt, attempt+1)
			})
			return
		}
		s.logger.Warn("antenne: parent project not found for task sync after retries",
			slog.String("project_facile_id", evt.Payload.ProjectFacileID),
			slog.String("task_facile_id", evt.Payload.FacileID))
		return
	}

	facileID := evt.Payload.FacileID
	status := schemas.NormalizeStatus(evt.Payload.Status)
	record := schemas.Task{
		ProjectID: project.ID,
		Name:      evt.Payload.Name,
		Status:    status,
		FacileID:  &facileID,
		ActorID:   s.resolveActorByEmail(evt.Payload.ActorEmail),
	}
	if err := s.orm.Create(&record).Error; err != nil {
		s.logger.Error("antenne: failed to create synced task", slog.Any("error", err))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced task created", slog.String("facile_id", facileID))
}

func (s *Service) resolveActorByEmail(email string) *int64 {
	if email == "" {
		return nil
	}
	var user schemas.User
	if err := s.orm.Select("id").Where("email = ?", email).First(&user).Error; err != nil {
		return nil
	}
	return &user.ID
}

func (s *Service) handleTaskUpdated(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt enveloppe.Event[enveloppe.Task]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode task.updated", slog.Any("error", err))
		return
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return
	}

	var record schemas.Task
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Info("antenne: task not found for update, creating it", slog.String("facile_id", evt.Payload.FacileID))
			s.upsertSyncedTask(&evt, 0)
			return
		}
		s.logger.Error("antenne: failed to find task for update", slog.Any("error", err))
		return
	}

	if evt.Payload.Name != "" {
		record.Name = evt.Payload.Name
	}
	if evt.Payload.Status != "" {
		record.Status = schemas.NormalizeStatus(evt.Payload.Status)
	}
	if err := s.orm.Save(&record).Error; err != nil {
		s.logger.Error("antenne: failed to update synced task", slog.Any("error", err))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced task updated", slog.String("facile_id", evt.Payload.FacileID))
}

func (s *Service) handleTaskDeleted(payload json.RawMessage, meta antenneclient.EventMeta) {
	var evt enveloppe.Event[enveloppe.Task]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode task.deleted", slog.Any("error", err))
		return
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return
	}

	result := s.orm.Where("facile_id = ?", evt.Payload.FacileID).Delete(&schemas.Task{})
	if result.Error != nil {
		s.logger.Error("antenne: failed to delete synced task", slog.Any("error", result.Error))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced task deleted", slog.String("facile_id", evt.Payload.FacileID))
}
