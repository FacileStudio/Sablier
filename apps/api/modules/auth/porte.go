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

// NewUserStore returns the user store backed by the given database.
func NewUserStore(orm *gorm.DB) *UserStore {
	return &UserStore{orm: orm}
}

var _ porte.UserStore = (*UserStore)(nil)

// UpsertFromOIDC matches on (provider, subject) first, falls back to a verified
// email, and creates the account when neither finds anything.
//
// porte has already done the matching it can: it looks up the identity row and
// only calls this when it needs a user id. The email fallback here is what
// links a pre-existing account to a subject signing in for the first time. It
// is refused for an unverified address because matching on one is an account
// takeover primitive, and refused when the account already holds another
// subject of the same provider, so a recycled address is not handed to a
// newcomer (porte v0.10.0's linking rule, which an app with its own users table
// has to apply itself).
//
// The identity provider is the source of truth for the profile, but an absent
// claim asserts nothing: a provider that stops sending a name or a locale is
// not allowed to blank the stored one.
func (s *UserStore) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	if email == "" {
		return 0, errors.Invalid("the identity provider returned no email")
	}

	user, err := s.resolve(ctx, claims, email)
	if err != nil {
		return 0, err
	}
	if user.IsAgent() {
		return 0, errors.Forbidden("agent accounts cannot sign in")
	}

	updates := map[string]any{"email": email, "oidc_picture_url": oidcavatar.PhotoURL(claims.Picture)}
	if displayName := claims.DisplayName(); displayName != "" {
		updates["name"] = displayName
	}
	if claims.Locale != "" {
		updates["locale"] = claims.Locale
	}
	if err := s.orm.WithContext(ctx).Model(&schemas.User{}).Where("id = ?", user.ID).Updates(updates).Error; err != nil {
		return 0, errors.Internal("failed to update the account", err)
	}
	return user.ID, nil
}

func (s *UserStore) resolve(ctx context.Context, claims porte.Claims, email string) (*schemas.User, error) {
	var user schemas.User
	err := s.orm.WithContext(ctx).
		Joins("JOIN porte_identities i ON i.user_id = users.id AND i.provider = ? AND i.subject = ?", claims.Provider, claims.Subject).
		First(&user).Error
	switch {
	case err == nil:
		return &user, nil
	case !stderrors.Is(err, gorm.ErrRecordNotFound):
		return nil, errors.Internal("failed to resolve the identity", err)
	}

	err = s.orm.WithContext(ctx).Where("lower(email) = ?", email).Order("id").First(&user).Error
	switch {
	case err == nil:
		return &user, s.checkLinkable(ctx, user.ID, claims)
	case stderrors.Is(err, gorm.ErrRecordNotFound):
		return s.create(ctx, email, claims.DisplayName(), claims.Locale)
	}
	return nil, errors.Internal("failed to look up the account", err)
}

func (s *UserStore) checkLinkable(ctx context.Context, userID int64, claims porte.Claims) error {
	if !claims.EmailVerified {
		return errors.Conflict("an account with this email already exists and the identity provider did not verify the address")
	}
	var held int64
	err := s.orm.WithContext(ctx).Table("porte_identities").
		Where("user_id = ? AND provider = ?", userID, claims.Provider).
		Count(&held).Error
	if err != nil {
		return errors.Internal("failed to check the account's identities", err)
	}
	if held > 0 {
		return errors.Conflict("an account with this email is already linked to another identity")
	}
	return nil
}

// create is the one place a Sablier account comes into existence, so the
// colour is assigned in exactly one place.
func (s *UserStore) create(ctx context.Context, email, name, locale string) (*schemas.User, error) {
	color, err := usercolor.NextAvailable(ctx, s.orm)
	if err != nil {
		return nil, errors.Internal("failed to choose a user colour", err)
	}
	user := schemas.User{Email: email, Name: name, Color: color, Locale: locale}
	if locale == "" {
		user.Locale = "en"
	}
	if err := s.orm.WithContext(ctx).Create(&user).Error; err != nil {
		return nil, errors.Internal("failed to create the account", err)
	}
	return &user, nil
}
