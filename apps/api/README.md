# API

Go backend for authentication, projects, tasks, sessions, profiles, and webhook settings.

## Responsibilities

- JWT authentication with optional OIDC login
- Shared project and task management
- Timer start/stop plus manual time entries
- User profiles with the Registre profile picture
- Per-user webhook configuration for timer events
- Health, readiness, and JSON docs endpoints

## Run locally

```sh
cp .env.example .env
go run .
```

Default address: `http://localhost:4000`

The API expects PostgreSQL at `DATABASE_URL` and runs migrations on startup.

## Useful endpoints

- `GET /health`
- `GET /ready`
- `GET /docs`
- `GET /auth/config`
- `GET /users/me`
- `GET /projects`
- `GET /projects/{id}/tasks`
- `GET /time-entries`
- `POST /time-entries/start`
- `POST /time-entries/stop`
- `GET /settings/`

## Environment

- `DATABASE_URL`: PostgreSQL connection string
- `DOMAINS`: comma-separated allowed origins
- `PORT`: HTTP port, default `4000`
- `LOG_LEVEL`: `debug`, `info`, `warn`, or `error`
- `STORAGE_DIR`: root of legacy uploaded avatars, default `./data`
- `OIDC_*`: enables OpenID Connect login
- `GET /version` serves the newest released heading of the root `CHANGELOG.md`, read by `scripts/version.sh` in the Docker build

See [`./.env.example`](./.env.example) for the local template.

## Build and test

```sh
go test ./...
go build -o bin/api .
```

## Notes

- Legacy uploaded avatar files are served by the API under `/files/*`.
- The production image stores uploaded files in `/data`.
