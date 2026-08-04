# Story 038: Make daily planning idempotent

**Status:** complete  
**Type:** hardening  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-04  
**Completed:** 2026-08-04

---

## Goal
Make production daily planning safe to rerun. Planning the same day twice must reuse or update the existing open occurrences instead of inserting duplicates, while preserving completed, skipped, and historical records.

## Verification
Calling the production daily planner repeatedly for the same user and date results in one open occurrence set per task/subtask/day. Replanning a later date after earlier days were planned does not duplicate earlier work. Completed and skipped occurrence history remains untouched.

## Scope — files this story may touch
- db/migrations/014_story_038_occurrence_idempotency.sql (new)
- internal/occurrences/types.go
- internal/occurrences/repository.go
- internal/occurrences/service.go
- internal/scheduler/service.go
- internal/scheduler/service_test.go
- internal/store/story002_integration_test.go
- internal/store/story014_integration_test.go
- internal/store/story037_integration_test.go (new, if needed)
- cmd/server/schedule_test.go

## Out of scope — do not touch
- Weekly preview behavior
- Telegram notification content
- Frontend visual design
- Task creation/edit fields
- Calendar sync behavior

## Dependencies
- Story 003
- Story 013
- Story 037

---

## Checklist
- [x] Reproduce the duplicate occurrence behavior with a regression test for planning the same day twice.
- [x] Define the stable open-occurrence identity using user, task, subtask, original scheduled date, and non-terminal status.
- [x] Add migration and repository support for unique open-occurrence identity or an equivalent safe upsert.
- [x] Update scheduler persistence so scheduled, overflow, and skipped writes do not duplicate existing open records.
- [x] Add tests covering reruns, prior-day backlog, overflow, and historical-record preservation.
- [x] Run backend scheduler/task/server tests through the available Go container and document results.

---

## Issues

- End-to-end verification on 2026-08-03 showed repeated single-day schedule calls created duplicate Laundry occurrences on the following Monday. This confirms production day planning is not fully idempotent.
- The pre-existing Story 014 migration fixture did not create the tables required by later migrations. Its minimal bootstrap was expanded within this story's scope so the full migration chain can be tested.
- `gofmt -l` still reports the pre-existing formatting issue in `cmd/server/schedule_test.go`; it was not changed because it was unrelated to this story's implementation.

---

## Completion Summary

Implemented open-occurrence idempotency for production planning. Occurrences now use the stable identity `(user, task, subtask, original scheduled date)` for non-terminal `pending` and `scheduled` rows. Migration 014 removes pre-existing duplicate open rows deterministically and adds a partial unique index, while repository/service save behavior reuses the existing open row before inserting and handles concurrent insert races.

Scheduler persistence now uses the idempotent save path for scheduled, overflow, and skipped results. Added regression coverage for same-day reruns, later-day backlog reuse, and preservation of completed history. Updated the Story 014 migration fixture to include the tables needed by the complete migration chain.

Verification: the full backend suite passes in the Go 1.25 container with `go test ./... -count=1`. Story 038 is implementation-ready for `/complete-story`; final closeout should separately decide whether to address the unrelated pre-existing `cmd/server/schedule_test.go` formatting warning.