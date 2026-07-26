# Story 020: Global design tokens and app shell

**Status:** complete  
**Type:** feature  
**Created:** 2026-07-26  
**Last accessed:** 2026-07-26  
**Completed:** 2026-07-26

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
- [x] Create `web/src/app.css` with the full sage/cream token set: colors, typography, spacing, radius, shadows, and focus states.
- [x] Load DM Serif Display and Outfit in `web/src/app.html`.
- [x] Create `web/src/routes/+layout.svelte` that imports `app.css` and provides the global app shell.
- [x] Create primitive components `Button`, `Input`, `Tile`, and `InfoBox` mapped to the design tokens.
- [x] Remove any conflicting `:global(body)` overrides from `web/src/routes/+page.svelte` so the global shell takes effect.
- [x] Run the dev server and visually confirm the landing page renders with the new palette and fonts.

---

## Issues

---

## Completion Summary

Established the sage/cream design system and global app shell:

1. Added `web/src/app.css` with named CSS custom properties for all colors, typography, spacing, radius, shadows, and motion tokens from the design reference.
2. Loaded `DM Serif Display`, `Outfit`, and `JetBrains Mono` in `web/src/app.html` with preconnect hints.
3. Created `web/src/routes/+layout.svelte` that imports `app.css` and wraps every route in `.app-shell`.
4. Built primitive components under `web/src/lib/components/design/`: `Button` (primary/secondary/text), `Input` (with error state), `Tile` (selected/unselected with check), and `InfoBox`.
5. Removed the conflicting `:global(body)` override from the landing page and also from the login page so the global shell background and typography take effect immediately.
6. Filled the design-system and brand docs with the confirmed sage/cream values and marked them `<!-- source: story-020 -->`.

Verification:
- `cd web && npm run check` passes with zero errors.
- `cd web && npm test` passes (26 tests).
- `go test ./...` passes.
- The dev server starts cleanly and serves SSR pages that include the new tokens and fonts.

No blockers. Ready for `/complete-story`.
