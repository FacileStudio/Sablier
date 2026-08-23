package schemas

import (
	"os"
	"testing"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// openPostgres returns a connection to a throwaway database, or skips.
//
// AdoptPorte is PostgreSQL — a DO block, to_regclass, ON CONFLICT — and it is
// skipped outright on any other dialect, so this app's in-memory SQLite tests
// prove nothing at all about it. It is also the one piece of this migration
// that can sign every user out, which makes an untested version worse than no
// version.
func openPostgres(t *testing.T) *gorm.DB {
	t.Helper()
	url := os.Getenv("SABLIER_TEST_DATABASE_URL")
	if url == "" {
		t.Skip("SABLIER_TEST_DATABASE_URL is unset")
	}
	db, err := gorm.Open(postgres.Open(url), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	if err := db.Exec("DROP SCHEMA public CASCADE; CREATE SCHEMA public").Error; err != nil {
		t.Fatalf("reset the schema: %v", err)
	}
	return db
}

const testIssuer = "https://porte.test/application/o/sablier/"

// seedPrePorte builds the shape production was in before porte: a users table,
// the old sessions table, the separate api_tokens table, and a federated
// identity recorded on the user row.
func seedPrePorte(t *testing.T, db *gorm.DB) {
	t.Helper()
	if err := db.AutoMigrate(&User{}); err != nil {
		t.Fatalf("seed users: %v", err)
	}
	statements := []string{
		`CREATE TABLE sessions (token text PRIMARY KEY, user_id bigint NOT NULL, expires_at timestamptz, created_at timestamptz)`,
		`CREATE TABLE api_tokens (token text PRIMARY KEY, user_id bigint NOT NULL, name text, created_at timestamptz)`,
		`INSERT INTO users (id, email, name, oidc_subject, oidc_access_token, oidc_refresh_token, profile_synced_at, color, password_hash, created_at)
		 VALUES (1, 'camille@facile.studio', 'Camille', 'sub-1', 'access', 'refresh', now(), '#fff', '', now())`,
		`INSERT INTO users (id, email, name, oidc_subject, color, password_hash, created_at)
		 VALUES (2, 'noah@facile.studio', 'Noah', NULL, '#eee', '$argon2id$fake', now())`,
		`INSERT INTO sessions (token, user_id, expires_at, created_at) VALUES
			('live', 1, now() + interval '10 days', now() - interval '40 days'),
			('dead', 1, now() - interval '1 day', now() - interval '31 days')`,
		`INSERT INTO api_tokens (token, user_id, name, created_at) VALUES ('cli-token', 2, 'CLI', now())`,
	}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			t.Fatalf("seed: %v\n%s", err, statement)
		}
	}
}

// Nobody may be signed out by this deploy, and the CLI must keep working. Both
// tables store the SHA-256 hex of a token and nothing else, so the rows move
// and the credentials already in a browser and in ~/.sablier.yml keep working.
//
// last_used_at is asserted to have been stamped rather than copied: carrying
// created_at over would put the browser session 40 days into the seven-day
// idle window and sign the user out on the deploy meant to keep them. The API
// token must become a labelled session with no expiry — the CLI has no login
// flow, so if that row does not survive there is no way to get another one.
func TestAdoptPorteKeepsEverybodySignedIn(t *testing.T) {
	db := openPostgres(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var carried struct {
		UserID     int64
		Label      string
		LastUsedAt time.Time
	}
	if err := db.Raw(`SELECT user_id, label, last_used_at FROM porte_sessions WHERE token_hash = 'live'`).Scan(&carried).Error; err != nil {
		t.Fatalf("read the carried session: %v", err)
	}
	if carried.UserID != 1 || carried.Label != "" {
		t.Fatalf("the browser session did not survive as an unlabelled session: %+v", carried)
	}
	if time.Since(carried.LastUsedAt) > time.Hour {
		t.Fatalf("last_used_at was copied instead of stamped: %v", carried.LastUsedAt)
	}

	var expired int64
	if err := db.Raw(`SELECT count(*) FROM porte_sessions WHERE token_hash = 'dead'`).Scan(&expired).Error; err != nil {
		t.Fatalf("count expired: %v", err)
	}
	if expired != 0 {
		t.Fatal("an already-expired session was carried over")
	}

	var token struct {
		UserID    int64
		Label     string
		ExpiresAt *time.Time
	}
	if err := db.Raw(`SELECT user_id, label, expires_at FROM porte_sessions WHERE token_hash = 'cli-token'`).Scan(&token).Error; err != nil {
		t.Fatalf("read the carried api token: %v", err)
	}
	if token.UserID != 2 || token.Label != "CLI" {
		t.Fatalf("the api token did not survive: %+v", token)
	}
	if token.ExpiresAt != nil {
		t.Fatalf("the api token was given an expiry: %v", token.ExpiresAt)
	}

	for _, table := range []string{"sessions", "api_tokens"} {
		var remaining *string
		if err := db.Raw(`SELECT to_regclass(?)::text`, table).Scan(&remaining).Error; err != nil {
			t.Fatalf("check %s: %v", table, err)
		}
		if remaining != nil {
			t.Fatalf("the legacy %s table survived", table)
		}
	}

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt is not idempotent: %v", err)
	}
}

