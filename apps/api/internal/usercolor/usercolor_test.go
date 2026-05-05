package usercolor_test

import (
	"context"
	"testing"

	"api/internal/usercolor"
	"api/schemas"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func newTestDB(t *testing.T) *gorm.DB {
	t.Helper()

	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	if err != nil {
		t.Fatalf("open db: %v", err)
	}
	if err := db.AutoMigrate(&schemas.User{}); err != nil {
		t.Fatalf("migrate db: %v", err)
	}
	return db
}

func TestNormalize(t *testing.T) {
	color, ok := usercolor.Normalize("#f09ed6")
	if !ok {
		t.Fatal("expected color to normalize")
	}
	if color != "F09ED6" {
		t.Fatalf("expected F09ED6, got %s", color)
	}
}

func TestBackfillMissingDistributesPalette(t *testing.T) {
	db := newTestDB(t)
	users := []schemas.User{
		{Email: "a@example.com", PasswordHash: "hash"},
		{Email: "b@example.com", PasswordHash: "hash"},
		{Email: "c@example.com", PasswordHash: "hash", Color: "#ee7e89"},
	}
	for _, user := range users {
		if err := db.Create(&user).Error; err != nil {
			t.Fatalf("create user: %v", err)
		}
	}

	if err := usercolor.BackfillMissing(context.Background(), db); err != nil {
		t.Fatalf("backfill colors: %v", err)
	}

	var rows []schemas.User
	if err := db.Order("id asc").Find(&rows).Error; err != nil {
		t.Fatalf("list users: %v", err)
	}

	if rows[0].Color != "AD9EF0" {
		t.Fatalf("expected first missing user color AD9EF0, got %s", rows[0].Color)
	}
	if rows[1].Color != "F09ED6" {
		t.Fatalf("expected second missing user color F09ED6, got %s", rows[1].Color)
	}
	if rows[2].Color != "EE7E89" {
		t.Fatalf("expected normalized existing color EE7E89, got %s", rows[2].Color)
	}
}
