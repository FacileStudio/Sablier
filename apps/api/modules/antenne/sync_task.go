package antenne

import (
	"encoding/json"
	"errors"
	"log/slog"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	antenneclient "github.com/FacileStudio/antenne-client/go"

	"gorm.io/gorm"
)

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
	at, coarse := eventTime(evt.Payload.UpdatedAt, evt.Timestamp)
	change := inboundChange[taskWire]{channel: channel, facileID: facileID, wire: evt.Payload, at: at, coarse: coarse}
	var adopted int64
	s.applyInTx(channel, evt.IdempotencyKey, func(tx *gorm.DB) error {
		var existing schemas.Task
		err := tx.Where("facile_id = ?", facileID).First(&existing).Error
		if err == nil {
			return s.updateTaskFromWire(tx, &existing, &change)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		adopted, err = s.createTaskFromWire(tx, &change)
		return err
	})
	if adopted != 0 {
		s.resendAdopted(0, adopted)
	}
}

func (s *Service) updateTaskFromWire(tx *gorm.DB, existing *schemas.Task, change *inboundChange[taskWire]) error {
	if isOlder(change.at, change.coarse, existing.UpdatedAt) {
		s.logger.Info("antenne: older task update ignored", slog.String("facile_id", change.facileID))
		return nil
	}
	t := change.wire
	updates := s.taskFieldUpdates(tx, existing, &t)
	updates["synced_at"] = time.Now().UTC()
	if !change.coarse {
		updates["updated_at"] = change.at
	}
	projectID, err := s.moveTask(tx, existing, t.ProjectFacileID, updates)
	if err != nil {
		return err
	}
	if err := renameTask(tx, existing, projectID, t.name(), updates); err != nil {
		return err
	}
	if err := tx.Model(existing).Updates(updates).Error; err != nil {
		return err
	}
	s.logger.Info("antenne: synced task updated", slog.String("facile_id", change.facileID))
	return nil
}

// taskFieldUpdates collects the fields Opus owns. A field left out keeps its value; an
// unknown status slug or priority keeps the current one.
func (s *Service) taskFieldUpdates(tx *gorm.DB, existing *schemas.Task, t *taskWire) map[string]any {
	updates := map[string]any{}
	if status, known := mapRemoteStatus(t.Status); known {
		updates["status"] = status
	}
	if actorID := s.resolveUser(tx, t.ActorSubject, t.ActorEmail); actorID != nil {
		updates["actor_id"] = *actorID
	}
	if t.Description != nil {
		updates["description"] = *t.Description
	}
	if due, ok := normalizeDueDate(t.DueDate.value); t.DueDate.set && ok {
		updates["due_date"] = due
	}
	if t.Priority != nil {
		if priority, ok := normalizePriority(*t.Priority); ok {
			updates["priority"] = priority
		}
	}
	return updates
}

// moveTask applies a changed project_facile_id: the task and its time entries move.
func (s *Service) moveTask(tx *gorm.DB, existing *schemas.Task, projectFacileID string, updates map[string]any) (int64, error) {
	if projectFacileID == "" {
		return existing.ProjectID, nil
	}
	var project schemas.Project
	if err := tx.Where("facile_id = ?", projectFacileID).First(&project).Error; err != nil {
		return 0, s.drop("task move target unknown, update dropped",
			slog.String("project_facile_id", projectFacileID), slog.String("task_id", *existing.FacileID))
	}
	if project.ID == existing.ProjectID {
		return project.ID, nil
	}
	updates["project_id"] = project.ID
	err := tx.Model(&schemas.TimeEntry{}).Where("task_id = ?", existing.ID).
		UpdateColumns(map[string]any{"project_id": project.ID, "space_id": project.SpaceID}).Error
	return project.ID, err
}

