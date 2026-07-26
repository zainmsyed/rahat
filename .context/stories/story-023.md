# Story 023: Onboarding profile page redesign

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

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
- [ ] Replace profile fields with design-system `Input` and labeled controls.
- [ ] Style the daily-budget slider with token thumb, track, and ticks.
- [ ] Use the summary-box component for the budget preview.
- [ ] Style validation messages with the `--rose` token.
- [ ] Preserve existing form state, validation, and submission logic.
- [ ] Verify the page matches the reference and still saves the profile.

---

## Issues

---

## Completion Summary
