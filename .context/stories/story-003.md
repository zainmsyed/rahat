# Story 003: Build the daily scheduling engine

**Status:** not-started  
**Type:** backend  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

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
- [ ] Generate due occurrences from interval and weekly-count cadences using last completion data
- [ ] Split each day’s task-time budget across morning, afternoon, and evening before fitting work
- [ ] Place tasks and subtasks by priority, overdue order, time-of-day preference, and min-gap rules
- [ ] Push overflow forward while tracking rollover counts, high-priority exceptions, and skip semantics
- [ ] Persist the resulting daily schedule and next-checkpoint state for delivery channels to read
- [ ] Add table-driven tests for normal days, overloaded days, and multistep examples like laundry

---

## Issues

---

## Completion Summary
