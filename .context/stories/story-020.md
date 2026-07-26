# Story 020: Global design tokens and app shell

**Status:** not-started  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** —

---

## Goal
Establish a single source of truth for the sage/cream visual language in the SvelteKit app and update the global layout shell so every page inherits the correct background, type, and max-width behavior.

## Verification
- A temporary or existing page renders with the cream background, sage primary, DM Serif/Outfit fonts, and centered 520 px card without relying on page-level overrides.
- The new token file is imported by the global layout and consumed by at least Button, Input, Tile, and InfoBox primitives.

## Scope — files this story may touch
- web/src/app.css (new)
- web/src/app.html
- web/src/routes/+layout.svelte (new)
- web/src/lib/components/design/Button.svelte (new)
- web/src/lib/components/design/Input.svelte (new)
- web/src/lib/components/design/Tile.svelte (new)
- web/src/lib/components/design/InfoBox.svelte (new)

## Out of scope — do not touch
- Page-specific markup or functional logic
- Rebuilding pages other than token verification

## Dependencies
- None

---

## Checklist
- [ ] Create `web/src/app.css` with the full sage/cream token set: colors, typography, spacing, radius, shadows, and focus states.
- [ ] Load DM Serif Display and Outfit in `web/src/app.html`.
- [ ] Create `web/src/routes/+layout.svelte` that imports `app.css` and provides the global app shell.
- [ ] Create primitive components `Button`, `Input`, `Tile`, and `InfoBox` mapped to the design tokens.
- [ ] Remove any conflicting `:global(body)` overrides from `web/src/routes/+page.svelte` so the global shell takes effect.
- [ ] Run the dev server and visually confirm the landing page renders with the new palette and fonts.

---

## Issues

---

## Completion Summary
