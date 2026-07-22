# Story 002: Model tasks, occurrences, channels, and event history

**Status:** complete  
**Type:** backend  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** 2026-07-07

---

## Goal
Define the single-user v1 data model in SQLite and Go so Rahat can persist user profile settings, tasks, subtasks, occurrences, channel preferences, pause state, and lightweight telemetry. This story provides the durable foundation that later scheduling, Telegram, calendar, onboarding, and email stories all build on.

## Verification
Automated tests or fixtures can create a sample user, task, multistep task, occurrence, channel preference, pause state, and event log row, then read them back through the repository layer without manual database editing.

## Scope — files this story may touch
- db/migrations/
- internal/store/
- internal/users/
- internal/tasks/
- internal/occurrences/
- internal/events/
- internal/notifications/preferences/

## Out of scope — do not touch
- Daily schedule generation rules
- Telegram or email delivery behavior
- Web onboarding UI

## Dependencies
- Story 001

---

## Checklist
- [x] Create SQLite migrations for users, tasks, subtasks, occurrences, channel preferences, pauses, and event logs
- [x] Model cadence types, priorities, time-of-day preferences, and subtask gap rules in Go domain types
- [x] Add repositories and services for CRUD and lookup of tasks, subtasks, and occurrences
- [x] Store global pause states and per-task pause flags needed by later scheduler and messaging stories
- [x] Seed a starter task library for the new-parent defaults from the PRD
- [x] Cover schema and repository behavior with repeatable tests or fixtures

---

## Issues

---

## Completion Summary
- Added SQL migrations for the v1 persistence model: users, tasks, subtasks, occurrences, channel preferences, pauses, event logs, and starter task template tables.
- Added a migration runner in `internal/store` plus shared ID/time helpers for repository code.
- Implemented Go domain types for cadence, priority, time-of-day preferences, and subtask gap rules.
- Added repository and service layers for users, tasks, occurrences, event logs, and notification preferences/pause records.
- Seeded the PRD starter task library, including the multi-step laundry workflow and other new-parent defaults.
- Added an integration test that applies migrations, creates sample records across the repository layer, verifies pause state and starter templates, and confirms readback behavior.
