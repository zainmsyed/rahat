# Story 026: Onboarding review page redesign

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Apply summary rows, info boxes, and primary button styling to the final onboarding review screen so testers see a clear, designed outcome.

## Verification
- The review page shows scheduled tasks and overflow/skipped feedback using summary rows and info boxes.
- The finish action uses the primary button style.

## Scope — files this story may touch
- web/src/routes/onboarding/review/+page.svelte

## Out of scope — do not touch
- Onboarding shell
- Finish/submit API

## Dependencies
- 020
- 021
- 023
- 024
- 025

---

## Checklist
- [ ] Style scheduled-task summaries with summary-row components.
- [ ] Show overflowed/skipped tasks and day-selection reasons in `InfoBox` styling.
- [ ] Apply the primary button for the finish action.
- [ ] Preserve finish, schedule preview fetch, and submission logic.
- [ ] Ensure the Telegram confirmation payload is unchanged by the redesign.
- [ ] Verify the page matches the reference review screen.

---

## Issues

---

## Completion Summary
