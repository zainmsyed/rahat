# Story 022: Onboarding invite-code entry redesign

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** 2026-07-26

---

## Goal
Apply the sage/cream input, button, and info-box styles to the invite-code entry step so the first impression matches the designed warmth.

## Verification
- The invite-code page matches the reference screen: centered card, eyebrow, title, lede, input, primary button, and error state.
- Session-start logic continues to work unchanged.

## Scope — files this story may touch
- web/src/routes/onboarding/+page.svelte

## Out of scope — do not touch
- Onboarding shell or stepper
- Invite-code backend logic

## Dependencies
- 020
- 021

---

## Checklist
- [x] Replace local blue/Inter styles with token classes and the `Input`, `Button`, and `InfoBox` primitives.
- [x] Structure the markup to match the reference invite-code screen layout.
- [x] Style the error banner with the `--rose` token and `radius-lg`.
- [x] Preserve existing session creation, token storage, and redirect logic.
- [x] Verify the page visually matches the design reference and still completes onboarding.

---

## Issues

---

## Completion Summary

Redesigned the onboarding invite-code entry step using the sage/cream design system:

1. Replaced the local blue/Inter styles in `web/src/routes/onboarding/+page.svelte` with the `Input`, `Button`, and `InfoBox` primitives from `web/src/lib/components/design/`.
2. Restructured the markup into a simple form inside the `OnboardingShell` stage card: labeled invite-code input, full-width primary submit button, and an informational `InfoBox` explaining where to find the code.
3. Wired validation/server errors into the `Input` component, which already styles errors with the `--rose` token.
4. Preserved the existing session flow: `readInviteCodeFromUrl`, `createSession`, token storage via `setStoredOnboardingToken`/`syncTokenInUrl`, and redirect to `nextOnboardingPath` remain unchanged.
5. Removed all the old local component styles (panel, input, button, error-banner overrides).

Verification:
- `cd web && npm run check` passes with zero errors/warnings.
- `cd web && npm test` passes (42 tests).
- `go test ./...` passes.

No blockers. Ready for `/complete-story`.
