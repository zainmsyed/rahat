# Rahat — Plan

**Created:** 2026-07-07  
**Last updated:** 2026-07-21

---

## What we're building
Rahat is a greenfield single-user assistant that schedules recurring household tasks around time budgets, calendar constraints, and stated priorities for overwhelmed parents, starting with new mothers and mothers with multiple young children. v1 uses a Go + SQLite backend, Telegram for the interactive daily loop with long polling as the default transport and webhook mode as an optional domain-backed upgrade, email for overview/recap only, read-only Google Calendar integration, and a lightweight SvelteKit web flow for onboarding plus a passive today/tomorrow lookahead page.

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
Make Telegram the primary interactive surface for daily lists, reminders, check-ins, snoozes, reschedules, and pause actions, with long polling as the default v1 delivery path for easier development and testing, while keeping email as a non-interactive overview/recap channel. When deployment settings include a suitable public domain, Telegram webhook mode can be enabled; otherwise the app should continue on long polling. Implemented by Stories 004 and 010.

### Feature 4: Minimal web surfaces
Provide a guided, hand-holding onboarding flow (core, Telegram, and Google Calendar) plus a passive today/tomorrow lookahead page without introducing a full dashboard. Implemented by Stories 006, 007, 008, and 009.

### Feature 5: Launch readiness
Add job wiring, telemetry, backups, and deployment/runbook support so the small testing group can use Rahat consistently over several weeks. Implemented by Story 011.

## Story queue
| Story | Title | Status | Blocks |
|---|---|---|---|
| 001 | Bootstrap the Rahat app workspace | not-started | — |
| 002 | Model tasks, occurrences, channels, and event history | not-started | 001 |
| 003 | Build the daily scheduling engine | not-started | 001, 002 |
| 004 | Deliver Telegram reminders and adaptive check-ins | not-started | 001, 002, 003 |
| 005 | Sync Google Calendar and apply schedule constraints | not-started | 001, 002, 003 |
| 006 | Build the guided core onboarding flow | not-started | 001, 002, 003 |
| 007 | Build the guided Telegram connection onboarding | not-started | 004, 006 |
| 008 | Build the guided Google Calendar connection onboarding | not-started | 005, 006 |
| 009 | Add the read-only today/tomorrow lookahead page | not-started | 003, 005, 006 |
| 010 | Send email overview/recap digests | not-started | 003, 006, 009 |
| 011 | Add ops, telemetry, backups, and launch tooling | not-started | 001–010 |

## Replanning log
- 2026-07-07: Initial plan created from the PRD plus clarified scope decisions: greenfield repo, Telegram as the interactive loop, email as recap-only, onboarding and read-only web view retained, SMS removed from v1.
- 2026-07-12: Reframed Telegram transport expectations so long polling is the default v1 path for development, testing, and early deployment simplicity; webhooks are optional and should be enabled only for domain-backed deployments, otherwise the app stays on long polling.
- 2026-07-21: Retired the original Story 006 after a live UX review showed the onboarding flow was too operator-centric for the actual beta testers (non-technical new mothers). Its one-shot implementation was rolled back to the pre-Story-006 checkout, and onboarding was replanned as three guided stories: core profile/tasks flow with step-by-step hand-holding, Telegram connection via short code and deep link (no raw chat IDs), and Google Calendar connection with honest optional/unavailable states.
- 2026-07-21: Renumbered the queue so the onboarding sequence takes 006–008 and the remaining stories shift up: 006 guided core onboarding, 007 guided Telegram onboarding, 008 guided Google Calendar onboarding, 009 lookahead page (was 007), 010 email digests (was 008), 011 ops/launch tooling (was 009). The retired original Story 006 is preserved in `.context/stories/archive/story-006.md`. Also decided v1 uses a single shared Telegram bot operated by the project owner: testers never create bots, each tester gets a private 1:1 chat with the shared bot, and per-tester link codes bind each chat to the right profile.
