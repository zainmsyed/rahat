# Story 011: Add ops, telemetry, backups, and launch tooling

**Status:** in-progress  
**Type:** ops  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-25  
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
- Story 009
- Story 010

---

## Checklist
- [x] Wire scheduled jobs for daily schedule generation, notification dispatch, recap sending, and calendar sync
- [x] Add lightweight reporting queries or exports for message_sent, message_type, and user_response events
- [x] Automate daily SQLite backups to a configurable object-storage target
- [x] Add seed or bootstrap tooling for onboarding testers and resetting non-production environments safely
- [x] Document Coolify or Hetzner deployment, secret management, and webhook setup for Telegram and Google
- [x] Create a launch smoke-check list covering scheduling, Telegram, email recap, calendar sync, and the read-only view

---

## Issues

---

## Completion Summary

Story 011 added an operator-focused ops layer around the existing Rahat app. The server now has CLI-driven ops commands for running named jobs, exporting event summaries/CSV, creating backups, seeding demo testers, and resetting non-production environments. Under `internal/jobs/`, jobs are registered by name so platform cron or operator scripts can run daily schedule generation, Telegram daily/window dispatch, recap generation, calendar sync, and backups without building an admin dashboard.

Delivery observability now has lightweight reporting through `internal/events` summary/filter helpers, exposed through the `ops:report-events` command and `scripts/report-events.sh`, so operators can inspect `message_sent`, `message_type`, and `user_response` activity. Backups are automated through `ops:backup` / `scripts/run-backup.sh`, producing gzipped SQLite snapshots to a configurable local/file target or an `s3://` target via the AWS CLI.

Tester provisioning no longer requires manual DB edits: `scripts/bootstrap-testers.sh` seeds demo users, starter tasks, and recap preferences, while `scripts/reset-nonprod.sh` safely wipes a non-production database only when `RAHAT_RESET_CONFIRM=reset-non-production` is explicitly set. Deployment and launch guidance now lives in `deploy/README.md` and `deploy/launch-smoke-checklist.md`, and the root `README.md` documents the new ops env vars and script entry points.

Verification completed with `go test ./...`, `cd web && npm run check`, and `cd web && npm test`. Operationally, the main caveat is that recap generation currently writes file-based outbox artifacts plus delivery events rather than sending SMTP mail directly, which keeps launch tooling usable while the broader auth/email rollout remains deferred.
