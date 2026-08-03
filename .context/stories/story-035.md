# Story 035: Fix daily-budget slider positioning and value alignment

**Status:** complete  
**Type:** bug  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-03  
**Completed:** 2026-08-03

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
- [x] Reproduce the slider misalignment at the supported desktop and narrow viewport sizes.
- [x] Align the slider track, thumb positions, tick labels, and summary value across the full `15`–`480` range.
- [x] Preserve proportional tick placement rather than evenly spacing labels independent of their values.
- [x] Add a regression test that asserts tick values and their rendered positions/structure, plus the summary update when the slider changes.
- [x] Verify the corrected layout manually in a browser at desktop and narrow widths.

---

## Issues

- The supplied desktop screenshot showed the `60` thumb positioned to the right of the `60` label because native range thumbs are centered inside an inset usable track width. The tick-label row previously used the full input width, so its endpoints and intermediate positions did not match the thumb geometry. The corrected tick row now explicitly uses the 12px thumb-radius inset.
- The budget control now also accepts direct whole-minute entry from 15 through 480; the slider and summary remain bound to the typed value.
- Manual desktop and narrow-width verification completed successfully in a private browser window after clearing the stale cached frontend bundle. The slider, tick labels, direct-entry field, and summary are aligned and synchronized.

---

## Completion Summary

Updated the profile budget control without changing the existing visual design. The tick-label row now explicitly uses the same 12px thumb-radius inset as the native range control, so the labels share the slider thumb's actual usable coordinate range while retaining numeric proportionality across 15–480 minutes. Added a compact direct-entry number field for any whole-minute value in that range; it stays synchronized with the slider and summary. Added stable tick markers and a page-level regression test covering proportional positions, endpoint structure, every supported tick value, direct entry, and summary updates. The profile test passes, all web tests pass (79 tests), and `npm run check` reports no Svelte diagnostics. Manual desktop and narrow-width browser verification also passed after refreshing the frontend cache. The story is ready for Vazir closeout and remains in-progress until that workflow runs.
