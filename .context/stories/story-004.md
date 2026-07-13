# Story 004: Deliver Telegram reminders and adaptive check-ins

**Status:** in-progress  
**Type:** integration  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-08  
**Completed:** —

---

## Goal
Make Telegram the primary interactive v1 surface by sending batched daily lists, window-based reminders, and adaptive follow-ups with inline actions for Done, Not Yet, snooze, reschedule, and pause controls. Prefer long polling as the default v1 transport so local development and testing stay simple, while keeping webhook support optional for later deployment hardening. This story should implement the guilt-reducing check-in loop described in the PRD without introducing SMS or a dashboard dependency.

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
- Webhook handling may remain available as an optional deployment mode, but it is not required to verify or ship Story 004.

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
- [x] Support Telegram bot configuration with long polling as the default runtime path; keep webhook handling optional if retained
- [x] Send one batched morning message plus per-window task or subtask reminders from scheduled occurrences
- [x] Capture Done and Not Yet responses and advance occurrence state correctly
- [x] Implement consecutive-No check-ins with snooze, reschedule, and skip limits from the PRD
- [x] Add pause-everything-today, pause-this-week, and per-task pause actions in Telegram
- [x] Log outbound messages and user responses for later retention and tone analysis

---

## Issues

---

## Completion Summary
- Added a Telegram notification package with an HTTP Bot API client plus message builders for daily batched lists and per-window reminders.
- Prioritized long polling as the default v1 transport so developers and testers can exercise Telegram flows without public webhook setup; webhook support can remain optional where useful.
- Added callback handling for Done, Not Yet, Snooze, Reschedule, Skip, Pause everything today, Pause this week, and Pause this task actions.
- Implemented the consecutive-No check-in loop so a second Not Yet triggers adaptive options, snooze pushes the occurrence out three days, reschedule moves it forward, and repeated reschedule offers eventually skip the occurrence.
- Logged outbound Telegram messages and inbound user responses through the existing event log table for later analysis.
- Added tests covering reminder sending plus callback-driven occurrence completion, snooze behavior, adaptive check-ins, and pause creation.
