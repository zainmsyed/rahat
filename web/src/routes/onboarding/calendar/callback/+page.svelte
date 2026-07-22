<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import {
		clearStoredOnboardingToken,
		getStoredOnboardingToken,
		syncTokenInUrl
	} from '$lib/api/onboarding';

	const apiBaseUrl = import.meta.env.VITE_API_BASE_URL || 'http://localhost:8080';

	let loading = true;
	let pageError = '';
	let sessionToken = '';

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

<div class="callback">
	{#if loading}
		<p class="loading">Completing Google Calendar connection…</p>
	{:else if pageError}
		<div class="error-panel">
			<h1>Calendar connection did not finish</h1>
			<p class="error-banner">{pageError}</p>
			<div class="actions">
				<button type="button" class="ghost" on:click={() => goto('/onboarding/calendar')}
					>Back to calendar step</button
				>
				<button type="button" on:click={restart}>Restart onboarding</button>
			</div>
		</div>
	{/if}
</div>

<style>
	.callback {
		min-height: 100vh;
		display: grid;
		place-items: center;
		padding: 1.5rem;
		font-family: Inter, system-ui, sans-serif;
		background: #f4f7fb;
		color: #14202c;
	}

	.loading {
		font-size: 1.1rem;
	}

	.error-panel {
		max-width: 520px;
		padding: 1.5rem;
		border-radius: 1.4rem;
		background: white;
		box-shadow: 0 16px 40px rgba(20, 32, 44, 0.08);
	}

	.error-panel h1 {
		margin: 0 0 1rem;
		font-size: 1.4rem;
	}

	.error-banner {
		color: #b42318;
		font-weight: 600;
		padding: 0.9rem 1rem;
		border-radius: 1rem;
		background: #fff1f0;
	}

	.actions {
		display: flex;
		gap: 1rem;
		margin-top: 1rem;
	}

	button {
		padding: 0.85rem 1.1rem;
		border-radius: 999px;
		border: none;
		background: #2a6df4;
		color: white;
		font: inherit;
		font-weight: 700;
		cursor: pointer;
	}

	button.ghost {
		background: white;
		color: #14202c;
		border: 1px solid #cbd5e1;
	}

	@media (max-width: 720px) {
		.actions {
			flex-direction: column;
		}
	}
</style>
