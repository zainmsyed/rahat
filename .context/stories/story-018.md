# Story 018: Make the scheduler timezone-aware

**Status:** not-started  
**Type:** feature / bug-fix  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-25  
**Completed:** —

---

## Goal
Compute all schedule dates, time-of-day windows, ready times, and deadline boundaries in the user's local timezone, so that "morning" means the user's local morning and plan dates match the user's local calendar day.

## Context
The scheduler currently treats dates and window boundaries as UTC. `localDateAsUTC` returns a UTC-midnight instant for the local date, and `daytime.StartTime` uses fixed UTC hours (08:00/12:00/16:00). This causes two visible problems:

1. For users behind UTC (e.g., America/Chicago), a plan date can appear to be the previous local day when rendered.
2. Ready times rendered in the user's timezone are offset by the timezone offset, so a "morning" task can be shown as 3:00 AM.

The Story 015 review caught both issues in the Telegram confirmation message. Fixing them requires the scheduler itself to be timezone-aware, not just adjusting message formatting.

## Verification
- A user in America/Chicago gets plan dates that match their local calendar day.
- "Morning" tasks are ready at the local morning start (e.g., 8:00 AM Chicago time), not 3:00 AM.
- `PreviewDay` and `PlanDay` return consistent dates and times for users in any timezone.
- Lookahead and onboarding confirmation surfaces show the correct local date and window names without manual conversion hacks.

## Scope — files this story may touch
- `internal/scheduler/service.go`
- `internal/scheduler/types.go`
- `internal/time/windows.go`
- `internal/store/schedule_state.go`
- `cmd/server/onboarding.go` (plan date calculation)
- `cmd/server/lookahead.go` (if applicable)

## Out of scope — do not touch
- Task distribution across days (Story 017)
- Calendar block synchronization (Story 005)
- Message copy or frontend formatting beyond using the corrected schedule data

## Dependencies
- Story 003
- Story 017

## Checklist
- [ ] Identify all places where UTC is used as the scheduling anchor
- [ ] Introduce timezone-aware date and window calculations
- [ ] Add tests for users in timezones ahead of, behind, and near UTC
- [ ] Update any callers that assume UTC-midnight plan dates
- [ ] Ensure existing scheduler tests still pass

## Issues

## Completion Summary
