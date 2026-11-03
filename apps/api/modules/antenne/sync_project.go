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

func (s *Service) handleProjectCreated(payload json.RawMessage, meta antenneclient.EventMeta) {
	s.applyProject("project.created", payload)
}

func (s *Service) handleProjectUpdated(payload json.RawMessage, meta antenneclient.EventMeta) {
	s.applyProject("project.updated", payload)
}

// inboundChange is one decoded inbound change with the instant it happened.
type inboundChange[T any] struct {
	channel  string
	facileID string
	wire     T
	at       time.Time
	coarse   bool
}

func (s *Service) applyProject(channel string, payload json.RawMessage) {
	evt, facileID, ok := decodeInbound(s, channel, payload, func(p *projectWire) string { return p.FacileID })
	if !ok {
		return
	}
	at, coarse := eventTime(evt.Payload.UpdatedAt, evt.Timestamp)
	change := inboundChange[projectWire]{channel: channel, facileID: facileID, wire: evt.Payload, at: at, coarse: coarse}
	var adopted int64
	s.applyInTx(channel, evt.IdempotencyKey, func(tx *gorm.DB) error {
		var record schemas.Project
		err := tx.Where("facile_id = ?", facileID).First(&record).Error
		if err == nil {
			return s.updateProjectFromWire(tx, &record, &change)
		}
		if !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		adopted, err = s.createProjectFromWire(tx, &change)
		return err
	})
	if adopted != 0 {
		s.resendAdopted(adopted, 0)
	}
}

func (s *Service) updateProjectFromWire(tx *gorm.DB, record *schemas.Project, change *inboundChange[projectWire]) error {
	if isOlder(change.at, change.coarse, record.UpdatedAt) {
		s.logger.Info("antenne: older project update ignored", slog.String("facile_id", change.facileID))
		return nil
	}
	p := change.wire
	updates := projectFieldUpdates(&p)
	if !change.coarse {
		updates["updated_at"] = change.at
	}
	if p.Space != nil {
		place, ok := s.resolvePlacement(tx, p.Space)
		if !ok {
			return s.drop("project space unresolved, update dropped", slog.String("facile_id", change.facileID))
		}
		if err := moveProject(tx, record, place, updates); err != nil {
			return err
		}
	}
	if err := tx.Model(record).Updates(updates).Error; err != nil {
		return err
	}
	s.logger.Info("antenne: synced project updated", slog.String("facile_id", change.facileID))
	return nil
}

// projectFieldUpdates collects the fields the wire carries; a field left out keeps its value.
func projectFieldUpdates(p *projectWire) map[string]any {
	updates := map[string]any{"synced_at": time.Now().UTC()}
	if p.name() != "" {
		updates["name"] = p.name()
	}
	if p.Description != nil {
		updates["description"] = *p.Description
	}
	if p.Icon != nil {
		updates["icon"] = normalizeIcon(p.Icon)
	}
	return updates
}

// moveProject adds a space change to updates and carries the project's time entries along.
func moveProject(tx *gorm.DB, record *schemas.Project, place placement, updates map[string]any) error {
	if place.same(record) {
		return nil
	}
	updates["space_id"] = place.spaceID
	updates["owner_id"] = place.ownerID
	return tx.Model(&schemas.TimeEntry{}).Where("project_id = ?", record.ID).
		UpdateColumn("space_id", place.spaceID).Error
}

// createProjectFromWire creates the project, or merges it into an unlinked local twin
// when the incoming id is the smaller one. It returns the id of an adopted twin.
func (s *Service) createProjectFromWire(tx *gorm.DB, change *inboundChange[projectWire]) (int64, error) {
	p, facileID := change.wire, change.facileID
	if change.channel != "project.created" {
		return 0, s.drop("update for unknown project dropped", slog.String("facile_id", facileID))
	}
	if !isStableFacileID(facileID) || p.name() == "" {
		return 0, s.drop("project with unstable facile_id or no name dropped", slog.String("facile_id", facileID))
	}
	place, ok := s.resolvePlacement(tx, p.Space)
	if !ok {
		return 0, s.drop("project space unresolved, create dropped", slog.String("facile_id", facileID))
	}
	twin, err := findProjectTwin(tx, place, p.name(), facileID)
	if err != nil || twin != nil {
		return s.mergeProjectTwin(tx, twin, change, err)
	}
	desc := ""
	if p.Description != nil {
		desc = *p.Description
	}
	now := time.Now().UTC()
	record := schemas.Project{Name: p.name(), Description: desc, Icon: normalizeIcon(p.Icon),
		OwnerID: place.ownerID, SpaceID: place.spaceID, FacileID: &facileID, SyncedAt: &now}
	if err := tx.Create(&record).Error; err != nil {
		return 0, err
	}
	s.logger.Info("antenne: synced project created", slog.String("facile_id", facileID))
	return 0, nil
}

func findProjectTwin(tx *gorm.DB, place placement, name, facileID string) (*schemas.Project, error) {
	query := tx.Where("lower(name) = lower(?) AND synced_at IS NULL AND (facile_id IS NULL OR facile_id <> ?)", name, facileID)
	if place.spaceID != nil {
		query = query.Where("space_id = ?", *place.spaceID)
	} else {
		query = query.Where("space_id IS NULL AND owner_id = ?", place.ownerID)
	}
	var twin schemas.Project
	err := query.Order("id").First(&twin).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	return &twin, err
}

// mergeProjectTwin applies the tie-break: both sides converge on the smaller facile_id.
func (s *Service) mergeProjectTwin(tx *gorm.DB, twin *schemas.Project, change *inboundChange[projectWire], err error) (int64, error) {
	if err != nil {
		return 0, err
	}
	if twin.FacileID != nil && change.facileID >= *twin.FacileID {
		s.logger.Info("antenne: project twin keeps its id", slog.String("facile_id", *twin.FacileID))
		return 0, nil
	}
	if err := adoptFacileID(tx, &schemas.Project{}, twin.ID, change.facileID); err != nil {
		return 0, err
	}
	twin.FacileID = &change.facileID
	return twin.ID, s.updateProjectFromWire(tx, twin, change)
}

func adoptFacileID(tx *gorm.DB, model any, id int64, facileID string) error {
	return tx.Model(model).Where("id = ?", id).
		UpdateColumns(map[string]any{"facile_id": facileID, "synced_at": time.Now().UTC()}).Error
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
		if err := tx.Model(&schemas.TimeEntry{}).Where("project_id = ? OR task_id IN (?)", project.ID, taskIDs).
			UpdateColumns(map[string]any{"task_id": 0, "project_id": 0}).Error; err != nil {
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
