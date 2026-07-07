# Story 002: Model tasks, occurrences, channels, and event history

**Status:** not-started  
**Type:** backend  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

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
- [ ] Create SQLite migrations for users, tasks, subtasks, occurrences, channel preferences, pauses, and event logs
- [ ] Model cadence types, priorities, time-of-day preferences, and subtask gap rules in Go domain types
- [ ] Add repositories and services for CRUD and lookup of tasks, subtasks, and occurrences
- [ ] Store global pause states and per-task pause flags needed by later scheduler and messaging stories
- [ ] Seed a starter task library for the new-parent defaults from the PRD
- [ ] Cover schema and repository behavior with repeatable tests or fixtures

---

## Issues

---

## Completion Summary
