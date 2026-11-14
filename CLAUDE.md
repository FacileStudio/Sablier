# Sablier

Self-hosted time tracker for small teams. Go API + SvelteKit frontend + PostgreSQL.

## Tech Stack

| Layer    | Stack                                                        |
| -------- | ------------------------------------------------------------ |
| API      | Go 1.26, Chi router, GORM, PostgreSQL 16                    |
| Client   | SvelteKit 5 (Svelte 5 runes), Tailwind CSS 4, shadcn-svelte |
| Build    | `go build` (vendored deps), Bun                             |
| Deploy   | Docker Compose, one distroless container serving API + client |
| Auth     | Sign in with Facile (Registre OIDC) only, no passwords      |

## Project Structure

```
apps/
  api/                  Go backend
    main.go             Entrypoint: env, DB, migrations, router, graceful shutdown
    modules/            Domain modules (auth, projects, timeentries, users, settings, spaces)
    internal/           Shared infra (database, middleware, logger, env, errors, etc.)
    migrations/         Ordered goose SQL, embedded and applied at boot by tronc/migrate
    schemas/            GORM models (the query layer) and the boot data steps
    vendor/             Vendored Go dependencies
  client/               SvelteKit frontend
    src/
      routes/           SvelteKit file-based routing
        (app)/          Authenticated layout group (dashboard, projects, users, settings, profile, spaces)
        login/          Login page
      lib/
        backend.ts      API client (fetch wrapper)
        components/     App components + shadcn-svelte ui/ primitives
Dockerfile              Multi-stage: bun (client) + golang:1.26-alpine (api) -> distroless
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
- `STORAGE_DIR` -- Legacy uploaded avatars, read-only since Registre owns pictures (default `./data`)
- `OIDC_*` -- OpenID Connect config (optional)
- `SABLIER_ADMINS` -- emails allowed to read and change the app-wide webhook and Antenne settings. Empty closes
  `/api/settings` and `/api/antenne*` to everyone (403). Sablier has no admin role of its own: any Registre
  account can sign in, so never put app-wide data behind plain `RequireAuth`; use `middleware.RequireAdmin`
- `MCP_RESOURCE` -- Canonical URI of the remote MCP server (`https://sablier.facile.studio/mcp`); mounts `POST /mcp` while OIDC is on
- `VAPID_PUBLIC_KEY`, `VAPID_PRIVATE_KEY`, `VAPID_SUBJECT` -- Web push (optional; push stays off while the keys are unset)
- `JOURNAL_URL`, `JOURNAL_TOKEN` -- Log shipping to Journal (optional; both must be set)
- `JOURNAL_BROWSER_URL`, `JOURNAL_BROWSER_KEY` -- Browser error reporting. The URL is **not**
  `JOURNAL_URL`: that one is the server SDK's and points inside the Docker network, which no
  browser can resolve. It must end in `/api`, and the app refuses to boot otherwise
- `BULLE_URL`, `BULLE_SERVICE_KEY` -- AI step of the attribution resolver (`internal/attribution`,
  `internal/decide`, through `tronc/bulle` `DecideAI`), only for users with `ai.access`. Either unset = off
- `SOURCEMAP_DIR` -- where the built client's maps live, **outside `CLIENT_DIR`** so nothing
  serves them. The API uploads them to Journal at boot, keyed on the release in
  `_app/version.json` — the exact string the browser reports. Set by the Dockerfile, which moves
  the maps out of `build/` after `vite build`; unset locally, so no upload happens on a laptop.
  The client emits them with `sourcemap: 'hidden'`, so no `sourceMappingURL` comment ships and no
  browser fetches one

## Key Endpoints

- `GET /health`, `GET /ready` -- Health and readiness probes
- `GET /docs` -- Auto-generated JSON API documentation
- `GET /files/*` -- Static file serving (legacy uploaded avatars)
- Auth: `/api/auth/config`, `/api/auth/oidc/*` (no register or login route)
- `GET /version` -- the newest released heading of `CHANGELOG.md`, baked in at build time by `scripts/version.sh` (never an env); `POST /api/registre/sync` and `POST /api/registre/deletions` are Registre's pushes
- Resources: `/api/projects`, `/api/time-entries`, `/api/users`, `/api/settings`
- Spaces (read-only, Registre owns them): `/api/spaces`, `/api/spaces/{id}`, `/api/spaces/{id}/members`
- `POST /api/registre/sync` -- Registre's anonymous nudge, answers 202; mounted only with OIDC
- `GET /api/tasks/search`, `GET /api/review`, `POST /api/review/{entry_id}` -- task picker and the Review inbox

