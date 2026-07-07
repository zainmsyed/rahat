# Story 004: Deliver Telegram reminders and adaptive check-ins

**Status:** not-started  
**Type:** integration  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

---

## Goal
Make Telegram the primary interactive v1 surface by sending batched daily lists, window-based reminders, and adaptive follow-ups with inline actions for Done, Not Yet, snooze, reschedule, and pause controls. This story should implement the guilt-reducing check-in loop described in the PRD without introducing SMS or a dashboard dependency.

## Verification
In a dev or staging Telegram chat, a tester can receive today’s list, respond through inline buttons, and see occurrence state update correctly through completion, consecutive No responses, snoozes, reschedules, and pause actions.

## Scope — files this story may touch
- internal/notifications/telegram/
- internal/webhooks/telegram/
- internal/checkins/
- internal/occurrences/
- internal/events/
- cmd/server/

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
- [ ] Register Telegram webhook handling and secure bot configuration
- [ ] Send one batched morning message plus per-window task or subtask reminders from scheduled occurrences
- [ ] Capture Done and Not Yet responses and advance occurrence state correctly
- [ ] Implement consecutive-No check-ins with snooze, reschedule, and skip limits from the PRD
- [ ] Add pause-everything-today, pause-this-week, and per-task pause actions in Telegram
- [ ] Log outbound messages and user responses for later retention and tone analysis

---

## Issues

---

## Completion Summary
