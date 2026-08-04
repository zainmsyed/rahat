# Story 039: Add a read-only weekly schedule preview

**Status:** in-progress  
**Type:** feature/hardening  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-04  
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
- [x] Add an authenticated or token-scoped weekly preview endpoint based on `PreviewRange` rather than repeated `PlanDay` calls.
- [x] Include clear JSON response fields for dates, windows, overflow, skips, reasons, and window budgets.
- [x] Add tests that count-cadence weekday/weekend tasks appear on the expected number and type of days.
- [x] Add tests that weekly previews do not persist occurrences or checkpoints.
- [x] Leave the existing two-day lookahead page unchanged because product confirmation for a weekly UI extension was not obtained.
- [x] Update smoke-check documentation for weekly preview verification.

---

## Issues

- Manual backend verification on 2026-08-03 used repeated production `PlanDay` calls to inspect a week, which incorrectly treated weekly-count Laundry as due repeatedly and persisted duplicates. A true weekly preview must be read-only and state-carrying.
- Product confirmation was not obtained for extending the existing two-day lookahead UI, so this story adds the weekly backend contract and documentation without changing the visible page range.

---

## Completion Summary

Implemented a token-scoped weekly preview through the existing `PreviewRange` state-carrying scheduler path. `GET /lookahead/plan?token=<token>&days=7` now returns seven day records while preserving the existing two-day default. Responses include range length, dates/labels, window schedules, explicit overflowed and skipped items, reasons, blocked windows, and per-window budgets.

The preview now advances calendar days with `AddDate`, carries planned preview commitments between days, and counts them toward weekly cadence without persisting occurrences or checkpoints. Added coverage for weekday/weekend count cadence, repeated-preview shape stability, zero persistence, unsupported ranges, and existing lookahead behavior. The current lookahead UI remains today/tomorrow-only pending product confirmation; the API and smoke documentation support weekly inspection.

Verification: `go test ./... -count=1`, `npm run check`, and the full frontend suite (24 files / 87 tests) pass. Story 039 is ready for `/complete-story`.