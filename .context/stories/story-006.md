# Story 006: Build the guided core onboarding flow

**Status:** in-progress  
**Type:** full-stack  
**Created:** 2026-07-21  
**Last accessed:** 2026-07-21  
**Completed:** —

---

## Goal
Provide a step-by-step SvelteKit onboarding flow that hand-holds a non-technical new-mom beta tester through creating her Rahat profile: name, timezone, daily time budget, an optional email address for recaps, and at least one starter or custom task. Every screen must state plainly what to do next and what is required versus optional, so a tester can finish without any technical knowledge or off-product detective work. Telegram and Google Calendar connections are handled separately in Stories 007 and 008.

## Verification
A brand-new non-technical tester can complete the core flow unaided: profile saved, optional email captured, at least one task added, and a first schedule seeded with a clear on-screen summary of what happens next — no raw IDs, no config knowledge, no manual database edits.

## Scope — files this story may touch
- web/src/routes/onboarding/
- web/src/lib/components/
- web/src/lib/api/
- internal/users/
- internal/tasks/
- cmd/server/

## Out of scope — do not touch
- Telegram connection flow (Story 007)
- Google Calendar connection flow (Story 008)
- A day-to-day task management dashboard
- Multi-user accounts or household assignment

## Dependencies
- Story 001
- Story 002
- Story 003

---

## Checklist
- [x] Add a minimal invite or token-based onboarding session flow without a full auth system
- [x] Present a persistent step-by-step plain-language walkthrough that states what is required versus optional at every step
- [x] Capture profile basics, local timezone, and daily task-time budget with friendly defaults and inline validation
- [x] Let the tester optionally add an email address for daily recaps, clearly framed as optional
- [x] Offer starter-library tasks plus manual task and subtask creation or editing with plain-language guidance
- [x] Finish onboarding by validating required data, seeding the first schedule, and showing a clear on-screen summary of what happens next

---

## Issues

---

## Completion Summary
- Added a minimal invite-code onboarding session flow in `cmd/server/` with in-memory session tokens, profile/task/finish endpoints, and permissive CORS so the SvelteKit onboarding UI can call the Go API during local development.
- Added onboarding-specific task helpers in `internal/tasks/` so the flow can clone starter-library templates and fully replace a task plus its subtasks during edits.
- Built a guided `/onboarding` SvelteKit experience with separate step pages, a persistent step list, required/optional labels, inline profile validation, optional email capture, starter-task shortcuts, and a custom task/subtask editor.
- Finished the flow with backend validation plus first-schedule seeding, then surfaced a plain-language completion summary that explains what Rahat saved and what happens next.
