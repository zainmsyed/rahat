# Story 025: Onboarding connection pages redesign

**Status:** not-started  
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
- [ ] Create or reuse a `ConnectTile` component matching the reference icon, name, status, and hover states.
- [ ] Apply the connect tile to the Telegram and Calendar onboarding pages.
- [ ] Use `InfoBox` for optional/unavailable provider messaging.
- [ ] Preserve deep-link code exchange, redirect, and status polling logic.
- [ ] Style the Calendar callback page inside the same shell.
- [ ] Verify connection status still updates correctly after redesign.

---

## Issues

---

## Completion Summary
