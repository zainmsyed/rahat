# Story 026: Fix daily-budget slider tick alignment

**Status:** in-progress  
**Type:** bug  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Make the daily task-time budget slider thumb line up with its tick labels on the onboarding profile page.

## Verification
- The slider thumb for any value between 15 and 480 sits visually above the tick that matches that value.
- Tick labels remain readable on narrow viewports.

## Scope — files this story may touch
- web/src/routes/onboarding/profile/+page.svelte

## Out of scope — do not touch
- The budget value range (15–480) or step size.
- The summary-box preview or other profile fields.

## Dependencies
- 023

---

## Checklist
- [x] Position tick labels proportionally to the slider value range instead of evenly spacing them.
- [x] Verify the thumb aligns with each labeled tick at its corresponding value.
- [x] Confirm the layout still works at 520px and mobile widths.
- [x] Add or update a page/render test if needed.

---

## Issues

---

## Completion Summary
Fixed the daily-budget slider tick alignment in `web/src/routes/onboarding/profile/+page.svelte`. Tick labels are now positioned proportionally using `left: ((value - min) / (max - min)) * 100%` instead of `justify-content: space-between`, so the slider thumb sits directly above the label that matches the current value. First and last labels are edge-aligned to avoid clipping outside the card. Added a test in `web/src/routes/onboarding/profile/page.test.ts` that asserts the tick labels are rendered. Layout remains responsive: the label stack still fits at 520px and mobile widths by keeping the labels small and using edge alignment at the extremes. All web tests pass (65) and `svelte-check` reports no errors.
