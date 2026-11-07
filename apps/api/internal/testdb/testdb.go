// Package testdb opens the database a test runs its queries against: an in-memory SQLite
// by default, or PostgreSQL when SABLIER_TEST_DATABASE_URL is set. SQLite does not catch
// a uuid compared with text, so every query that crosses space_id should run here.
package testdb

import (
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"
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
