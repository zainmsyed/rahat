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
- `LOOKAHEAD_TOKEN_SECRET`: required in production for signed read-only lookahead links
- `LOOKAHEAD_TOKEN_ISSUER_ENABLED`: explicit opt-in for the non-production lookahead token helper
- `WEB_SESSION_SECRET`: required in production for hashing beta access grants and durable web sessions
- `EMAIL_RECAP_OUTBOX_DIR`: file outbox path used by the recap job
- `BACKUP_TARGET_URI`: backup destination, either a local path / `file://` path or an `s3://` URI

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

The web app starts on `http://localhost:5200` by default.

### 4. Ops commands and scripts

Story 011 adds operator-facing commands plus thin shell wrappers under `scripts/`.

Examples:

```bash
# Run one job immediately
bash scripts/run-daily-schedule.sh
bash scripts/run-telegram-daily.sh
bash scripts/run-telegram-window.sh
bash scripts/run-calendar-sync.sh
bash scripts/run-email-recap.sh
bash scripts/run-backup.sh

# Report event activity (JSON by default, or REPORT_FORMAT=csv)
bash scripts/report-events.sh

# Issue a setup link for a new tester; opens onboarding automatically
bash scripts/issue-onboarding-link.sh

# Issue a one-time beta access link for an existing tester
bash scripts/issue-beta-access.sh tester.one@example.com

# Seed demo testers into a non-production database
bash scripts/bootstrap-testers.sh

# Reset a non-production database (guarded)
RAHAT_RESET_CONFIRM=reset-non-production bash scripts/reset-nonprod.sh
```

All scripts source `.env` automatically when it exists and call the matching `go run ./cmd/server ops:...` command.

## Beta web sessions

Story 012 adds durable beta browser sessions backed by secure HttpOnly cookies rather than permanent onboarding tokens.

- New testers can start from `bash scripts/issue-onboarding-link.sh`; the generated `/onboarding?invite=...` URL starts setup automatically without asking them to type the invite code.
- Visual check: open the generated setup URL and confirm it briefly shows “Opening your onboarding steps…” before moving straight to the profile step, with no invite-code typing required.
- Finishing onboarding creates a web session on that browser.
- A signed-out tester can regain access through an operator-issued one-time link that opens `/login?token=...`.
- Generate a recovery link with `bash scripts/issue-beta-access.sh <user-id-or-email>`.
- In production, set `WEB_SESSION_SECRET` to a strong random secret and serve the frontend from the configured `WEB_ORIGIN` so cookie origin checks succeed.

### 5. Run the local verification target

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

- `http://localhost:5200` for the placeholder Rahat page
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

Story 011 adds deploy/runbook notes under:

- `deploy/README.md`
- `deploy/launch-smoke-checklist.md`

Use these docs for Coolify/Hetzner secret management, cron wiring, backups, Telegram webhook setup, Google OAuth setup, and launch smoke checks.

The included `Dockerfile` packages the Go API into a single container image suitable for an initial Coolify or Hetzner deployment. It expects a writable `/data` volume for SQLite and honors `PORT` or `RAHAT_HTTP_ADDR` at runtime.

For an initial deployment flow:

1. build the image from the repo root
2. mount persistent storage at `/data`
3. set `APP_ENV=production`
4. expose port `8080` or provide `PORT` from the platform

The frontend currently runs as a separate Node-based dev shell during development and can later be deployed independently or folded into a broader production setup in a future story.
