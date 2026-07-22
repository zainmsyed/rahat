# Story 005: Sync Google Calendar and apply schedule constraints

**Status:** complete  
**Type:** integration  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-16  
**Completed:** 2026-07-16

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
- [x] Add Google OAuth and token storage using the read-only calendar scope
- [x] Sync calendar events into local blocks keyed by user-local date and timezone
- [x] Classify events as small, medium, or large using the PRD rules
- [x] Apply calendar blocking rules inside the scheduling engine, including large-day small-task filtering
- [x] Surface human-readable blocked-window reasons for Telegram and web consumers
- [x] Add tests for timezone handling, all-day events, and medium-window blocking behavior

---

## Issues

---

## Completion Summary
- Added Google Calendar OAuth plumbing with read-only scope support, token persistence, and sync endpoints in `cmd/server` for auth URL generation, connection completion, and day sync.
- Added calendar connection and calendar block persistence, including transactional daily block replacement and a migration for Google token plus block storage.
- Synced Google events into local per-day blocks using the user timezone, with classification into small, medium, and large blocks plus all-day handling.
- Updated the scheduler to apply calendar constraints by zeroing blocked window budgets for medium or window-sized large events, preserving a large-day small-task-only mode, and surfacing blocked-window reasons in the plan result.
- Added a schedule planning endpoint so downstream Telegram or web consumers can read blocked-window explanations from the generated plan.
- Added tests covering timezone conversion, all-day event storage, medium-window blocking, and large-day small-task filtering, and verified the repository with `go test ./...`.
