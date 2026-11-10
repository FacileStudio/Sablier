package auth

import (
	"context"
	"log/slog"
	"net/http"

	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/porte/session"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/gorm"
)

// Service is what is left of Sablier's authentication after porte took the
// credential: the profile lookup the rest of the app reads and the token
// issuer the API key routes call.
type Service struct {
	orm      *gorm.DB
	sessions *session.Manager
	logger   *slog.Logger
}

// NewService wires the auth service onto the given database and session manager.
func NewService(orm *gorm.DB, sessions *session.Manager, logger *slog.Logger) *Service {
	return &Service{orm: orm, sessions: sessions, logger: logger}
}

// RequireAuth is porte's session middleware, re-exported so the eight module
// routers keep passing this one service to middleware.RequireAuth.
func (service *Service) RequireAuth(next http.Handler) http.Handler {
	return service.sessions.RequireAuth(next)
}

// Identity turns the user id porte authenticated into the identity the rest
// of Sablier reads. It is porte/kit's Resolver with Sablier's one extra rule:
// an agent user resolves to no row, so whatever credential names one, no
// request is ever served as an agent. A missing row means the session
// outlived the user and is treated as unauthenticated.
func (service *Service) Identity(ctx context.Context, userID int64) (kit.Identity, error) {
	var out struct {
		ID             int64
		Email          string
		Name           string
		OIDCPictureURL string
	}
	err := service.orm.WithContext(ctx).
		Model(&schemas.User{}).
		Select("id", "coalesce(email, '') AS email", "coalesce(name, '') AS name",
			"coalesce(oidc_picture_url, '') AS oidc_picture_url").
		Where("id = ? AND kind <> ?", userID, schemas.UserKindAgent).
		Scan(&out).Error
	if err != nil {
		return kit.Identity{}, errors.Internal("failed to load the account", err)
	}
	if out.ID == 0 {
		return kit.Identity{}, errors.Unauthorized("invalid auth token")
	}
	return kit.Identity{UserID: out.ID, Email: out.Email, Name: out.Name, PictureURL: out.OIDCPictureURL}, nil
}

// Issue mints a named API token: a porte session with a label and no
// expiry, which is what the separate api_tokens table used to be.
func (service *Service) Issue(ctx context.Context, userID int64, label string) (string, porte.Session, error) {
	return service.sessions.Issue(ctx, userID, label)
}

// Sessions exposes the manager for the modules that list or revoke tokens.
func (service *Service) Sessions() *session.Manager { return service.sessions }
