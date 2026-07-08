# Story 003: Build the daily scheduling engine

**Status:** complete  
**Type:** backend  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-08  
**Completed:** 2026-07-08

---

## Goal
Implement Rahat’s daily scheduling engine so it can generate a sane, realistic day plan from recurring tasks, priorities, time windows, daily budget, and multistep task rules. This story owns occurrence generation, window budgeting, overflow handling, rollover caps, skip semantics, and the persisted schedule state that downstream delivery channels will consume.

## Verification
Deterministic tests show that sample task sets produce the expected daily schedule, rollover behavior, and skip outcomes across normal days, overloaded days, and multistep cases like laundry.

## Scope — files this story may touch
- internal/scheduler/
- internal/occurrences/
- internal/tasks/
- internal/time/
- internal/store/
- db/migrations/
- testdata/

## Out of scope — do not touch
- Telegram, email, or web message formatting
- Google Calendar OAuth and sync
- Web onboarding forms

## Dependencies
- Story 001
- Story 002

---

## Checklist
- [x] Generate due occurrences from interval and weekly-count cadences using last completion data
- [x] Split each day’s task-time budget across morning, afternoon, and evening before fitting work
- [x] Place tasks and subtasks by priority, overdue order, time-of-day preference, and min-gap rules
- [x] Push overflow forward while tracking rollover counts, high-priority exceptions, and skip semantics
- [x] Persist the resulting daily schedule and next-checkpoint state for delivery channels to read
- [x] Add table-driven tests for normal days, overloaded days, and multistep examples like laundry

---

## Issues

---

## Completion Summary
- Added the scheduling engine in `internal/scheduler` to generate due task occurrences for interval and weekly-count cadences, merge backlog work, and produce a persisted day plan.
- Added window helpers and schedule-checkpoint persistence so each generated day stores its next checkpoint time and scheduled occurrence count for downstream delivery code.
- Implemented budget splitting across morning, afternoon, and evening, plus window-based fitting by priority, overdue age, and multistep subtask order.
- Added rollover handling that forwards overflowed work, auto-skips non-high-priority occurrences once the rollover cap is hit, and preserves high-priority exceptions.
- Extended task and occurrence services with list helpers needed by the scheduler and added a migration for schedule checkpoint storage.
- Added table-driven scheduler tests covering a normal day, an overloaded day with rollover/skip behavior, and a multistep laundry plan spanning morning, afternoon, and evening.
