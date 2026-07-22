# Story 009: Add the read-only today/tomorrow lookahead page

**Status:** not-started  
**Type:** frontend  
**Created:** 2026-07-07  
**Last accessed:** 2026-07-07  
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
- [ ] Add a simple token-link mechanism for read-only schedule access
- [ ] Build mobile-friendly today and tomorrow schedule views grouped by morning, afternoon, and evening
- [ ] Show blocked windows and conservative calendar explanations when tasks are omitted or limited
- [ ] Remove edit controls so the page stays passive and low-maintenance
- [ ] Cover token access rules and rendered schedule states with tests or smoke checks

---

## Issues

---

## Completion Summary
