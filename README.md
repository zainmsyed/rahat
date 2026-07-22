# Rahat Workspace

This repository now includes a minimal Rahat starter workspace with:

- a Go API in `cmd/server`
- shared backend packages in `internal/`
- SQLite bootstrap with WAL mode
- a SvelteKit shell in `web/`
- local developer commands in `Makefile`
- a baseline backend `Dockerfile` for Coolify/Hetzner-style deployment

## Requirements

- Go 1.23+
- Node.js 20+
- npm 10+

## Environment setup

Copy the example environment file and adjust values if needed:

```bash
cp .env.example .env
```

Key variables:

- `APP_ENV`: `development` or `production`
- `LOG_LEVEL`: `debug`, `info`, `warn`, or `error`
- `RAHAT_HTTP_ADDR`: backend listen address for local runs
- `PORT`: optional deployment override used by platforms like Coolify
- `DATABASE_PATH`: SQLite file location
- `WEB_ORIGIN`: allowed CORS origin for the web app
- `VITE_API_BASE_URL`: frontend API base URL
- `TELEGRAM_BOT_TOKEN`: Telegram bot token for interactive reminders
- `TELEGRAM_BOT_USERNAME`: Telegram bot username (optional; fetched from Telegram if omitted)
- `TELEGRAM_WEBHOOK_SECRET` / `TELEGRAM_WEBHOOK_URL`: optional webhook mode for Telegram updates
- `TELEGRAM_API_BASE_URL`: optional custom Telegram API base URL

## Local development

### 1. Initialize SQLite

This creates the database directory, opens the SQLite file, enables WAL mode, and verifies the connection.

```bash
make db-setup
```

### 2. Run the API

```bash
set -a && source .env && set +a
make api
```

The API starts on `http://localhost:8080` by default.

### 3. Run the web app

Install frontend dependencies once:

```bash
make web-install
```

Then start the SvelteKit dev server:

```bash
set -a && source .env && set +a
make web
```

The web app starts on `http://localhost:5173` by default.

### 4. Run the local verification target

```bash
make ci
```

This runs backend tests plus the frontend install, check, and production build steps from a clean dependency state.

## Verification

Once both services are running, confirm the workspace bootstraps correctly:

```bash
curl http://localhost:8080/healthz
curl http://localhost:8080/readyz
```

Then open:

- `http://localhost:5173` for the placeholder Rahat page
- `http://localhost:8080/healthz` for backend health
- `http://localhost:8080/readyz` for backend readiness

## Backend layout

- `cmd/server/main.go`: application entrypoint
- `internal/config`: environment-backed config loading
- `internal/db`: SQLite initialization and bootstrap
- `internal/app`: HTTP routing, readiness checks, and structured request logging

## Frontend layout

- `web/package.json`: SvelteKit scripts and dependencies
- `web/src/routes/+page.svelte`: starter placeholder page

## Deployment baseline

The included `Dockerfile` packages the Go API into a single container image suitable for an initial Coolify or Hetzner deployment. It expects a writable `/data` volume for SQLite and honors `PORT` or `RAHAT_HTTP_ADDR` at runtime.

For an initial deployment flow:

1. build the image from the repo root
2. mount persistent storage at `/data`
3. set `APP_ENV=production`
4. expose port `8080` or provide `PORT` from the platform

The frontend currently runs as a separate Node-based dev shell during development and can later be deployed independently or folded into a broader production setup in a future story.
