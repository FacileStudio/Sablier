package users

import (
	"context"
	stderrors "errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// TokenIssuer is the slice of the auth service this module needs: minting a
// labelled session, reaching the manager that lists and revokes them, and the
// two halves of the password write.
//
// SetPassword and ChangePassword are two methods because porte made them two:
// one gives a first password to an account that has none, the other replaces
// an existing one and cannot be called without it. One method serving both is
// how this app shipped a settings screen that changed a password without ever
// asking for the old one.
type TokenIssuer interface {
	Issue(ctx context.Context, userID int64, label string) (string, porte.Session, error)
	Sessions() *session.Manager
	SetPassword(ctx context.Context, userID int64, password string) error
	ChangePassword(ctx context.Context, w http.ResponseWriter, r *http.Request, userID int64, current, next string) (string, int64, error)
}

// Service backs the users module: reading, updating and deleting accounts,
// their avatars and their API tokens.
type Service struct {
	orm        *gorm.DB
	storageDir string
	tokens     TokenIssuer
	controller *Controller
}

// NewService wires the users service with a fresh route controller.
func NewService(orm *gorm.DB, storageDir string, tokens TokenIssuer) *Service {
	service := &Service{orm: orm, storageDir: storageDir, tokens: tokens}
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
	if err := service.ensureUserColor(context, &record); err != nil {
		return nil, err
	}

	return mapUser(record), nil
}

func (service *Service) listUsers(context context.Context) ([]User, error) {
	if err := usercolor.BackfillMissing(context, service.orm); err != nil {
		return nil, errors.Internal("failed to backfill user colors", err)
	}

	var records []schemas.User
	if err := service.orm.WithContext(context).Order("name asc, email asc, id asc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list users", err)
	}

	users := make([]User, 0, len(records))
	for _, record := range records {
		users = append(users, *mapUser(record))
	}

	return users, nil
}

// applyPassword writes the password half of PATCH /users/me and returns the
// rotated session token, empty when nothing was rotated.
//
// The password is porte's, not a column on this row: writing users.password_hash
// would look like it worked and change nothing, because porte reads the identity
// table — so the old password would keep signing in and the new one would never
// work.
//
// Which of porte's two calls applies is decided by the current password alone.
// Without one this can only be a first password, and an account that already
// has one lands on porte.ErrPasswordSet — answered here as a 400 naming the
// field the caller left out, rather than as porte's conflict, because the
// caller's mistake is an omission and not a lost race.
func (service *Service) applyPassword(context context.Context, w http.ResponseWriter, r *http.Request, userID int64, current string, next string) (string, error) {
	if current == "" {
		err := service.tokens.SetPassword(context, userID, next)
		if stderrors.Is(err, porte.ErrPasswordSet) {
			return "", errors.Invalid("current_password is required to change an existing password")
		}
		return "", err
	}
	token, _, err := service.tokens.ChangePassword(context, w, r, userID, current, next)
	return token, err
}

