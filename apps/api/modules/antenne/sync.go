package antenne

import (
	"encoding/json"
	"errors"
	"log/slog"

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

type projectWire struct {
	enveloppe.Project
	Title string `json:"title"`
}

type taskWire struct {
	enveloppe.Task
	Title string `json:"title"`
}

func (p projectWire) name() string {
	if p.Name != "" {
		return p.Name
	}
	return p.Title
}

func (t taskWire) name() string {
	if t.Name != "" {
		return t.Name
	}
	return t.Title
}

func decodeInbound[T any](s *Service, channel string, payload json.RawMessage, facileID func(*T) string) (*enveloppe.Event[T], string, bool) {
	var evt enveloppe.Event[T]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode "+channel, slog.Any("error", err))
		return nil, "", false
	}
	if evt.App == enveloppe.AppSablier {
		return nil, "", false
	}
	if !s.IsPoolEventEnabled(channel) {
		return nil, "", false
	}
	if s.alreadyProcessed(evt.IdempotencyKey) {
		return nil, "", false
	}
	id := facileID(&evt.Payload)
	if id == "" {
		id = evt.FacileID
	}
	if id == "" {
		s.logger.Warn("antenne: event without facile_id dropped", slog.String("channel", channel))
		return nil, "", false
	}
	return &evt, id, true
}

func (s *Service) handleProjectCreated(payload json.RawMessage, meta antenneclient.EventMeta) {
	s.applyProject("project.created", payload)
}

func (s *Service) handleProjectUpdated(payload json.RawMessage, meta antenneclient.EventMeta) {
	s.applyProject("project.updated", payload)
}

func (s *Service) applyProject(channel string, payload json.RawMessage) {
	evt, facileID, ok := decodeInbound(s, channel, payload, func(p *projectWire) string { return p.FacileID })
	if !ok {
		return
	}
	p := evt.Payload

	var record schemas.Project
	err := s.orm.Where("facile_id = ?", facileID).First(&record).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("antenne: failed to find project", slog.Any("error", err))
		return
	}

	if errors.Is(err, gorm.ErrRecordNotFound) {
		if p.name() == "" {
			s.logger.Warn("antenne: project without name dropped", slog.String("facile_id", facileID))
			return
		}
		desc := ""
		if p.Description != nil {
			desc = *p.Description
		}
		record = schemas.Project{
			Name:        p.name(),
			Description: desc,
			Icon:        normalizeIcon(p.Icon),
			OwnerID:     1,
			FacileID:    &facileID,
		}
		if err := s.orm.Create(&record).Error; err != nil {
			s.logger.Error("antenne: failed to create synced project", slog.Any("error", err))
			return
		}
		s.markProcessed(evt.IdempotencyKey)
		s.logger.Info("antenne: synced project created", slog.String("facile_id", facileID))
		return
	}

	if p.name() != "" {
		record.Name = p.name()
	}
	if p.Description != nil {
		record.Description = *p.Description
	}
	if p.Icon != nil {
		record.Icon = normalizeIcon(p.Icon)
	}
	if err := s.orm.Save(&record).Error; err != nil {
		s.logger.Error("antenne: failed to update synced project", slog.Any("error", err))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced project updated", slog.String("facile_id", facileID))
}

func (s *Service) handleProjectDeleted(payload json.RawMessage, meta antenneclient.EventMeta) {
	evt, facileID, ok := decodeInbound(s, "project.deleted", payload, func(p *projectWire) string { return p.FacileID })
	if !ok {
		return
	}
	err := s.orm.Transaction(func(tx *gorm.DB) error {
		var project schemas.Project
		if err := tx.Where("facile_id = ?", facileID).First(&project).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		taskIDs := tx.Model(&schemas.Task{}).Select("id").Where("project_id = ?", project.ID)
		if err := tx.Model(&schemas.TimeEntry{}).Where("task_id IN (?)", taskIDs).Update("task_id", 0).Error; err != nil {
			return err
		}
		if err := tx.Where("project_id = ?", project.ID).Delete(&schemas.Task{}).Error; err != nil {
			return err
		}
		return tx.Delete(&project).Error
	})
	if err != nil {
		s.logger.Error("antenne: failed to delete synced project", slog.Any("error", err))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced project deleted", slog.String("facile_id", facileID))
}

func (s *Service) handleTaskCreated(payload json.RawMessage, meta antenneclient.EventMeta) {
	s.applyTask("task.created", payload)
}

func (s *Service) handleTaskUpdated(payload json.RawMessage, meta antenneclient.EventMeta) {
	s.applyTask("task.updated", payload)
}

func (s *Service) applyTask(channel string, payload json.RawMessage) {
	evt, facileID, ok := decodeInbound(s, channel, payload, func(t *taskWire) string { return t.FacileID })
	if !ok {
		return
	}
	t := evt.Payload
	status, known := mapRemoteStatus(t.Status)

	var existing schemas.Task
	err := s.orm.Where("facile_id = ?", facileID).First(&existing).Error
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		s.logger.Error("antenne: failed to find task", slog.Any("error", err))
		return
	}

	if err == nil {
		updates := map[string]interface{}{}
		if name := t.name(); name != "" && name != existing.Name {
			updates["name"] = name
		}
		if known && status != existing.Status {
			updates["status"] = status
		}
		if actorID := s.resolveActorByEmail(t.ActorEmail); actorID != nil {
			if existing.ActorID == nil || *existing.ActorID != *actorID {
				updates["actor_id"] = *actorID
			}
		}
		if len(updates) > 0 {
			if err := s.orm.Model(&existing).Updates(updates).Error; err != nil {
				s.logger.Error("antenne: failed to update synced task", slog.Any("error", err))
				return
			}
			s.logger.Info("antenne: synced task updated", slog.String("facile_id", facileID))
		}
		s.markProcessed(evt.IdempotencyKey)
		return
	}

	var project schemas.Project
	if err := s.orm.Where("facile_id = ?", t.ProjectFacileID).First(&project).Error; err != nil {
		s.logger.Warn("antenne: parent project unknown, task dropped",
			slog.String("project_facile_id", t.ProjectFacileID),
			slog.String("task_facile_id", facileID))
		return
	}
	if t.name() == "" {
		s.logger.Warn("antenne: task without name dropped", slog.String("facile_id", facileID))
		return
	}
	if !known {
		status = schemas.StatusTodo
	}
	record := schemas.Task{
		ProjectID: project.ID,
		Name:      t.name(),
		Status:    status,
		FacileID:  &facileID,
		ActorID:   s.resolveActorByEmail(t.ActorEmail),
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

func (s *Service) handleTaskDeleted(payload json.RawMessage, meta antenneclient.EventMeta) {
	evt, facileID, ok := decodeInbound(s, "task.deleted", payload, func(t *taskWire) string { return t.FacileID })
	if !ok {
		return
	}
	err := s.orm.Transaction(func(tx *gorm.DB) error {
		var task schemas.Task
		if err := tx.Where("facile_id = ?", facileID).First(&task).Error; err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return nil
			}
			return err
		}
		if err := tx.Model(&schemas.TimeEntry{}).Where("task_id = ?", task.ID).Update("task_id", 0).Error; err != nil {
			return err
		}
		return tx.Delete(&task).Error
	})
	if err != nil {
		s.logger.Error("antenne: failed to delete synced task", slog.Any("error", err))
		return
	}
	s.markProcessed(evt.IdempotencyKey)
	s.logger.Info("antenne: synced task deleted", slog.String("facile_id", facileID))
}
