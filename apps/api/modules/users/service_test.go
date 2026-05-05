package users

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"api/internal/authcontext"
	"api/schemas"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
)

func TestPersistAvatarFileWritesToStorageDir(t *testing.T) {
	storageDir := t.TempDir()
	service := NewService(nil, storageDir)
	payload := bytes.Repeat([]byte{0x89, 0x50, 0x4e, 0x47}, 256)

	relativePath, absolutePath, err := service.persistAvatarFile(42, bytes.NewReader(payload), "image/png")
	if err != nil {
		t.Fatalf("persist avatar file: %v", err)
	}

	if !strings.HasPrefix(relativePath, "avatars"+string(filepath.Separator)+"user-42-") || !strings.HasSuffix(relativePath, ".png") {
		t.Fatalf("unexpected relative path: %s", relativePath)
	}

	if !strings.HasPrefix(absolutePath, storageDir+string(filepath.Separator)) {
		t.Fatalf("expected absolute path under storage dir, got %s", absolutePath)
	}

	info, err := os.Stat(absolutePath)
	if err != nil {
		t.Fatalf("stat avatar file: %v", err)
	}
	if info.Size() != int64(len(payload)) {
		t.Fatalf("unexpected avatar size: got %d want %d", info.Size(), len(payload))
	}
}

func TestRemoveAvatarFileDeletesManagedAvatarOnly(t *testing.T) {
	storageDir := t.TempDir()
	service := NewService(nil, storageDir)

	managedPath := filepath.Join(storageDir, "avatars", "managed.png")
	if err := os.MkdirAll(filepath.Dir(managedPath), 0o755); err != nil {
		t.Fatalf("mkdir managed path: %v", err)
	}
	if err := os.WriteFile(managedPath, []byte("avatar"), 0o644); err != nil {
		t.Fatalf("write managed avatar: %v", err)
	}

	externalPath := filepath.Join(storageDir, "outside.txt")
	if err := os.WriteFile(externalPath, []byte("keep"), 0o644); err != nil {
		t.Fatalf("write external file: %v", err)
	}

	service.removeAvatarFile("/files/avatars/managed.png")
	service.removeAvatarFile("/files/../outside.txt")

	if _, err := os.Stat(managedPath); !os.IsNotExist(err) {
		t.Fatalf("expected managed avatar removed, stat err=%v", err)
	}
	if _, err := os.Stat(externalPath); err != nil {
		t.Fatalf("expected external file preserved, stat err=%v", err)
	}
}

func TestGetUserAssignsLeastUsedColorWhenMissing(t *testing.T) {
	service := newDatabaseBackedService(t)
	seedUser(t, service.orm, "alpha@example.com", "AD9EF0")
	missing := seedUser(t, service.orm, "beta@example.com", "")

	user, err := service.getUser(context.Background(), fmt.Sprintf("%d", missing.ID))
	if err != nil {
		t.Fatalf("get user: %v", err)
	}

	if user.Color != "F09ED6" {
		t.Fatalf("expected F09ED6, got %s", user.Color)
	}

	var record schemas.User
	if err := service.orm.Where("id = ?", missing.ID).First(&record).Error; err != nil {
		t.Fatalf("read updated user: %v", err)
	}
	if record.Color != "F09ED6" {
		t.Fatalf("expected stored color F09ED6, got %s", record.Color)
	}
}

func TestUpdateUserPersistsChosenColor(t *testing.T) {
	service := newDatabaseBackedService(t)
	user := seedUser(t, service.orm, "color@example.com", "AD9EF0")
	color := "7EEEDB"

	updated, err := service.updateUser(context.Background(), fmt.Sprintf("%d", user.ID), nil, nil, nil, &color)
	if err != nil {
		t.Fatalf("update user: %v", err)
	}

	if updated.Color != color {
		t.Fatalf("expected color %s, got %s", color, updated.Color)
	}
}

func TestListUsersBackfillsMissingColors(t *testing.T) {
	service := newDatabaseBackedService(t)
	seedUser(t, service.orm, "one@example.com", "")
	seedUser(t, service.orm, "two@example.com", "")

	users, err := service.listUsers(context.Background())
	if err != nil {
		t.Fatalf("list users: %v", err)
	}

	if len(users) != 2 {
		t.Fatalf("expected 2 users, got %d", len(users))
	}
	if users[0].Color == "" || users[1].Color == "" {
		t.Fatalf("expected colors assigned, got %#v", users)
	}
	if users[0].Color == users[1].Color {
		t.Fatalf("expected distributed colors, got %#v", users)
	}
}

func TestUpdateMeRejectsInvalidColor(t *testing.T) {
	controller := &Controller{}
	color := "FFFFFF"
	req := &UpdateRequest{Color: &color}
	ctx := authcontext.WithIdentity(context.Background(), authcontext.Identity{
		UserID: "1",
		Email:  "user@example.com",
	})

	_, err := controller.updateMe(ctx, req)
	if err == nil || err.Error() != "color must be one of: AD9EF0, F09ED6, EE7E89, EEB47E, A9EE7E, 7EEEDB" {
		t.Fatalf("expected invalid color error, got %v", err)
	}
}

func newDatabaseBackedService(t *testing.T) *Service {
	t.Helper()

	orm, err := gorm.Open(sqlite.Open(fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))), &gorm.Config{})
	if err != nil {
		t.Fatalf("open test database: %v", err)
	}

	if err := orm.AutoMigrate(&schemas.User{}); err != nil {
		t.Fatalf("migrate test database: %v", err)
	}

	return NewService(orm, t.TempDir())
}

func seedUser(t *testing.T, orm *gorm.DB, email string, color string) schemas.User {
	t.Helper()

	user := schemas.User{
		Email:        email,
		Color:        color,
		PasswordHash: "hash",
	}
	if err := orm.Create(&user).Error; err != nil {
		t.Fatalf("create user: %v", err)
	}
	return user
}