// renameTask applies the name Opus sent, keeping it as synced_name when the local name
// needs a " (n)" suffix so the suffix never goes back on the wire.
func renameTask(tx *gorm.DB, existing *schemas.Task, projectID int64, name string, updates map[string]any) error {
	wireName := existing.Name
	if existing.SyncedName != "" {
		wireName = existing.SyncedName
	}
	if name == "" || (name == wireName && projectID == existing.ProjectID) {
		return nil
	}
	unique, err := uniqueTaskName(tx, projectID, existing.ID, name)
	if err != nil {
		return err
	}
	updates["name"] = unique
	updates["synced_name"] = syncedName(unique, name)
	return nil
}

func syncedName(local, wire string) string {
	if local == wire {
		return ""
	}
	return wire
}

// createTaskFromWire creates the task, or merges it into an unlinked twin in the same
// project when the incoming id is the smaller one. It returns the id of an adopted twin.
func (s *Service) createTaskFromWire(tx *gorm.DB, change *inboundChange[taskWire]) (int64, error) {
	t, facileID := change.wire, change.facileID
	if !isStableFacileID(facileID) {
		return 0, s.drop("task with unstable facile_id dropped", slog.String("facile_id", facileID))
	}
	var project schemas.Project
	if err := tx.Where("facile_id = ?", t.ProjectFacileID).First(&project).Error; err != nil {
		return 0, s.drop("parent project unknown, task dropped",
			slog.String("project_facile_id", t.ProjectFacileID), slog.String("task_facile_id", facileID))
	}
	if t.name() == "" {
		return 0, s.drop("task without name dropped", slog.String("facile_id", facileID))
	}
	twin, err := findTaskTwin(tx, project.ID, t.name(), facileID)
	if err != nil || twin != nil {
		return s.mergeTaskTwin(tx, twin, change, err)
	}
	return 0, s.insertTask(tx, project.ID, change)
}

func (s *Service) insertTask(tx *gorm.DB, projectID int64, change *inboundChange[taskWire]) error {
	t, facileID := change.wire, change.facileID
	name, err := uniqueTaskName(tx, projectID, 0, t.name())
	if err != nil {
		return err
	}
	status, known := mapRemoteStatus(t.Status)
	if !known {
		status = schemas.StatusTodo
	}
	now := time.Now().UTC()
	record := schemas.Task{ProjectID: projectID, Name: name, Status: status, FacileID: &facileID,
		ActorID: s.resolveUser(tx, t.ActorSubject, t.ActorEmail)}
	record.SyncedName = syncedName(name, t.name())
	record.SyncedAt = &now
	if err := tx.Create(&record).Error; err != nil {
		return err
	}
	if updates := s.taskFieldUpdates(tx, &record, &t); len(updates) > 0 {
		if err := tx.Model(&record).Updates(updates).Error; err != nil {
			return err
		}
	}
	s.logger.Info("antenne: synced task created", slog.String("facile_id", facileID))
	return nil
}

func findTaskTwin(tx *gorm.DB, projectID int64, name, facileID string) (*schemas.Task, error) {
	var twin schemas.Task
	err := tx.Where("project_id = ? AND lower(name) = lower(?) AND synced_at IS NULL", projectID, name).
		Where("facile_id IS NULL OR facile_id <> ?", facileID).Order("id").First(&twin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &twin, err
}

// mergeTaskTwin applies the tie-break: both sides converge on the smaller facile_id.
func (s *Service) mergeTaskTwin(tx *gorm.DB, twin *schemas.Task, change *inboundChange[taskWire], err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	if twin.FacileID != nil && change.facileID >= *twin.FacileID {
		s.logger.Info("antenne: task twin keeps its id", slog.String("facile_id", *twin.FacileID))
		return 0, nil
	}
	if err := adoptFacileID(tx, &schemas.Task{}, twin.ID, change.facileID); err != nil {
		return 0, err
	}
	twin.FacileID = &change.facileID
	return twin.ID, s.updateTaskFromWire(tx, twin, change)
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
		if err := tx.Model(&schemas.TimeEntry{}).Where("task_id = ?", task.ID).UpdateColumn("task_id", 0).Error; err != nil {
			return err
		}
		if err := tx.Delete(&task).Error; err != nil {
			return err
		}
		s.logger.Info("antenne: synced task deleted", slog.String("facile_id", facileID))
		return nil
	})
}
