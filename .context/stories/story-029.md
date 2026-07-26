# Story 029: Lookahead and landing pages redesign

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

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
- [ ] Redesign the landing page to the centered card with display type, lede, and sage primary action.
- [ ] Apply token styling to `LookaheadDay` cards and window lists.
- [ ] Use `InfoBox` for empty-day and blocked-window explanations.
- [ ] Preserve lookahead token-based access and API calls.
- [ ] Confirm no local `:global(body)` style overrides remain in any page.
- [ ] Run the full web test suite and visually verify both pages.

---

## Issues

---

## Completion Summary
