# Story 005: Sync Google Calendar and apply schedule constraints

**Status:** not-started  
**Type:** integration  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

---

## Goal
Connect Rahat to Google Calendar with read-only access, sync daily event blocks, classify them by size, and feed those constraints into the scheduler and user-facing explanations. This story makes schedules adapt around real commitments while keeping the conservative v1 blocking rules from the PRD.

## Verification
A connected calendar with representative events changes the generated schedule as expected, and downstream Telegram or web surfaces can explain why a window or day was limited by calendar commitments.

## Scope — files this story may touch
- internal/calendar/
- internal/calendar/google/
- internal/scheduler/
- internal/store/
- db/migrations/
- cmd/server/

## Out of scope — do not touch
- Calendar write-back or edits
- Non-Google calendar providers
- Fine-grained subdivision inside medium blocked windows

## Dependencies
- Story 001
- Story 002
- Story 003

---

## Checklist
- [ ] Add Google OAuth and token storage using the read-only calendar scope
- [ ] Sync calendar events into local blocks keyed by user-local date and timezone
- [ ] Classify events as small, medium, or large using the PRD rules
- [ ] Apply calendar blocking rules inside the scheduling engine, including large-day small-task filtering
- [ ] Surface human-readable blocked-window reasons for Telegram and web consumers
- [ ] Add tests for timezone handling, all-day events, and medium-window blocking behavior

---

## Issues

---

## Completion Summary
