# Story 027: Login page redesign

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Apply the centered card, input, button, and typography tokens to the beta access-link login page.

## Verification
- The login page renders as a centered sage/cream card with token inputs and buttons.
- Token exchange and redirect behavior continue to work.

## Scope — files this story may touch
- web/src/routes/login/+page.svelte

## Out of scope — do not touch
- Global tokens and layout
- Auth service

## Dependencies
- 020

---

## Checklist
- [ ] Replace local blue/Inter styles with token classes and `Input`/`Button` primitives.
- [ ] Structure the page as a centered 520 px stage card matching the onboarding shell.
- [ ] Style loading and error states with token colors and `InfoBox`.
- [ ] Preserve access-link token exchange and redirect logic.
- [ ] Verify trusted-origin CSRF behavior is unchanged.
- [ ] Confirm the page matches the reference calm screen aesthetic.

---

## Issues

---

## Completion Summary
