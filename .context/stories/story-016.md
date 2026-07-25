# Story 016: Make the daily scheduler fit realistic task combinations

**Status:** not-started  
**Type:** feature / bug-fix  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-25  
**Completed:** —

---

## Goal
Fix the daily scheduler so that when a set of tasks clearly fits within the user’s daily time budget, it actually gets scheduled instead of being pushed forward because of an artificial per-window budget split.

## Context
While testing Story 015, a user with a 60-minute daily budget and these routines:

- Laundry — 25 min, 2×/week (Wash morning, Move afternoon, Fold evening)
- Clean kitchen — 20 min, every day, evening
- Grocery run — 60 min, 1×/week, afternoon
- Meal prep — 45 min, 2×/week, afternoon

saw the first day push **all 6 sub-tasks/occurrences to the next day**, and the next day push them all forward again with **0 items scheduled**.

Laundry (25 min) + Clean kitchen (20 min) = 45 min, which fits easily in a 60-minute day. The current scheduler splits the 60-minute budget across windows proportionally by demand, leaving morning with almost no room, so the laundry group fails as a unit and everything else cascades into overflow.

## Verification
- A 60-minute day with Laundry + Clean kitchen schedules both items.
- A day with Grocery run (60 min) alone schedules it.
- A day with Meal prep (45 min) + Clean kitchen (20 min) either schedules both or honestly overflows only what truly does not fit.
- The scheduler still respects time-of-day preferences when possible, but falls back to other windows when the preferred window is full and the total day budget allows it.
- Overflow and skip counts are still accurate and honest.
- Existing scheduler tests continue to pass; new tests cover realistic multi-window combinations.

## Scope — files this story may touch
- `internal/scheduler/service.go`
- `internal/scheduler/service_test.go`
- `internal/scheduler/types.go`
- `internal/time/windows.go`
- `db/migrations/` (if new persisted state is needed)

## Out of scope — do not touch
- Onboarding confirmation message text
- Telegram service
- Frontend onboarding screens
- Event logging

## Dependencies
- Story 015

## Checklist
- [ ] Diagnose the exact budget-split logic that causes feasible combinations to fail
- [ ] Implement a scheduler strategy that fits tasks when total duration <= daily budget, while still respecting windows and preferences
- [ ] Add tests for the scenarios above
- [ ] Ensure existing scheduler tests still pass
- [ ] Update this story with the fix and any migration notes

## Issues

## Completion Summary
