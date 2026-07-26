<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import {
		apiBaseUrl,
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		getStoredOnboardingToken,
		syncTokenInUrl
	} from '$lib/api/onboarding';

	let loading = true;
	let pageError = '';
	let sessionToken = '';

	$: steps = buildOnboardingSteps(
		{
			has_profile: true,
			telegram_linked: false,
			calendar_connected: false,
			tasks: [],
			starter_templates: []
		},
		!!sessionToken
	);

	onMount(async () => {
		sessionToken = getStoredOnboardingToken();
		if (!sessionToken) {
			await goto('/onboarding');
			return;
		}

		const params = new URL(window.location.href).searchParams;
		const state = params.get('state');
		const code = params.get('code');
		const error = params.get('error');

		if (error) {
			loading = false;
			pageError = `Google returned an error (${error}). You can skip this step or try again.`;
			return;
		}

		if (!state || !code) {
			loading = false;
			pageError =
				'The Google response is missing information needed to connect your calendar. Please try again.';
			return;
		}

		try {
			const response = await fetch(
				`${apiBaseUrl}/calendar/google/connect?state=${encodeURIComponent(state)}&code=${encodeURIComponent(code)}`,
				{
					method: 'POST',
					headers: { 'Content-Type': 'application/json' },
					body: JSON.stringify({})
				}
			);
			if (!response.ok) {
				throw new Error(await response.text());
			}
			syncTokenInUrl(sessionToken);
			await goto('/onboarding/calendar');
		} catch (error) {
			loading = false;
			pageError = error instanceof Error ? error.message : 'Could not connect Google Calendar.';
		}
	});

	function restart() {
		clearStoredOnboardingToken();
		goto('/onboarding');
	}
</script>

<OnboardingShell
	{steps}
	currentStep={3}
	title="Completing Google Calendar connection."
	intro="We're exchanging the secure code from Google. This should only take a moment."
>
	<div class="callback">
		{#if loading}
			<p class="loading">Completing Google Calendar connection…</p>
		{:else if pageError}
			<InfoBox title="Calendar connection did not finish">{pageError}</InfoBox>
			<div class="actions">
				<Button variant="secondary" on:click={() => goto('/onboarding/calendar')}>
					Back to calendar step
				</Button>
				<Button variant="primary" on:click={restart}>Restart onboarding</Button>
			</div>
		{/if}
	</div>
</OnboardingShell>

<style>
	.callback {
		display: grid;
		gap: var(--space-5);
		justify-items: start;
	}

	.loading {
		font-size: 1.1rem;
		color: var(--ink-2);
	}

	.actions {
		display: flex;
		gap: var(--space-4);
		min-width: 0;
	}

	@media (max-width: 540px) {
		.actions {
			flex-direction: column;
			align-items: stretch;
			width: 100%;
		}
	}
</style>
