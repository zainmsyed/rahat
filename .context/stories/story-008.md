# Story 008: Send email overview and recap digests

**Status:** not-started  
**Type:** integration  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
**Completed:** —

---

## Goal
Keep email in v1 as a non-interactive overview and recap channel that complements Telegram by summarizing the current schedule and linking to the passive lookahead page. This story should not attempt email reply parsing or replace Telegram as the real-time check-in surface.

## Verification
A tester with email enabled receives a readable overview or recap that matches the current schedule, includes a safe link to the read-only page, and can disable email without affecting Telegram reminders.

## Scope — files this story may touch
- internal/notifications/email/
- internal/notifications/preferences/
- internal/events/
- web/src/routes/onboarding/
- web/src/lib/api/
- internal/tokens/

## Out of scope — do not touch
- Email reply parsing or inline task actions
- SMS delivery
- Replacing Telegram as the interactive loop

## Dependencies
- Story 003
- Story 006
- Story 007

---

## Checklist
- [ ] Add SMTP-backed email sending with template support and delivery logging
- [ ] Compose overview or recap emails that summarize today’s schedule and pending check-in context
- [ ] Include a safe link to the read-only today and tomorrow page instead of interactive task controls
- [ ] Let users opt in or out of email recaps without changing Telegram reminder behavior
- [ ] Add tests or preview fixtures for empty days, blocked-window days, and mixed-task days

---

## Issues

---

## Completion Summary
