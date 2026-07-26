# Story 022: Onboarding invite-code entry redesign

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

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
- [ ] Replace local blue/Inter styles with token classes and the `Input`, `Button`, and `InfoBox` primitives.
- [ ] Structure the markup to match the reference invite-code screen layout.
- [ ] Style the error banner with the `--rose` token and `radius-lg`.
- [ ] Preserve existing session creation, token storage, and redirect logic.
- [ ] Verify the page visually matches the design reference and still completes onboarding.

---

## Issues

---

## Completion Summary
