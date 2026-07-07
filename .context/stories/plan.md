# Rahat — Plan

**Created:** 2026-07-07  
**Last updated:** 2026-07-07

---

## What we're building
Rahat is a greenfield single-user assistant that schedules recurring household tasks around time budgets, calendar constraints, and stated priorities for overwhelmed parents, starting with new mothers and mothers with multiple young children. v1 uses a Go + SQLite backend, Telegram for the interactive daily loop, email for overview/recap only, read-only Google Calendar integration, and a lightweight SvelteKit web flow for onboarding plus a passive today/tomorrow lookahead page.

## What we're not building (v1 scope)
- Multi-user household assignment
- SMS / Twilio delivery
- Push notifications or a PWA
- Calendar write-back
- Yearly cadence tasks
- A full interactive dashboard for day-to-day task management
- Machine-learned personalization beyond the PRD rules

## Features
### Feature 1: Platform and domain foundation
Establish the greenfield repo structure, runtime skeleton, SQLite persistence, and core domain models for users, tasks, subtasks, occurrences, channel preferences, pauses, and event history. Implemented by Stories 001 and 002.

### Feature 2: Scheduling and calendar-aware task planning
Implement the daily scheduling engine, cadence generation, window budgeting, overflow behavior, rollover caps, skip semantics, and Google Calendar read-only constraints. Implemented by Stories 003 and 005.

### Feature 3: Adaptive messaging loop
Make Telegram the primary interactive surface for daily lists, reminders, check-ins, snoozes, reschedules, and pause actions, while keeping email as a non-interactive overview/recap channel. Implemented by Stories 004 and 008.

### Feature 4: Minimal web surfaces
Provide a lightweight onboarding flow plus a passive today/tomorrow lookahead page without introducing a full dashboard. Implemented by Stories 006 and 007.

### Feature 5: Launch readiness
Add job wiring, telemetry, backups, and deployment/runbook support so the small testing group can use Rahat consistently over several weeks. Implemented by Story 009.

## Story queue
| Story | Title | Status | Blocks |
|---|---|---|---|
| 001 | Bootstrap the Rahat app workspace | not-started | — |
| 002 | Model tasks, occurrences, channels, and event history | not-started | 001 |
| 003 | Build the daily scheduling engine | not-started | 001, 002 |
| 004 | Deliver Telegram reminders and adaptive check-ins | not-started | 001, 002, 003 |
| 005 | Sync Google Calendar and apply schedule constraints | not-started | 001, 002, 003 |
| 006 | Build the onboarding web flow | not-started | 001, 002, 005 |
| 007 | Add the read-only today/tomorrow lookahead page | not-started | 003, 005, 006 |
| 008 | Send email overview/recap digests | not-started | 003, 006, 007 |
| 009 | Add ops, telemetry, backups, and launch tooling | not-started | 001–008 |

## Replanning log
- 2026-07-07: Initial plan created from the PRD plus clarified scope decisions: greenfield repo, Telegram as the interactive loop, email as recap-only, onboarding and read-only web view retained, SMS removed from v1.
