# Story 023: Onboarding profile page redesign

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** 2026-07-26

---

## Goal
Redesign the profile setup page to use the design system's inputs, slider, and summary tiles while keeping the same form state and submission flow.

## Verification
- The profile page uses the reference input, slider, and summary-box styling.
- Daily-budget selection displays in DM Serif with the sage thumb and track.

## Scope — files this story may touch
- web/src/routes/onboarding/profile/+page.svelte

## Out of scope — do not touch
- Onboarding shell
- Profile API contract

## Dependencies
- 020
- 021

---

## Checklist
- [x] Replace profile fields with design-system `Input` and labeled controls.
- [x] Style the daily-budget slider with token thumb, track, and ticks.
- [x] Use the summary-box component for the budget preview.
- [x] Style validation messages with the `--rose` token.
- [x] Preserve existing form state, validation, and submission logic.
- [x] Verify the page matches the reference and still saves the profile.

---

## Issues

---

## Completion Summary
Redesigned `web/src/routes/onboarding/profile/+page.svelte` to use the design-system `Input` and `Button` components, a custom-styled range slider for daily budget (sage thumb, track, and tick labels), and a local `summary-box` preview that shows the selected minutes in DM Serif. Validation messages now use the `--rose` token, and the original form state, validation rules, and `saveProfile` submission flow remain unchanged. Added `web/src/routes/onboarding/profile/page.test.ts` covering the happy save path, validation errors, and the no-token redirect. All web tests pass and `svelte-check` reports no errors.
