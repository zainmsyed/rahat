# Story 024: Onboarding tasks page and TaskEditor redesign

**Status:** in-progress  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Apply the tile, input, and button components to starter-task selection and the manual task editor so adding routines feels calm and consistent.

## Verification
- Starter-task tiles show selected/unselected states matching the reference.
- `TaskEditor` inputs, buttons, and subtask rows use the token set.

## Scope — files this story may touch
- web/src/routes/onboarding/tasks/+page.svelte
- web/src/lib/components/TaskEditor.svelte

## Out of scope — do not touch
- Onboarding shell
- Task persistence API

## Dependencies
- 020
- 021

---

## Checklist
- [x] Style starter-task tiles with selection, icon, title, subtitle, and check states from the design reference.
- [x] Apply `Input`, `Button`, and `Tile` primitives inside `TaskEditor`.
- [x] Ensure subtask rows use consistent spacing, radius, and typography tokens.
- [x] Preserve add, remove, edit, and starter-template selection logic.
- [x] Verify responsive layout and interaction on narrow viewports.

---

## Issues

---

## Completion Summary
Redesigned `web/src/routes/onboarding/tasks/+page.svelte` to use `Tile` for starter-task selection, `Button` for all actions, and design-system tokens for saved-task cards and layout. Redesigned `web/src/lib/components/TaskEditor.svelte` to use the `Input`, `Button`, and `InfoBox` primitives, design-system colors/spacing/typography tokens, and token-styled selects and textareas. Subtask rows now render as rounded cards with consistent spacing. The original add/remove/edit/save logic and starter-template selection flow were preserved. Added `web/src/routes/onboarding/tasks/page.test.ts` covering the no-token redirect, starter-template rendering, and tile-click addition. Extended `Input.svelte` to support numeric `value` and `min`/`max` props. All web tests pass (52) and `svelte-check` reports no errors.
