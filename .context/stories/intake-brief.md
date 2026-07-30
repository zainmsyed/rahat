# Intake Brief

**Last updated:** 2026-07-26

## Planning brief
Add production-ready single-container Docker deployment for Rahat on the existing Coolify/Hetzner infrastructure. The container should build and serve both the Go backend and the SvelteKit frontend from the same origin, keep Telegram on long polling for now, and persist the SQLite database on a mounted volume.

## Source files
- .context/intake/prd/rahat-prd.md (14465 bytes)
- .context/intake/references/rahat design file.html (66635 bytes)
- .context/intake/references/rahat.html (45120 bytes)

## Distilled notes
### .context/intake/prd/rahat-prd.md
- Hosting: Coolify on a Hetzner VPS; SQLite in WAL mode; Telegram long polling is acceptable when no stable domain is available.
- Architecture: Go backend, SvelteKit frontend, SQLite database.
- Existing launch tooling (Story 011) already documents Coolify/Hetzner steps and required env vars, but the repo currently has only a backend-only Dockerfile.
- The frontend is intentionally lightweight (onboarding + routine maintenance + passive lookahead), so serving static files from the backend is viable.

### Deployment decisions (clarified)
- **Single container** is preferred over split backend/frontend services.
- **Coolify can deploy the single Dockerfile** as one service.
- **Telegram stays on long polling** in v1, so the container does not require a public HTTPS webhook path.
- **Frontend static files will be served by the Go backend** from the same origin, making API calls origin-relative.

## Planning rules
- Treat listed source files as user-authored planning inputs unless they are explicitly marked as generated artifacts.
- Vazir-generated files in .context/stories/ are replan context, not primary intake.
- Read all text-based planning sources before asking questions.
- Ask only implementation-blocking delta questions after reviewing this brief and any raw files you actually need.
- State safe default assumptions briefly so the user can correct them.
- Surface contradictions instead of resolving them silently.
