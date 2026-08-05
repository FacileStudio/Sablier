# Client

SvelteKit frontend for Sablier.

## Responsibilities

- Login, registration, and OIDC entry flow
- Dashboard, sessions, projects, users, settings, and profile pages
- Live timer controls and manual session editing
- Avatar rendering and project/user activity summaries

## Run locally

```sh
bun install
bun run dev
```

Default dev URL: `http://localhost:5173`

## Scripts

```sh
bun run dev
bun run build
bun run preview
bun run check
```

## Configuration

The client calls the API on its own origin under `/api`, so it needs no build-time
configuration. `bun run dev` proxies `/api` and `/files` to `http://localhost:4000`; in
production the Go binary serves this build directly.
