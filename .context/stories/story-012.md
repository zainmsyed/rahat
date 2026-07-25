# Story 012: Establish durable beta web sessions

**Status:** in-progress  
**Type:** auth  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-25  
**Completed:** —

---

## Goal
Give each beta tester a durable, revocable authenticated web session after onboarding so post-onboarding pages can identify the current user without reusing onboarding tokens, exposing bearer tokens in normal page URLs, or requiring the deferred email-delivery system.

## Verification
A tester can finish onboarding and remain signed in on that browser, sign out, and regain access through a single-use operator-issued beta access link. Protected endpoints reject missing, expired, revoked, or cross-user sessions, and no protected page accepts a raw `user_id` as authorization.

## Scope — files this story may touch
- db/migrations/
- internal/auth/
- internal/users/
- cmd/server/
- scripts/
- web/src/lib/api/
- web/src/routes/login/
- web/src/routes/onboarding/
- web/src/hooks.*
- README.md
- deploy/

## Out of scope — do not touch
- Password authentication
- Google or social login
- Sending login links by email
- Multi-user household accounts or roles
- A general admin dashboard
- Post-onboarding task CRUD UI

## Dependencies
- Story 006
- Story 007
- Story 011

---

## Checklist
- [x] Add migrations and repositories for hashed single-use access grants and revocable web sessions with expirations
- [x] Establish an authenticated session when onboarding finishes without treating the onboarding token as the permanent credential
- [x] Add an operator CLI/script that issues a short-lived, single-use beta access link for an existing tester
- [x] Add access-link exchange, current-session, and logout endpoints using secure HttpOnly cookies
- [x] Add shared middleware/helpers that resolve the authenticated user and reject raw `user_id` authorization on protected routes
- [x] Apply secure cookie, SameSite, origin/CSRF, expiration, and revocation behavior appropriate to local and production environments
- [x] Add a minimal login/access-link page and authenticated-route handling in the SvelteKit app
- [x] Add tests for onboarding session promotion, one-time exchange, expiry, revocation/logout, cookie flags, missing sessions, and cross-user access attempts
- [x] Document beta access issuance, session secrets, production cookie requirements, and recovery procedure

---

## Issues

---

## Completion Summary

Story 012 added a durable beta-session layer without introducing passwords or email delivery. Backend persistence now includes hashed single-use `beta_access_grants` and revocable `web_sessions` with expirations. The new `internal/auth` service issues access links, exchanges them exactly once into durable opaque browser sessions, verifies session cookies, and revokes sessions on logout.

Onboarding completion now promotes the user into a real web session by setting a secure HttpOnly `rahat_session` cookie instead of treating the onboarding token as a permanent credential. A new auth handler exposes access-link exchange, current-session, and logout endpoints, applies trusted-origin checks on state-changing cookie flows, and provides shared authenticated-user helpers for protected routes. Existing protected integration points such as schedule planning and Google Calendar auth/sync now resolve ownership from the authenticated session rather than accepting a raw `user_id` query parameter.

On the web side, SvelteKit now has a minimal `/login` page for access-link exchange and sign-out, a server hook that enforces authenticated routing for protected pages, and onboarding finish now stores the session cookie and clears the temporary onboarding token path. New-tester setup is simpler for the MVP: `scripts/issue-onboarding-link.sh` prints a `/onboarding?invite=...` URL that starts onboarding automatically without asking the tester to type the invite code. Operator recovery uses `scripts/issue-beta-access.sh <user-id-or-email>` / `go run ./cmd/server ops:issue-access-link ...` to produce single-use beta access links for existing testers.

Verification completed with `go test ./...`, `cd web && npm run check`, and `cd web && npm test`. Documentation was updated in `README.md` and `deploy/README.md` to cover `WEB_SESSION_SECRET`, beta access issuance, cookie/origin requirements, and tester recovery flow.
