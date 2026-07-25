# Story 009: Add the read-only today/tomorrow lookahead page

**Status:** in-progress  
**Type:** frontend  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-24  
**Completed:** —

---

## Goal
Deliver the passive web page linked from Rahat notifications so users can inspect today and tomorrow without managing tasks in a dashboard. The page should show scheduled tasks, time windows, and blocked-calendar explanations only, keeping the product low-maintenance and trust-building.

## Verification
A tester can open a tokenized link on mobile and read the current today and tomorrow plan, including blocked-window explanations, without being able to edit tasks or schedule state.

## Scope — files this story may touch
- web/src/routes/lookahead/
- web/src/lib/components/schedule/
- web/src/lib/api/
- internal/tokens/
- internal/scheduler/
- cmd/server/

## Out of scope — do not touch
- Editing tasks or marking them done from the web
- Historical analytics pages
- Full authentication or account settings

## Dependencies
- Story 003
- Story 005
- Story 006

---

## Checklist
- [x] Add a simple token-link mechanism for read-only schedule access
- [x] Build mobile-friendly today and tomorrow schedule views grouped by morning, afternoon, and evening
- [x] Keep required multistep task chains together so later steps do not appear without earlier steps
- [x] Support hidden soft-follow-up subtask metadata so cleanup steps can defer without blocking required chains
- [x] Show blocked windows and conservative calendar explanations when tasks are omitted or limited
- [x] Remove edit controls so the page stays passive and low-maintenance
- [x] Cover token access rules and rendered schedule states with tests or smoke checks

---

## Issues

---

## Completion Summary

Added a read-only lookahead flow backed by signed HMAC tokens. The server now exposes `GET /lookahead/plan?token=...` for today/tomorrow schedule previews and a development token helper at `GET /lookahead/token?user_id=...` when token issuing is allowed. Lookahead uses `scheduler.PreviewDay`, which computes the same scheduled, overflowed, skipped, blocked-window, and budget information as the scheduler without persisting occurrences or checkpoints.

The web app now has a passive `/lookahead` page that reads the token from the URL, loads the plan, and renders today and tomorrow grouped by morning, afternoon, and evening. The schedule component shows blocked calendar explanations, small-task-only warnings, omitted/limited tasks with conservative reasons, and intentionally has no edit, complete, or reschedule controls. Required multistep task chains now stay together: Rahat will not show a later step such as “move to dryer” unless the earlier required step can also be scheduled. The scheduler also supports hidden internal subtask dependency metadata, so future task templates can mark cleanup steps like “Fold laundry” as soft follow-ups that may defer without blocking the required wash/dry chain.

Coverage includes token issue/verify/expiry/tamper tests, lookahead handler token-access and response tests, scheduler preview read behavior through handler tests, and frontend rendering tests for grouped tasks, blocked windows, omitted items, and absence of edit controls. `go test ./...`, `cd web && npm run check`, and `cd web && npm test` pass.
