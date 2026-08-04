# Story 040: Prevent Telegram polling conflicts from operator jobs

**Status:** in-progress  
**Type:** bug/hardening  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-04  
**Completed:** not yet  

---

## Goal
Make CLI/operator commands safe to run alongside the long-running server. Operator jobs such as `telegram-daily`, `telegram-window`, and reports must not start a second Telegram long-polling loop, because Telegram rejects concurrent `getUpdates` consumers with `409` errors.

## Verification
Running `rahat-api ops:run-job telegram-daily`, `telegram-window`, or `ops:report-events` while the server is polling completes the requested job/report without starting long polling and without creating `409 getUpdates` conflicts in the service logs.

## Scope — files this story may touch
- cmd/server/main.go
- cmd/server/main_test.go
- cmd/server/ops.go
- cmd/server/ops_test.go
- scripts/run-job.sh
- scripts/run-telegram-daily.sh
- scripts/run-telegram-window.sh
- scripts/report-events.sh
- deploy/README.md
- deploy/launch-smoke-checklist.md

## Out of scope — do not touch
- Telegram message content or routing behavior
- Webhook transport selection
- Long-running server polling behavior
- Notification scheduling rules

## Dependencies
- Story 004
- Story 011
- Story 014
- Story 033

---

## Checklist
- [x] Reproduce the `409 getUpdates` conflict with an operator job while a server is polling. (Observed in end-to-end testing on 2026-08-04; reproduced by the previous `main()` path which started long polling before handling any `ops:` command.)
- [x] Refactor CLI command setup so operator jobs construct only the services required for that job and do not configure the Telegram update transport.
- [x] Keep production server startup on long polling by default.
- [x] Add tests or smoke coverage proving report/job command setup does not invoke polling/webhook setup.
- [x] Update deployment smoke guidance to call out the single-poller rule.

---

## Issues

- End-to-end notification testing on 2026-08-04 showed `telegram getUpdates returned status 409` after operator jobs were executed inside the running container. The jobs themselves completed, but their processes incorrectly enabled a second long-polling loop.

---

## Completion Summary

Refactored `cmd/server/main.go` so operator commands are dispatched before any long-running Telegram transport is configured. The server path still creates the bot, registers webhooks/long-polling, and starts the HTTP server; the operator path creates only the services needed for the requested command and uses a Telegram service that can send messages but never consumes updates.

Key changes:
- Added an early `ops:` branch in `main()` that builds an `opsRuntime` with `buildTelegramServiceForOps` and dispatches via `runOpsCommand`, then returns before the server setup begins.
- Removed the previous operator-command switch from the server path so `configureTelegramTransport` is unreachable during CLI runs.
- Added `TestOpsCommandsDoNotConfigureTelegramTransport` to prove `ops:run-job telegram-daily` and `ops:report-events` complete without calling `DeleteWebhook`, `SetWebhook`, or starting the poller.
- Updated `deploy/README.md` and `deploy/launch-smoke-checklist.md` to document the single-poller rule and the smoke check that verifies no `409 getUpdates` errors after running an operator job.

Verification: `go test ./... -count=1` passes. Production server startup remains on long polling by default. Story 040 is ready for `/complete-story`.