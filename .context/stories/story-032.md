# Story 032: Build a single-container Dockerfile for backend and frontend

**Status:** in-progress  
**Type:** ops  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-30  
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
- [x] Rewrite `Dockerfile` with a Node stage that builds `web/` and a Go stage that builds `cmd/server`.
- [x] Copy `web/build` into the final image and set `WEB_STATIC_DIR` to its path.
- [x] Ensure the final image exposes port 8080 and runs as a non-root user if practical.
- [x] Update `.dockerignore` to exclude `web/node_modules`, `.svelte-kit`, `var/`, and local env files.
- [ ] Run a local `docker build` and container smoke test against `/healthz` and the landing page.
- [x] Document any new build-time env vars in `deploy/README.md`.

---

## Issues

- **Local Docker build not executed.** The `docker` CLI is not available in this environment, so the `docker build -t rahat .` verification and container smoke test against `/healthz` and the landing page could not be run. The frontend build was verified locally (`cd web && npm run build`), but the full multi-stage image still needs to be built and run elsewhere.

---

## Completion Summary

Replaced the backend-only `Dockerfile` with a three-stage build: a Node stage compiles the SvelteKit frontend, a Go stage builds `cmd/server`, and a slim Debian runtime image copies the static build into `/app/web/static` and sets `WEB_STATIC_DIR` to that path. The runtime image exposes port 8080 and runs as a non-root `rahat` user with a writable `/data` directory for the SQLite database. Added a `.dockerignore` that excludes frontend build artifacts, dependencies, local env files, runtime data, and VCS metadata. Updated `deploy/README.md` with build-time notes about the multi-stage Dockerfile and the optional `VITE_API_BASE_URL` build argument. The only remaining checklist item is the local Docker build and container smoke test, which is blocked by the absence of the Docker CLI in this environment.
