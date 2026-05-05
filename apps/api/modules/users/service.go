package users

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"api/internal/authcrypto"
	"api/internal/errors"
	"api/schemas"

	"gorm.io/gorm"
)

type Service struct {
	orm        *gorm.DB
	storageDir string
	controller *Controller
}

func NewService(orm *gorm.DB, storageDir string) *Service {
	service := &Service{orm: orm, storageDir: storageDir}
	service.controller = newController(service)
	return service
}

func (service *Service) getUser(context context.Context, userID string) (*User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}

	var record schemas.User
	if err := service.orm.WithContext(context).Where("id = ?", id).First(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal("failed to read user", err)
	}

	return mapUser(record), nil
}

func (service *Service) updateUser(context context.Context, userID string, name *string, email *string, password *string) (*User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}

	updates := map[string]any{}
	if name != nil {
		updates["name"] = *name
	}
	if email != nil {
		updates["email"] = *email
	}
	if password != nil {
		hash, err := authcrypto.HashPassword(*password)
		if err != nil {
			return nil, errors.Invalid("invalid password")
		}
		updates["password_hash"] = hash
	}

	result := service.orm.WithContext(context).
		Model(&schemas.User{}).
		Where("id = ?", id).
		Updates(updates)
	if result.Error != nil {
		if stderrors.Is(result.Error, gorm.ErrDuplicatedKey) {
			return nil, errors.Conflict("email already registered")
		}
		return nil, errors.Internal("failed to update user", result.Error)
	}
	var record schemas.User
	if err := service.orm.WithContext(context).Where("id = ?", id).First(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal("failed to read user", err)
	}

	return &User{
		ID:        strconv.FormatInt(record.ID, 10),
		Email:     record.Email,
		Name:      record.Name,
		AvatarURL: record.AvatarURL,
	}, nil
}

func (service *Service) storeAvatar(context context.Context, userID string, reader io.Reader, contentType string) (*User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}

	extension, ok := avatarExtension(contentType)
	if !ok {
		return nil, errors.Invalid("avatar must be a PNG, JPEG, GIF, or WebP image")
	}

	var record schemas.User
	if err := service.orm.WithContext(context).Where("id = ?", id).First(&record).Error; err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.NotFound("user not found")
		}
		return nil, errors.Internal("failed to read user", err)
	}

	filename := fmt.Sprintf("user-%d-%d%s", id, time.Now().UnixNano(), extension)
	relativePath := filepath.Join("avatars", filename)
	absolutePath := filepath.Join(service.storageDir, relativePath)

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return nil, errors.Internal("failed to prepare avatar storage", err)
	}

	file, err := os.Create(absolutePath)
	if err != nil {
		return nil, errors.Internal("failed to create avatar file", err)
	}
	if _, err := io.Copy(file, reader); err != nil {
		_ = file.Close()
		return nil, errors.Internal("failed to write avatar file", err)
	}
	if err := file.Close(); err != nil {
		return nil, errors.Internal("failed to finalize avatar file", err)
	}

	newAvatarURL := "/files/" + strings.ReplaceAll(relativePath, string(filepath.Separator), "/")
	oldAvatarURL := record.AvatarURL
	record.AvatarURL = newAvatarURL

	if err := service.orm.WithContext(context).Save(&record).Error; err != nil {
		_ = os.Remove(absolutePath)
		return nil, errors.Internal("failed to save avatar", err)
	}

	if oldAvatarURL != "" {
		oldPath := strings.TrimPrefix(oldAvatarURL, "/files/")
		oldAbsolutePath := filepath.Join(service.storageDir, filepath.Clean(oldPath))
		if strings.HasPrefix(oldAbsolutePath, filepath.Clean(filepath.Join(service.storageDir, "avatars"))) {
			_ = os.Remove(oldAbsolutePath)
		}
	}

	return mapUser(record), nil
}

func mapUser(record schemas.User) *User {
	return &User{
		ID:        strconv.FormatInt(record.ID, 10),
		Email:     record.Email,
		Name:      record.Name,
		AvatarURL: record.AvatarURL,
	}
}

func avatarExtension(contentType string) (string, bool) {
	switch contentType {
	case "image/png":
		return ".png", true
	case "image/jpeg":
		return ".jpg", true
	case "image/gif":
		return ".gif", true
	case "image/webp":
		return ".webp", true
	default:
		return "", false
	}
}
