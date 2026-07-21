<script lang="ts">
	import OnboardingStepper from '$lib/components/OnboardingStepper.svelte';
	import type { OnboardingStep } from '$lib/api/onboarding';

	export let steps: OnboardingStep[] = [];
	export let currentStep = 0;
	export let title = '';
	export let intro = '';
	export let finished = false;
</script>

<svelte:head>
	<title>Rahat onboarding</title>
	<meta
		name="description"
		content="Guided Rahat onboarding for profile setup, starter tasks, and first schedule seeding."
	/>
</svelte:head>

<div class="page">
	<OnboardingStepper {steps} {currentStep} {finished} />

	<main class="content">
		<header class="hero">
			<p class="eyebrow">Welcome to Rahat</p>
			<h1>{title}</h1>
			<p>{intro}</p>
		</header>

		<slot />
	</main>
</div>

<style>
	:global(body) {
		margin: 0;
		font-family: Inter, system-ui, sans-serif;
		background: #f4f7fb;
		color: #14202c;
	}

	.page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 1.5rem;
		display: grid;
		grid-template-columns: 320px minmax(0, 1fr);
		gap: 1.5rem;
	}

	.content {
		display: grid;
		gap: 1.25rem;
	}

	.hero,
	:global(.panel),
	:global(.success) {
		padding: 1.4rem;
		border-radius: 1.4rem;
		background: white;
		box-shadow: 0 16px 40px rgba(20, 32, 44, 0.08);
	}

	.hero h1,
	:global(.panel h2),
	:global(.panel h3),
	:global(.panel h4),
	:global(.success h3),
	:global(.success h4) {
		margin: 0;
	}

	.hero p,
	:global(.panel p),
	:global(.success p),
	:global(.panel li),
	:global(.success li) {
		line-height: 1.55;
	}

	.eyebrow,
	:global(.label) {
		margin: 0 0 0.4rem;
		font-size: 0.85rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #4f46e5;
	}

	@media (max-width: 960px) {
		.page {
			grid-template-columns: 1fr;
		}
	}
</style>
