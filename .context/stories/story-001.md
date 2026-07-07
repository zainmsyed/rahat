# Story 001: Bootstrap the Rahat app workspace

**Status:** complete  
**Type:** platform  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** 2026-07-07

---

## Goal
Set up this greenfield repo as a runnable Rahat workspace with a Go backend, SQLite foundation, and SvelteKit frontend shell so later stories can add product logic without first inventing structure. This story establishes the top-level layout, shared configuration, local run commands, and a basic deployment baseline for the planned Coolify/Hetzner environment.

## Verification
A developer can clone the repo, copy the env example, run the backend and web app locally, create or open the SQLite database in WAL mode, and hit a basic health endpoint plus placeholder web page using only documented commands.

## Scope — files this story may touch
- go.mod
- cmd/server/
- internal/app/
- internal/config/
- internal/db/
- web/
- Makefile
- Dockerfile
- .env.example
- README.md

## Out of scope — do not touch
- Scheduling rules and occurrence generation
- Telegram, email, or Google Calendar integrations
- Real onboarding forms beyond placeholders

## Dependencies
- None

---

## Checklist
- [x] Create the base repo layout for the Go backend, SvelteKit frontend, shared config, and docs
- [x] Add local development commands for running the API, web app, and SQLite setup
- [x] Wire SQLite initialization with WAL mode and a basic connection health check
- [x] Add a minimal HTTP server with health and readiness endpoints plus structured logging
- [x] Add frontend shell files so the web app boots with a placeholder Rahat page
- [x] Document required environment variables, local startup, and deployment assumptions

---

## Issues

---

## Completion Summary
- Bootstrapped the repo with a Go API entrypoint, shared config package, SQLite bootstrap helper, and HTTP app package.
- Added SQLite initialization that creates the database path, enables WAL mode and foreign keys, creates a bootstrap table, and verifies connectivity.
- Added `/healthz` and `/readyz` JSON endpoints with structured request logging via `log/slog`.
- Created a SvelteKit frontend shell with a placeholder Rahat landing page linking to the backend health endpoints.
- Added `Makefile` commands, `.env.example`, a backend `Dockerfile`, and README setup/deployment guidance so a developer can start both services locally.
