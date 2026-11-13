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
| `STORAGE_DIR` | no | `./data` | Root of avatars uploaded before Registre owned pictures, served under `/files/`. Nothing writes there any more |
| `CLIENT_DIR` | no | `./client` | Directory holding the built SPA. The image sets `/client` explicitly |
| `SABLIER_ADMINS` | no | — | Comma-separated emails allowed to read and change the app-wide webhook and Antenne settings (`/api/settings`, `/api/antenne*`). Empty closes those routes to everyone |

Any Registre account can sign in to Sablier, so the app-wide settings answer 403 to every
account whose email is not on `SABLIER_ADMINS`. The email compared is the account's own,
which the sign-in rewrites from Registre's verified address; `PATCH /api/users/me` cannot
change it.

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
| `OIDC_MACHINE_AUDIENCE` | no | — | This app's own Registre client id. Lets porte verify a bearer JWT minted for Sablier, such as the 5 minute delegated token Bulle obtains for a user. Needs `OIDC_ISSUER` |
| `OIDC_SCOPE_PREFIX` | no | — | Set to `sablier` to gate those JWTs by method: `GET`, `HEAD` and `OPTIONS` need `sablier.read` or `sablier.write`, anything else needs `sablier.write`. Needs `OIDC_MACHINE_AUDIENCE`. Cookie and opaque sessions are never gated |
| `OIDC_MOBILE_REDIRECT` | no | `studio.facile.sablier://auth` | Where porte's `?flow=mobile` login hands the one-time code back to the Android app. Must match the scheme in `apps/client/capacitor.config.ts` |
| `MCP_RESOURCE` | no | compose: `https://sablier.facile.studio/mcp` | Canonical URI of the remote MCP server and the RFC 8707 resource Registre binds its tokens to. Mounts `POST /mcp` and its RFC 9728 metadata while OIDC is on; empty turns it off. Registre must declare the resource with the scopes `sablier.read` and `sablier.write` |

The callback route is registered inside the `/api` group, so the redirect URL is
`https://<host>/api/auth/oidc/callback`. Both `.env.example` files still show it without
the `/api` prefix — the router is the source of truth.

The Android app signs in through the system browser with PKCE and gets a bearer token
from `POST /api/auth/oidc/exchange`. While SSO is on, the API adds the Capacitor origin
`https://localhost` to the CORS list by itself, so a deploy that sets `CORS_ALLOWED_ORIGINS` explicitly cannot lock the app out. The Registre
client needs no new redirect URI: the browser still returns to `/api/auth/oidc/callback`.
The app keeps that token in the Android Keystore through `capacitor-secure-storage-plugin`;
the first launch of an updated app moves a token an older release left in plain
`@capacitor/preferences` across and deletes the old key. The desktop app keeps its token in
`~/.sablier.yml` (mode 0600), the file `sablier login` writes, so the window, the tray and
the CLI share one sign-in.

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
| `ANTENNE_URL` | no | — | Pool instance URL, used only as a fallback. `ANTENNE_INSTANCE` is read as a synonym; `ANTENNE_URL` wins |
| `ANTENNE_SECRET` | no | — | Pool shared secret, used only as a fallback |

These are read through antenne-client's `ConfigFromEnv` in `modules/antenne/service.go` and are used
**only when no `app_settings` row exists yet**. Once settings have been saved through
`PUT /api/antenne`, the database wins and the environment is ignored. Even in the
fallback path the connection is only attempted when both values are non-empty.

## AI attribution

| Variable | Required | Default | What it does |
|---|---|---|---|
| `BULLE_URL` | no | — | Bulle base URL. Sablier calls `POST {BULLE_URL}/api/ai/decide` |
| `BULLE_SERVICE_KEY` | no | — | Bearer key Bulle issues to Sablier. Each call also sends `X-Bulle-Acting-For: <user email>` |
| `JEV_API_KEY` | no | — | Jev key for a Sablier deployed without Bulle |
| `JEV_ENDPOINT` | no | `systemone` | `systemone` (`/v1/systemone`) or `decisions` (`/api/alpha/decisions`) |
| `JEV_BASE_URL` | no | `https://api.typesafe.ai` (`https://openrouter.ai` for `decisions`) | Jev host |
| `JEV_MODEL` | no | `jev-latest` | Jev model name |

The backend is chosen once at boot and logged as `attribution: AI backend chosen`:
Bulle when `BULLE_URL` is set, else direct Jev when `JEV_API_KEY` is set, else off. Bulle
wins when both are set, and a failed Bulle call never falls back to Jev. Each call has a
5 second timeout, and any error, an older Bulle's 404 or a 402 `payment_required` included, counts as no answer: the
work then waits in Review. The AI sees only the session title, the project hint, the
entity and the candidate labels (`Space › Project › Task`), never message content.

## Ideas

With `BULLE_URL` and `OIDC_ISSUER` set, the API serves `/api/ideas*`, a proxy to Bulle's Ideas
API. It exchanges the signed-in user's Registre refresh token for a Bulle token (audience
`bulle`, scopes `bulle.read` and `bulle.write`), so `PORTE_TOKEN_KEY` should be set to seal the
stored refresh tokens. `?space_id=` takes a Sablier space id the user belongs to and is sent to
Bulle as the linked Registre space; without it the personal space is used. A plan without
`ai.access` gets `ai_access: false` on the list and a 402 on generate, run and approve, and
Bulle is not called. Entitlements come from the Registre directory snapshot and are kept in the
`forfait_state` row, so a restart never fails open again.

## Compose-only

`POSTGRES_USER`, `POSTGRES_PASSWORD`, and `POSTGRES_DB` are consumed by the `db` service
in `docker-compose.yml`. The API never reads them — it only reads `DATABASE_URL`.

## Client

The client has no build-time configuration. `apps/client/src/lib/backend.ts` sets its
base URL to the empty string and calls `/api/...` relative to whatever origin served it.
In development, `apps/client/vite.config.ts` proxies `/api` and `/files` to
`http://localhost:4000`.