// updateUser applies the non-nil profile fields. The password is not among
// them: it lives in porte's identity table and goes through applyPassword,
// which needs the response writer this method does not have.
func (service *Service) updateUser(context context.Context, userID string, name *string, email *string, color *string, rate *float64, rateType *string, workdayHours *float64) (*User, error) {
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
	if color != nil {
		updates["color"] = *color
	}
	if rate != nil {
		updates["rate"] = *rate
	}
	if rateType != nil {
		updates["rate_type"] = *rateType
	}
	if workdayHours != nil {
		updates["workday_hours"] = *workdayHours
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
	if err := service.ensureUserColor(context, &record); err != nil {
		return nil, err
	}

	return mapUser(record), nil
}

// storeAvatar saves an uploaded image as the account's avatar. Uploading is
// the fallback for people the identity provider has no photo for, so a photo
// in Porte makes this endpoint unavailable rather than merely outranked:
// accepting the file and never showing it is the worse failure, since the
// user sees a success and no change.
func (service *Service) storeAvatar(context context.Context, userID string, reader io.Reader, contentType string) (*User, error) {
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

	if record.OIDCPictureURL != "" {
		return nil, errors.Invalid("your photo is managed in Porte — change it there")
	}

	relativePath, absolutePath, err := service.persistAvatarFile(id, reader, contentType)
	if err != nil {
		return nil, err
	}

	newUploadPath := strings.ReplaceAll(relativePath, string(filepath.Separator), "/")
	oldUploadPath := record.AvatarUploadPath
	record.AvatarUploadPath = newUploadPath

	if err := service.orm.WithContext(context).Save(&record).Error; err != nil {
		_ = os.Remove(absolutePath)
		return nil, errors.Internal("failed to save avatar", err)
	}

	if oldUploadPath != "" {
		service.removeAvatarFile("/files/" + oldUploadPath)
	}

	if err := service.ensureUserColor(context, &record); err != nil {
		return nil, err
	}

	return mapUser(record), nil
}

// clearAvatar removes the uploaded avatar. Only the upload is the user's to
// clear: the Porte photo is not deleted from here, because it is not ours and
// the next sync would bring it straight back.
func (service *Service) clearAvatar(context context.Context, userID string) (*User, error) {
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

	oldUploadPath := record.AvatarUploadPath
	record.AvatarUploadPath = ""
	if err := service.orm.WithContext(context).Save(&record).Error; err != nil {
		return nil, errors.Internal("failed to clear avatar", err)
	}

	if oldUploadPath != "" {
		service.removeAvatarFile("/files/" + oldUploadPath)
	}

	if err := service.ensureUserColor(context, &record); err != nil {
		return nil, err
	}

	return mapUser(record), nil
}

func (service *Service) persistAvatarFile(userID int64, reader io.Reader, contentType string) (string, string, error) {
	extension, ok := avatarExtension(contentType)
	if !ok {
		return "", "", errors.Invalid("avatar must be a PNG, JPEG, GIF, or WebP image")
	}

	filename := fmt.Sprintf("user-%d-%d%s", userID, time.Now().UnixNano(), extension)
	relativePath := filepath.Join("avatars", filename)
	absolutePath := filepath.Join(service.storageDir, relativePath)

	if err := os.MkdirAll(filepath.Dir(absolutePath), 0o755); err != nil {
		return "", "", errors.Internal("failed to prepare avatar storage", err)
	}

	file, err := os.Create(absolutePath)
	if err != nil {
		return "", "", errors.Internal("failed to create avatar file", err)
	}
	if _, err := io.Copy(file, reader); err != nil {
		_ = file.Close()
		return "", "", errors.Internal("failed to write avatar file", err)
	}
	if err := file.Close(); err != nil {
		return "", "", errors.Internal("failed to finalize avatar file", err)
	}

	return relativePath, absolutePath, nil
}

func (service *Service) removeAvatarFile(avatarURL string) {
	oldPath := strings.TrimPrefix(avatarURL, "/files/")
	oldAbsolutePath := filepath.Join(service.storageDir, filepath.Clean(oldPath))
	if strings.HasPrefix(oldAbsolutePath, filepath.Clean(filepath.Join(service.storageDir, "avatars"))) {
		_ = os.Remove(oldAbsolutePath)
	}
}

func (service *Service) ensureUserColor(context context.Context, record *schemas.User) error {
	color, ok := usercolor.Normalize(record.Color)
	if ok && record.Color == color {
		return nil
	}

	color, err := usercolor.EnsureForUser(context, service.orm, record.ID)
	if err != nil {
		if stderrors.Is(err, gorm.ErrRecordNotFound) {
			return errors.NotFound("user not found")
		}
		return errors.Internal("failed to assign user color", err)
	}

	record.Color = color
	return nil
}

func mapUser(record schemas.User) *User {
	rateType := record.RateType
	if rateType == "" {
		rateType = "daily"
	}
	workdayHours := record.WorkdayHours
	if workdayHours <= 0 {
		workdayHours = 8
	}
	return &User{
		ID:           strconv.FormatInt(record.ID, 10),
		Email:        record.Email,
		Name:         record.Name,
		AvatarURL:    record.Avatar(),
		AvatarSource: record.AvatarOrigin(),
		Color:        record.Color,
		Rate:         record.Rate,
		RateType:     rateType,
		WorkdayHours: workdayHours,
		CreatedAt:    record.CreatedAt.UTC().Format(time.RFC3339),
	}
}

// The API token is a porte session with a label and no expiry. It used to be
// its own table with its own lookup in the auth path; porte.Session.Label
// exists precisely so that a named credential is not a second mechanism, and
// folding it in means one place issues, one place verifies, one place revokes.
//
// One token per user, as before: creating replaces.
func (service *Service) createApiToken(context context.Context, userID string, name string) (string, *porte.Session, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return "", nil, errors.Internal("failed to parse user id", err)
	}
	if err := service.revokeLabelled(context, id); err != nil {
		return "", nil, err
	}
	rawToken, issued, err := service.tokens.Issue(context, id, name)
	if err != nil {
		return "", nil, err
	}
	return rawToken, &issued, nil
}

func (service *Service) getApiToken(context context.Context, userID string) (*porte.Session, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	held, err := service.tokens.Sessions().List(context, id)
	if err != nil {
		return nil, errors.Internal("failed to read api token", err)
	}
	for _, candidate := range held {
		if candidate.IsAPIToken() {
			return &candidate, nil
		}
	}
	return nil, nil
}

func (service *Service) deleteApiToken(context context.Context, userID string) error {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return errors.Internal("failed to parse user id", err)
	}
	return service.revokeLabelled(context, id)
}

// revokeLabelled drops every named token this user holds, and only those: an
// interactive login has no label and must survive minting a new API token.
func (service *Service) revokeLabelled(context context.Context, userID int64) error {
	held, err := service.tokens.Sessions().List(context, userID)
	if err != nil {
		return errors.Internal("failed to read api tokens", err)
	}
	for _, candidate := range held {
		if !candidate.IsAPIToken() {
			continue
		}
		if err := service.tokens.Sessions().Revoke(context, userID, candidate.ID); err != nil {
			return errors.Internal("failed to revoke api token", err)
		}
	}
	return nil
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
