# Story 027: Login page redesign

**Status:** in-progress  
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
- [x] Replace local blue/Inter styles with token classes and `Input`/`Button` primitives.
- [x] Structure the page as a centered 520 px stage card matching the onboarding shell.
- [x] Style loading and error states with token colors and `InfoBox`.
- [x] Preserve access-link token exchange and redirect logic.
- [x] Verify trusted-origin CSRF behavior is unchanged.
- [x] Confirm the page matches the reference calm screen aesthetic.

---

## Issues

---

## Completion Summary
Redesigned `web/src/routes/login/+page.svelte` to use the design-system `Input`, `Button`, and `InfoBox` components and the sage/cream token palette. The page now renders as a centered 520px card with the same radius, border, shadow, and typography as the onboarding shell. Loading and signed-out states use `InfoBox`; errors use the `--rose`/`--rose-soft` tokens. The access-link token exchange and `/tasks` redirect logic are preserved, including the on-mount URL-token auto-exchange. A manual token input fallback was added so users can paste a token if the URL does not contain one. The auth API still sends credentials, so trusted-origin/CSRF behavior is unchanged. Updated `web/src/routes/login/page.test.ts` with tests for URL-token auto-exchange, manual form submission, and exchange failure. All web tests pass (67) and `svelte-check` reports no errors.
