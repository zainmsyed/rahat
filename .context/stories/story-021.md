# Story 021: Onboarding shell and stepper redesign

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Replace the current two-column onboarding shell with the 520 px centered stage card, progress bar, and sage/cream stepper from the design reference.

## Verification
- Every onboarding step renders inside the new shell with the correct progress bar, stage card, display heading, and eyebrow.
- The stepper shows active, completed, and pending states using the token set.

## Scope — files this story may touch
- web/src/lib/components/OnboardingShell.svelte
- web/src/lib/components/OnboardingStepper.svelte

## Out of scope — do not touch
- Inner content of individual onboarding steps
- Functional onboarding state machine

## Dependencies
- 020

---

## Checklist
- [ ] Update `OnboardingShell.svelte` to the centered 520 px stage-card layout with topbar wordmark and progress bar.
- [ ] Update `OnboardingStepper.svelte` to the design reference step indicators, labels, and active/completed states.
- [ ] Wire title and intro into the stage header using display/eyebrow/lede typography tokens.
- [ ] Ensure mobile padding, card radius, and stage-enter animation match the reference.
- [ ] Keep all existing props and slot behavior so later stories only change inner markup.
- [ ] Verify the shell renders correctly across all onboarding routes.

---

## Issues

---

## Completion Summary
