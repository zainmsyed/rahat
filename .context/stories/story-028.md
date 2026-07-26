# Story 028: Task management page redesign

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Redesign the authenticated routines page to use the design system's cards, task groups, and toggle/secondary actions.

## Verification
- The routines page uses token cards, task-group headers, and toggle/secondary-button actions.
- Pause, edit, and archive actions remain functional.

## Scope — files this story may touch
- web/src/routes/tasks/+page.svelte
- web/src/lib/components/tasks/TaskGroup.svelte

## Out of scope — do not touch
- Global tokens and layout
- Task management API

## Dependencies
- 020

---

## Checklist
- [ ] Apply the global shell and token typography to the authenticated tasks page.
- [ ] Style `TaskGroup` headers and individual routine rows with the reference card/list treatment.
- [ ] Use the toggle component for pause/resume state.
- [ ] Style edit and archive actions with text/secondary button variants.
- [ ] Preserve ownership checks, archive confirmation, and pause logic.
- [ ] Verify the page matches the reference and still passes its tests.

---

## Issues

---

## Completion Summary
