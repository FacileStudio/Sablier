// Package testdb opens the PostgreSQL database a test runs its queries
// against, built by the same migrations a booting server applies. It is
// PostgreSQL and nothing else: the schema ships as Postgres DDL under
// migrations/, and SQLite would build a different one from the struct tags.
package testdb

import (
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/binary"
	"encoding/hex"
	"fmt"
	"log/slog"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/migrations"
	"github.com/FacileStudio/tronc/migrate"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// EnvVar names the connection string the database tests read.
const EnvVar = "SABLIER_TEST_DATABASE_URL"

var unsafeName = regexp.MustCompile(`[^a-z0-9_]+`)

// migratedSchema is the schema this test binary migrated, once: a package
// runs many tests against the same shape.
var migratedSchema = sync.OnceValues(migrateSchema)

func migrateSchema() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(dir))
	name := "m_" + unsafeName.ReplaceAllString(strings.ToLower(filepath.Base(dir)), "_")
	name = name[:min(len(name), 40)] + "_" + hex.EncodeToString(sum[:4])
	admin, err := gorm.Open(postgres.Open(os.Getenv(EnvVar)), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return "", err
	}
	defer closeGorm(admin)
	if err := admin.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE; CREATE SCHEMA %s", name, name)).Error; err != nil {
		return "", err
	}
	orm, err := gorm.Open(postgres.Open(WithSearchPath(os.Getenv(EnvVar), name)), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return "", err
	}
	defer closeGorm(orm)
	sqlDB, err := orm.DB()
	if err != nil {
		return "", err
	}
	return name, migrations.Up(context.Background(), Config(sqlDB, name))
}

// Config is the migrations configuration for a test schema. The advisory lock
// is scoped to the database, not the schema, so each schema derives its own
// lock id instead of queueing every test package on the default one.
func Config(db *sql.DB, schema string) migrate.Config {
	cfg := migrations.Config(db, slog.New(slog.DiscardHandler))
	sum := sha256.Sum256([]byte(schema))
	cfg.LockID = int64(binary.BigEndian.Uint64(sum[:8]) >> 1)
	return cfg
}

func closeGorm(orm *gorm.DB) {
	if sqlDB, err := orm.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

// WithSearchPath returns raw with search_path set to schema. The search_path
// travels in the connection string rather than in a SET: GORM hands out
// pooled connections, so a SET would bind one of them only.
func WithSearchPath(raw, schema string) string {
	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

// emptySchema deletes every row and rewinds every sequence in one
// transaction, the migration ledger excepted. TRUNCATE would be the obvious
// statement, but it costs seconds per call on an fsync-bound disk, and these
// tables hold a handful of rows.
const emptySchema = `DO $$
DECLARE name text;
BEGIN
	SET LOCAL session_replication_role = replica;
	FOR name IN SELECT tablename FROM pg_tables WHERE schemaname = current_schema()
		AND tablename <> '` + migrate.TableName + `' LOOP
		EXECUTE format('DELETE FROM %I', name);
	END LOOP;
	FOR name IN SELECT sequencename FROM pg_sequences WHERE schemaname = current_schema()
		AND sequencename <> '` + migrate.TableName + `_id_seq' LOOP
		EXECUTE format('ALTER SEQUENCE %I RESTART', name);
	END LOOP;
END $$`

// Migrated returns a PostgreSQL database brought up to date by the migrations
// a booting server applies, and emptied for t. The schema is migrated once per
// test binary and emptied before each test, so tests in one package must not
// run in parallel. It skips the test when SABLIER_TEST_DATABASE_URL is unset.
func Migrated(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv(EnvVar) == "" {
		t.Skip(EnvVar + " is unset")
	}
	schema, err := migratedSchema()
	if err != nil {
		t.Fatalf("migrate the test database: %v", err)
	}
	orm, err := gorm.Open(postgres.Open(WithSearchPath(os.Getenv(EnvVar), schema)), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open the test database: %v", err)
	}
	t.Cleanup(func() { closeGorm(orm) })
	if err := orm.Exec(emptySchema).Error; err != nil {
		t.Fatalf("empty the test database: %v", err)
	}
	return orm
}
