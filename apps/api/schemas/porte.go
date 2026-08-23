package schemas

import "gorm.io/gorm"

// AdoptPorte moves Sablier's credentials onto porte's tables.
//
// issuer is the configured OIDC_ISSUER, and it has to be passed in because it
// is half of the account matching key: porte finds an identity by (provider,
// subject), and the provider is the issuer. Backfilling with the wrong value —
// or with a placeholder — would leave every existing SSO user unmatched, which
// degrades silently to the email fallback rather than failing.
//
// An empty issuer skips only the identity backfill. The sessions and the API
// tokens still move, because they are what keeps people signed in.
//
// Every statement below is PostgreSQL — a DO block, to_regclass, a partial
// index. The app's own tests run against in-memory SQLite, and porte/pg is
// Postgres-only anyway, so on any other dialect there is nothing here worth
// half-executing and the call returns early.
func AdoptPorte(db *gorm.DB, issuer string) error {
	if db.Dialector.Name() != "postgres" {
		return nil
	}
	statements := []string{porteSchema, carryLegacySessionsOver, carryAPITokensOver}
	for _, statement := range statements {
		if err := db.Exec(statement).Error; err != nil {
			return err
		}
	}
	if issuer == "" {
		return nil
	}
	return db.Exec(adoptOIDCSubjects, issuer).Error
}

// porteSchema is porte/pg's Schema with its porte_users table left out and the
// foreign keys repointed at Sablier's own users.
//
// porte offers UserStore as the escape hatch for exactly this. Sablier's users
// table carries a display colour, a billing rate and a workday length, and it
// is referenced across the whole codebase; moving it would be a rewrite of
// everything except authentication. The other three stores come from porte/pg
// unchanged — they only ever touch the tables below.
//
// The one UPDATE is porte v0.3.0's, copied along with the rest: a password
// identity is keyed on the account id now, not on the address it can move away
// from. It is idempotent, since the predicate is false for every row once it
// has run, it leaves federated subjects alone, and it is allowed to fail — the
// only way it can is an account holding two password identities, which is
// exactly the state the old key let this app reach and which nothing should
// paper over by picking one.
const porteSchema = `
CREATE TABLE IF NOT EXISTS porte_identities (
	user_id         bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	provider        text NOT NULL,
	subject         text NOT NULL,
	password_hash   text NOT NULL DEFAULT '',
	access_token    text NOT NULL DEFAULT '',
	refresh_token   text NOT NULL DEFAULT '',
	token_expiry    timestamptz,
	roles           jsonb,
	roles_synced_at timestamptz,
	synced_at       timestamptz,
	created_at      timestamptz DEFAULT now(),
	PRIMARY KEY (provider, subject)
);
CREATE INDEX IF NOT EXISTS porte_identities_user_idx ON porte_identities (user_id);
ALTER TABLE porte_identities ADD COLUMN IF NOT EXISTS created_at timestamptz;
ALTER TABLE porte_identities ALTER COLUMN created_at SET DEFAULT now();

UPDATE porte_identities SET subject = user_id::text
 WHERE provider = 'local' AND subject <> user_id::text;

CREATE TABLE IF NOT EXISTS porte_sessions (
	id           bigserial PRIMARY KEY,
	token_hash   text NOT NULL UNIQUE,
	user_id      bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	label        text NOT NULL DEFAULT '',
	created_at   timestamptz NOT NULL DEFAULT now(),
	last_used_at timestamptz NOT NULL DEFAULT now(),
	expires_at   timestamptz
);
CREATE INDEX IF NOT EXISTS porte_sessions_user_idx ON porte_sessions (user_id);
CREATE INDEX IF NOT EXISTS porte_sessions_expiry_idx ON porte_sessions (expires_at)
	WHERE expires_at IS NOT NULL;

CREATE TABLE IF NOT EXISTS porte_login_codes (
	code_hash   text PRIMARY KEY,
	user_id     bigint NOT NULL REFERENCES users(id) ON DELETE CASCADE,
	expires_at  timestamptz NOT NULL,
	consumed_at timestamptz
);
`

// carryLegacySessionsOver moves the pre-porte sessions table across instead of
// dropping it, so adopting porte does not sign every existing user out.
//
// Both tables store the SHA-256 hex of a 32-byte token and never the token
// itself, so a row copied here keeps authenticating the credential already in
// somebody's browser. last_used_at is stamped now rather than copied from
// created_at: porte retires a cookie session idle for seven days and the old
// table recorded no use at all, so carrying created_at over would sign out
// everyone who last signed in more than a week ago, on the deploy meant to
// keep them.
const carryLegacySessionsOver = `
DO $$
BEGIN
	IF to_regclass('sessions') IS NOT NULL THEN
		INSERT INTO porte_sessions (token_hash, user_id, created_at, last_used_at, expires_at)
		SELECT s.token, s.user_id, s.created_at, now(), s.expires_at
		  FROM sessions s JOIN users u ON u.id = s.user_id
		 WHERE s.expires_at > now()
		ON CONFLICT (token_hash) DO NOTHING;
		DROP TABLE sessions;
	END IF;
END
$$;
`

// carryAPITokensOver folds the separate api_tokens table into porte's sessions
// as labelled rows.
//
// porte.Session.Label exists for precisely this: two apps in the suite grew
// their own ApiToken type and table for what is a session with a name on it
// and no expiry, which is one more mechanism than the problem needs. A null
// expires_at never expires, and the idle window deliberately does not apply to
// the bearer transport, so a token wired into a script keeps working.
const carryAPITokensOver = `
DO $$
BEGIN
	IF to_regclass('api_tokens') IS NOT NULL THEN
		INSERT INTO porte_sessions (token_hash, user_id, label, created_at, last_used_at, expires_at)
		SELECT t.token, t.user_id, COALESCE(NULLIF(t.name, ''), 'API token'), t.created_at, now(), NULL
		  FROM api_tokens t JOIN users u ON u.id = t.user_id
		ON CONFLICT (token_hash) DO NOTHING;
		DROP TABLE api_tokens;
	END IF;
END
$$;
`

// adoptOIDCSubjects moves the federated identity off the user row and into the
// identity table porte reads.
//
// Without it nothing breaks loudly: porte finds no identity for (issuer,
// subject), falls back to matching the verified email, finds the same user and
// links it on the next login. But the email fallback is the weaker path by
// design — it is the one an identity provider that lets a user set any address
// can abuse — and relying on it for every existing account, on the one deploy
// where it would go unnoticed, is not a trade worth making. The OIDC tokens
// come across in the same statement so a profile refresh does not need a
// second login.
const adoptOIDCSubjects = `
INSERT INTO porte_identities (user_id, provider, subject, access_token, refresh_token, token_expiry, synced_at)
SELECT id, ?, oidc_subject,
       COALESCE(oidc_access_token, ''), COALESCE(oidc_refresh_token, ''),
       NULLIF(oidc_token_expiry, '0001-01-01 00:00:00+00'::timestamptz),
       NULLIF(profile_synced_at, '0001-01-01 00:00:00+00'::timestamptz)
  FROM users
 WHERE oidc_subject IS NOT NULL AND oidc_subject <> ''
ON CONFLICT (provider, subject) DO NOTHING;
`
