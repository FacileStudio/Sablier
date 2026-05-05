# Sablier

Sablier is a monorepo containing a Go API and a SvelteKit frontend.

- `apps/api`: Go backend
- `apps/client`: SvelteKit frontend

## Quick Start

### 1. Prerequisites

- [Go 1.24+](https://go.dev/dl/)
- [Node.js](https://nodejs.org/) (LTS recommended)
- [Docker](https://www.docker.com/)

### 2. Infrastructure

Start the development database using Docker Compose:

```sh
docker compose up db -d
```

The database will be available at `localhost:5432`.

### 3. API Setup

Navigate to the API directory and start the server:

```sh
cd apps/api
go run main.go
```

The API server runs on `http://localhost:4000` by default. It will automatically run database migrations on startup.

### 4. Frontend Setup

Navigate to the client directory, install dependencies, and start the development server:

```sh
cd apps/client
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173` (or `http://localhost:3000` depending on your local config).

## Environment Configuration

Both the API and Client can be configured via environment variables.

- **API**: See `apps/api/.env.example`. Since the API does not load `.env` files automatically, ensure variables are exported in your shell or provided via a tool like `direnv`.
- **Client**: See `apps/client/.env.example` (or `apps/api/.env.example` for shared values). SvelteKit/Vite will load `.env` files automatically.

## Docker Compose

To run the entire stack (API, Client, and DB) using Docker:

```sh
docker compose up --build
```

## Common Commands

### API (`apps/api`)

- `go run main.go`: Run the API in development mode.
- `go build -o bin/api .`: Build the API binary.
- `go test ./...`: Run API tests.

### Client (`apps/client`)

- `npm run dev`: Start the development server.
- `npm run build`: Build the frontend for production.
- `npm run check`: Run type and svelte checks.
- `npm run preview`: Preview the production build locally.
