# Story 031: Serve static frontend assets from the Go backend

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Add a static-file handler to the Go backend so it can serve the built SvelteKit frontend from the same origin. API routes must remain unaffected, and client-side routes should fall back to `index.html` so SvelteKit's router can take over.

## Verification
Running the server with `WEB_STATIC_DIR=./web/build` serves the landing page at `/`, the onboarding/login pages, and the read-only lookahead page, while `/healthz`, `/readyz`, and all API routes continue to respond correctly.

## Scope — files this story may touch
- cmd/server/main.go
- internal/config/config.go
- internal/web/static.go (new)
- web/build/ (generated)

## Out of scope — do not touch
- Frontend source code or build configuration
- Telegram or Google Calendar handlers
- SQLite persistence layer

## Dependencies
- Story 030

---

## Checklist
- [ ] Add a `WEB_STATIC_DIR` config option for the path to built static files.
- [ ] Implement a static-file handler that serves files from `WEB_STATIC_DIR` on `/`.
- [ ] Ensure API and ops routes are registered before the static handler so they take precedence.
- [ ] Add a catch-all fallback to `index.html` for unmatched non-API routes to support SvelteKit client-side routing.
- [ ] Verify `/healthz` and `/readyz` still return 200 when static serving is enabled.
- [ ] Manually smoke-test `/`, `/login`, and `/lookahead?token=...` with a local build.

---

## Issues

---

## Completion Summary
