# Story 036: Redesign onboarding step six to match the design system

**Status:** complete  
**Type:** design  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-03  
**Completed:** 2026-08-03

---

## Goal
Bring onboarding step six, **Review and finish**, into the established Rahat design language. Replace the visibly ad hoc panel, hardcoded colors, native button styling, and inconsistent card treatment with the shared layout, spacing, typography, colors, and CTA patterns used by the other onboarding steps.

The review step must remain clear on desktop and narrow screens, show profile/task/calendar summaries without technical clutter, and keep the finish action prominent without destabilizing the loading, error, or success states.

## Verification
The review page matches the shared onboarding shell and design tokens at desktop and narrow widths. Profile, task, and calendar summaries use consistent cards and typography; Back and Finish onboarding use the shared Button component; validation/error and finishing states remain usable; and the completed schedule summary remains readable. Automated page-level tests cover the pre-filled review state, finish action, error state, and completed summary state.

## Scope — files this story may touch
- web/src/routes/onboarding/review/+page.svelte
- web/src/routes/onboarding/review/page.test.ts
- web/src/lib/components/design/Button.svelte (only if a missing shared CTA capability is required)
- web/src/lib/components/design/Button.test.ts (only if Button changes)
- web/src/lib/components/design/InfoBox.svelte (only if a shared status surface is required)
- web/src/lib/components/design/InfoBox.test.ts (only if InfoBox changes)

## Out of scope — do not touch
- Onboarding API contracts or scheduler behavior
- Telegram `/edit` behavior
- Profile slider behavior
- Global stylesheet or unrelated pages unless explicitly approved as an exception

## Dependencies
- Story 022
- Story 024
- Story 033

---

## Checklist
- [x] Compare step six against the shared onboarding shell and current design-system components, documenting the visual gaps.
- [x] Replace hardcoded page-specific colors, borders, spacing, and native CTA styles with shared design tokens/components.
- [x] Redesign profile, task, calendar, error, and success summary surfaces for consistent hierarchy and responsive layout.
- [x] Preserve the finish flow, loading/finishing state, error handling, and redirect behavior.
- [x] Add or update page-level tests for pre-filled render, finish action, error state, and completed summary state.
- [x] Manually verify step six at desktop and narrow widths against the design language used by the other onboarding steps.

---

## Issues

- The original review page used hardcoded blue borders, buttons, spacing, and status colors instead of the established Rahat tokens and shared Button component. These were replaced with the existing onboarding shell, Button, and InfoBox patterns.
- The design guide had unspecified spacing, visual style, and hard-constraint fields. Those gaps were filled from the user's direction to preserve the current warm, readable design language and viewport behavior, with `<!-- source: story-036 -->` attribution.
- Automated and manual verification are complete. The profile and task summaries stack at full card width for easier reading, with the calendar summary below them. The latest stacked layout was verified at desktop and narrow widths at the LAN test URL.

---

## Completion Summary

Redesigned onboarding step six to use the existing Rahat onboarding shell and design language. Replaced the ad hoc panel and native CTA styles with token-based summary cards, responsive full-width profile/task/calendar surfaces, shared Button CTAs, and an InfoBox error surface. Preserved loading, finish-in-progress, error, and redirect behavior while making the success schedule summary structured and readable. Added page-level tests for pre-filled review data, summary card structure, finish action and redirect, error handling, and completed schedule summary; all web tests pass (23 files, 83 tests) and `npm run check` reports no Svelte diagnostics. The latest stacked layout was manually verified at desktop and narrow widths. The story is ready for `/complete-story`; its status remains in-progress until Vazir runs the closeout workflow.
