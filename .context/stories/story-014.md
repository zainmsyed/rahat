# Story 014: Enable self-service web access through Telegram

**Status:** complete  
**Type:** auth/integration  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-25  
**Completed:** 2026-07-25

---

## Goal
Let an existing beta tester securely return to Rahat without email delivery, a saved invite URL, or operator assistance. A tester whose Telegram chat is already linked to exactly one Rahat account can send `/edit` to the shared bot and receive a private, short-lived, single-use link that signs that account into the web app and opens routine management.

## Verification
From each of two separately linked Telegram test chats, sending `/edit` returns a different private “Manage my routines” link. Opening a link once creates a browser session for only the Rahat user linked to that Telegram chat and redirects to `/tasks`; replay, expiry, forwarding after use, unknown chats, and attempts to ambiguously link one Telegram chat to multiple users are handled safely. A browser with a valid durable session can continue revisiting the normal app URL without requesting another link.

## Scope — files this story may touch
- db/migrations/
- internal/auth/
- internal/users/
- internal/notifications/telegram/
- internal/webhooks/telegram/
- cmd/server/
- web/src/routes/login/
- web/src/hooks.*
- README.md
- deploy/

## Out of scope — do not touch
- Email delivery or email magic links
- Password authentication
- PocketBase or another replacement identity provider
- Household sharing, assignment, or roles
- Permanent per-user bearer links
- Changes to task-management behavior beyond redirecting authenticated users to `/tasks`

## Dependencies
- Story 004
- Story 007
- Story 012
- Story 013

---

## Checklist
- [x] Enforce a one-to-one identity link so a non-empty Telegram `chat_id` can belong to only one Rahat user, with migration and relink-conflict handling for existing data
- [x] Route the `/edit` command through every enabled Telegram update transport, using the verified incoming chat identity rather than a supplied `user_id`, email, or URL parameter
- [x] Resolve the linked backend user from Telegram `chat_id` and return safe guidance for unknown or unlinked chats without leaking account details
- [x] Reuse the existing access-grant/session exchange system to issue a fresh cryptographically random token bound server-side to that user
- [x] Make Telegram-issued grants single-use and short-lived (target 10–15 minutes), store only token hashes, and avoid logging raw tokens
- [x] Reply privately with clear expiry/do-not-forward copy and an inline “Manage my routines” URL button
- [x] Exchange the link into the existing secure HttpOnly browser session and redirect successfully authenticated users to `/tasks`
- [x] Preserve the normal revisit flow: users with a valid browser session open the app directly and do not need to send `/edit` each time
- [x] Log issuance and exchange outcomes without storing bearer secrets or exposing whether unrelated accounts exist
- [x] Add automated tests for two-user isolation, unknown chats, duplicate Telegram linkage, expiry, single use/replay, wrong-user access, command parsing, and both polling/webhook update routing
- [x] Document the tester-facing return flow, session lifetime, new-device/cookie-loss recovery, and operator fallback

---

## Issues

---

## Completion Summary

Story 014 lets a linked beta tester return to the web app by sending `/edit` to the Rahat Telegram bot.

Database changes:
- Added migration `011_story_014_telegram_self_service.sql` which records pre-existing duplicate Telegram chat IDs in `telegram_identity_conflicts`, keeps the earliest-created user linked, nullifies the duplicates, and adds a partial unique index enforcing one non-empty `telegram_chat_id` per user.

Backend changes:
- `internal/users` gained `GetByTelegramChatID` and `LinkTelegramChat`. `LinkTelegramChat` uses a transaction to detect an existing link to another user and returns `ErrTelegramChatLinked`, preventing ambiguous identity links.
- `internal/notifications/telegram` added `EditCommandHandler`. It handles private `/edit` messages by resolving the user from the verified `chat_id`, issuing a 15-minute `AccessGrant` through the existing auth service, and replying with a private inline button labeled “Manage my routines”. Raw tokens are never logged; only the grant selector and expiry are logged.
- `cmd/server/telegram_router.go` routes `/edit` to the edit handler and all other messages (e.g. `/start` onboarding codes) to onboarding. The router is passed to both the long-polling poller and the webhook handler, so `/edit` works on every enabled transport.
- `cmd/server/onboarding.go` now links Telegram via `LinkTelegramChat` so duplicate attempts during onboarding are rejected with a clear private reply instead of corrupting the one-to-one mapping.
- `cmd/server/auth.go` already exchanges grants into secure HttpOnly sessions.

Frontend changes:
- `web/src/routes/login/+page.svelte` now redirects to `/tasks` after a successful access-link exchange, matching the self-service return flow.

Tests:
- Added backend tests for one-to-one enforcement, duplicate linkage, unknown-chat guidance, 15-minute expiry, single-use/replay, wrong-user exchange, command parsing, poller/webhook routing, and migration conflict resolution.
- Updated the login page test to expect redirection to `/tasks`.

Documentation:
- Updated `README.md` and `deploy/README.md` with the `/edit` self-service flow, link lifetime, session behavior, new-device/cookie-loss recovery, and operator fallback.

All Go tests and web tests pass.