// The federated identity moves off the user row. Without it porte finds no
// identity, falls back to matching the verified email and relinks on the next
// login — which works, but leans the whole existing user base on the weaker of
// the two matching paths, on the one deploy where nobody would notice. The
// user seeded without a subject must not gain one, so exactly one identity row
// is expected.
func TestAdoptPorteMovesTheOIDCSubject(t *testing.T) {
	db := openPostgres(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var identity struct {
		UserID       int64
		AccessToken  string
		RefreshToken string
		SyncedAt     *time.Time
	}
	err := db.Raw(
		`SELECT user_id, access_token, refresh_token, synced_at FROM porte_identities WHERE provider = ? AND subject = 'sub-1'`,
		testIssuer,
	).Scan(&identity).Error
	if err != nil {
		t.Fatalf("read the identity: %v", err)
	}
	if identity.UserID != 1 {
		t.Fatal("the oidc subject was not adopted")
	}
	if identity.AccessToken != "access" || identity.RefreshToken != "refresh" {
		t.Fatalf("the provider tokens did not come across: %+v", identity)
	}
	if identity.SyncedAt == nil {
		t.Fatal("profile_synced_at did not come across, so the next request refreshes needlessly")
	}

	var rows int64
	if err := db.Raw(`SELECT count(*) FROM porte_identities`).Scan(&rows).Error; err != nil {
		t.Fatalf("count: %v", err)
	}
	if rows != 1 {
		t.Fatalf("expected exactly one identity, got %d", rows)
	}
}

// An empty issuer is a deployment with SSO switched off. The sessions and the
// API tokens still have to move — they are what keeps people signed in — but
// there is no provider to key an identity against.
func TestAdoptPorteWithoutAnIssuerStillMovesTheCredentials(t *testing.T) {
	db := openPostgres(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, ""); err != nil {
		t.Fatalf("adopt: %v", err)
	}

	var sessions, identities int64
	if err := db.Raw(`SELECT count(*) FROM porte_sessions`).Scan(&sessions).Error; err != nil {
		t.Fatalf("count sessions: %v", err)
	}
	if err := db.Raw(`SELECT count(*) FROM porte_identities`).Scan(&identities).Error; err != nil {
		t.Fatalf("count identities: %v", err)
	}
	if sessions != 2 {
		t.Fatalf("expected the live session and the api token, got %d", sessions)
	}
	if identities != 0 {
		t.Fatalf("an identity was keyed against no provider: %d rows", identities)
	}
}

// A password identity created under porte v0.2 was keyed on the email address,
// which is the mutable half of the account: changing an address left the
// credential behind on the old one, and setting a password afterwards inserted
// a *second* identity rather than replacing the first, because Save upserts on
// (provider, subject). The re-key is part of the schema, so a second AdoptPorte
// is the production path and not a contrivance.
//
// The federated subject must not move with it. That one belongs to the identity
// provider, and re-keying it would unmatch every SSO account on the same deploy.
func TestAdoptPorteRekeysPasswordIdentitiesOntoTheAccountID(t *testing.T) {
	db := openPostgres(t)
	seedPrePorte(t, db)

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("adopt: %v", err)
	}
	if err := db.Exec(
		`INSERT INTO porte_identities (user_id, provider, subject, password_hash)
		 VALUES (2, 'local', 'noah@facile.studio', '$argon2id$fake')`).Error; err != nil {
		t.Fatalf("seed the v0.2 password identity: %v", err)
	}

	if err := AdoptPorte(db, testIssuer); err != nil {
		t.Fatalf("re-adopt: %v", err)
	}

	var local struct {
		Subject string
		UserID  int64
	}
	if err := db.Raw(`SELECT subject, user_id FROM porte_identities WHERE provider = 'local'`).Scan(&local).Error; err != nil {
		t.Fatalf("read the password identity: %v", err)
	}
	if local.Subject != "2" || local.UserID != 2 {
		t.Fatalf("the password identity is keyed on %q, not on the account id: %+v", local.Subject, local)
	}

	var federated string
	if err := db.Raw(`SELECT subject FROM porte_identities WHERE provider = ?`, testIssuer).Scan(&federated).Error; err != nil {
		t.Fatalf("read the federated identity: %v", err)
	}
	if federated != "sub-1" {
		t.Fatalf("the federated subject was re-keyed too, unmatching the SSO account: %q", federated)
	}
}
