# Story 038: Make daily planning idempotent

**Status:** not-started  
**Type:** hardening  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-03  

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
- [ ] Reproduce the duplicate occurrence behavior with a regression test for planning the same day twice.
- [ ] Define the stable open-occurrence identity using user, task, subtask, original scheduled date, and non-terminal status.
- [ ] Add migration and repository support for unique open-occurrence identity or an equivalent safe upsert.
- [ ] Update scheduler persistence so scheduled, overflow, and skipped writes do not duplicate existing open records.
- [ ] Add tests covering reruns, prior-day backlog, overflow, and historical-record preservation.
- [ ] Run backend scheduler/task/server tests through the available Go container and document results.

---

## Issues

- End-to-end verification on 2026-08-03 showed repeated single-day schedule calls created duplicate Laundry occurrences on the following Monday. This confirms production day planning is not fully idempotent.

---

## Completion Summary
