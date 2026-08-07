# Sablier — Development

Getting a local Sablier running, the tasks that exist, and the quality gate that runs
before every push.

## Prerequisites

| Tool | Version | Why |
|---|---|---|
| Go | 1.24 | `apps/api/go.mod` declares `go 1.24.0`; `mise.toml` pins `go = "1.24"` |
| Bun | any recent | Client package manager and dev server |
| Docker | any recent | Postgres 16 for local development |
| mise | optional | Runs the tasks below; each one is a one-line shell command you can run by hand |

## Setup

```sh
mise run hooks
mise run install
docker compose up db -d
```

`mise run hooks` points `core.hooksPath` at `.githooks`, which is what wires up the
pre-push quality gate. `mise run install` runs `bun install --frozen-lockfile` in
`apps/client`.

## Running

Two terminals. The API:

```sh
cd apps/api
cp .env.example .env
go run .
```

The client:

```sh
cd apps/client
bun run dev
```

The API listens on `:4000` (the `.env.example` pins `PORT=4000`, since `tronc` defaults to
`8080` and the Vite proxy targets `4000`). The client listens on `:5173` and proxies
`/api` and `/files` through to the API, so the browser only ever talks to one origin.

Migrations run automatically at startup via `schemas.Migrate`. There is no separate
migration tool and no migration files — GORM's `AutoMigrate` plus two backfill functions
do the work.

## Tasks

| Task | Command | What it does |
|---|---|---|
| `mise run install` | `bun install --frozen-lockfile` in `apps/client` | Client dependencies |
| `mise run check` | `sh ./scripts/check.sh` | The full quality gate: Go, then the client |
| `mise run check-go` | `sh ./scripts/check.sh --go-only` | Go half only |
| `mise run format` | `sh ./scripts/check.sh --format` | `go fmt ./...`, rewriting files in place |
| `mise run hooks` | `git config core.hooksPath .githooks` | Enables the tracked hooks in this clone |

Client scripts, run from `apps/client`: `bun run dev`, `bun run build`, `bun run preview`,
`bun run check` (`svelte-check` against `tsconfig.json`).

## The quality gate

`scripts/check.sh` is the gate, and `.githooks/pre-push` does nothing but exec it. It
reports and never rewrites, except under `--format`. Three steps per Go module, then the
client:

1. `gofmt -l .`, ignoring `vendor/`. Any listed file fails the gate.
2. `go vet ./...`
3. `go test ./...`
4. `bun run check` in `apps/client`, skipped with a warning if `bun` is not on `PATH`.

Two deliberate details worth knowing before you "fix" the script:

- **It is not invoked through mise.** `mise run` resolves every tool in the merged config
  before running any task body, so one broken tool in your global mise config would take
  the gate down with it. The hook calls `sh` directly.
- **It resolves the toolchain from `GOROOT`.** mise exports `GOROOT` for the pinned
  version but can leave an unrelated `go` earlier on `PATH`; mixing them produces
  `compile: version "X" does not match go tool version "Y"`. The script prefers
  `$GOROOT/bin/go` and `$GOROOT/bin/gofmt` when they exist.

Bypass once with `git push --no-verify`.

## Tests

```sh
cd apps/api
go test ./...
```

Tests live next to the code they cover: `schemas/migrate_test.go` (color backfill),
`modules/projects/service_test.go`, `modules/timeentries/service_test.go`,
`modules/users/service_test.go`, and `modules/antenne/agent_sessions_test.go`. They open
an in-memory SQLite database through `gorm.io/driver/sqlite`, so no Postgres is needed to
run them. There is no client-side test setup.

## Dependencies

Go dependencies are **vendored** in `apps/api/vendor`, and the Docker build runs
`go build -mod=vendor`. After changing anything in `go.mod`:

```sh
cd apps/api
go mod tidy
go mod vendor
```

Forgetting `go mod vendor` produces a build that works locally and fails in the image.

## Conventions

- Svelte 5 runes are enforced through `dynamicCompileOptions` in `svelte.config.js`. No
  Svelte 4 syntax.
- UI primitives come from shadcn-svelte in `src/lib/components/ui/`.
- Each API module keeps the same file layout: `router.go` wires paths, `controller.go`
  validates and shapes responses, `service.go` holds the logic and the GORM calls,
  `types.go` holds the request and response structs, `documentation.go` feeds `/docs`.
- No linter or formatter beyond `gofmt` is configured.
