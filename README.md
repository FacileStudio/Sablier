# Sablier

Sablier is a self-hosted time tracker for small teams. It pairs a Go API with a SvelteKit frontend and keeps the setup boring on purpose.

## What it does

- Email/password auth with optional OIDC SSO and `SSO_ONLY` mode
- Shared projects and tasks for the whole workspace
- Live timers, manual sessions, and per-user session editing
- Dashboard, user directory, and user activity views
- Profile names, colors, and avatar uploads
- Outbound timer webhooks on start and stop

## Stack

- `apps/api`: Go, Chi, GORM, PostgreSQL
- `apps/client`: SvelteKit 5, Tailwind CSS 4, Bun
- `docker-compose.yml`: PostgreSQL plus production-style API and client services

## Quick start

### Docker

1. Copy the root env file and adjust values if needed:

```sh
cp .env.example .env
```

2. Start the full stack:

```sh
docker compose up --build
```

3. Open the app:

- Client: `http://localhost`
- API: `http://localhost:4000`
- API docs payload: `http://localhost:4000/docs`

### Local development

1. Start PostgreSQL:

```sh
docker compose up db -d
```

2. Start the API:

```sh
cd apps/api
cp .env.example .env
go run .
```

3. Start the client in another terminal:

```sh
cd apps/client
bun install
bun run dev
```

The client defaults to `http://localhost:5173` and talks to `http://localhost:4000`.

## Configuration

Main environment variables:

- `DATABASE_URL`: PostgreSQL connection string
- `DOMAINS`: allowed frontend origins for CORS
- `PORT`: API port, default `4000`
- `LOG_LEVEL`: `debug`, `info`, `warn`, or `error`
- `STORAGE_DIR`: local file storage for avatars, default `./data`
- `OIDC_ISSUER`, `OIDC_CLIENT_ID`, `OIDC_CLIENT_SECRET`, `OIDC_REDIRECT_URL`: enable OIDC login
- `OIDC_SUCCESS_URL`: post-login redirect, defaults to the first `DOMAINS` entry
- `SSO_ONLY=true`: hide password login and registration
- `VITE_API_BASE_URL`: client-side API base URL for production builds

See [`.env.example`](.env.example) and [`apps/api/.env.example`](apps/api/.env.example) for examples.

## Architecture

Source: [`docs/architecture.d2`](docs/architecture.d2)

![Sablier architecture](docs/architecture.svg)

## Repo map

- [`apps/api/README.md`](apps/api/README.md): backend setup and API overview
- [`apps/client/README.md`](apps/client/README.md): frontend setup and build notes
