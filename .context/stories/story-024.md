# Story 024: Onboarding tasks page and TaskEditor redesign

**Status:** not-started  
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
- [ ] Style starter-task tiles with selection, icon, title, subtitle, and check states from the design reference.
- [ ] Apply `Input`, `Button`, and `Tile` primitives inside `TaskEditor`.
- [ ] Ensure subtask rows use consistent spacing, radius, and typography tokens.
- [ ] Preserve add, remove, edit, and starter-template selection logic.
- [ ] Verify responsive layout and interaction on narrow viewports.

---

## Issues

---

## Completion Summary
