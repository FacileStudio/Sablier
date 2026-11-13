# Sablier — Architecture

How a request reaches the database, what the tables look like, how a session is
authenticated, and how Sablier talks to the rest of the suite.

## Runtime topology

```
Internet ──▶ Traefik ──▶ Go binary (:4000) ──┬──▶ /health /ready       liveness, readiness
                                              ├──▶ /api/*              eight modules
                                              ├──▶ /docs /openapi      Scalar + OpenAPI
                                              ├──▶ /files/*            STORAGE_DIR
                                              └──▶ /*                  SPA catch-all
                                                              │
                                                        Postgres 16
                                                              │
                        Antenne (WebSocket) ◀── outbox drainer ──▶ antenne_outbox
                        Journal (HTTP)        ◀── slog handler
```

One process, one container, one Traefik router. The SvelteKit client is built to static
files at image build time and served by the same Go binary through `tronc`'s `spa`
package, so the client never needs an API base URL — it calls `/api/...` on its own
origin. `apps/client/src/lib/backend.ts` hard-codes `backendBaseUrl = ''` for exactly
that reason.

## Components

| Piece | Where | What it does |
|---|---|---|
| Entrypoint | `apps/api/main.go` | Loads config, opens the DB, migrates, builds the router, starts two workers, shuts down on `SIGINT`/`SIGTERM` |
| Router | `tronc/httpx.NewRouter` | Chi router with the suite's logging, recovery, and CORS middleware |
| Modules | `apps/api/modules/*` | One package per domain, each with `router.go`, `controller.go`, `service.go`, `types.go`, `documentation.go` |
| Migrations | `apps/api/migrations` | Ordered goose SQL, embedded and applied at boot by `tronc/migrate` |
| Schemas | `apps/api/schemas` | GORM models (the query layer) plus `Migrate`, which applies the migrations and then the data steps of `Reconcile` |
| Internal | `apps/api/internal/*` | `database`, `middleware`, `env`, `authcontext`, `authcrypto`, `usercolor`, `webhook`, `worker`, `documentation` |
| Client | `apps/client` | SvelteKit 5 with `adapter-static` and `fallback: index.html` |

The eight modules are `auth`, `projects`, `timeentries`, `users`, `settings`, `spaces`,
`antenne`, and `notifications`. All of them are registered inside a single
`router.Route("/api", ...)` block, so every application route lives under `/api`.

## Request lifecycle

1. Traefik terminates TLS and forwards to the container on port `4000`.
2. `tronc/httpx` middleware attaches the request logger, recovers panics, and applies
   CORS using `CORS_ALLOWED_ORIGINS`.
3. `health.Mount` answers `/health`, `/ready`, `/api/health`, and `/api/ready` before
   anything else. `/ready` pings Postgres with a two-second timeout and returns 503 as
   `{"status":"not_ready"}` when the ping fails.
4. `/api/*` dispatches into a module router. Protected routes run
   `internal/middleware.RequireAuth`: porte/kit's `RequireAuth` resolves the cookie or bearer
   credential, refuses an agent user, and the middleware copies the identity onto the request
   context as `authcontext.Identity{UserID, Email}`, with the user id as a decimal string.
5. The module controller validates and delegates to its service, which talks to GORM.
6. Responses are written by `tronc/httpjson`. Errors constructed with `tronc/errors`
   (`Invalid`, `Unauthorized`, `Conflict`, `Internal`) map to their status codes.
7. Anything that is not `/api`, `/files`, `/docs`, `/openapi`, or a health route falls
   through to the SPA handler, which serves `index.html` for extensionless paths and
   404s for missing assets.

## Authentication

Two token families, both presented as `Authorization: Bearer <token>` and both stored
hashed:

- **Session tokens** — created by the OIDC callback. Rows in `sessions`, valid for 30 days.
- **API tokens** — created by `POST /api/users/me/api-token`, one per user, no expiry.
  Rows in `api_tokens`.

`authenticateRequest` hashes the presented token, looks for a matching session first,
then falls back to an API token. A session past `expires_at` is rejected as
`expired auth token`.

### OIDC

