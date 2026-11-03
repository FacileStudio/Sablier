package antenne

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
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

var stableFacileID = regexp.MustCompile(`^(fac_[0-9a-f]{20}|[A-Z2-7]{26}|[0-9a-f]{32})$`)

var errDropped = errors.New("antenne: event dropped")

const ledgerRetention = 35 * 24 * time.Hour

func isStale(timestamp string) bool {
	at, err := time.Parse(time.RFC3339, timestamp)
	return err != nil || time.Since(at) > ledgerRetention
}

func isStableFacileID(id string) bool {
	return stableFacileID.MatchString(id)
}

func claimEvent(tx *gorm.DB, key string) (bool, error) {
	if key == "" {
		return true, nil
	}
	result := tx.Clauses(clause.OnConflict{DoNothing: true}).Create(&schemas.AntenneProcessedEvent{IdempotencyKey: key})
	return result.RowsAffected == 1, result.Error
}

func (s *Service) applyInTx(channel, key string, fn func(tx *gorm.DB) error) {
	err := s.orm.Transaction(func(tx *gorm.DB) error {
		fresh, err := claimEvent(tx, key)
		if err != nil || !fresh {
			return err
		}
		return fn(tx)
	})
	if err != nil && !errors.Is(err, errDropped) {
		s.logger.Error("antenne: failed to apply "+channel, slog.Any("error", err))
	}
}

func (s *Service) drop(reason string, attrs ...any) error {
	s.logger.Warn("antenne: "+reason, attrs...)
	return errDropped
}

func uniqueTaskName(tx *gorm.DB, projectID, selfID int64, name string) (string, error) {
	candidate := name
	for n := 2; ; n++ {
		var count int64
		err := tx.Model(&schemas.Task{}).
			Where("project_id = ? AND lower(name) = lower(?) AND id <> ?", projectID, candidate, selfID).
			Count(&count).Error
		if err != nil {
			return "", err
		}
		if count == 0 {
			return candidate, nil
		}
		candidate = fmt.Sprintf("%s (%d)", name, n)
	}
}

func decodeInbound[T any](s *Service, channel string, payload json.RawMessage, facileID func(*T) string) (*enveloppe.Event[T], string, bool) {
	var evt enveloppe.Event[T]
	if err := json.Unmarshal(payload, &evt); err != nil {
		s.logger.Error("antenne: failed to decode "+channel, slog.Any("error", err))
		return nil, "", false
	}
	if evt.App != enveloppe.AppOpus {
		return nil, "", false
	}
	if !s.IsPoolEventEnabled(channel) {
		return nil, "", false
	}
	if isStale(evt.Timestamp) {
		s.logger.Warn("antenne: stale event dropped", slog.String("channel", channel), slog.String("timestamp", evt.Timestamp))
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

	s.applyInTx(channel, evt.IdempotencyKey, func(tx *gorm.DB) error {
		var record schemas.Project
		err := tx.Where("facile_id = ?", facileID).First(&record).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if errors.Is(err, gorm.ErrRecordNotFound) {
			if channel != "project.created" {
				return s.drop("update for unknown project dropped", slog.String("facile_id", facileID))
			}
			if !isStableFacileID(facileID) {
				return s.drop("project with unstable facile_id dropped", slog.String("facile_id", facileID))
			}
			if p.name() == "" {
				return s.drop("project without name dropped", slog.String("facile_id", facileID))
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
			if err := tx.Create(&record).Error; err != nil {
				return err
			}
			s.logger.Info("antenne: synced project created", slog.String("facile_id", facileID))
			return nil
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
		if err := tx.Save(&record).Error; err != nil {
			return err
		}
		s.logger.Info("antenne: synced project updated", slog.String("facile_id", facileID))
		return nil
	})
}

func (s *Service) handleProjectDeleted(payload json.RawMessage, meta antenneclient.EventMeta) {
	evt, facileID, ok := decodeInbound(s, "project.deleted", payload, func(p *projectWire) string { return p.FacileID })
	if !ok {
		return
	}
	s.applyInTx("project.deleted", evt.IdempotencyKey, func(tx *gorm.DB) error {
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
		if err := tx.Delete(&project).Error; err != nil {
			return err
		}
		s.logger.Info("antenne: synced project deleted", slog.String("facile_id", facileID))
		return nil
	})
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

	s.applyInTx(channel, evt.IdempotencyKey, func(tx *gorm.DB) error {
		var existing schemas.Task
		err := tx.Where("facile_id = ?", facileID).First(&existing).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}

		if err == nil {
			updates := map[string]interface{}{}
			if name := t.name(); name != "" && name != existing.Name {
				unique, err := uniqueTaskName(tx, existing.ProjectID, existing.ID, name)
				if err != nil {
					return err
				}
				if unique != existing.Name {
					updates["name"] = unique
				}
			}
			if known && status != existing.Status {
				updates["status"] = status
			}
			if actorID := s.resolveActorByEmail(tx, t.ActorEmail); actorID != nil {
				if existing.ActorID == nil || *existing.ActorID != *actorID {
					updates["actor_id"] = *actorID
				}
			}
			if len(updates) > 0 {
				if err := tx.Model(&existing).Updates(updates).Error; err != nil {
					return err
				}
				s.logger.Info("antenne: synced task updated", slog.String("facile_id", facileID))
			}
			return nil
		}

		if !isStableFacileID(facileID) {
			return s.drop("task with unstable facile_id dropped", slog.String("facile_id", facileID))
		}
		var project schemas.Project
		if err := tx.Where("facile_id = ?", t.ProjectFacileID).First(&project).Error; err != nil {
			return s.drop("parent project unknown, task dropped",
				slog.String("project_facile_id", t.ProjectFacileID),
				slog.String("task_facile_id", facileID))
		}
		if t.name() == "" {
			return s.drop("task without name dropped", slog.String("facile_id", facileID))
		}
		if !known {
			status = schemas.StatusTodo
		}
		name, err := uniqueTaskName(tx, project.ID, 0, t.name())
		if err != nil {
			return err
		}
		record := schemas.Task{
			ProjectID: project.ID,
			Name:      name,
			Status:    status,
			FacileID:  &facileID,
			ActorID:   s.resolveActorByEmail(tx, t.ActorEmail),
		}
		if err := tx.Create(&record).Error; err != nil {
			return err
		}
		s.logger.Info("antenne: synced task created", slog.String("facile_id", facileID))
		return nil
	})
}

func (s *Service) resolveActorByEmail(tx *gorm.DB, email string) *int64 {
	if email == "" {
		return nil
	}
	var user schemas.User
	if err := tx.Select("id").Where("email = ? AND kind <> ?", email, schemas.UserKindAgent).First(&user).Error; err != nil {
		return nil
	}
	return &user.ID
}

func (s *Service) handleTaskDeleted(payload json.RawMessage, meta antenneclient.EventMeta) {
	evt, facileID, ok := decodeInbound(s, "task.deleted", payload, func(t *taskWire) string { return t.FacileID })
	if !ok {
		return
	}
	s.applyInTx("task.deleted", evt.IdempotencyKey, func(tx *gorm.DB) error {
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
		if err := tx.Delete(&task).Error; err != nil {
			return err
		}
		s.logger.Info("antenne: synced task deleted", slog.String("facile_id", facileID))
		return nil
	})
}
