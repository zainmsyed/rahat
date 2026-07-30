# Story 032: Build a single-container Dockerfile for backend and frontend

**Status:** not-started  
**Type:** ops  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Replace the backend-only Dockerfile with a multi-stage build that compiles the SvelteKit frontend, copies the static output into the Go runtime image, and starts the backend configured to serve those static files from a single container.

## Verification
`docker build -t rahat .` succeeds from the repo root, and `docker run -p 8080:8080 -e TELEGRAM_BOT_TOKEN=... rahat` serves the landing page and API on port 8080.

## Scope — files this story may touch
- Dockerfile
- .dockerignore
- web/package.json (lockfile)
- deploy/README.md

## Out of scope — do not touch
- Application source code
- Database schema or migrations
- Coolify service metadata (covered in Story 033)

## Dependencies
- Story 030
- Story 031

---

## Checklist
- [ ] Rewrite `Dockerfile` with a Node stage that builds `web/` and a Go stage that builds `cmd/server`.
- [ ] Copy `web/build` into the final image and set `WEB_STATIC_DIR` to its path.
- [ ] Ensure the final image exposes port 8080 and runs as a non-root user if practical.
- [ ] Update `.dockerignore` to exclude `web/node_modules`, `.svelte-kit`, `var/`, and local env files.
- [ ] Run a local `docker build` and container smoke test against `/healthz` and the landing page.
- [ ] Document any new build-time env vars in `deploy/README.md`.

---

## Issues

---

## Completion Summary
