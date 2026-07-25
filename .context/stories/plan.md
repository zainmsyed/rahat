# Rahat — Plan

**Created:** 2026-07-07  
**Last updated:** 2026-07-25

---

## What we're building
Rahat is a greenfield per-user assistant that schedules recurring household tasks around time budgets, calendar constraints, and stated priorities for overwhelmed parents, starting with new mothers and mothers with multiple young children. Multiple beta testers have isolated accounts, while household sharing remains out of scope. v1 uses a Go + SQLite backend, Telegram for the interactive daily loop and self-service web sign-in recovery, long polling as the default transport with webhook mode as an optional domain-backed upgrade, read-only Google Calendar integration, and lightweight SvelteKit web surfaces for onboarding, authenticated routine maintenance, and a passive today/tomorrow lookahead page. Email recap delivery remains deferred.

## What we're not building (v1 scope)
- Multi-user household assignment
- SMS / Twilio delivery
- Push notifications or a PWA
- Calendar write-back
- Yearly cadence tasks
- A full interactive scheduling dashboard or direct manipulation of generated occurrences
- Machine-learned personalization beyond the PRD rules

## Features
### Feature 1: Platform and domain foundation
Establish the greenfield repo structure, runtime skeleton, SQLite persistence, and core domain models for users, tasks, subtasks, occurrences, channel preferences, pauses, and event history. Implemented by Stories 001 and 002.

### Feature 2: Scheduling and calendar-aware task planning
Implement the daily scheduling engine, cadence generation, window budgeting, overflow behavior, rollover caps, skip semantics, and Google Calendar read-only constraints. Implemented by Stories 003 and 005.

### Feature 3: Adaptive messaging loop
Make Telegram the primary interactive surface for daily lists, reminders, check-ins, snoozes, reschedules, pause actions, onboarding confirmation, and self-service return to authenticated web settings. Long polling remains the default v1 delivery path; webhook mode is enabled only when deployment settings clearly support it. Email overview/recap delivery remains deferred. Implemented by Story 004 and extended by Stories 014–015; the original email-digest Story 010 has been retired pending replanning.

### Feature 4: Minimal web surfaces
Provide a guided, hand-holding onboarding flow (core, Telegram, and Google Calendar), a passive today/tomorrow lookahead page, and focused post-onboarding routine maintenance without introducing a full scheduling dashboard. Onboarding and lookahead are implemented by Stories 006–009; durable beta sessions and routine maintenance are planned in Stories 012 and 013.

### Feature 5: Launch readiness
Add job wiring, telemetry, backups, and deployment/runbook support so the small testing group can use Rahat consistently over several weeks. Implemented by Story 011.

### Feature 6: Durable beta account access
Use revocable browser sessions as the authorization boundary for each isolated beta account. Story 012 implemented durable sessions and operator-issued recovery links as an initial bridge. Story 014 replaces operator dependence in the normal return flow: a tester sends `/edit` from their uniquely linked Telegram chat and receives a fresh short-lived, single-use web access link bound to their backend user. Passwords, email delivery, and permanent bearer links remain out of scope.

### Feature 7: Onboarding outcome confirmation
After onboarding persists routines and generates the first schedule, confirm the outcome both on screen and in the tester's linked Telegram chat. The Telegram message should summarize saved routines and actual scheduled windows, report overflow honestly, and explain the `/edit` return path. Planned in Story 015.

