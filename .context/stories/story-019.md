# Story 019: Calendar-aware day selection and load balancing

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-26  
**Completed:** 2026-07-26

---

## Goal
Use synced Google Calendar blocks not only to block windows, but also to choose better days and windows for recurring tasks, so heavier chores land on lighter days and the schedule respects the user's real commitments.

## Context
Story 005 added calendar sync and constraint application. The scheduler currently zeroes window budgets for blocked windows and filters to small tasks on large-commitment days. That prevents scheduling into busy time, but it does not help the scheduler spread work across the week or prefer less-busy days.

For the target user, Rahat should look ahead and place a 60-minute grocery run on a day with no afternoon commitments, rather than repeatedly trying and failing to fit it into an already-busy afternoon.

## Verification
- With calendar events on some days, the scheduler prefers scheduling recurring tasks on days with more available budget.
- Large calendar days still restrict to small tasks, but the scheduler can move larger tasks to the next suitable day instead of leaving them as overflow.
- The schedule explanation can tell the user why a task landed on a particular day (e.g., "Your calendar was full Friday, so I moved it to Saturday").
- Existing calendar blocking behavior (Story 005) continues to work.

## Scope — files this story may touch
- `internal/scheduler/service.go`
- `internal/scheduler/types.go`
- `internal/store/calendar_state.go`
- `internal/calendar/service.go`
- `internal/scheduler/service_test.go`

## Out of scope — do not touch
- Calendar OAuth or sync plumbing (Story 005)
- Task distribution logic independent of calendar (Story 017)
- Timezone-aware windows (Story 018)

## Dependencies
- Story 005
- Story 017
- Story 018

## Checklist
- [x] Add tests that show the scheduler preferring less-busy days for heavier tasks
- [x] Implement calendar-aware day selection within the planning horizon
- [x] Surface day-selection reasons in the plan result or schedule explanation
- [x] Ensure large-day small-task filtering still works
- [x] Verify existing scheduler and calendar tests still pass

## Issues

## Completion Summary

Implemented calendar-aware day selection and load balancing in `internal/scheduler/service.go`:

1. **Horizon calendar load lookup**: `PlanDay` and `previewDayWithOccurrences` now load calendar constraints for the next 7 days via `loadCalendarConstraintsForRange`.

2. **Budget-score day selection**: Added `availableBudgetScore` and `windowDurationMinutes` to score each day by how much chore-window time is not blocked by calendar events. Days with small-task-only restrictions score negatively for heavy tasks, so they are avoided.

3. **Calendar-aware target dates**: `intervalTaskScheduleDates` and `countTaskScheduleDates` now use `chooseBestDates` to pick the eligible day(s) with the highest score. They keep the default cadence dates when scores are tied, so the Story 017 spread is preserved when no calendar events are present.

4. **Day-selection reasons**: Added a `Reasons map[string]string` field to `PlanResult`. When a task is moved because of a blocked calendar day, the scheduler records a reason such as "Your calendar was full on 2026-08-04, so we moved it to 2026-08-05." This can be surfaced in explanations without manual conversion hacks.

5. **Tests**: Added `TestSchedulerCalendarAwareDaySelection` with two subtests:
   - A 60-minute grocery run moves from an afternoon-blocked day to the next free day.
   - A 60-minute grocery run moves from an all-day large-event day to the next free day.

All existing scheduler and calendar-blocking tests continue to pass, and `go test ./...` / `npm test` are green. No migrations were needed.
