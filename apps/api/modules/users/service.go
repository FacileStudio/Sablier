package users

import (
	"context"
	"database/sql"
	stderrors "errors"
	"strconv"
	"time"

	"github.com/FacileStudio/Sablier/apps/api/internal/middleware"
	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// TokenIssuer is the slice of the auth service this module needs: minting a
// labelled session and reaching the manager that lists and revokes them.
type TokenIssuer interface {
	Issue(ctx context.Context, userID int64, label string) (string, porte.Session, error)
	Sessions() *session.Manager
}

// Service backs the users module: reading, updating and deleting accounts,
// their profiles and their API tokens.
type Service struct {
	orm        *gorm.DB
	tokens     TokenIssuer
	admins     middleware.Admins
	controller *Controller
}

// NewService wires the users service with a fresh route controller. admins is
// SABLIER_ADMINS, which /users/me reports as is_admin so the client can show
// the app-wide settings to the people allowed to change them.
func NewService(orm *gorm.DB, tokens TokenIssuer, admins middleware.Admins) *Service {
	service := &Service{orm: orm, tokens: tokens, admins: admins}
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

// visibleUsers narrows query to the accounts callerID may see: itself, its
// own agents, and everyone who shares a space with it. Nobody else's email or
// rate is any of the caller's business.
func visibleUsers(query *gorm.DB, callerID int64) *gorm.DB {
	peers := query.Session(&gorm.Session{NewDB: true}).Model(&schemas.SpaceMember{}).
		Select("user_id").Where("space_id IN (?)", schemas.MemberSpaces(query, callerID))
	return query.Where("users.id = @caller OR (users.kind = @agent AND users.owner_user_id = @caller) OR users.id IN (@peers)",
		sql.Named("caller", callerID), sql.Named("agent", schemas.UserKindAgent), sql.Named("peers", peers))
}

func (service *Service) listUsers(context context.Context, callerID string) ([]Peer, error) {
	caller, err := strconv.ParseInt(callerID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	if err := usercolor.BackfillMissing(context, service.orm); err != nil {
		return nil, errors.Internal("failed to backfill user colors", err)
	}

	var records []schemas.User
	query := visibleUsers(service.orm.WithContext(context), caller)
	if err := query.Order("name asc, email asc, id asc").Find(&records).Error; err != nil {
		return nil, errors.Internal("failed to list users", err)
	}

	users := make([]Peer, 0, len(records))
	for _, record := range records {
		users = append(users, mapUser(record).peer())
	}

	return users, nil
}

// getPeer returns userID when callerID may see it, and not_found otherwise, so
// the answer does not reveal whether an invisible account exists.
func (service *Service) getPeer(context context.Context, callerID, userID string) (*Peer, error) {
	caller, err := strconv.ParseInt(callerID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.NotFound("user not found")
	}
	var count int64
	if err := visibleUsers(service.orm.WithContext(context).Model(&schemas.User{}), caller).
		Where("users.id = ?", id).Count(&count).Error; err != nil {
		return nil, errors.Internal("failed to read user", err)
	}
	if count == 0 {
		return nil, errors.NotFound("user not found")
	}
	user, err := service.getUser(context, userID)
	if err != nil {
		return nil, err
	}
	peer := user.peer()
	return &peer, nil
}

// updateUser applies the non-nil profile fields.
func (service *Service) updateUser(context context.Context, userID string, name *string, color *string, rate *float64, rateType *string, workdayHours *float64) (*User, error) {
	id, err := strconv.ParseInt(userID, 10, 64)
	if err != nil {
		return nil, errors.Internal("failed to parse user id", err)
	}

	updates := map[string]any{}
	if name != nil {
		updates["name"] = *name
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

func (user *User) peer() Peer {
	return Peer{
		ID:           user.ID,
		Email:        user.Email,
		Name:         user.Name,
		AvatarURL:    user.AvatarURL,
		Color:        user.Color,
		Rate:         user.Rate,
		RateType:     user.RateType,
		WorkdayHours: user.WorkdayHours,
		CreatedAt:    user.CreatedAt,
		Kind:         user.Kind,
		OwnerUserID:  user.OwnerUserID,
	}
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
		Locale:       record.Locale,
		Rate:         record.Rate,
		RateType:     rateType,
		WorkdayHours: workdayHours,
		CreatedAt:    record.CreatedAt.UTC().Format(time.RFC3339),
		Kind:         record.Kind,
		OwnerUserID:  record.OwnerUserID,
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
