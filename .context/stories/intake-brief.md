# Intake Brief

**Last updated:** 2026-07-26

## Planning brief
Apply the sage/cream design system from `.context/intake/references/rahat design file.html` and `.context/intake/references/rahat.html` to all existing SvelteKit web surfaces. Replace the current ad-hoc blue/Inter styling with the defined tokens, typography, components, and screen layouts so the web experience matches the designed warmth of the product.

## Source files
- .context/intake/prd/rahat-prd.md
- .context/intake/references/rahat design file.html
- .context/intake/references/rahat.html

## Distilled notes
### Scope of the replan
- All existing SvelteKit pages and shared components are in scope: landing (`+page.svelte`), login, onboarding (invite-code entry, profile, tasks, Telegram connection, Google Calendar connection, review), task management, and the today/tomorrow lookahead page.
- The design system uses a sage primary green (`#7a9b76`), warm cream background (`#faf7f2`), near-black ink (`#1f1d1a`), DM Serif Display for display type, Outfit for body/UI type, 520 px max-width centered cards for standalone flows, and a named 8 pt spacing/radius/shadow scale.
- Reusable components defined in the reference: buttons (primary, secondary, text), inputs, select, tiles, toggles, sliders, connect tiles, summary boxes, info boxes, progress bars, and screen shells.
- Tone and behavior remain unchanged; this is a visual redesign and component-alignment pass, not a functional rewrite.

### Safe default assumptions
- Start with global design tokens and layout primitives, then apply the shell to onboarding, then to authenticated maintenance surfaces, then to the landing/lookahead pages.
- Keep each page's existing logic and API contracts; only change markup, styles, and shared component usage.
- No new pages are added unless required by the design reference.
- Story numbers continue from 020; existing stories 001–019 are preserved as history.
