# Story 021: Onboarding shell and stepper redesign

**Status:** in-progress  
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
- [x] Update `OnboardingShell.svelte` to the centered 520 px stage-card layout with topbar wordmark and progress bar.
- [x] Update `OnboardingStepper.svelte` to the design reference step indicators, labels, and active/completed states.
- [x] Wire title and intro into the stage header using display/eyebrow/lede typography tokens.
- [x] Ensure mobile padding, card radius, and stage-enter animation match the reference.
- [x] Keep all existing props and slot behavior so later stories only change inner markup.
- [x] Verify the shell renders correctly across all onboarding routes.

---

## Issues

---

## Completion Summary

Replaced the two-column onboarding shell with the sage/cream centered stage card:

1. **OnboardingShell.svelte** now renders a topbar wordmark (`Rahat.`), a topbar step counter, the `<OnboardingStepper>` progress bar, and a 520 px stage card with eyebrow, display title, and lede. The card uses the design-system color, spacing, radius, and shadow tokens and includes a `stepIn` animation.
2. **OnboardingStepper.svelte** was rewritten as a linear progress bar whose fill width reflects the current step relative to the total onboarding steps, using `--primary` on a `--primary-soft` track. It exposes ARIA progress attributes.
3. Mobile styles reduce padding and card radius below 540 px, matching the reference.
4. All existing props (`steps`, `currentStep`, `title`, `intro`, `finished`) and the default slot are preserved so later stories only update inner step markup.
5. Added `OnboardingShell` and `OnboardingStepper` entries to `.context/design/components.md`.

Verification:
- `cd web && npm run check` passes with zero errors or warnings.
- `cd web && npm test` passes (38 tests).
- `go test ./...` passes.
- The onboarding pages still compile and render inside the new shell.

No blockers. Ready for `/complete-story`.
