# Sablier

Self-hosted time tracker for small teams. Go API, SvelteKit frontend, boring on purpose.

One Go binary serves both the JSON API and the built client, so a deployment is a single
container behind a single Traefik router.

Live at [sablier.facile.studio](https://sablier.facile.studio).

## What it does

- Email/password auth with optional OIDC SSO and an `SSO_ONLY` mode
- Spaces that scope projects, tasks, and time entries to a group of members
- Live timers with pause and resume, plus manually entered sessions
- Per-project task boards with `to-do`, `in-progress`, `in-review`, and `done` statuses
- Profiles with names, colors, avatar uploads, billing rate, and workday length
- Long-lived API tokens for scripting, alongside browser session tokens
- Web push reminders for timers that are still running
- Two-way project and task sync over the Antenne, plus outbound webhooks

## Stack

| Layer | Tech |
|---|---|
| API | Go 1.24, Chi v5, GORM, PostgreSQL 16, [tronc](https://github.com/FacileStudio/tronc) v0.6.0 |
| Client | SvelteKit 5 (runes), Tailwind CSS 4, shadcn-svelte, bits-ui |
| Deploy | Docker Compose, one distroless container behind Traefik |

## Quick start

`docker-compose.yml` is deployment-shaped: it publishes no host port and expects the
external `dokploy-network` with Traefik in front of it. Use it to deploy, not to browse
locally.

```sh
cp .env.example .env
docker compose up -d --build
```

### Local development

Start Postgres, then the API and the client in separate terminals.

```sh
mise run install
docker compose up db -d
```

```sh
cd apps/api
cp .env.example .env
go run .
```

```sh
cd apps/client
bun run dev
```

The client runs on <http://localhost:5173> and proxies `/api` and `/files` to the API on
`:4000`. Migrations run on API startup, so there is no separate migration step.

## Configuration

| Variable | What it does |
|---|---|
| `DATABASE_URL` | Postgres connection string. Required — the API exits 1 without it |
| `CORS_ALLOWED_ORIGINS` | Comma-separated browser origins allowed to call the API |
| `PORT` | HTTP listen port. `tronc` defaults to `8080`; this repo pins `4000` everywhere |
| `STORAGE_DIR` | Root for uploaded avatars, served back under `/files/` |
| `APP_ENV` | `development`, `staging`, or `production` |
| `LOG_LEVEL` | `debug`, `info`, `warn`, or `error` |

Full reference: [docs/configuration.md](docs/configuration.md).

## Structure

```
apps/
  api/       Go backend — modules/ (auth, projects, timeentries, users, settings,
             spaces, antenne, notifications), schemas/ (GORM models + migrations)
  client/    SvelteKit 5 SPA, built into the API image and served by it
scripts/     check.sh, the quality gate the pre-push hook runs
docs/        Architecture, configuration, development, deployment, API
```

## Documentation

| Doc | What's in it |
|---|---|
| [Architecture](docs/architecture.md) | Request flow, data model, how the pieces fit |
| [Configuration](docs/configuration.md) | Every environment variable and default |
| [Development](docs/development.md) | Local setup, tests, the quality gate |
| [Deployment](docs/deployment.md) | Docker Compose, Dokploy, Traefik routing |
| [API](docs/api.md) | HTTP endpoints and payloads |

---

Part of the [Facile Suite](https://facile.studio) — self-hosted tools for creative studios
and freelancers. One login, zero cloud dependency.
