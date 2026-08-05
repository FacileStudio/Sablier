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
                        Nook Pool (WebSocket) ◀── outbox worker ──▶ pool_outbox
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
| Schemas | `apps/api/schemas` | GORM models plus `Migrate`, which runs `AutoMigrate` and the backfills |
| Internal | `apps/api/internal/*` | `database`, `middleware`, `env`, `authcontext`, `authcrypto`, `usercolor`, `oidcavatar`, `webhook`, `worker`, `documentation` |
| Client | `apps/client` | SvelteKit 5 with `adapter-static` and `fallback: index.html` |

The eight modules are `auth`, `projects`, `timeentries`, `users`, `settings`, `spaces`,
`nookpool`, and `notifications`. All of them are registered inside a single
`router.Route("/api", ...)` block, so every application route lives under `/api`.

## Request lifecycle

1. Traefik terminates TLS and forwards to the container on port `4000`.
2. `tronc/httpx` middleware attaches the request logger, recovers panics, and applies
   CORS using `CORS_ALLOWED_ORIGINS`.
3. `health.Mount` answers `/health`, `/ready`, `/api/health`, and `/api/ready` before
   anything else. `/ready` pings Postgres with a two-second timeout and returns 503 as
   `{"status":"not_ready"}` when the ping fails.
4. `/api/*` dispatches into a module router. Protected routes run
   `internal/middleware.RequireAuth`, which resolves the bearer token and puts an
   `authcontext.Identity{UserID, Email}` on the request context.
5. The module controller validates and delegates to its service, which talks to GORM.
6. Responses are written by `tronc/httpjson`. Errors constructed with `tronc/errors`
   (`Invalid`, `Unauthorized`, `Conflict`, `Internal`) map to their status codes.
7. Anything that is not `/api`, `/files`, `/docs`, `/openapi`, or a health route falls
   through to the SPA handler, which serves `index.html` for extensionless paths and
   404s for missing assets.

## Authentication

Two token families, both presented as `Authorization: Bearer <token>` and both stored
hashed:

- **Session tokens** — created by `POST /api/auth/register`, `POST /api/auth/login`, and
  the OIDC callback. Rows in `sessions`, valid for 30 days.
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
- `GET /api/auth/oidc/callback` verifies state and the ID token, upserts the user by
  email, creates a session, and redirects to `OIDC_SUCCESS_URL` — which defaults to the
  first entry of `CORS_ALLOWED_ORIGINS`.
- `POST /api/auth/sync-profile` re-pulls the provider profile for the logged-in user.

The scopes requested are `openid`, `email`, `profile`, and `offline_access`. Access,
refresh, and expiry are stored on the user row, along with `profile_synced_at`. A
provider `picture` claim is downloaded into `STORAGE_DIR/avatars` and exposed as
`/files/...`, with `avatar_source` set to `oidc`.

`SSO_ONLY=true` removes `POST /api/auth/register` and `POST /api/auth/login` from the
router entirely — they 404 rather than 403.

## Data model

| Table | Key columns | Notes |
|---|---|---|
| `users` | `id`, `email` unique, `name`, `color`, `password_hash`, `rate`, `rate_type`, `workday_hours` | Also holds `avatar_url`, `avatar_source`, `oidc_picture_url`, the OIDC tokens, and `profile_synced_at` |
| `sessions` | `token` PK (hashed), `user_id`, `expires_at` | 30-day browser sessions |
| `api_tokens` | `token` PK (hashed), `user_id`, `name` | One long-lived token per user |
| `spaces` | `id` UUID PK, `name`, `description` | UUID generated in `BeforeCreate` |
| `space_members` | `id` UUID PK, unique `(space_id, user_id)`, `role` | Roles are `owner`, `admin`, `member` |
| `projects` | `id`, `name`, `owner_id`, `space_id`, `facile_id` unique | `facile_id` is the cross-app identity used by pool sync |
| `tasks` | `id`, unique `(project_id, name)`, `status`, `actor_id`, `space_id`, `facile_id` | Status is one of `to-do`, `in-progress`, `in-review`, `done` |
| `time_entries` | `id`, `project_id`, `task_id`, `user_id`, `started_at`, `stopped_at`, `paused_at`, `paused_duration_ms` | A running entry has a null `stopped_at`; `last_notification_at` throttles push reminders |
| `app_settings` | single row `id = 1` | Webhook URL and secret header/value, plus the Nook Pool URL, secret, enabled flag, and per-event toggles |
| `push_subscriptions` | `user_id` unique, `endpoint`, `p256dh`, `auth` | One Web Push subscription per user |
| `pool_outbox` | `id`, `channel`, `payload`, `attempts`, `last_error` | Outbound events waiting for the pool |
| `pool_processed_events` | `idempotency_key` PK, `processed_at` | Inbound de-duplication ledger |

`schemas.Migrate` runs `AutoMigrate` over all of them on every boot, then two backfills:
`usercolor.BackfillMissing` assigns a color to users created before colors existed, and
`backfillTimeEntryTasks` turns the legacy free-text `description` column into real `tasks`
rows so every entry points at a task.

## Background workers

- **Notification worker** (`internal/worker`) ticks every minute and calls
  `notifications.SendActiveTimerReminders`, pushing a reminder to users whose timer is
  still running. It is a no-op when the VAPID keys are unset.
- **Pool outbox worker** (`modules/nookpool`) ticks every two seconds, drains
  `pool_outbox` onto the WebSocket connection, and prunes `pool_processed_events` older
  than 35 days once a day.

## Cross-app integration

Sablier is one of the apps genuinely wired into the Nook event bus. It uses
`github.com/FacileStudio/pool/go` as the client and `github.com/FacileStudio/enveloppe/go`
as the event contract.

- **Emits** `project.created`, `project.updated`, `project.deleted`, `task.created`,
  `task.updated`, `task.deleted`, `time_entry.created`, `time_entry.updated`.
- **Listens** for `project.*`, `task.*`, and `agent_session.created` /
  `agent_session.updated`.

Inbound project and task events upsert by `facile_id`, so a re-emitted history is
idempotent. Inbound `agent_session` events (from Jardin) are materialized as time
entries: a project whose name matches the session's project receives the entry under its
`Agent sessions` task, and anything unmatched is parked in a shared `Agent work` project
so no session is dropped.

Connection settings live in `app_settings` and are editable through
`/api/nook-pool`. If no row exists, `NOOK_POOL_URL` and `NOOK_POOL_SECRET` are read from
the environment as a fallback. `nook.yaml` at the repo root documents the same event
lists; it is a description, not the source of truth — `modules/nookpool/service.go`
builds the config inline.

Logs ship to Journal when both `JOURNAL_URL` and `JOURNAL_TOKEN` are set; the Journal SDK
wraps the `slog` handler, so nothing else in the code changes.

Outbound webhooks are separate from the pool. Starting a timer fires a `timer_started`
event and stopping one fires `timer_stopped`; `internal/webhook.Fire` posts
`{"event": ..., "data": ...}` to the URL in `app_settings`, adding the configured secret
header when both the header name and value are set. The `data` object carries the entry
id, project id and name, task id and name, user id and email, `started_at`, and
`stopped_at`. It is fire-and-forget with a ten-second timeout and no retry.
