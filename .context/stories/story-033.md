# Story 033: Update Coolify deployment docs for single-container long-polling deployment

**Status:** in-progress  
**Type:** ops  
**Created:** 2026-07-26  
**Last accessed:** 2026-08-03  
**Completed:** —

---

## Goal
Bring the deployment runbook in line with the single-container, long-polling setup so the app can be deployed on Coolify without needing split services, a separate frontend domain, or Telegram webhook configuration.

## Verification
A teammate can follow `deploy/README.md` to create a Coolify service from the repo-root Dockerfile, mount a persistent `/data` volume, set the required environment variables, and have a running Rahat instance that serves the web UI and responds to Telegram long polling.

## Scope — files this story may touch
- deploy/README.md
- deploy/launch-smoke-checklist.md
- .env.example (new)

## Out of scope — do not touch
- Application code or tests
- The Dockerfile itself (covered in Story 032)

## Dependencies
- Story 032

---

## Checklist
- [x] Update `deploy/README.md` Coolify section for a single service built from the repo-root `Dockerfile`.
- [x] Document the required `/data` volume mount for SQLite persistence.
- [x] List the minimum environment variables for long-polling mode and note which are optional.
- [x] Add or update `deploy/launch-smoke-checklist.md` to include static-asset serving and the landing page.
- [x] Create a `.env.example` file at the repo root that shows production env vars without secrets.
- [x] Verify the docs mention that Telegram webhooks remain optional and off by default.

---

## Issues

---

## Completion Summary

Updated the deployment runbook for the single-container Coolify deployment. `deploy/README.md` now documents one service built from the repo-root Dockerfile, port 8080, a persistent `/data` mount for SQLite/WAL/runtime data, the minimum production environment variables, optional integrations, and long-polling as the default Telegram transport with webhooks off unless explicitly configured. Updated `deploy/launch-smoke-checklist.md` with landing-page, health, static-asset, persistence, onboarding, session, and long-polling checks. Replaced `.env.example` with a production-oriented, secret-free runtime configuration template.
