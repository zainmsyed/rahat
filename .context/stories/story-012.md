# Story 012: Establish durable beta web sessions

**Status:** not-started  
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
- [ ] Add migrations and repositories for hashed single-use access grants and revocable web sessions with expirations
- [ ] Establish an authenticated session when onboarding finishes without treating the onboarding token as the permanent credential
- [ ] Add an operator CLI/script that issues a short-lived, single-use beta access link for an existing tester
- [ ] Add access-link exchange, current-session, and logout endpoints using secure HttpOnly cookies
- [ ] Add shared middleware/helpers that resolve the authenticated user and reject raw `user_id` authorization on protected routes
- [ ] Apply secure cookie, SameSite, origin/CSRF, expiration, and revocation behavior appropriate to local and production environments
- [ ] Add a minimal login/access-link page and authenticated-route handling in the SvelteKit app
- [ ] Add tests for onboarding session promotion, one-time exchange, expiry, revocation/logout, cookie flags, missing sessions, and cross-user access attempts
- [ ] Document beta access issuance, session secrets, production cookie requirements, and recovery procedure

---

## Issues

---

## Completion Summary
