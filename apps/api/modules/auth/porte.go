package auth

import (
	"context"
	stderrors "errors"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/internal/oidcavatar"
	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// UserStore resolves a login to a row in Sablier's own users table.
//
// It is porte's escape hatch, taken deliberately. porte/pg ships a porte_users
// table and Sablier could have moved onto it, but this users row carries a
// billing rate, a workday length and an assigned display colour, and it is
// referenced by projects, tasks, time entries, spaces and settings. Moving it
// would be a rewrite of everything except authentication.
//
// The colour is the reason porte asks the app to own this method rather than
// owning the write itself: choosing one requires knowing which colours the
// other users already hold, which is a question only this app can answer.
type UserStore struct {
	orm *gorm.DB
}

func NewUserStore(orm *gorm.DB) *UserStore {
	return &UserStore{orm: orm}
}

var (
	_ porte.UserStore         = (*UserStore)(nil)
	_ porte.PasswordUserStore = (*UserStore)(nil)
)

// UpsertFromOIDC matches on (provider, subject) first, falls back to a verified
// email, and creates the account when neither finds anything.
//
// porte has already done the matching it can: it looks up the identity row and
// only calls this when it needs a user id. The email fallback here is what
// links a pre-existing account to a subject signing in for the first time, and
// it is refused for an unverified address because matching on one is an account
// takeover primitive.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return 0, errors.Invalid("the identity provider returned no email")
	}

	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	switch {
	case err == nil:
	case stderrors.Is(err, gorm.ErrRecordNotFound):
		err = s.orm.WithContext(ctx).Where("email = ?", email).First(&user).Error
		switch {
		case err == nil:
			if !claims.EmailVerified {
				return 0, errors.Conflict("an account with this email already exists and the identity provider did not verify the address")
			}
		case stderrors.Is(err, gorm.ErrRecordNotFound):
			created, createErr := s.create(ctx, email, claims.DisplayName())
			if createErr != nil {
				return 0, createErr
			}
			user = *created
		default:
			return 0, errors.Internal("failed to look up the account", err)
		}
	default:
		return 0, errors.Internal("failed to resolve the identity", err)
	}

	// The identity provider is the source of truth for the profile, but an
	// absent claim asserts nothing: a provider that stops sending a name
	// should not blank it here.
	updates := map[string]any{"email": email, "oidc_picture_url": oidcavatar.PhotoURL(claims.Picture)}
	if displayName := claims.DisplayName(); displayName != "" {
		updates["name"] = displayName
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return 0, errors.Internal("failed to update the account", err)
	}
	return user.ID, nil
}

// CreateFromPassword creates a user row for a local registration. porte has
// already validated the address and hashed the password.
func (s *UserStore) CreateFromPassword(ctx context.Context, email, name string) (int64, error) {
	user, err := s.create(ctx, email, name)
	if err != nil {
		return 0, err
	}
	return user.ID, nil
}

// FindByEmail returns the user id for an address, or porte.ErrNotFound.
func (s *UserStore) FindByEmail(ctx context.Context, email string) (int64, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).Select("id").Where("email = ?", email).First(&user).Error
	if stderrors.Is(err, gorm.ErrRecordNotFound) {
		return 0, porte.ErrNotFound
	}
	if err != nil {
		return 0, errors.Internal("failed to look up the account", err)
	}
	return user.ID, nil
}

// CountUsers backs porte/local's registration gate.
func (s *UserStore) CountUsers(ctx context.Context) (int64, error) {
	var count int64
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Count(&count).Error; err != nil {
		return 0, errors.Internal("failed to count users", err)
	}
	return count, nil
}

// create is the one place a Sablier account comes into existence, so the
// colour is assigned in exactly one place regardless of how the human arrived.
func (s *UserStore) create(ctx context.Context, email, name string) (*schemas.User, error) {
	color, err := usercolor.NextAvailable(ctx, s.orm)
	if err != nil {
		return nil, errors.Internal("failed to choose a user colour", err)
	}
	user := schemas.User{Email: email, Name: name, Color: color}
	if err := s.orm.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, errors.Internal("failed to create the account", err)
	}
	return &user, nil
}
