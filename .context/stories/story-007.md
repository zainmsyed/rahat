# Story 007: Build the guided Telegram connection onboarding

**Status:** not-started  
**Type:** full-stack  
**Created:** 2026-07-21  
**Last accessed:** 2026-07-21  
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
- [ ] Replace manual Telegram chat ID entry with a short alphanumeric code the tester sends to the bot
- [ ] Provide a one-tap Telegram deep link (and a QR code if feasible) that opens the bot with the code prefilled
- [ ] Match the incoming bot message to the onboarding session, save the chat ID automatically, and confirm the connection on-screen without a page reload
- [ ] Send an in-chat welcome and confirmation message after a successful link
- [ ] Handle Telegram being unavailable or declined with clear guidance and an email-only fallback path

---

## Issues

---

## Completion Summary
