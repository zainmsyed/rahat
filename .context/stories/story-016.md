# Story 016: Make the daily scheduler fit realistic task combinations

**Status:** complete  
**Type:** feature / bug-fix  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-25  
**Completed:** 2026-07-25

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
- [x] Diagnose the exact budget-split logic that causes feasible combinations to fail
- [x] Implement a scheduler strategy that fits tasks when total duration <= daily budget, while still respecting windows and preferences
- [x] Add tests for the scenarios above
- [x] Ensure existing scheduler tests still pass
- [x] Update this story with the fix and any migration notes

## Issues

## Completion Summary

Implemented two scheduler changes in `internal/scheduler/service.go`:

1. **Demand-aware window budgeting** (`splitWindowBudgets`): when the total demand of the day's candidates fits within the daily budget, each window is allocated exactly its demand instead of a proportional split. Leftover minutes are distributed to windows not blocked by the calendar. This lets feasible multi-window combinations like Laundry + Clean kitchen (45 min in a 60-min day) schedule successfully.

2. **Fallback window placement** (`tryFitCandidate`): when a task cannot fit its preferred window, the scheduler now tries other windows in order before overflowing, as long as the total day budget allows it and subtask ordering is respected.

Added `TestSchedulerFitsRealisticCombinations` in `internal/scheduler/service_test.go` covering:
- Laundry + Clean kitchen in a 60-minute day schedules all four items.
- A 60-minute Grocery run alone schedules.
- Meal prep + Clean kitchen honestly overflows one item because their combined 65 minutes exceeds the 60-minute budget.

All existing scheduler tests and the full `go test ./...` / `npm test` suites pass. No new migrations were needed.
