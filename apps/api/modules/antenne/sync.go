package antenne

import (
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"regexp"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
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
