package schemas

import (
	"strings"
	"time"

	"github.com/FacileStudio/antenne-client/go/outbox"

	"gorm.io/gorm"
)

// AntenneProcessedEvent records a pooled event's idempotency key once it has
// been handled.
type AntenneProcessedEvent struct {
	IdempotencyKey string    `gorm:"column:idempotency_key;primaryKey"`
	ProcessedAt    time.Time `gorm:"column:processed_at;autoCreateTime"`
}

func (AntenneProcessedEvent) TableName() string { return "antenne_processed_events" }

// OutboxSchema is the Up half of antenne-client's outbox migration: the
// antenne_outbox table exactly as the library's drainer expects it.
func OutboxSchema() string {
	up, _, _ := strings.Cut(outbox.Migration, "-- +goose Down")
	return strings.TrimPrefix(strings.TrimSpace(up), "-- +goose Up")
}

const outboxReshape = `
ALTER TABLE antenne_outbox DROP COLUMN IF EXISTS attempts, DROP COLUMN IF EXISTS last_error,
	ADD COLUMN IF NOT EXISTS idempotency_key text, ADD COLUMN IF NOT EXISTS sent_at timestamptz;
UPDATE antenne_outbox SET idempotency_key = COALESCE(NULLIF(payload::jsonb->>'idempotency_key', ''), 'sablier-outbox:' || id)
	WHERE idempotency_key IS NULL;
UPDATE antenne_outbox o SET idempotency_key = o.idempotency_key || ':' || o.id
	WHERE EXISTS (SELECT 1 FROM antenne_outbox p WHERE p.idempotency_key = o.idempotency_key AND p.id < o.id);
UPDATE antenne_outbox SET created_at = now() WHERE created_at IS NULL;
UPDATE antenne_outbox SET channel = '' WHERE channel IS NULL;
ALTER TABLE antenne_outbox ALTER COLUMN payload TYPE jsonb USING payload::jsonb,
	ALTER COLUMN payload SET NOT NULL, ALTER COLUMN channel SET NOT NULL,
	ALTER COLUMN idempotency_key SET NOT NULL,
	ALTER COLUMN created_at SET DEFAULT now(), ALTER COLUMN created_at SET NOT NULL;
DO $$ BEGIN
	IF NOT EXISTS (SELECT 1 FROM pg_constraint WHERE conname = 'antenne_outbox_idempotency_key_key'
		AND conrelid = 'antenne_outbox'::regclass) THEN
		ALTER TABLE antenne_outbox ADD CONSTRAINT antenne_outbox_idempotency_key_key UNIQUE (idempotency_key);
	END IF;
END $$;
CREATE INDEX IF NOT EXISTS antenne_outbox_unsent ON antenne_outbox (id) WHERE sent_at IS NULL;
CREATE INDEX IF NOT EXISTS antenne_outbox_sent ON antenne_outbox (sent_at) WHERE sent_at IS NOT NULL;
`

// migrateAntenneOutbox gives antenne_outbox the library's shape. A new
// database gets the library's own DDL. The table Sablier's hand-written outbox
// used (text payload, attempts, last_error, rows deleted once sent) is
// converted in place, its unsent rows kept and keyed by the idempotency key
// their envelope already carries. The conversion runs again whenever the
// shape drifts, which is what booting an older release on this database does.
// PostgreSQL only: the outbox is PostgreSQL SQL.
func migrateAntenneOutbox(db *gorm.DB) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	migrator := db.Migrator()
	if !migrator.HasTable("antenne_outbox") {
		return db.Exec(OutboxSchema()).Error
	}
	var drifted bool
	err := db.Raw(`SELECT EXISTS (SELECT 1 FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = 'antenne_outbox'
		AND ((column_name = 'payload' AND data_type <> 'jsonb') OR column_name IN ('attempts', 'last_error')))
		OR NOT EXISTS (SELECT 1 FROM information_schema.columns WHERE table_schema = current_schema()
		AND table_name = 'antenne_outbox' AND column_name = 'sent_at')`).Scan(&drifted).Error
	if err != nil || !drifted {
		return err
	}
	return db.Transaction(func(tx *gorm.DB) error { return tx.Exec(outboxReshape).Error })
}
