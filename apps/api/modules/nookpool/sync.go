package nookpool

import (
	"encoding/json"
	"log/slog"
	"time"

	enveloppe "github.com/FacileStudio/enveloppe/go"
	pool "github.com/FacileStudio/pool/go"
	"api/schemas"

	"gorm.io/gorm"
)

func (s *Service) handleProjectCreated(payload json.RawMessage, meta pool.EventMeta) {
	var evt enveloppe.Event[enveloppe.Project]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("pool: failed to decode project.created", slog.Any("error", err))
		return
	}

	var existing schemas.Project
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&existing).Error; err == nil {
		return
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
		s.logger.Error("pool: failed to create synced project", slog.Any("error", err))
		return
	}
	s.logger.Info("pool: synced project created", slog.String("facile_id", facileID))
}

func (s *Service) handleProjectUpdated(payload json.RawMessage, meta pool.EventMeta) {
	var evt enveloppe.Event[enveloppe.Project]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("pool: failed to decode project.updated", slog.Any("error", err))
		return
	}

	var record schemas.Project
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("pool: project not found for update", slog.String("facile_id", evt.Payload.FacileID))
			return
		}
		s.logger.Error("pool: failed to find project for update", slog.Any("error", err))
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
		s.logger.Error("pool: failed to update synced project", slog.Any("error", err))
		return
	}
	s.logger.Info("pool: synced project updated", slog.String("facile_id", evt.Payload.FacileID))
}

func (s *Service) handleProjectDeleted(payload json.RawMessage, meta pool.EventMeta) {
	var evt enveloppe.Event[enveloppe.Project]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("pool: failed to decode project.deleted", slog.Any("error", err))
		return
	}

	result := s.orm.Where("facile_id = ?", evt.Payload.FacileID).Delete(&schemas.Project{})
	if result.Error != nil {
		s.logger.Error("pool: failed to delete synced project", slog.Any("error", result.Error))
		return
	}
	s.logger.Info("pool: synced project deleted", slog.String("facile_id", evt.Payload.FacileID))
}

func (s *Service) handleTaskCreated(payload json.RawMessage, meta pool.EventMeta) {
	var evt enveloppe.Event[enveloppe.Task]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("pool: failed to decode task.created", slog.Any("error", err))
		return
	}

	var existing schemas.Task
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&existing).Error; err == nil {
		return
	}

	var project schemas.Project
	var found bool
	for attempt := 0; attempt < 5; attempt++ {
		if err := s.orm.Where("facile_id = ?", evt.Payload.ProjectFacileID).First(&project).Error; err == nil {
			found = true
			break
		}
		time.Sleep(1 * time.Second)
	}
	if !found {
		s.logger.Warn("pool: parent project not found for task sync after retries",
			slog.String("project_facile_id", evt.Payload.ProjectFacileID),
			slog.String("task_facile_id", evt.Payload.FacileID))
		return
	}

	facileID := evt.Payload.FacileID
	status := normalizeStatus(evt.Payload.Status)
	record := schemas.Task{
		ProjectID: project.ID,
		Name:      evt.Payload.Name,
		Status:    status,
		FacileID:  &facileID,
	}
	if err := s.orm.Create(&record).Error; err != nil {
		s.logger.Error("pool: failed to create synced task", slog.Any("error", err))
		return
	}
	s.logger.Info("pool: synced task created", slog.String("facile_id", facileID))
}

func (s *Service) handleTaskUpdated(payload json.RawMessage, meta pool.EventMeta) {
	var evt enveloppe.Event[enveloppe.Task]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("pool: failed to decode task.updated", slog.Any("error", err))
		return
	}

	var record schemas.Task
	if err := s.orm.Where("facile_id = ?", evt.Payload.FacileID).First(&record).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			s.logger.Warn("pool: task not found for update", slog.String("facile_id", evt.Payload.FacileID))
			return
		}
		s.logger.Error("pool: failed to find task for update", slog.Any("error", err))
		return
	}

	if evt.Payload.Name != "" {
		record.Name = evt.Payload.Name
	}
	if evt.Payload.Status != "" {
		record.Status = normalizeStatus(evt.Payload.Status)
	}
	if err := s.orm.Save(&record).Error; err != nil {
		s.logger.Error("pool: failed to update synced task", slog.Any("error", err))
		return
	}
	s.logger.Info("pool: synced task updated", slog.String("facile_id", evt.Payload.FacileID))
}

func (s *Service) handleTaskDeleted(payload json.RawMessage, meta pool.EventMeta) {
	var evt enveloppe.Event[enveloppe.Task]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("pool: failed to decode task.deleted", slog.Any("error", err))
		return
	}

	result := s.orm.Where("facile_id = ?", evt.Payload.FacileID).Delete(&schemas.Task{})
	if result.Error != nil {
		s.logger.Error("pool: failed to delete synced task", slog.Any("error", result.Error))
		return
	}
	s.logger.Info("pool: synced task deleted", slog.String("facile_id", evt.Payload.FacileID))
}