## Story queue
| Story | Title | Status | Blocks |
|---|---|---|---|
| 001 | Bootstrap the Rahat app workspace | complete | — |
| 002 | Model tasks, occurrences, channels, and event history | complete | 001 |
| 003 | Build the daily scheduling engine | complete | 001, 002 |
| 004 | Deliver Telegram reminders and adaptive check-ins | complete | 001, 002, 003 |
| 005 | Sync Google Calendar and apply schedule constraints | complete | 001, 002, 003 |
| 006 | Build the guided core onboarding flow | complete | 001, 002, 003 |
| 007 | Build the guided Telegram connection onboarding | complete | 004, 006 |
| 008 | Build the guided Google Calendar connection onboarding | complete | 005, 006 |
| 009 | Add the read-only today/tomorrow lookahead page | complete | 003, 005, 006 |
| 010 | Send email overview/recap digests | retired | 003, 006, 009 |
| 011 | Add ops, telemetry, backups, and launch tooling | complete | 001–009 |
| 012 | Establish durable beta web sessions | complete | 006, 007, 011 |
| 013 | Add post-onboarding task management | in-progress | 003, 006, 009, 012 |
| 014 | Enable self-service web access through Telegram | not-started | 004, 007, 012, 013 |
| 015 | Confirm onboarding completion in Telegram | not-started | 003, 004, 006, 007, 014 |

## Replanning log
- 2026-07-07: Initial plan created from the PRD plus clarified scope decisions: greenfield repo, Telegram as the interactive loop, email as recap-only, onboarding and read-only web view retained, SMS removed from v1.
- 2026-07-12: Reframed Telegram transport expectations so long polling is the default v1 path for development, testing, and early deployment simplicity; webhooks are optional and should be enabled only for domain-backed deployments, otherwise the app stays on long polling.
- 2026-07-21: Retired the original Story 006 after a live UX review showed the onboarding flow was too operator-centric for the actual beta testers (non-technical new mothers). Its one-shot implementation was rolled back to the pre-Story-006 checkout, and onboarding was replanned as three guided stories: core profile/tasks flow with step-by-step hand-holding, Telegram connection via short code and deep link (no raw chat IDs), and Google Calendar connection with honest optional/unavailable states.
- 2026-07-21: Renumbered the queue so the onboarding sequence takes 006–008 and the remaining stories shift up: 006 guided core onboarding, 007 guided Telegram onboarding, 008 guided Google Calendar onboarding, 009 lookahead page (was 007), 010 email digests (was 008), 011 ops/launch tooling (was 009). The retired original Story 006 is preserved in `.context/stories/archive/story-006.md`. Also decided v1 uses a single shared Telegram bot operated by the project owner: testers never create bots, each tester gets a private 1:1 chat with the shared bot, and per-tester link codes bind each chat to the right profile.
- 2026-07-25: Marked Stories 001–009 complete in the plan. Story 009’s read-only lookahead page is implemented and closed at the story level, but broader rollout is intentionally paused until better auth/OAuth and email delivery work is in place; that dependency is expected to be addressed by the remaining auth/digest work rather than by reopening Story 009.
- 2026-07-25: Retired Story 010 without implementation. Email overview/recap digests are intentionally deferred until Rahat has better auth/OAuth groundwork and a clearer delivery/rollout path; if the feature returns, it should be replanned as a new future story rather than resumed from the old scoped digest story.
- 2026-07-25: Added Stories 012 and 013 after identifying that testers cannot maintain routines once onboarding ends. Story 012 first establishes durable, revocable beta browser sessions and operator-issued single-use access links without depending on deferred email delivery. Story 013 then adds an authenticated, focused task-management page that reuses the onboarding editor and preserves completed history when routines are removed. This remains intentionally smaller than a full scheduling dashboard.
- 2026-07-25: Replanned returning-user access after recognizing that operator-issued links do not provide a viable multi-tester recovery flow. Story 014 makes the uniquely linked Telegram chat the trusted self-service entry point: `/edit` resolves the backend user and issues a fresh short-lived, single-use link into the existing durable session system. Permanent invite/access links and frontend-supplied user identity are explicitly rejected.
- 2026-07-25: Added Story 015 so onboarding has a visible outcome in the primary product channel. Successful completion will send one idempotent Telegram summary based on persisted routines and the actual first schedule, including honest overflow/skipped feedback and the `/edit` return instruction; on-screen completion remains available and Telegram delivery failure must not roll back saved onboarding data.
