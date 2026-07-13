# Story 004: Deliver Telegram reminders and adaptive check-ins

**Status:** in-progress  
**Type:** integration  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-13  
**Completed:** —

---

## Goal
Make Telegram the primary interactive v1 surface by sending batched daily lists, window-based reminders, and adaptive follow-ups with inline actions for Done, Not Yet, snooze, reschedule, and pause controls. Prefer long polling as the default v1 transport so local development and testing stay simple, while allowing webhook mode when a user has domain-backed deployment infrastructure. This story should implement the guilt-reducing check-in loop described in the PRD without introducing SMS or a dashboard dependency.

## Verification
In a dev or staging Telegram chat, a tester can start the app without public webhook infrastructure, receive today’s list, respond through inline buttons, and see occurrence state update correctly through completion, consecutive No responses, snoozes, reschedules, and pause actions.

## Scope — files this story may touch
- internal/notifications/telegram/
- internal/webhooks/telegram/
- internal/checkins/
- internal/occurrences/
- internal/events/
- cmd/server/

## Transport note
- Long polling is the default and preferred v1 path for development, testing, and early production simplicity.
- If deployment settings provide a suitable public domain and webhook configuration, webhook mode may be enabled.
- Without that domain-backed webhook setup, the app should fall back to long polling rather than treating webhook infrastructure as required.

## Out of scope — do not touch
- SMS or Twilio delivery
- Email recap generation
- Full onboarding UI

## Dependencies
- Story 001
- Story 002
- Story 003

---

## Checklist
- [x] Support Telegram bot configuration with long polling as the default runtime path, and allow webhook mode only when deployment settings support it
- [x] Select Telegram transport at runtime: use webhook mode only when domain-backed webhook settings are present; otherwise fall back to long polling
- [x] Verify local/dev startup works end-to-end without public webhook infrastructure
- [x] Send one batched morning message plus per-window task or subtask reminders from scheduled occurrences
- [x] Capture Done and Not Yet responses and advance occurrence state correctly
- [x] Implement consecutive-No check-ins with snooze, reschedule, and skip limits from the PRD
- [x] Add pause-everything-today, pause-this-week, and per-task pause actions in Telegram
- [x] Log outbound messages and user responses for later retention and tone analysis

---

## Issues
- 2026-07-12: Story direction updated to keep Telegram long polling as the default path for development and early deployment, while allowing webhook mode only for users with domain-backed webhook infrastructure. Remaining implementation work should stay inside Story 004 and continue through the normal Vazir story review loop.

---

## Completion Summary
- Added a Telegram notification package with an HTTP Bot API client plus message builders for daily batched lists and per-window reminders.
- Added runtime Telegram transport selection so the app enables webhook mode only when `TELEGRAM_WEBHOOK_URL` and `TELEGRAM_WEBHOOK_SECRET` describe a secure domain-backed endpoint; otherwise it deletes any stale webhook and falls back to long polling.
- Added a long-polling update loop for Telegram callback queries so local and dev startup work without public webhook infrastructure.
- Prioritized long polling as the default v1 transport so developers and testers can exercise Telegram flows without public webhook setup, while preserving webhook mode as an optional path for domain-backed deployments.
- Added callback handling for Done, Not Yet, Snooze, Reschedule, Skip, Pause everything today, Pause this week, and Pause this task actions.
- Implemented the consecutive-No check-in loop so a second Not Yet triggers adaptive options, snooze pushes the occurrence out three days, reschedule moves it forward, and repeated reschedule offers eventually skip the occurrence.
- Logged outbound Telegram messages and inbound user responses through the existing event log table for later analysis.
- Added tests covering reminder sending plus callback-driven occurrence completion, snooze behavior, adaptive check-ins, pause creation, runtime transport selection, and long-poll callback handling.