## Conventions

- **Go modules are vendored** -- run `go mod vendor` after changing dependencies.
- **Migrations are ordered SQL** in `apps/api/migrations/`, applied at boot by `tronc/migrate` and embedded in the
  binary. GORM is the query layer only: `AutoMigrate` never runs, and `scripts/check.sh` fails on an
  `.AutoMigrate(` outside a test file. To change the schema, add the next numbered file with `-- +goose Up` and
  `-- +goose Down`, update the struct in `schemas/` by hand (`schemas/columns_test.go` fails otherwise), and
  regenerate `apps/api/schema.sql` (`docs/development.md`). `schemas.Reconcile` runs the idempotent data steps
  after the migrations at every boot. `/api migrate status` (also `up`, `version`, `baseline <n>`) answers
  without starting a server.
- **Sign-in is porte/kit** (`modules/auth/store.go`, `kit.NewStore`). `store_test.go` pins `Unwritten()` and
  checks a kit account has the same NULL columns as a GORM one: a new nullable `users` column must be filled in
  `fillNewAccount` or accepted there. The agent-user refusal lives app-side in `Accounts.UpsertFromOIDC`.
- **The bus is antenne-client v1** (`modules/antenne`). `antenne_outbox` is the library's table, created by
  the baseline migration and pinned to the library's DDL by `migrations_test.go`; there is no GORM struct for it. The drainer re-encodes jsonb rows
  through their payload struct so the wire bytes match `modules/antenne/testdata/wire`; a new published payload
  type needs a case in `canonical` and a fixture.
- **Database tests** use `internal/testdb.Migrated(t)`: one PostgreSQL schema per package, built by the
  migrations, emptied before each test, skipped without `SABLIER_TEST_DATABASE_URL`. Never `t.Parallel()` in those
  packages. There is no SQLite: `scripts/check.sh` fails on a non-PostgreSQL driver in `go.mod`.
- **Client uses static adapter** -- output is plain HTML/CSS/JS served by the Go binary through tronc's `spa` package, mounted as the catch-all after `/api`.
- **shadcn-svelte** provides the UI component primitives in `src/lib/components/ui/`. The Nova style is configured.
- **Svelte 5 runes** are enforced (`$state`, `$props`, `$derived`, `$effect`). No legacy Svelte 4 syntax.
- **No linter or formatter configured** in the repo currently.
- **No test setup on the client side** -- only the API has `go test`.
- Profile pictures come from Registre: `oidc_picture_url` holds `<issuer>/avatars/<sub>`, backfilled at boot from `porte_identities`. Sablier has no upload; the profile page links to `https://sso.facile.studio/account`. Old uploads in `STORAGE_DIR/avatars/` still render for accounts with no Registre identity.
- Shared spaces are Registre's (`modules/spaces/registre_*.go`, `porte/directory`). Boot imports unlinked spaces once, retrying, and only then starts the snapshot sync. Apply never touches members without an identity under `OIDC_ISSUER`, unlinked spaces, or linked spaces missing from the snapshot. No write route for spaces exists here; the UI links to `https://sso.facile.studio/spaces`.
- `/api/users` and `/api/users/{id}` show only the caller, its agents and its space peers (`visibleUsers` in
  `modules/users/service.go`). Secrets are write-only: a GET reports `*_secret_set`, a PUT without the field keeps it.
- Space visibility goes through `schemas.VisibleSpaces` / `schemas.ScopeSpace` with `schemas.Projects` or `schemas.Entries`. Personal rows (no space) are private to their owner: projects by `owner_id`, entries by `user_id` or the owner's agents (`space_id` absent = personal, `all`, or a member space; a non-member space is 404). They compare with `CAST(space_id AS TEXT)` because `space_members.space_id` is text. Query tests use `internal/testdb`, which runs on PostgreSQL in its own schema when `SABLIER_TEST_DATABASE_URL` is set.
