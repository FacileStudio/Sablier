package migrations_test

import (
	"context"
	"io/fs"
	"os"
	"slices"
	"strings"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/internal/testdb"
	"github.com/FacileStudio/Sablier/apps/api/migrations"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func openSchema(t *testing.T, schema string, prepare func(*gorm.DB) error) *gorm.DB {
	t.Helper()
	raw := os.Getenv(testdb.EnvVar)
	if raw == "" {
		t.Skip(testdb.EnvVar + " is unset")
	}
	admin, err := gorm.Open(postgres.Open(raw), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := admin.Exec(`DROP SCHEMA IF EXISTS ` + schema + ` CASCADE; CREATE SCHEMA ` + schema).Error; err != nil {
		t.Fatalf("reset %s: %v", schema, err)
	}
	if sqlDB, err := admin.DB(); err == nil {
		_ = sqlDB.Close()
	}
	db, err := gorm.Open(postgres.Open(testdb.WithSearchPath(raw, schema)), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open %s: %v", schema, err)
	}
	t.Cleanup(func() {
		if sqlDB, err := db.DB(); err == nil {
			_ = sqlDB.Close()
		}
	})
	if prepare != nil {
		if err := prepare(db); err != nil {
			t.Fatalf("prepare %s: %v", schema, err)
		}
	}
	return db
}

func runUp(t *testing.T, db *gorm.DB) error {
	t.Helper()
	sqlDB, err := db.DB()
	if err != nil {
		return err
	}
	var schema string
	if err := db.Raw(`SELECT current_schema()`).Scan(&schema).Error; err != nil {
		return err
	}
	return migrations.Up(context.Background(), testdb.Config(sqlDB, schema))
}

func appliedVersions(t *testing.T, db *gorm.DB) []int64 {
	t.Helper()
	var versions []int64
	if err := db.Raw(`SELECT version_id FROM goose_db_version WHERE version_id > 0 ORDER BY id`).Scan(&versions).Error; err != nil {
		t.Fatalf("read the ledger: %v", err)
	}
	return versions
}

func columnDefault(t *testing.T, db *gorm.DB, table, column string) string {
	t.Helper()
	var value *string
	err := db.Raw(`SELECT column_default FROM information_schema.columns
		WHERE table_schema = current_schema() AND table_name = ? AND column_name = ?`, table, column).Scan(&value).Error
	if err != nil {
		t.Fatalf("read the default of %s.%s: %v", table, column, err)
	}
	if value == nil {
		return ""
	}
	return *value
}

func sections(t *testing.T, name string) (string, string) {
	t.Helper()
	raw, err := fs.ReadFile(migrations.FS, name)
	if err != nil {
		t.Fatalf("read %s: %v", name, err)
	}
	_, rest, found := strings.Cut(string(raw), "-- +goose Up")
	if !found {
		t.Fatalf("%s has no Up section", name)
	}
	up, down, found := strings.Cut(rest, "-- +goose Down")
	if !found {
		t.Fatalf("%s has no Down section", name)
	}
	return up, down
}

func assertVersions(t *testing.T, db *gorm.DB) {
	t.Helper()
	if got := appliedVersions(t, db); !slices.Equal(got, []int64{1, 2}) {
		t.Fatalf("applied versions %v, want [1 2]", got)
	}
}

func tableShape(t *testing.T, tx *gorm.DB, table string) []string {
	t.Helper()
	var shape []string
	err := tx.Raw(`SELECT column_name || ' ' || data_type || ' ' || is_nullable || ' ' ||
			coalesce(regexp_replace(column_default, 'nextval\(.*\)', 'nextval'), '')
		FROM information_schema.columns WHERE table_schema = current_schema() AND table_name = ?
		UNION ALL SELECT indexname || ' ' || replace(indexdef, current_schema() || '.', '')
		FROM pg_indexes WHERE schemaname = current_schema() AND tablename = ?
		UNION ALL SELECT conname || ' ' || pg_get_constraintdef(oid) FROM pg_constraint
		WHERE conrelid = to_regclass(?)
		ORDER BY 1`, table, table, table).Scan(&shape).Error
	if err != nil {
		t.Fatalf("read the shape of %s: %v", table, err)
	}
	return shape
}

func assertRowKept(t *testing.T, db *gorm.DB) {
	t.Helper()
	var kept struct {
		Name   string
		Locale string
	}
	if err := db.Raw(`SELECT name, locale FROM users WHERE email = 'ada@facile.studio'`).Scan(&kept).Error; err != nil {
		t.Fatalf("read the row: %v", err)
	}
	if kept.Name != "Ada" || kept.Locale != "fr" {
		t.Fatalf("the existing row changed: %+v", kept)
	}
}
