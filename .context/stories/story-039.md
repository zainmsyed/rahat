# Story 039: Add a read-only weekly schedule preview

**Status:** not-started  
**Type:** feature/hardening  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-03  
**Completed:** not yet  

---

## Goal
Provide a backend weekly schedule preview that uses the existing `PreviewRange` state-carrying logic without persisting occurrences. Weekly inspection should show how recurring work spreads across days while leaving the database unchanged.

## Verification
A caller can request a seven-day preview and receive a day-by-day schedule where count-cadence tasks appear only their intended number of times and weekday/weekend filters are honored. Repeated preview calls produce the same shape without creating occurrences, advancing rollover counts, or changing checkpoints.

## Scope — files this story may touch
- cmd/server/main.go
- cmd/server/schedule_test.go
- cmd/server/lookahead.go
- cmd/server/lookahead_test.go
- internal/scheduler/service.go
- internal/scheduler/service_test.go
- internal/scheduler/types.go
- web/src/lib/api/lookahead.ts
- web/src/routes/lookahead/+page.svelte
- web/src/lib/components/schedule/LookaheadDay.svelte
- web/src/lib/components/schedule/LookaheadDay.test.ts
- web/src/routes/lookahead/page.test.ts
- README.md
- deploy/launch-smoke-checklist.md

## Out of scope — do not touch
- Occurrence persistence behavior (Story 038)
- Editing generated occurrences
- Telegram notification behavior
- Google Calendar sync
- A full interactive scheduling dashboard

## Dependencies
- Story 003
- Story 009
- Story 037
- Story 038

---

## Checklist
- [ ] Add an authenticated or token-scoped weekly preview endpoint based on `PreviewRange` rather than repeated `PlanDay` calls.
- [ ] Include clear JSON response fields for dates, windows, overflow, skips, reasons, and window budgets.
- [ ] Add tests that count-cadence weekday/weekend tasks appear on the expected number and type of days.
- [ ] Add tests that weekly previews do not persist occurrences or checkpoints.
- [ ] Optionally extend the existing read-only lookahead page from two days to the supported weekly range only if product confirmation is obtained.
- [ ] Update smoke-check documentation for weekly preview verification.

---

## Issues

- Manual backend verification on 2026-08-03 used repeated production `PlanDay` calls to inspect a week, which incorrectly treated weekly-count Laundry as due repeatedly and persisted duplicates. A true weekly preview must be read-only and state-carrying.

---

## Completion Summary
