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

- `VITE_API_BASE_URL`: API base URL used by the frontend, default `http://localhost:4000`

The production Docker build injects `VITE_API_BASE_URL` at build time and serves the static output with Nginx.
