# Story 006: Build the onboarding web flow

**Status:** in-progress  
**Type:** full-stack  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-17  
**Completed:** —

---

## Goal
Provide a lightweight SvelteKit onboarding flow so a tester can create a Rahat profile, set timezone and daily budget, add Telegram and email contact details, connect Google Calendar, and load starter or custom tasks without needing a full account system or persistent dashboard. This story should make first-use setup realistic for the initial testing group.

## Verification
A brand-new tester can finish the onboarding flow and end with a saved profile, at least one task, optional calendar connection, and enough data for the next daily schedule run without manual database edits.

## Scope — files this story may touch
- web/src/routes/onboarding/
- web/src/lib/components/
- web/src/lib/api/
- internal/users/
- internal/tasks/
- internal/calendar/
- cmd/server/

## Out of scope — do not touch
- A day-to-day task management dashboard
- Multi-user accounts or household assignment
- SMS setup

## Dependencies
- Story 001
- Story 002
- Story 005

---

## Checklist
- [x] Add a minimal invite or token-based onboarding session flow without a full auth system
- [x] Capture profile basics, local timezone, and daily task-time budget
- [x] Let the user supply Telegram and email contact details with clear channel roles
- [x] Offer starter-library tasks plus manual task and subtask creation or editing
- [x] Expose Google Calendar connect and disconnect steps using the read-only integration
- [x] Finish onboarding by validating required data and triggering the first schedule seed or run
- [x] Delete the Google calendar connection row on disconnect via `internal/store` `CalendarConnectionRepository.Delete` (scope exception for `internal/store/` explicitly approved by the user; originally implemented as token-clearing to stay in scope)

---

## Issues

---

## Completion Summary

Implemented the full onboarding flow across the Go API and the SvelteKit frontend.

**Backend** (`cmd/server/onboarding.go`, wiring in `cmd/server/main.go`):
- Invite-gated session flow with no auth system: `POST /onboarding/start` validates an invite token from `ONBOARDING_INVITE_TOKENS` (comma-separated; defaults to `rahat-tester` in development), creates the user, and returns a stateless HMAC-SHA256-signed session token (7-day TTL, secret from `ONBOARDING_SESSION_SECRET`, dev default in development). All other `/onboarding/*` endpoints require `Authorization: Bearer <token>`; no session table or migration was needed.
- Profile endpoint validates display name, IANA timezone (`time.LoadLocation`), and a 10–960 minute daily budget.
- Contacts endpoint stores Telegram chat ID and email on the user and upserts channel preferences with explicit roles: Telegram = interactive check-ins/reminders (primary when set), email = daily recaps.
- Task endpoints: list starter library, add-from-starter (copies template subtasks), manual create/edit/delete with subtasks. Editing replaces subtasks atomically via the new transactional `tasks.Service.ReplaceSubtasks`. Ownership is checked against the session user.
- Calendar endpoints: status, auth-url, connect (verifies the OAuth state's user matches the session), and disconnect via new `calendar.Service.DisconnectGoogle` / `GoogleConnectionStatus`. Disconnect deletes the connection row entirely through a new idempotent `CalendarConnectionRepository.Delete` in `internal/store/calendar_state.go` — a scope exception for `internal/store/` that the user explicitly approved during implementation.
- `POST /onboarding/complete` validates required data (name, valid timezone, budget > 0, at least one contact channel, at least one task) returning a `missing` list on failure, best-effort syncs Google Calendar when connected, then runs `scheduler.PlanDay` for today in the user's timezone to seed the first schedule.
- Added `withCORS` middleware (origin from existing `WEB_ORIGIN` config) so the SvelteKit dev server can call the API, including OPTIONS preflights.

**Frontend** (`web/src/routes/onboarding/`, `web/src/lib/api/`, `web/src/lib/components/`):
- Six-step wizard (Invite → Profile → Contacts → Tasks → Calendar → Finish) with a progress indicator; the session token is kept in `localStorage` so testers can resume, and `?step=` deep-links support returning from the OAuth redirect.
- Typed API client (`$lib/api/client.ts`, `$lib/api/types.ts`) with shared cadence/priority/time-of-day options.
- Profile step offers all IANA timezones via `Intl.supportedValuesOf`, defaulting to the browser timezone; contacts step explains channel roles; tasks step lists existing tasks with edit/delete, a starter library with one-click add, and a custom task form with dynamic subtask rows; calendar step handles connect (redirect) and disconnect; finish step shows a readiness summary and calls `/onboarding/complete`, reporting scheduled/overflowed/skipped counts.
- OAuth callback route `/onboarding/calendar/callback` exchanges `state`/`code` through the authenticated API and links back to the calendar step.

**Verification:** `go test ./...` passes (new `cmd/server/onboarding_test.go` covers token signing/tampering/expiry, invite parsing, auth enforcement, and a full HTTP flow through completion); `npm run check` and `npm run build` are clean; a live smoke test against the real server completed onboarding end-to-end and seeded 3 scheduled occurrences.
