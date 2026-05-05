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
	if err := orm.AutoMigrate(&User{}, &Session{}, &Project{}, &Task{}, &TimeEntry{}, &UserSetting{}); err != nil {
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
