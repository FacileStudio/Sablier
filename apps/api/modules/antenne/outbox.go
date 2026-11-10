package antenne

import (
	"context"
	"database/sql"
	"encoding/json"
	stderrors "errors"
	"log/slog"
	"strings"
	"time"

	antenneclient "github.com/FacileStudio/antenne-client/go"
	"github.com/FacileStudio/antenne-client/go/catalog"
	"github.com/FacileStudio/antenne-client/go/envelope"
	"github.com/FacileStudio/antenne-client/go/outbox"
	"github.com/FacileStudio/tronc/errors"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
)

const ledgerPruneEvery = 24 * time.Hour

// enqueueOutbox queues an event in antenne-client's durable outbox under the
// envelope's own idempotency key. A key already queued is the same event, so
// it is not an error.
func (s *Service) enqueueOutbox(ctx context.Context, channel catalog.Channel, evt envelope.Event[any]) error {
	db, err := s.orm.DB()
	if err != nil {
		return errors.Internal("failed to open the outbox", err)
	}
	tx, err := db.BeginTx(ctx, nil)
	if err != nil {
		return errors.Internal("failed to open the outbox", err)
	}
	defer func() { _ = tx.Rollback() }()
	err = outbox.Enqueue(ctx, tx, channel, evt, evt.IdempotencyKey)
	if err != nil && !stderrors.Is(err, outbox.ErrDuplicate) {
		return err
	}
	return tx.Commit()
}

// RunOutboxWorker drains the outbox to the bus through the live client, in id
// order, stopping a pass at the first refused event and retrying it later. It
// also prunes the inbound idempotency ledger once a day.
func (s *Service) RunOutboxWorker(ctx context.Context) {
	db, err := s.orm.DB()
	if err != nil {
		s.logger.Error("antenne: outbox worker not started", slog.Any("error", err))
		return
	}
	go s.drainer(db).Run(ctx)
	ticker := time.NewTicker(ledgerPruneEvery)
	defer ticker.Stop()
	for {
		s.orm.Where("processed_at < ?", time.Now().Add(-ledgerRetention)).Delete(&schemas.AntenneProcessedEvent{})
		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		}
	}
}

func (s *Service) drainer(db *sql.DB) *outbox.Drainer {
	return outbox.NewDrainer(db, s.emitQueued,
		outbox.WithMaxBackoff(5*time.Second),
		outbox.WithOnError(func(err error) {
			if !stderrors.Is(err, antenneclient.ErrNotConnected) {
				s.logger.Error("antenne: outbox pass failed", slog.Any("error", err))
			}
		}))
}

// emitQueued sends one queued event through the current client, refusing
// while there is none, so the drainer leaves the row queued.
func (s *Service) emitQueued(channel string, payload any) error {
	s.mu.RLock()
	client := s.client
	s.mu.RUnlock()
	if client == nil {
		return antenneclient.ErrNotConnected
	}
	raw, ok := payload.(json.RawMessage)
	if !ok {
		return client.EmitNow(channel, payload)
	}
	event, err := canonical(channel, raw)
	if err != nil {
		return err
	}
	return client.EmitNow(channel, event)
}

// canonical gives a queued event back the bytes it was marshalled with. The
// outbox keeps payloads as jsonb, which stores the same object with its keys
// reordered, so reading a row back yields different bytes. Decoding it into
// the struct it came from and encoding it again restores the original order,
// which is what receivers have always been sent.
func canonical(channel string, raw json.RawMessage) (json.RawMessage, error) {
	object, _, _ := strings.Cut(channel, ".")
	switch envelope.ObjectType(object) {
	case envelope.ObjectProject:
		return reencode[projectPayload](raw)
	case envelope.ObjectTask:
		return reencode[taskPayload](raw)
	case envelope.ObjectTimeEntry:
		return reencode[timeEntryPayload](raw)
	}
	return raw, nil
}

func reencode[T any](raw json.RawMessage) (json.RawMessage, error) {
	var evt envelope.Event[T]
	if err := json.Unmarshal(raw, &evt); err != nil {
		return nil, err
	}
	return json.Marshal(evt)
}
