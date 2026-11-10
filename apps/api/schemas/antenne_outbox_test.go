package schemas

import (
	"testing"

	"gorm.io/gorm"
)

const legacyOutbox = `CREATE TABLE antenne_outbox (
	id bigserial PRIMARY KEY, channel text, payload text,
	attempts bigint NOT NULL DEFAULT 0, last_error text, created_at timestamptz)`

type outboxColumn struct {
	Name     string
	Type     string
	Nullable string
	Default  string
}

func outboxShape(t *testing.T, db *gorm.DB) ([]outboxColumn, []string) {
	t.Helper()
	var columns []outboxColumn
	if err := db.Raw(`SELECT column_name AS name, data_type AS type, is_nullable AS nullable,
		coalesce(regexp_replace(column_default, 'nextval\(.*\)', 'nextval'), '') AS default
		FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = 'antenne_outbox'
		ORDER BY column_name`).Scan(&columns).Error; err != nil {
		t.Fatal(err)
	}
	var indexes []string
	if err := db.Raw(`SELECT indexdef FROM pg_indexes WHERE schemaname = current_schema()
		AND tablename = 'antenne_outbox' ORDER BY indexname`).Scan(&indexes).Error; err != nil {
		t.Fatal(err)
	}
	return columns, indexes
}

func exec(t *testing.T, db *gorm.DB, statement string, args ...any) {
	t.Helper()
	if err := db.Exec(statement, args...).Error; err != nil {
		t.Fatalf("%s: %v", statement, err)
	}
}

func TestTheLegacyOutboxBecomesTheLibraryOutboxKeepingUnsentRows(t *testing.T) {
	db := openPostgres(t)
	exec(t, db, OutboxSchema())
	wantColumns, wantIndexes := outboxShape(t, db)
	exec(t, db, `DROP TABLE antenne_outbox`)

	exec(t, db, legacyOutbox)
	event := `{"version":1,"app":"Sablier","payload":{"name":"a"},"idempotency_key":"project.created:fac_1:42"}`
	exec(t, db, `INSERT INTO antenne_outbox (channel, payload, attempts, last_error) VALUES
		('project.created', ?, 3, 'refused'), ('project.updated', ?, 0, NULL), ('task.created', '{}', 0, NULL)`,
		event, event)
	for round := range 2 {
		if err := migrateAntenneOutbox(db); err != nil {
			t.Fatalf("round %d: %v", round, err)
		}
		columns, indexes := outboxShape(t, db)
		if len(columns) != len(wantColumns) || len(indexes) != len(wantIndexes) {
			t.Fatalf("round %d: shape %v %v, want %v %v", round, columns, indexes, wantColumns, wantIndexes)
		}
		for i := range columns {
			if columns[i] != wantColumns[i] {
				t.Errorf("round %d: column %+v, want %+v", round, columns[i], wantColumns[i])
			}
		}
		for i := range indexes {
			if indexes[i] != wantIndexes[i] {
				t.Errorf("round %d: index %q, want %q", round, indexes[i], wantIndexes[i])
			}
		}
		exec(t, db, `ALTER TABLE antenne_outbox ALTER COLUMN payload TYPE text, ADD COLUMN attempts bigint NOT NULL DEFAULT 0`)
	}

	var rows []struct {
		Key  string
		Same bool
		Sent bool
	}
	if err := migrateAntenneOutbox(db); err != nil {
		t.Fatal(err)
	}
	if err := db.Raw(`SELECT idempotency_key AS key, payload = ?::jsonb AS same, sent_at IS NOT NULL AS sent
		FROM antenne_outbox ORDER BY id`, event).Scan(&rows).Error; err != nil {
		t.Fatal(err)
	}
	want := []string{"project.created:fac_1:42", "project.created:fac_1:42:2", "sablier-outbox:3"}
	if len(rows) != 3 {
		t.Fatalf("rows = %+v", rows)
	}
	for i, row := range rows {
		if row.Key != want[i] || row.Sent || (i < 2 && !row.Same) {
			t.Errorf("row %d = %+v, want key %q, unsent, payload kept", i, row, want[i])
		}
	}
}
