# Components

<!-- source: seed -->
<!-- Living registry of established components. Populated incrementally by UI stories. -->

## Button
<!-- source: story-020 -->
- Location: `web/src/lib/components/design/Button.svelte`
- Props: `variant: 'primary' | 'secondary' | 'text'`, `type`, `disabled`, `fullWidth`
- Use: Primary actions, secondary actions, and low-emphasis text links.

## Input
<!-- source: story-020 -->
- Location: `web/src/lib/components/design/Input.svelte`
- Props: `id`, `label`, `value`, `placeholder`, `type`, `required`, `error`
- Use: Form fields with labeled controls and optional error messaging.

## Tile
<!-- source: story-020 -->
- Location: `web/src/lib/components/design/Tile.svelte`
- Props: `title`, `subtitle`, `icon`, `selected`
- Use: Selectable list items such as starter tasks or connection options.

## InfoBox
<!-- source: story-020 -->
- Location: `web/src/lib/components/design/InfoBox.svelte`
- Props: `title` (optional)
- Use: Inline informational or contextual notes with an icon and a text slot.

## SummaryBox
<!-- source: story-023 -->
- Location: `web/src/lib/components/design/SummaryBox.svelte`
- Props: `id`, `value`, `unit`, `hint`
- Use: Highlight a numeric value with a unit and a short hint, such as a budget preview.

## OnboardingShell
<!-- source: story-021 -->
- Location: `web/src/lib/components/OnboardingShell.svelte`
- Props: `steps`, `currentStep`, `title`, `intro`, `finished`
- Use: Centered 520 px stage-card wrapper for every onboarding step, including topbar wordmark, progress bar, and animated stage header.

## OnboardingStepper
<!-- source: story-021 -->
- Location: `web/src/lib/components/OnboardingStepper.svelte`
- Props: `steps`, `currentStep`, `finished`
- Use: Linear progress indicator rendered below the topbar; fill width reflects current onboarding progress.
