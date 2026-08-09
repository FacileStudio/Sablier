# Sablier — Deployment

How the image is built, what Compose declares, and how Sablier is routed on la ruche.

## The image

`Dockerfile` is a four-stage build producing one small binary plus the built client:

1. `oven/bun:1` installs `apps/client` dependencies with `--frozen-lockfile` and runs
   `bun run build`, producing static files in `/client/build`.
2. `golang:1.26-alpine` builds the API with
   `CGO_ENABLED=0 go build -mod=vendor -trimpath -ldflags="-s -w"`. Dependencies come from
   the committed `apps/api/vendor` directory, so the build never reaches the network.
3. A throwaway stage creates `/data/avatars`, since a distroless image has no shell to
   `mkdir` with at runtime.
4. `gcr.io/distroless/static-debian12` receives the binary at `/api`, the client at
   `/client`, and the prepared `/data`.

The runtime stage sets `ENV CLIENT_DIR=/client` explicitly. A distroless base can carry
its own `WorkingDir`, which would make the relative default `./client` resolve somewhere
the SPA is not.

`EXPOSE 4000`, `VOLUME ["/data"]`, `ENTRYPOINT ["/api"]`.

## Compose topology

Two services, and no published host ports — this file is written for Dokploy and Traefik,
not for browsing on your laptop.

| Service | Image | Notes |
|---|---|---|
| `db` | `postgres:16-alpine` | `expose: 5432`, `pg_isready` healthcheck every 5s, volume `sablier_db_data` |
| `api` | built from `Dockerfile` | `expose: 4000`, joins `default` and the external `dokploy-network`, volume `sablier_api_data` at `/data` |

`api` waits on `db` being healthy through `depends_on: condition: service_healthy`.

### Healthcheck

```yaml
test: ["CMD", "/api", "healthcheck"]
```

The image is distroless: no shell, no `curl`, no `wget`. `tronc/healthcheck` handles this
by making the binary probe itself — `main` checks `os.Args` first and, when the argument
is `healthcheck`, requests `http://127.0.0.1:$PORT/health` with a three-second timeout and
exits 0 or 1. It targets `127.0.0.1` rather than `localhost` on purpose: in these
containers `localhost` resolves to `::1` first while the server binds `0.0.0.0`, so a
`localhost` probe fails against a perfectly healthy process.

### Traefik labels

One hostname, one service, two routers — plain HTTP redirecting to HTTPS:

```yaml
traefik.enable: "true"
traefik.docker.network: dokploy-network
traefik.http.routers.sablier-web.rule: "Host(`sablier.facile.studio`)"
traefik.http.routers.sablier-web.entrypoints: web
traefik.http.routers.sablier-web.middlewares: redirect-to-https@file
traefik.http.routers.sablier-web.service: sablier-svc
traefik.http.routers.sablier-secure.rule: "Host(`sablier.facile.studio`)"
traefik.http.routers.sablier-secure.entrypoints: websecure
traefik.http.routers.sablier-secure.tls.certresolver: letsencrypt
traefik.http.routers.sablier-secure.service: sablier-svc
traefik.http.services.sablier-svc.loadbalancer.server.port: "4000"
```

This is the suite's one-container / one-router / one-hostname rule. There is no separate
frontend service and no `/api` path-prefix router, because the same binary answers both
the API and the SPA. Do not add a second router for `/api` — you would be splitting
traffic that is already unified.

The load balancer port must match `PORT`. Both are `4000` here; change one and you must
change the other.

## Deploying to la ruche

Sablier runs on la ruche behind the Dokploy panel at `gare.facile.studio`. Prefer the
`dokploy` CLI over SSH plus `docker`:

```sh
dokploy compose --help
```

Environment values are set in the Dokploy project, not committed. `.env.example` is the
template of what needs filling in — at minimum `DATABASE_URL` and
`CORS_ALLOWED_ORIGINS`, which for a same-origin deployment is just the public hostname.

## Migrations

There is no migration step to run. `schemas.Migrate` executes on every boot: GORM
`AutoMigrate` over all twelve models, then the user-color backfill and the legacy
time-entry-to-task backfill. Both backfills are idempotent — they only touch rows that
still need it — so restarts are cheap.

A failed migration aborts startup and the container exits 1, which Dokploy surfaces as a
failed deploy rather than a half-migrated database.

## Verifying a deploy

`/health` answers as soon as the process is serving and touches nothing, so a green
`/health` proves the binary started and nothing more. `/ready` pings Postgres and is the
one that tells you the database is reachable. Neither says anything about the SPA — for
that, request `/` and confirm you get the client's `index.html` rather than a 404, which
is what a missing or misplaced `CLIENT_DIR` looks like.

## Persistent state

| Volume | Mounted at | Holds |
|---|---|---|
| `sablier_db_data` | `/var/lib/postgresql/data` | The database |
| `sablier_api_data` | `/data` | Uploaded and OIDC-fetched avatars under `/data/avatars` |

Avatars are files on disk, not rows. Losing `sablier_api_data` loses every uploaded
avatar; OIDC-sourced ones come back on the next profile sync, uploaded ones do not.
