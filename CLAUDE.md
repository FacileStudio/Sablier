# Sablier

Self-hosted time tracker for small teams. Go API + SvelteKit frontend + PostgreSQL.

## Tech Stack

| Layer    | Stack                                                        |
| -------- | ------------------------------------------------------------ |
| API      | Go 1.24, Chi router, GORM, PostgreSQL 16                    |
| Client   | SvelteKit 5 (Svelte 5 runes), Tailwind CSS 4, shadcn-svelte |
| Build    | `go build` (vendored deps), Bun                             |
| Deploy   | Docker Compose, one distroless container serving API + client |
| Auth     | JWT + optional OIDC SSO, `SSO_ONLY` mode                    |

## Project Structure

```
apps/
  api/                  Go backend
    main.go             Entrypoint: env, DB, migrations, router, graceful shutdown
    modules/            Domain modules (auth, projects, timeentries, users, settings, spaces)
    internal/           Shared infra (database, middleware, logger, env, errors, etc.)
    schemas/            GORM models and migrations (auto-run on startup)
    vendor/             Vendored Go dependencies
  client/               SvelteKit frontend
    src/
      routes/           SvelteKit file-based routing
        (app)/          Authenticated layout group (dashboard, projects, users, settings, profile, spaces)
        login/          Login page
      lib/
        backend.ts      API client (fetch wrapper)
        components/     App components + shadcn-svelte ui/ primitives
Dockerfile              Multi-stage: bun (client) + golang:1.24-alpine (api) -> distroless
docker-compose.yml      Full stack: db + api (serves the client too)
.env.example            Root-level env template (production)
```

## Commands

### API (`apps/api/`)

```sh
cp .env.example .env
go run .                    # Dev server on :4000
go test ./...               # Run tests
go build -o bin/api .       # Production binary
```

### Client (`apps/client/`)

```sh
bun install                 # Install dependencies
bun run dev                 # Dev server on :5173
bun run build               # Static build to build/
bun run preview             # Preview production build
bun run check               # Svelte type checking
```

### Full Stack (Docker)

```sh
cp .env.example .env
docker compose up --build           # Everything
docker compose up db -d             # Just PostgreSQL for local dev
```

## Environment Variables

Core variables (see `.env.example` and `apps/api/.env.example` for full list):

- `DATABASE_URL` -- **required**, PostgreSQL connection string. No default since the tronc/env adoption
- `CORS_ALLOWED_ORIGINS` -- Comma-separated allowed CORS origins. `DOMAINS` is still read as a fallback, so deployments need no rename
- `APP_ENV` -- `development`, `staging`, `production`. Never gates security behaviour
- `PORT` -- API port (tronc default `8080`; compose and both `.env.example` pin `4000`)
- `LOG_LEVEL` -- `debug`, `info`, `warn`, `error`
- `STORAGE_DIR` -- Avatar file storage (default `./data`)
- `OIDC_*` -- OpenID Connect config (optional)
- `SSO_ONLY` -- Hide password auth when `true`
- `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, `VAPID_SUBJECT` -- Web push (optional; push stays off while the keys are unset)
- `JOURNAL_URL`, `JOURNAL_TOKEN` -- Log shipping to Journal (optional; both must be set)

## Key Endpoints

- `GET /health`, `GET /ready` -- Health and readiness probes
- `GET /docs` -- Auto-generated JSON API documentation
- `GET /files/*` -- Static file serving (avatars)
- Auth: `/api/auth/register`, `/api/auth/login`, `/api/auth/config`, `/api/auth/oidc/*`
- Resources: `/api/projects`, `/api/time-entries`, `/api/users`, `/api/settings`
- Spaces: `/api/spaces`, `/api/spaces/{id}`, `/api/spaces/{id}/members`, `/api/spaces/{id}/leave`

## Conventions

- **Go modules are vendored** -- run `go mod vendor` after changing dependencies.
- **Migrations run on startup** via `schemas.Migrate(db)`. No separate migration tool.
- **Client uses static adapter** -- output is plain HTML/CSS/JS served by the Go binary through tronc's `spa` package, mounted as the catch-all after `/api`.
- **shadcn-svelte** provides the UI component primitives in `src/lib/components/ui/`. The Nova style is configured.
- **Svelte 5 runes** are enforced (`$state`, `$props`, `$derived`, `$effect`). No legacy Svelte 4 syntax.
- **No linter or formatter configured** in the repo currently.
- **No test setup on the client side** -- only the API has `go test`.
- Avatar uploads are stored on disk at `STORAGE_DIR/avatars/` and served under `/files/`.
