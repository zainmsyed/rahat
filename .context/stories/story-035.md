# Story 035: Fix daily-budget slider positioning and value alignment

**Status:** not-started  
**Type:** bug  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-03  
**Completed:** —

---

## Goal
Make the onboarding daily task-time budget slider accurate and visually aligned. The slider thumb, displayed value, and tick labels (`15`, `60`, `120`, `240`, `480`) must represent the same linear value range, remain aligned at the ends, and render correctly at supported viewport widths.

## Verification
On the onboarding profile step, moving the slider to each supported tick places the thumb over the corresponding label and updates the summary value to the same number. The first and last labels remain within the track bounds, labels are positioned proportionally to the slider's numeric range, and the behavior is covered by a page-level regression test plus a manual browser check at desktop and narrow widths.

## Scope — files this story may touch
- web/src/routes/onboarding/profile/+page.svelte
- web/src/routes/onboarding/profile/page.test.ts

## Out of scope — do not touch
- Backend profile validation or persistence
- Shared design-system components unless explicitly approved as an exception
- Other onboarding steps
- Docker or deployment configuration

## Dependencies
- Story 026
- Story 033

---

## Checklist
- [ ] Reproduce the slider misalignment at the supported desktop and narrow viewport sizes.
- [ ] Align the slider track, thumb positions, tick labels, and summary value across the full `15`–`480` range.
- [ ] Preserve proportional tick placement rather than evenly spacing labels independent of their values.
- [ ] Add a regression test that asserts tick values and their rendered positions/structure, plus the summary update when the slider changes.
- [ ] Verify the corrected layout manually in a browser at desktop and narrow widths.

---

## Issues

---

## Completion Summary
