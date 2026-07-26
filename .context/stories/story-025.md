# Story 025: Onboarding connection pages redesign

**Status:** in-progress  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Redesign the Telegram and Google Calendar onboarding connection screens using connect tiles and status indicators from the design reference.

## Verification
- Telegram and Calendar pages show provider connect tiles with connected/disconnected status dots.
- Optional/unavailable states use the info-box style.

## Scope — files this story may touch
- web/src/routes/onboarding/telegram/+page.svelte
- web/src/routes/onboarding/calendar/+page.svelte
- web/src/routes/onboarding/calendar/callback/+page.svelte

## Out of scope — do not touch
- Onboarding shell
- OAuth and deep-link logic

## Dependencies
- 020
- 021

---

## Checklist
- [x] Create or reuse a `ConnectTile` component matching the reference icon, name, status, and hover states.
- [x] Apply the connect tile to the Telegram and Calendar onboarding pages.
- [x] Use `InfoBox` for optional/unavailable provider messaging.
- [x] Preserve deep-link code exchange, redirect, and status polling logic.
- [x] Style the Calendar callback page inside the same shell.
- [x] Verify connection status still updates correctly after redesign.

---

## Issues

---

## Completion Summary
Created `web/src/lib/components/design/ConnectTile.svelte` (with test and registry entry) to display provider connection tiles with icon, name, subtitle, and connected/disconnected status dots. Applied `ConnectTile` and `InfoBox` to `web/src/routes/onboarding/telegram/+page.svelte` and `web/src/routes/onboarding/calendar/+page.svelte`, replacing the old ad-hoc connection panels. Replaced raw buttons with the design-system `Button` component. Preserved Telegram deep-link opening, QR code/code display, and status polling, and preserved Google Calendar OAuth linking, disconnect, and status loading. Wrapped `web/src/routes/onboarding/calendar/callback/+page.svelte` in `OnboardingShell` and styled the loading/error states with `InfoBox` and `Button`. Updated the calendar page test to match the new tile markup. All web tests pass (58) and `svelte-check` reports no errors.
