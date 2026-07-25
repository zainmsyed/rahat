# Story 019: Calendar-aware day selection and load balancing

**Status:** planned  
**Type:** feature  
**Created:** 2026-07-25  
**Completed:** —

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
- [ ] Add tests that show the scheduler preferring less-busy days for heavier tasks
- [ ] Implement calendar-aware day selection within the planning horizon
- [ ] Surface day-selection reasons in the plan result or schedule explanation
- [ ] Ensure large-day small-task filtering still works
- [ ] Verify existing scheduler and calendar tests still pass

## Issues

## Completion Summary
