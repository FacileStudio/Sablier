# Sablier — Configuration

Every environment variable the API actually reads, taken from `apps/api/internal/env`,
`tronc/env`, `tronc/spa`, and `modules/antenne/service.go`.

Configuration is read once at startup. `godotenv` loads a `.env` file from the working
directory first if one exists, so `apps/api/.env` is picked up by `go run .` but plays no
role in the container, where everything comes from the process environment.

## Core

These come from `tronc/env`'s `LoadCore`, shared by every Go app in the suite.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `DATABASE_URL` | yes | — | Postgres connection string |
| `PORT` | no | `8080` | HTTP listen port. Must be a valid TCP port or startup fails |
| `APP_ENV` | no | `development` | `development`, `staging`, or `production`. Never gates security behavior |
| `LOG_LEVEL` | no | `info` | `debug`, `info`, `warn`, or `error`. Any other value fails startup |
| `CORS_ALLOWED_ORIGINS` | no | — | Comma-separated browser origins allowed to call the API |
| `JOURNAL_URL` | no | — | Journal ingest URL for log shipping |
| `JOURNAL_TOKEN` | no | — | Journal per-app key. Logs ship only when both this and `JOURNAL_URL` are set |

**The trap:** `DATABASE_URL` is required, and a missing or blank value makes `run()`
return an error and the process exit 1. There is no fallback DSN.

**`PORT` is 8080 by default but Sablier is pinned to 4000** in `docker-compose.yml`, both
`.env.example` files, the Dockerfile's `EXPOSE`, and the Vite dev proxy. Changing it means
changing the Traefik service port too.

### CORS name fallbacks

`tronc/env` reads the first of these that is set, in order, and splits it on commas:
`CORS_ALLOWED_ORIGINS`, `ALLOWED_ORIGINS`, `DOMAINS`, `DOMAIN`, `CORS_ORIGINS`,
`TRUSTED_ORIGINS`, `CLIENT_ORIGIN`. Sablier's compose file still passes `DOMAINS`, so
deployments that predate the rename keep working. An unset list means no cross-origin
caller is allowed — which is fine in production, where the client is served from the same
origin, and fatal in dev, where Vite runs on `:5173`.

## Sablier-specific

| Variable | Required | Default | What it does |
|---|---|---|---|
| `STORAGE_DIR` | no | `./data` | Root for uploaded files. `STORAGE_DIR/avatars` is created at startup and served under `/files/` |
| `SSO_ONLY` | no | `false` | When true, the register and login routes are not registered at all |
| `CLIENT_DIR` | no | `./client` | Directory holding the built SPA. The image sets `/client` explicitly |

`CLIENT_DIR` is read by `tronc/spa`. If the directory has no `index.html`, the SPA
catch-all is simply not mounted and the binary serves the API alone.

## OIDC

Optional and additive. Setting `OIDC_ISSUER` turns it on; the other three then become
required, and startup fails with a clear error if any of them is blank.

| Variable | Required | Default | What it does |
|---|---|---|---|
| `OIDC_ISSUER` | no | — | Discovery URL, for example `https://porte.facile.studio/application/o/sablier/` |
| `OIDC_CLIENT_ID` | with issuer | — | Client id from the provider |
| `OIDC_CLIENT_SECRET` | with issuer | — | Client secret from the provider |
| `OIDC_REDIRECT_URL` | with issuer | — | Must point at `/api/auth/oidc/callback` on the public hostname |
| `OIDC_SUCCESS_URL` | no | first `CORS_ALLOWED_ORIGINS` entry | Where the callback redirects after a successful login |

The callback route is registered inside the `/api` group, so the redirect URL is
`https://<host>/api/auth/oidc/callback`. Both `.env.example` files still show it without
the `/api` prefix — the router is the source of truth.

## Web push

| Variable | Required | Default | What it does |
|---|---|---|---|
| `VAPID_PUBLIC_KEY` | no | — | VAPID public key, also served by `GET /api/notifications/vapid-public-key` |
| `VAPID_PRIVATE_KEY` | no | — | VAPID private key |
| `VAPID_SUBJECT` | no | `mailto:admin@example.com` | `mailto:` or URL contact for the push service |

Push stays off while the keys are unset. The reminder worker still ticks, it just has
nothing to send.

## Antenne

| Variable | Required | Default | What it does |
|---|---|---|---|
| `ANTENNE_URL` | no | — | Pool instance URL, used only as a fallback |
| `ANTENNE_SECRET` | no | — | Pool shared secret, used only as a fallback |

These are read directly with `os.Getenv` in `modules/antenne/service.go` and are used
**only when no `app_settings` row exists yet**. Once settings have been saved through
`PUT /api/antenne`, the database wins and the environment is ignored. Even in the
fallback path the connection is only attempted when both values are non-empty.

## Compose-only

`POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` are consumed by the `db` service
in `docker-compose.yml`. The API never reads them — it only reads `DATABASE_URL`.

## Client

The client has no build-time configuration. `apps/client/src/lib/backend.ts` sets its
base URL to the empty string and calls `/api/...` relative to whatever origin served it.
In development, `apps/client/vite.config.ts` proxies `/api` and `/files` to
`http://localhost:4000`.
