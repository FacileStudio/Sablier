// Package testdb opens the database a test runs its queries against: an in-memory SQLite
// by default, or PostgreSQL when SABLIER_TEST_DATABASE_URL is set. SQLite does not catch
// a uuid compared with text, so every query that crosses space_id should run here.
package testdb

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"testing"

	"github.com/FacileStudio/Sablier/apps/api/schemas"

	"gorm.io/driver/postgres"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var unsafeName = regexp.MustCompile(`[^a-z0-9_]+`)

// Open returns a migrated database private to t. On PostgreSQL every test gets its own
// schema through search_path, so it never touches public, which the schemas tests drop.
func Open(t *testing.T) *gorm.DB {
	t.Helper()
	models := []any{&schemas.User{}, &schemas.Project{}, &schemas.Task{}, &schemas.TimeEntry{},
		&schemas.Space{}, &schemas.SpaceMember{}, &schemas.Heartbeat{}, &schemas.DraftDismissal{}}
	orm := open(t)
	if err := orm.AutoMigrate(models...); err != nil {
		t.Fatalf("migrate the test database: %v", err)
	}
	return orm
}

func open(t *testing.T) *gorm.DB {
	t.Helper()
	name := "t_" + unsafeName.ReplaceAllString(strings.ToLower(t.Name()), "_")
	raw := os.Getenv("SABLIER_TEST_DATABASE_URL")
	if raw == "" {
		orm, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", name)), &gorm.Config{})
		if err != nil {
			t.Fatalf("open sqlite: %v", err)
		}
		return orm
	}
	return openPostgres(t, raw, name[:min(len(name), 60)])
}

func openPostgres(t *testing.T, raw, schema string) *gorm.DB {
	t.Helper()
	admin, err := gorm.Open(postgres.Open(raw), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open postgres: %v", err)
	}
	reset := fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE; CREATE SCHEMA %s", schema, schema)
	if err := admin.Exec(reset).Error; err != nil {
		t.Fatalf("reset schema %s: %v", schema, err)
	}
	if sqlDB, err := admin.DB(); err == nil {
		_ = sqlDB.Close()
	}
	parsed, err := url.Parse(raw)
	if err != nil {
		t.Fatalf("parse SABLIER_TEST_DATABASE_URL: %v", err)
	}
	query := parsed.Query()
	query.Set("search_path", schema)
	parsed.RawQuery = query.Encode()
	orm, err := gorm.Open(postgres.Open(parsed.String()), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open postgres schema %s: %v", schema, err)
	}
	return orm
}

// migratedSchema is the schema this test binary migrated, once: schemas.Migrate
// takes seconds, and a package runs many tests against the same shape.
var migratedSchema = sync.OnceValues(migrateSchema)

func migrateSchema() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256([]byte(dir))
	name := "m_" + unsafeName.ReplaceAllString(strings.ToLower(filepath.Base(dir)), "_")
	name = name[:min(len(name), 40)] + "_" + hex.EncodeToString(sum[:4])
	admin, err := gorm.Open(postgres.Open(os.Getenv("SABLIER_TEST_DATABASE_URL")), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return "", err
	}
	defer closeGorm(admin)
	if err := admin.Exec(fmt.Sprintf("DROP SCHEMA IF EXISTS %s CASCADE; CREATE SCHEMA %s", name, name)).Error; err != nil {
		return "", err
	}
	orm, err := gorm.Open(postgres.Open(withSearchPath(os.Getenv("SABLIER_TEST_DATABASE_URL"), name)),
		&gorm.Config{Logger: logger.Discard})
	if err != nil {
		return "", err
	}
	defer closeGorm(orm)
	return name, schemas.Migrate(orm, "")
}

func closeGorm(orm *gorm.DB) {
	if sqlDB, err := orm.DB(); err == nil {
		_ = sqlDB.Close()
	}
}

func withSearchPath(raw, schema string) string {
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
// transaction. TRUNCATE would be the obvious statement, but it costs seconds
// per call on an fsync-bound disk, and these tables hold a handful of rows.
const emptySchema = `DO $$
DECLARE name text;
BEGIN
	SET LOCAL session_replication_role = replica;
	FOR name IN SELECT tablename FROM pg_tables WHERE schemaname = current_schema() LOOP
		EXECUTE format('DELETE FROM %I', name);
	END LOOP;
	FOR name IN SELECT sequencename FROM pg_sequences WHERE schemaname = current_schema() LOOP
		EXECUTE format('ALTER SEQUENCE %I RESTART', name);
	END LOOP;
END $$`

// Migrated returns a PostgreSQL database brought up to date by schemas.Migrate
// exactly as a booting server does, porte tables included, and emptied for t.
// The schema is migrated once per test binary and emptied before each test,
// so tests in one package must not run in parallel. It skips the test when
// SABLIER_TEST_DATABASE_URL is unset: what it proves is PostgreSQL behaviour.
func Migrated(t *testing.T) *gorm.DB {
	t.Helper()
	if os.Getenv("SABLIER_TEST_DATABASE_URL") == "" {
		t.Skip("SABLIER_TEST_DATABASE_URL is unset")
	}
	schema, err := migratedSchema()
	if err != nil {
		t.Fatalf("migrate the test database: %v", err)
	}
	orm, err := gorm.Open(postgres.Open(withSearchPath(os.Getenv("SABLIER_TEST_DATABASE_URL"), schema)),
		&gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open the test database: %v", err)
	}
	t.Cleanup(func() { closeGorm(orm) })
	if err := orm.Exec(emptySchema).Error; err != nil {
		t.Fatalf("empty the test database: %v", err)
	}
	return orm
}
