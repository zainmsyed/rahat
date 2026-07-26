# Story 029: Lookahead and landing pages redesign

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** 2026-07-26

---

## Goal
Apply the sage/cream shell to the today/tomorrow lookahead and the landing page so all public and authenticated web surfaces share the same visual language.

## Verification
- The landing page becomes a centered card with sage primary CTA and concise explanation.
- The lookahead page uses token day cards, window lists, and info boxes.

## Scope — files this story may touch
- web/src/routes/+page.svelte
- web/src/routes/lookahead/+page.svelte
- web/src/lib/components/schedule/LookaheadDay.svelte

## Out of scope — do not touch
- Global tokens and layout
- Lookahead API

## Dependencies
- 020

---

## Checklist
- [x] Redesign the landing page to the centered card with display type, lede, and sage primary action.
- [x] Apply token styling to `LookaheadDay` cards and window lists.
- [x] Use `InfoBox` for empty-day and blocked-window explanations.
- [x] Preserve lookahead token-based access and API calls.
- [x] Confirm no local `:global(body)` style overrides remain in any page.
- [x] Run the full web test suite and visually verify both pages.

---

## Issues

---

## Completion Summary
Redesigned the public landing page (`web/src/routes/+page.svelte`) and the read-only lookahead page (`web/src/routes/lookahead/+page.svelte`) to share the sage/cream design-system language. The landing page now renders as a single centered 520px card with display typography, a lede, and a sage primary CTA anchor. The lookahead page removes its local `:global(body)` override, uses token cards and spacing, and surfaces loading and error states with `InfoBox`. `LookaheadDay` was rewritten with token day cards, window sections styled on the soft background, and `InfoBox` used for small-task-only notes, blocked-window reasons, empty-window notes, and omitted-item explanations. The token-based access flow and API calls are unchanged. Added `web/src/routes/page.test.ts` covering the landing page's card, heading, lede, and CTA. Full web test suite passes (73 tests) and `svelte-check` reports no errors.