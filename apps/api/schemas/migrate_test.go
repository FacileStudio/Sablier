package schemas

import (
	"fmt"
	"strings"
	"testing"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestMigrateBackfillsMissingUserColors(t *testing.T) {
	orm := openTestDatabase(t)
	if err := orm.AutoMigrate(&User{}, &Session{}, &Project{}, &Task{}, &TimeEntry{}, &AppSetting{}); err != nil {
		t.Fatalf("prepare schema: %v", err)
	}

	alpha := User{Email: "alpha@example.com", PasswordHash: "hash", Color: "AD9EF0"}
	beta := User{Email: "beta@example.com", PasswordHash: "hash"}
	gamma := User{Email: "gamma@example.com", PasswordHash: "hash"}
	for _, user := range []*User{&alpha, &beta, &gamma} {
		if err := orm.Create(user).Error; err != nil {
			t.Fatalf("create user %s: %v", user.Email, err)
		}
	}

	if err := Migrate(orm); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var users []User
	if err := orm.Order("email asc").Find(&users).Error; err != nil {
		t.Fatalf("read users: %v", err)
	}

	expected := map[string]string{
		"alpha@example.com": "AD9EF0",
		"beta@example.com":  "F09ED6",
		"gamma@example.com": "EE7E89",
	}
	for _, user := range users {
		if user.Color != expected[user.Email] {
			t.Fatalf("expected %s to have color %s, got %s", user.Email, expected[user.Email], user.Color)
		}
	}
}

func openTestDatabase(t *testing.T) *gorm.DB {
	t.Helper()

	orm, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))), &gorm.Config{})
	if err != nil {
		t.Fatalf("open database: %v", err)
	}

	return orm
}

func TestMigrateRenamesNookPoolColumnsKeepingValues(t *testing.T) {
	orm := openTestDatabase(t)
	if err := orm.Exec(`CREATE TABLE app_settings (
		id integer PRIMARY KEY AUTOINCREMENT,
		nook_pool_url text NOT NULL DEFAULT '',
		nook_pool_secret text NOT NULL DEFAULT '',
		nook_pool_enabled numeric NOT NULL DEFAULT false
	)`).Error; err != nil {
		t.Fatalf("create legacy table: %v", err)
	}
	if err := orm.Exec(`INSERT INTO app_settings (nook_pool_url, nook_pool_secret, nook_pool_enabled)
		VALUES (?, ?, ?)`, "https://antenne.facile.studio", "s3cret", true).Error; err != nil {
		t.Fatalf("seed legacy row: %v", err)
	}

	if err := Migrate(orm); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var setting AppSetting
	if err := orm.First(&setting).Error; err != nil {
		t.Fatalf("read migrated row: %v", err)
	}
	if setting.AntenneURL != "https://antenne.facile.studio" || setting.AntenneSecret != "s3cret" || !setting.AntenneEnabled {
		t.Fatalf("values lost in rename: %+v", setting)
	}
	if orm.Migrator().HasColumn(&AppSetting{}, "nook_pool_url") {
		t.Fatal("legacy column still present")
	}
}

func TestMigrateRenamesPoolTablesKeepingRows(t *testing.T) {
	orm := openTestDatabase(t)
	if err := orm.Exec(`CREATE TABLE pool_outbox (
		id integer PRIMARY KEY AUTOINCREMENT,
		channel text,
		payload text,
		attempts integer NOT NULL DEFAULT 0,
		last_error text,
		created_at datetime
	)`).Error; err != nil {
		t.Fatalf("create legacy outbox: %v", err)
	}
	if err := orm.Exec(`INSERT INTO pool_outbox (channel, payload) VALUES (?, ?)`,
		"project.updated", `{"id":1}`).Error; err != nil {
		t.Fatalf("seed outbox row: %v", err)
	}

	if err := Migrate(orm); err != nil {
		t.Fatalf("migrate: %v", err)
	}

	var pending []AntenneOutbox
	if err := orm.Find(&pending).Error; err != nil {
		t.Fatalf("read migrated outbox: %v", err)
	}
	if len(pending) != 1 || pending[0].Channel != "project.updated" {
		t.Fatalf("in-flight events lost in table rename: %+v", pending)
	}
	if orm.Migrator().HasTable("pool_outbox") {
		t.Fatal("legacy table still present")
	}
}
