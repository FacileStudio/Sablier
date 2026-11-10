package main

import (
	"context"
	"log/slog"
	"net/http"
	"strconv"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/internal/authcontext"
	"github.com/FacileStudio/Sablier/apps/api/internal/env"
	"github.com/FacileStudio/Sablier/apps/api/internal/plans"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte/exchange"
	"github.com/FacileStudio/tronc/bulle"
	"github.com/FacileStudio/tronc/errors"
	"github.com/FacileStudio/tronc/httpjson"

	"github.com/go-chi/chi/v5"
	"gorm.io/gorm"
)

type ideasConfig struct {
	URL        string
	UpgradeURL string
	Token      bulle.TokenSource
	Forget     func(context.Context)
	Gate       *plans.Gate
}

type registreSpaceKey struct{}

func newIdeasConfig(exchanger *exchange.Exchanger, gate *plans.Gate, appEnv *env.Config) *ideasConfig {
	bulleURL := strings.TrimSpace(appEnv.BulleURL)
	if exchanger == nil || bulleURL == "" || appEnv.OIDC == nil {
		return nil
	}
	source := exchanger.Source(exchange.Target{Audience: "bulle", Scopes: []string{"bulle.read", "bulle.write"}})
	return &ideasConfig{
		URL:        bulleURL,
		UpgradeURL: strings.TrimRight(appEnv.OIDC.Issuer, "/") + "/spaces",
		Token:      source.Token,
		Forget:     source.Forget,
		Gate:       gate,
	}
}

func newExchanger(store exchange.Identities, appEnv *env.Config, logger *slog.Logger) (*exchange.Exchanger, error) {
	if appEnv.OIDC == nil || strings.TrimSpace(appEnv.BulleURL) == "" {
		return nil, nil
	}
	return exchange.New(exchange.Config{
		Issuer:       appEnv.OIDC.Issuer,
		ClientID:     appEnv.OIDC.ClientID,
		ClientSecret: appEnv.OIDC.ClientSecret,
		Identities:   store,
		Logger:       logger,
	})
}

func mountIdeas(api chi.Router, db *gorm.DB, cfg *ideasConfig, authenticate func(http.Handler) http.Handler) error {
	if cfg == nil {
		return nil
	}
	var mountErr error
	api.Group(func(group chi.Router) {
		group.Use(authenticate)
		group.Use(scopeIdeasSpace(db))
		mountErr = bulle.MountIdeas(group, bulle.Options{
			App:         "sablier",
			URL:         cfg.URL,
			TokenSource: cfg.Token,
			Forget:      cfg.Forget,
			Space:       ideasSpace,
			AIAccess:    ideasAIAccess(cfg.Gate),
			UpgradeURL:  cfg.UpgradeURL,
		})
	})
	return mountErr
}

func ideasSpace(r *http.Request) string {
	id, _ := r.Context().Value(registreSpaceKey{}).(string)
	return id
}

func ideasAIAccess(gate *plans.Gate) func(*http.Request) bool {
	return func(r *http.Request) bool {
		identity, ok := authcontext.IdentityFromContext(r.Context())
		if !ok {
			return false
		}
		userID, err := strconv.ParseInt(identity.UserID, 10, 64)
		if err != nil {
			return false
		}
		allowed, err := gate.HasAI(r.Context(), userID)
		return err == nil && allowed
	}
}

func scopeIdeasSpace(db *gorm.DB) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			spaceID := r.URL.Query().Get("space_id")
			if spaceID == "" {
				next.ServeHTTP(w, r)
				return
			}
			registreID, err := registreSpace(r.Context(), db, spaceID)
			if err != nil {
				httpjson.WriteError(w, errors.NotFound("space not found"))
				return
			}
			next.ServeHTTP(w, r.WithContext(context.WithValue(r.Context(), registreSpaceKey{}, registreID)))
		})
	}
}

func registreSpace(ctx context.Context, db *gorm.DB, spaceID string) (string, error) {
	identity, ok := authcontext.IdentityFromContext(ctx)
	if !ok {
		return "", gorm.ErrRecordNotFound
	}
	userID, err := strconv.ParseInt(identity.UserID, 10, 64)
	if err != nil {
		return "", err
	}
	if err := schemas.RequireMember(db.WithContext(ctx), userID, spaceID); err != nil {
		return "", err
	}
	var space schemas.Space
	if err := db.WithContext(ctx).Where("id = ?", spaceID).First(&space).Error; err != nil {
		return "", err
	}
	if space.RegistreSpaceID == nil || *space.RegistreSpaceID == "" {
		return "", gorm.ErrRecordNotFound
	}
	return *space.RegistreSpaceID, nil
}
