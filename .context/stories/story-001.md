# Story 001: Bootstrap the Rahat app workspace

**Status:** not-started  
**Type:** platform  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

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
- [ ] Create the base repo layout for the Go backend, SvelteKit frontend, shared config, and docs
- [ ] Add local development commands for running the API, web app, and SQLite setup
- [ ] Wire SQLite initialization with WAL mode and a basic connection health check
- [ ] Add a minimal HTTP server with health and readiness endpoints plus structured logging
- [ ] Add frontend shell files so the web app boots with a placeholder Rahat page
- [ ] Document required environment variables, local startup, and deployment assumptions

---

## Issues

---

## Completion Summary
