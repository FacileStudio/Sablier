# Sablier

Self-hosted time tracker for small teams. Go API + SvelteKit frontend + PostgreSQL.

## Tech Stack

| Layer    | Stack                                                        |
| -------- | ------------------------------------------------------------ |
| API      | Go 1.24, Chi router, GORM, PostgreSQL 16                    |
| Client   | SvelteKit 5 (Svelte 5 runes), Tailwind CSS 4, shadcn-svelte |
| Build    | `go build` (vendored deps), Bun                             |
| Deploy   | Docker Compose (API via distroless, client via Nginx)        |
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
    Dockerfile          Multi-stage: golang:1.24-alpine -> distroless
  client/               SvelteKit frontend
    src/
      routes/           SvelteKit file-based routing
        (app)/          Authenticated layout group (dashboard, projects, users, settings, profile, spaces)
        login/          Login page
      lib/
        backend.ts      API client (fetch wrapper)
        components/     App components + shadcn-svelte ui/ primitives
    Dockerfile          Multi-stage: oven/bun -> nginx:alpine (static adapter)
docker-compose.yml      Full stack: db + api + client
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

- `DATABASE_URL` -- PostgreSQL connection string (default: local postgres)
- `DOMAINS` -- Comma-separated allowed CORS origins
- `PORT` -- API port (default `4000`)
- `LOG_LEVEL` -- `debug`, `info`, `warn`, `error`
- `STORAGE_DIR` -- Avatar file storage (default `./data`)
- `OIDC_*` -- OpenID Connect config (optional)
- `SSO_ONLY` -- Hide password auth when `true`
- `VITE_API_BASE_URL` -- Client-side API URL (build-time, default `http://localhost:4000`)

## Key Endpoints

- `GET /health`, `GET /ready` -- Health and readiness probes
- `GET /docs` -- Auto-generated JSON API documentation
- `GET /files/*` -- Static file serving (avatars)
- Auth: `/auth/register`, `/auth/login`, `/auth/config`, `/auth/oidc/*`
- Resources: `/projects`, `/time-entries`, `/users`, `/settings`
- Spaces: `/spaces`, `/spaces/{id}`, `/spaces/{id}/members`, `/spaces/{id}/leave`

## Conventions

- **Go modules are vendored** -- run `go mod vendor` after changing dependencies.
- **Migrations run on startup** via `schemas.Migrate(db)`. No separate migration tool.
- **Client uses static adapter** -- output is plain HTML/CSS/JS served by Nginx in production.
- **shadcn-svelte** provides the UI component primitives in `src/lib/components/ui/`. The Nova style is configured.
- **Svelte 5 runes** are enforced (`$state`, `$props`, `$derived`, `$effect`). No legacy Svelte 4 syntax.
- **No linter or formatter configured** in the repo currently.
- **No test setup on the client side** -- only the API has `go test`.
- Avatar uploads are stored on disk at `STORAGE_DIR/avatars/` and served under `/files/`.
