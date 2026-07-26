# Story 017: Spread recurring tasks across days and weeks

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-26  
**Completed:** 2026-07-26

---

## Goal
Make the daily scheduler distribute recurring tasks across the available days and weeks according to each task's cadence, priority, and duration, instead of piling every newly-due task onto the first day and rolling overflow forward indefinitely.

## Context
For the target user (a new or overwhelmed mother), the scheduler must feel like it is planning around her life, not dumping a week of chores onto one day. During Story 015 testing, a user with Laundry (2×/week), Meal prep (2×/week), Grocery run (1×/week), and Clean kitchen (daily) saw every task attempted on day 1, overflow to day 2, and then overflow again because the scheduler kept trying to fit the whole backlog into a single day.

The current engine generates all tasks for a day when they are due, tries to fit them into that day's budget, and only moves what does not fit to the next day. It does not intentionally spread tasks across multiple days even when the cadence allows it.

## Verification
- Given a 60-minute daily budget and the routines above, the scheduler places no more than fits each day and spreads remaining tasks over the following days.
- A 2×/week task is not scheduled every day; it is scheduled on separate days within the week when possible.
- A daily task is scheduled every day, but never forces other tasks to overflow if they can be placed on other days.
- Higher-priority tasks are scheduled first within a day, but the scheduler still spreads lower-priority recurring tasks across days instead of packing them together.
- Overflow counts are honest and only represent work that truly could not fit anywhere in the planning horizon.

## Scope — files this story may touch
- `internal/scheduler/service.go`
- `internal/scheduler/types.go`
- `internal/scheduler/service_test.go`
- `internal/tasks/types.go` (if cadence metadata is needed)
- `db/migrations/` (if new persisted state is needed)

## Out of scope — do not touch
- Per-window budget split logic (Story 016)
- Timezone-aware windows and dates (Story 018)
- Calendar-aware day selection (Story 019)
- Onboarding or confirmation message text

## Dependencies
- Story 003
- Story 016

## Checklist
- [x] Document the current "schedule everything due on day 1" behavior with a regression test that would have failed before the spread logic
- [x] Implement distribution of recurring tasks across days/weeks based on cadence and priority
- [x] Add tests for daily, 2×/week, and 1×/week tasks with mixed durations
- [x] Ensure overflow only occurs when no day in the horizon has room
- [x] Verify existing scheduler tests still pass

## Issues

## Completion Summary

Implemented recurring-task spreading in `internal/scheduler/service.go` by changing when tasks become schedule candidates:

1. **Horizon-aware target dates**: replaced the old "every due task is due today" check with `taskScheduleDates`, which computes target dates inside a 7-day planning horizon.
   - Interval tasks now schedule on a specific target date instead of becoming due every day after their interval passes.
   - Count-based weekly tasks now generate spaced target dates inside the week instead of defaulting to day 1.
   - Brand-new recurring tasks use a stable task-name hash to stagger their first occurrence so similar weekly chores do not all start on the same day.

2. **Cadence-aware daily prioritization**: added `taskSortRank` so daily work (for example `CadenceTypeInterval` with value `1`) is scheduled before less-frequent recurring work of the same priority. This keeps daily chores like Clean kitchen from being displaced by weekly tasks that can be deferred or honestly overflowed.

3. **Tests**:
   - Updated `TestSchedulerFitsRealisticCombinations` to force due dates explicitly with seeded history so the Story 016 regressions remain stable under the new spread logic.
   - Added `TestSchedulerSpreadsRecurringTasksAcrossWeek`, which simulates consecutive `PlanDay` runs across the week and verifies:
     - daily cleaning is scheduled each day,
     - a 2×/week laundry follow-up is placed later in the week instead of every day,
     - 1×/week grocery honestly overflows when it cannot fit anywhere alongside daily cleaning,
     - no day exceeds the 60-minute budget.

All scheduler tests, `go test ./...`, and `npm test` pass. No migrations were needed.
