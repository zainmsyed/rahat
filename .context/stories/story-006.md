# Story 006: Build the onboarding web flow

**Status:** not-started  
**Type:** full-stack  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

---

## Goal
Provide a lightweight SvelteKit onboarding flow so a tester can create a Rahat profile, set timezone and daily budget, add Telegram and email contact details, connect Google Calendar, and load starter or custom tasks without needing a full account system or persistent dashboard. This story should make first-use setup realistic for the initial testing group.

## Verification
A brand-new tester can finish the onboarding flow and end with a saved profile, at least one task, optional calendar connection, and enough data for the next daily schedule run without manual database edits.

## Scope — files this story may touch
- web/src/routes/onboarding/
- web/src/lib/components/
- web/src/lib/api/
- internal/users/
- internal/tasks/
- internal/calendar/
- cmd/server/

## Out of scope — do not touch
- A day-to-day task management dashboard
- Multi-user accounts or household assignment
- SMS setup

## Dependencies
- Story 001
- Story 002
- Story 005

---

## Checklist
- [ ] Add a minimal invite or token-based onboarding session flow without a full auth system
- [ ] Capture profile basics, local timezone, and daily task-time budget
- [ ] Let the user supply Telegram and email contact details with clear channel roles
- [ ] Offer starter-library tasks plus manual task and subtask creation or editing
- [ ] Expose Google Calendar connect and disconnect steps using the read-only integration
- [ ] Finish onboarding by validating required data and triggering the first schedule seed or run

---

## Issues

---

## Completion Summary
