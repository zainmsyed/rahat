# Story 007: Build the guided Telegram connection onboarding

**Status:** in-progress  
**Type:** full-stack  
**Created:** 2026-07-21  
**Last accessed:** 2026-07-22  
**Completed:** —

---

## Goal
Let a non-technical beta tester connect Telegram during onboarding without ever seeing or typing a chat ID. The product uses a single shared Rahat bot operated by the project owner — every tester talks to the same bot in a private 1:1 chat, testers never create or configure bots, and no tester can see another tester's messages. The tester taps one button to open the Rahat bot, sends a short code, and watches the onboarding screen confirm the connection automatically. Telegram is presented as the recommended channel for interactive reminders, with a graceful email-only fallback when Telegram is unavailable or the tester prefers not to connect it.

## Verification
A tester with only a phone can connect Telegram in under a minute: open the bot via the deep link, send the short code (or tap start with it prefilled), see on-screen confirmation, and receive an in-chat welcome message. If Telegram is not configured on the server, the screen says so plainly and the tester can continue with email only.

## Scope — files this story may touch
- web/src/routes/onboarding/
- web/src/lib/components/
- web/src/lib/api/
- internal/users/
- internal/notifications/telegram/
- internal/webhooks/telegram/
- cmd/server/

## Out of scope — do not touch
- Google Calendar connection flow (Story 008)
- Daily reminder/check-in message content and scheduling logic (Story 004)
- SMS setup
- BotFather or operator bot provisioning (operator docs, not tester UX)

## Dependencies
- Story 004
- Story 006

---

## Checklist
- [x] Replace manual Telegram chat ID entry with a short alphanumeric code the tester sends to the bot
- [x] Provide a one-tap Telegram deep link (and a QR code if feasible) that opens the bot with the code prefilled
- [x] Match the incoming bot message to the onboarding session, save the chat ID automatically, and confirm the connection on-screen without a page reload
- [x] Send an in-chat welcome and confirmation message after a successful link
- [x] Handle Telegram being unavailable or declined with clear guidance and an email-only fallback path

---

## Issues

---

## Completion Summary
Added a new "Connect Telegram" onboarding step between profile and task selection. The backend now generates a short 6-character alphanumeric code per onboarding session, exposes a status endpoint (`GET /onboarding/telegram`) and a skip endpoint (`POST /onboarding/telegram/skip`), and routes incoming private Telegram messages through both the long-polling and webhook transports. When a tester opens the bot via the deep link and sends the code (or `/start <code>`), the server links the chat ID to the user, upserts Telegram as the primary interactive channel preference, and sends a welcome message. The frontend polls for the linked state every 2.5 seconds so confirmation appears without a reload. If Telegram is not configured, the screen plainly says so and offers an email-only continue path.

Files changed:
- Backend: `cmd/server/onboarding.go`, `cmd/server/onboarding_test.go`, `cmd/server/main.go`, `cmd/server/main_test.go`
- Telegram transport: `internal/notifications/telegram/types.go`, `internal/notifications/telegram/client.go`, `internal/notifications/telegram/poller.go`, `internal/notifications/telegram/poller_test.go`
- Webhook handler: `internal/webhooks/telegram/handler.go`
- Frontend: `web/src/lib/api/onboarding.ts`, `web/src/routes/onboarding/profile/+page.svelte`, `web/src/routes/onboarding/tasks/+page.svelte`, `web/src/routes/onboarding/review/+page.svelte`, `web/src/routes/onboarding/+page.svelte`, plus new `web/src/routes/onboarding/telegram/+page.svelte`
- Docs: `.env.example`, `README.md`

Verification:
- `go test ./...` passes
- `cd web && npm run check && npm run test && npm run build` pass

Note: the QR code is rendered via an external QR API for now; a fully local QR encoder can be swapped in later if desired.
