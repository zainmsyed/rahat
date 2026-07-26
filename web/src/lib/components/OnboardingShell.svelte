<script lang="ts">
	import OnboardingStepper from './OnboardingStepper.svelte';
	import type { OnboardingStep } from '$lib/api/onboarding';

	export let steps: OnboardingStep[] = [];
	export let currentStep = 0;
	export let title = '';
	export let intro = '';
	export let finished = false;

	$: counter = finished ? 'Done' : steps.length > 0 ? `${currentStep + 1} of ${steps.length}` : '';
	$: eyebrow = finished ? 'All set' : steps.length > 0 ? `Step ${currentStep + 1} of ${steps.length}` : '';
</script>

<svelte:head>
	<title>Rahat onboarding</title>
	<meta
		name="description"
		content="Guided Rahat onboarding for profile setup, starter tasks, and first schedule seeding."
	/>
</svelte:head>

<div class="page">
	<header class="topbar">
		<div class="wordmark">Rahat<span>.</span></div>
		<div class="topbar-meta">{counter}</div>
	</header>

	<OnboardingStepper {steps} {currentStep} {finished} />

	<main class="stage">
		<section class="step active">
			<p class="step-eyebrow">{eyebrow}</p>
			<h1 class="step-title">{title}</h1>
			<p class="step-lede">{intro}</p>
			<slot />
		</section>
	</main>
</div>

<style>
	.page {
		min-height: 100vh;
		display: flex;
		flex-direction: column;
	}

	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--space-3);
		flex-wrap: wrap;
		max-width: var(--surface-max-width);
		width: 100%;
		margin: 0 auto;
		padding: var(--space-6) var(--space-5) var(--space-3);
	}

	.wordmark {
		font-family: var(--font-display);
		font-size: 24px;
		font-weight: 400;
		letter-spacing: -0.01em;
		color: var(--ink);
	}

	.wordmark span {
		color: var(--primary-2);
	}

	.topbar-meta {
		font-size: 12px;
		color: var(--ink-3);
		letter-spacing: 0.04em;
		font-feature-settings: 'tnum';
	}

	.stage {
		flex: 1;
		width: 100%;
		max-width: var(--surface-max-width);
		margin: 0 auto;
		padding: 0 var(--space-5) var(--space-10);
		position: relative;
		min-width: 0;
	}

	.step {
		display: none;
		animation: stepIn 0.5s var(--ease-out);
		background: var(--paper);
		border: 1px solid var(--line);
		border-radius: var(--radius-3xl);
		box-shadow: var(--shadow-md);
		padding: var(--space-8) var(--space-6);
		min-width: 0;
		overflow-wrap: break-word;
	}

	.step.active {
		display: block;
	}

	@keyframes stepIn {
		0% {
			opacity: 0;
			transform: translateY(8px);
		}
		100% {
			opacity: 1;
			transform: translateY(0);
		}
	}

	.step-eyebrow {
		font-size: 11px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--primary-2);
		font-weight: 600;
		margin-bottom: var(--space-3);
	}

	.step-title {
		font-family: var(--font-display);
		font-size: 34px;
		line-height: 1.1;
		letter-spacing: -0.005em;
		color: var(--ink);
		margin-bottom: var(--space-3);
		font-weight: 400;
	}

	.step-lede {
		font-size: 15.5px;
		color: var(--ink-2);
		line-height: 1.6;
		margin-bottom: var(--space-6);
	}

	@media (max-width: 540px) {
		.topbar {
			padding: var(--space-5) var(--space-4) var(--space-3);
		}

		.stage {
			padding: 0 var(--space-4) var(--space-8);
		}

		.step {
			padding: var(--space-6) var(--space-4);
			border-radius: var(--radius-2xl);
		}

		.step-title {
			font-size: 28px;
		}
	}
</style>
