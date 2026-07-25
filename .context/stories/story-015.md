# Story 015: Confirm onboarding completion in Telegram

**Status:** in-progress  
**Type:** integration  
**Created:** 2026-07-25  
**Last accessed:** 2026-07-25  
**Completed:** —

---

## Goal
Give a newly onboarded tester an immediate, concrete confirmation in the channel where Rahat will interact with them. After onboarding successfully saves the profile, routines, and first schedule, send the linked Telegram chat a calm summary of what Rahat recorded, what is planned next, and how to return to routine editing later.

## Verification
A tester who linked Telegram and completes onboarding sees the existing on-screen completion state and receives one Telegram message naming their saved routines, summarizing the first scheduled day by time window, noting overflow or unscheduled items honestly, and explaining that `/edit` opens routine management. The message reflects the persisted schedule rather than a separate approximation, does not expose internal IDs or dependency semantics, and repeated finish requests do not send duplicate confirmations. If Telegram delivery fails, onboarding remains successfully persisted and the UI clearly explains that the confirmation could not be delivered.

## Scope — files this story may touch
- db/migrations/
- internal/notifications/telegram/
- internal/events/
- internal/scheduler/
- cmd/server/
- web/src/lib/api/
- web/src/routes/onboarding/review/
- README.md

## Out of scope — do not touch
- Changing the scheduling algorithm
- Email confirmations or recaps
- A full scheduling dashboard
- Sending a confirmation before onboarding is successfully persisted
- Exposing required-chain versus soft-follow-up dependency semantics
- Account authentication implementation beyond referring users to Story 014’s `/edit` command

## Dependencies
- Story 003
- Story 004
- Story 006
- Story 007
- Story 014

---

## Checklist
- [x] Build the confirmation from the persisted onboarding result and first generated schedule, not from duplicated frontend scheduling logic
- [x] Include the saved routine names and a concise first-day schedule grouped by morning, afternoon, and evening as applicable
- [x] Clearly report overflowed, skipped, or currently unscheduled work without implying everything fit when it did not
- [x] Use warm, plain-language Telegram copy that explains what happens next and tells the tester to send `/edit` whenever they need routine settings on a new or signed-out device
- [x] Send only to the Telegram chat already verified and linked to the authenticated onboarding user
- [x] Make completion notification idempotent so retries or repeated finish requests do not send duplicate summaries
- [x] Record delivery success/failure in event history without logging sensitive links or credentials
- [x] Keep successful onboarding data and browser-session creation intact if Telegram delivery fails, while returning a machine-detectable delivery result for honest UI feedback
- [x] Update the onboarding completion screen to state whether the Telegram summary was delivered and retain the same useful schedule summary on screen
- [x] Add tests for linked and unlinked users, exact routine/schedule mapping, empty windows, overflow/skips, duplicate finish requests, send failure, and safe message contents
- [x] Add Telegram handler/service tests covering the new completion message type

---

## Issues

---

## Completion Summary

Implemented the Telegram onboarding confirmation flow end-to-end.

- Added `onboarding_telegram_confirmations` table migration (`db/migrations/012_story_015_onboarding_telegram_confirmation.sql`) and a small store repository (`internal/store/onboarding_confirmation.go`) to make the notification idempotent per user.
- Added `SendOnboardingConfirmation` to the Telegram service (`internal/notifications/telegram/service.go`). It builds a warm, plain-text summary from the persisted `scheduler.PlanResult` and linked task definitions, groups scheduled items by morning/afternoon/evening, honestly reports overflowed/skipped work, and reminds the tester to send `/edit`. It records delivery success/failure in `event_logs` without sensitive data and uses the confirmation repository to prevent duplicate sends.
- Wired the confirmation into the onboarding finish handler (`cmd/server/onboarding.go`). After the schedule is persisted and the web session cookie is set, the handler attempts delivery only when a verified Telegram chat is linked. A delivery failure is logged but does not fail onboarding; the response now includes `telegram_delivered` for UI feedback.
- Updated the onboarding review completion screen (`web/src/routes/onboarding/review/+page.svelte`) and its API type (`web/src/lib/api/onboarding.ts`) to display whether the Telegram summary was delivered, failed, or skipped because Telegram wasn't linked.
- Added unit/service tests in `internal/notifications/telegram/service_test.go` covering linked/unlinked users, idempotency, send failure, overflow/skips, empty windows, and safe message contents.
- Added handler-level integration tests in `cmd/server/onboarding_test.go` covering finish with a linked Telegram chat, duplicate finish without resending, and finish without Telegram linked.
- All Go tests and web tests pass.
