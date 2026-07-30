# Story 031: Serve static frontend assets from the Go backend

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-30  
**Completed:** 2026-07-30

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
- [x] Add a `WEB_STATIC_DIR` config option for the path to built static files.
- [x] Implement a static-file handler that serves files from `WEB_STATIC_DIR` on `/`.
- [x] Ensure API and ops routes are registered before the static handler so they take precedence.
- [x] Add a catch-all fallback to `index.html` for unmatched non-API routes to support SvelteKit client-side routing.
- [x] Verify `/healthz` and `/readyz` still return 200 when static serving is enabled.
- [x] Manually smoke-test `/`, `/login`, and `/lookahead?token=...` with a local build.

---

## Issues

---

## Completion Summary
Added `WEB_STATIC_DIR` to `internal/config/config.go` and implemented `internal/web/static.go` to serve built SvelteKit files. The handler serves existing files (including hashed assets under `/_app`) and falls back to `index.html` for any non-API path, enabling SvelteKit client-side routing. Known API prefixes (`/healthz`, `/readyz`, `/auth`, `/onboarding`, `/tasks`, `/calendar`, `/schedule`, `/lookahead`, `/telegram`, `/webhooks`) are passed through to the existing mux. `cmd/server/main.go` wraps the CORS-enabled mux with the static handler when `WEB_STATIC_DIR` is set. Added `internal/web/static_test.go` covering file serving, index.html fallback, API passthrough, non-GET passthrough, and the disabled case. All Go tests pass, and a manual smoke test confirmed `/healthz` and `/readyz` return 200 while `/`, `/login`, `/lookahead?token=...`, and `/_app/version.json` are served correctly.