# Story 009: Add ops, telemetry, backups, and launch tooling

**Status:** not-started  
**Type:** ops  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

---

## Goal
Make Rahat operable for the small testing group by wiring scheduled jobs, delivery observability, backups, and deployment or reset tooling needed for several weeks of real use on the planned infrastructure. This story turns the feature set from a codebase into something the team can dogfood, monitor, and recover safely.

## Verification
Scheduled jobs can run automatically, delivery and response events are queryable, backups execute on a schedule, and a tester environment can be provisioned or reset without manual database surgery.

## Scope — files this story may touch
- internal/jobs/
- internal/events/
- scripts/
- deploy/
- README.md
- cmd/server/
- db/

## Out of scope — do not touch
- Paid analytics tooling
- A complex admin dashboard
- Production multi-tenant scale work

## Dependencies
- Story 001
- Story 002
- Story 003
- Story 004
- Story 005
- Story 006
- Story 007
- Story 008

---

## Checklist
- [ ] Wire scheduled jobs for daily schedule generation, notification dispatch, recap sending, and calendar sync
- [ ] Add lightweight reporting queries or exports for message_sent, message_type, and user_response events
- [ ] Automate daily SQLite backups to a configurable object-storage target
- [ ] Add seed or bootstrap tooling for onboarding testers and resetting non-production environments safely
- [ ] Document Coolify or Hetzner deployment, secret management, and webhook setup for Telegram and Google
- [ ] Create a launch smoke-check list covering scheduling, Telegram, email recap, calendar sync, and the read-only view

---

## Issues

---

## Completion Summary
