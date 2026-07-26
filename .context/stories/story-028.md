# Story 028: Task management page redesign

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** 2026-07-26

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
- [x] Apply the global shell and token typography to the authenticated tasks page.
- [x] Style `TaskGroup` headers and individual routine rows with the reference card/list treatment.
- [x] Use the toggle component for pause/resume state.
- [x] Style edit and archive actions with text/secondary button variants.
- [x] Preserve ownership checks, archive confirmation, and pause logic.
- [x] Verify the page matches the reference and still passes its tests.

---

## Issues

---

## Completion Summary
Redesigned `web/src/routes/tasks/+page.svelte` and `web/src/lib/components/tasks/TaskGroup.svelte` to use the sage/cream design-system tokens and shared primitives. The routines page now centers a token-styled card, uses `Button` for the primary "Add a routine" action, and surfaces errors with a token-colored banner. `TaskGroup` headers use display typography with a count badge, and each routine row renders as a bordered card with an eyebrow status, title, summary, description, and subtasks. Pause/resume is controlled by a new reusable `Toggle` component (`web/src/lib/components/design/Toggle.svelte`) with a thumb-and-track switch, while Edit and Remove are styled as `Button` text variants (Remove in rose). The existing confirmation modal and pause/resume/archive API flow are preserved. Added tests for `Toggle` and updated `web/src/routes/tasks/page.test.ts` to target the switch by its accessible label. Full web test suite passes (72 tests) and `svelte-check` reports no errors.