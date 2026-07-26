<script lang="ts">
	import OnboardingShell from './OnboardingShell.svelte';
	import type { OnboardingStep } from '$lib/api/onboarding';

	export let steps: OnboardingStep[] = [];
	export let currentStep = 0;
	export let title = '';
	export let intro = '';
	export let finished = false;
</script>

<OnboardingShell {steps} {currentStep} {title} {intro} {finished}>
	<p class="slot-content">Slot content rendered.</p>
</OnboardingShell>

<style>
	.slot-content {
		color: var(--ink-3);
	}
</style>
