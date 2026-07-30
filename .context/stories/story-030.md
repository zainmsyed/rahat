# Story 030: Configure SvelteKit for static export and origin-relative API calls

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Make the SvelteKit frontend buildable as static files that can be served by the Go backend from the same origin. Ensure all frontend API calls work when the backend and frontend share a host by supporting an empty or origin-relative API base URL.

## Verification
`npm run build` produces static files under `web/build`, all API modules continue to work in dev when a base URL is supplied, and they use origin-relative requests in production when no base URL is supplied. `npm test` and `npm run check` pass with no errors.

## Scope — files this story may touch
- web/svelte.config.js
- web/package.json
- web/package-lock.json
- web/src/lib/api/config.ts
- web/src/lib/api/auth.ts
- web/src/lib/api/onboarding.ts
- web/src/lib/api/tasks.ts
- web/src/lib/api/lookahead.ts
- web/src/lib/api/calendar.ts

## Out of scope — do not touch
- Backend API routes
- Go binary or Dockerfile
- Runtime server-side rendering

## Dependencies
- Story 020
- Story 029

---

## Checklist
- [ ] Install and configure `@sveltejs/adapter-static` with SPA fallback for `web/`.
- [ ] Update `web/src/lib/api/config.ts` so an empty `VITE_API_BASE_URL` falls back to origin-relative requests.
- [ ] Audit each API module for hardcoded `http://localhost:8080` or `import.meta.env` fallbacks and make them relative-aware.
- [ ] Verify `npm run build` outputs a `web/build` directory containing `index.html` and assets.
- [ ] Run `npm test` and `npm run check` and fix any regressions.

---

## Issues

---

## Completion Summary