OIDC is additive. `env.Load` builds an `OIDCConfig` only when `OIDC_ISSUER` is set, and
then requires `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, and `OIDC_REDIRECT_URL` alongside
it. When it is configured, three extra routes appear:

- `GET /api/auth/oidc` sets a short-lived `oidc_state` cookie (10 minutes) and redirects
  to the provider.
- `GET /api/auth/oidc/callback` verifies state and the ID token, upserts the user through
  `modules/auth.Accounts` (porte/kit's store over the `users` table), creates a session, and
  redirects to `OIDC_SUCCESS_URL`, which defaults to the first entry of `CORS_ALLOWED_ORIGINS`.
  The store links by identity, then by a verified email, then creates the account; it
  rewrites the stored email only when the provider marks it verified, and refuses any
  sign-in that would land on an agent user. A new account gets its colour in the kit's
  `OnCreate` hook.
- `POST /api/auth/sync-profile` re-pulls the provider profile for the logged-in user.

The scopes requested are `openid`, `email`, `profile`, and `offline_access`. Access,
refresh, and expiry are stored on the user row, along with `profile_synced_at`. The
provider `picture` claim is stored as a URL in `oidc_picture_url`, never downloaded, and
boot backfills `<issuer>/avatars/<subject>` for every identity under the configured issuer,
so users who have not signed in since still get their Registre picture.

There is no local sign-up and no password: porte mounts no register or login route while
`OIDC_ISSUER` is set, and the API refuses to boot in production without it.

## Data model

| Table | Key columns | Notes |
|---|---|---|
| `users` | `id`, `email` unique, `name`, `color`, `password_hash`, `rate`, `rate_type`, `workday_hours` | Also holds `avatar_url`, `avatar_source`, `oidc_picture_url`, the OIDC tokens, and `profile_synced_at` |
| `sessions` | `token` PK (hashed), `user_id`, `expires_at` | 30-day browser sessions |
| `api_tokens` | `token` PK (hashed), `user_id`, `name` | One long-lived token per user |
| `spaces` | `id` UUID PK, `name`, `description`, `registre_space_id` (nullable, unique) | UUID generated in `BeforeCreate`; linked rows mirror Registre |
| `space_members` | `id` UUID PK, unique `(space_id, user_id)`, `role` | Roles are `owner`, `admin`, `member` |
| `projects` | `id`, `name`, `owner_id`, `space_id`, `facile_id` unique | `facile_id` is the cross-app identity used by pool sync |
| `tasks` | `id`, unique `(project_id, name)`, `status`, `actor_id`, `space_id`, `facile_id` | Status is one of `to-do`, `in-progress`, `in-review`, `done` |
| `time_entries` | `id`, `project_id`, `task_id`, `user_id`, `started_at`, `stopped_at`, `paused_at`, `paused_duration_ms` | A running entry has a null `stopped_at`; `last_notification_at` throttles push reminders |
| `app_settings` | single row `id = 1` | Webhook URL and secret header/value, plus the Antenne URL, secret, enabled flag, and per-event toggles |
| `push_subscriptions` | `user_id` unique, `endpoint`, `p256dh`, `auth` | One Web Push subscription per user |
| `antenne_outbox` | `id`, `channel`, `payload` jsonb, `idempotency_key` unique, `sent_at` | antenne-client's outbox: events waiting for the bus, and sent ones for 7 days |
| `pool_processed_events` | `idempotency_key` PK, `processed_at` | Inbound de-duplication ledger |

The schema is the ordered SQL in `apps/api/migrations/`: `00001_baseline.sql` is the
production schema GORM `AutoMigrate` had built by v1.20.0, and every later file is an
ordinary migration. `apps/api/schema.sql` is the normalized dump of the whole set.
`schemas.Migrate` applies them on every boot, then `schemas.Reconcile` runs the data steps,
among them `usercolor.BackfillMissing`, which assigns a color to users created before
colors existed, and `backfillTimeEntryTasks`, which turns the legacy free-text
`description` column into real `tasks` rows so every entry points at a task.

## Background workers

- **Notification worker** (`internal/worker`) ticks every minute and calls
  `notifications.SendActiveTimerReminders`, pushing a reminder to users whose timer is
  still running. It is a no-op when the VAPID keys are unset.
- **Outbox drainer** (`modules/antenne`, antenne-client's `outbox.Drainer`) sends unsent
  `antenne_outbox` rows in id order every two seconds, backs off to five seconds while the
  bus is down, and prunes rows sent over 7 days ago. Before sending, it re-encodes each row
  through the struct it was marshalled from, because jsonb reorders keys: receivers get the
  exact bytes Sablier always sent (`modules/antenne/testdata/wire`). The same worker prunes
  `antenne_processed_events` older than 35 days once a day.

## Cross-app integration

Sablier is one of the apps genuinely wired into the Antenne event bus. It uses
`github.com/FacileStudio/antenne-client/go` v1 for all of it: the client (with persistent
connect, so a bus that is down at boot is reached when it returns), `envelope` for the event
contract, `catalog` for channel names and `outbox` for durability.

- **Emits** `project.created`, `project.updated`, `project.deleted`, `task.created`,
  `task.updated`, `task.deleted`, `time_entry.created`, `time_entry.updated`, `time_entry.deleted`.
- **Listens** for `project.*`, `task.*`, and `agent_session.created` /
  `agent_session.updated`.

Inbound project and task events upsert by `facile_id`, so a re-emitted history is
idempotent. Inbound `agent_session` events (from Mycelium) are materialized as time
entries: a project whose name matches the session's project receives the entry under its
`Agent sessions` task, and anything unmatched is parked in a shared `Agent work` project
so no session is dropped.

Connection settings live in `app_settings` and are editable through
`/api/antenne`. If no row exists, `ANTENNE_URL` (or its older synonym `ANTENNE_INSTANCE`) and
`ANTENNE_SECRET` are read from
the environment as a fallback. `antenne.yaml` at the repo root documents the same event
lists; it is a description, not the source of truth — `modules/antenne/service.go`
builds the config inline.

Logs ship to Journal when both `JOURNAL_URL` and `JOURNAL_TOKEN` are set; the Journal SDK
wraps the `slog` handler, so nothing else in the code changes.

Outbound webhooks are separate from the pool. Starting a timer fires a `timer_started`
event and stopping one fires `timer_stopped`; `internal/webhook.Fire` posts
`{"event": ..., "data": ...}` to the URL in `app_settings`, adding the configured secret
header when both the header name and value are set. The `data` object carries the entry
id, project id and name, task id and name, user id and email, `started_at`, and
`stopped_at`. It is fire-and-forget with a ten-second timeout and no retry.
