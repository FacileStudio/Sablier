package auth

import (
	"context"
	"database/sql"
	stderrors "errors"
	"log/slog"
	"strings"

	"github.com/FacileStudio/Sablier/apps/api/internal/usercolor"
	"github.com/FacileStudio/Sablier/apps/api/schemas"
	"github.com/FacileStudio/porte"
	"github.com/FacileStudio/porte/kit"
	"github.com/FacileStudio/tronc/errors"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// columns maps porte/kit onto Sablier's users table. The picture lives in
// oidc_picture_url; there is no admin flag.
func columns() kit.Columns {
	return kit.Columns{Picture: "oidc_picture_url"}
}

// Accounts is the porte user store over Sablier's own users table: porte/kit's
// store, plus the one rule the kit has no hook for on an existing row. An
// agent user never signs in, whether the sign-in would reach it by subject or
// by email.
type Accounts struct {
	store *kit.Store
	db    *sql.DB
}

var _ porte.UserStore = (*Accounts)(nil)

// NewAccounts builds the store. It reads the users table once, so it runs
// after the schema migration.
func NewAccounts(ctx context.Context, db *sql.DB, logger *slog.Logger) (*Accounts, error) {
	store, err := kit.NewStore(ctx, kit.Store{
		DB: db, Columns: columns(), OnCreate: fillNewAccount, EraseData: eraseData, Logger: logger,
	})
	if err != nil {
		return nil, err
	}
	return &Accounts{store: store, db: db}, nil
}

// Store is the kit store underneath, for the directory's erase hook.
func (a *Accounts) Store() *kit.Store { return a.store }

// UpsertFromOIDC refuses a sign-in that would land on an agent user, then
// hands the rest to the kit: identity first, then a verified email, then a
// new account.
func (a *Accounts) UpsertFromOIDC(ctx context.Context, claims porte.Claims) (int64, error) {
	agent, err := a.reachesAgent(ctx, claims)
	if err != nil {
		return 0, errors.Internal("failed to resolve the identity", err)
	}
	if agent {
		return 0, errors.Forbidden("agent accounts cannot sign in")
	}
	return a.store.UpsertFromOIDC(ctx, claims)
}

const reachedKindSQL = `SELECT kind FROM (
	(SELECT u.kind, 0 AS rank FROM porte_identities i JOIN users u ON u.id = i.user_id
	  WHERE i.provider IN ($1, $2) AND i.subject = $3 ORDER BY i.provider = $1 DESC LIMIT 1)
	UNION ALL
	(SELECT kind, 1 FROM users WHERE lower(email) = $4 ORDER BY id LIMIT 1)
) reached ORDER BY rank LIMIT 1`

func (a *Accounts) reachesAgent(ctx context.Context, claims porte.Claims) (bool, error) {
	other := claims.Provider + "/"
	if trimmed, ok := strings.CutSuffix(claims.Provider, "/"); ok {
		other = trimmed
	}
	email := strings.ToLower(strings.TrimSpace(claims.Email))
	var kind string
	err := a.db.QueryRowContext(ctx, reachedKindSQL, claims.Provider, other, claims.Subject, email).Scan(&kind)
	if stderrors.Is(err, sql.ErrNoRows) {
		return false, nil
	}
	return kind == schemas.UserKindAgent, err
}

const fillLegacyColumns = `UPDATE users SET color = $2, avatar_url = '', avatar_source = '',
	avatar_upload_path = '', oidc_access_token = '', oidc_refresh_token = '', password_hash = '',
	oidc_token_expiry = '0001-01-01 00:00:00+00', profile_synced_at = '0001-01-01 00:00:00+00'
	WHERE id = $1`

// fillNewAccount gives an account the kit just created what a GORM insert
// used to write: a display colour, chosen against the colours the other users
// hold, and the empty values the profile columns read as.
func fillNewAccount(ctx context.Context, tx *sql.Tx, userID int64) error {
	orm, err := gorm.Open(postgres.New(postgres.Config{Conn: tx}), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		return errors.Internal("failed to open the account transaction", err)
	}
	color, err := usercolor.NextAvailable(ctx, orm)
	if err != nil {
		return errors.Internal("failed to choose a user colour", err)
	}
	if _, err := tx.ExecContext(ctx, fillLegacyColumns, userID, color); err != nil {
		return errors.Internal("failed to fill the new account", err)
	}
	return nil
}
