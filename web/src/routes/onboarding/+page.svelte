<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import {
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		createSession,
		getState,
		getStoredOnboardingToken,
		nextOnboardingPath,
		readInviteCodeFromUrl,
		readTokenFromUrl,
		setStoredOnboardingToken,
		syncTokenInUrl,
		type OnboardingState
	} from '$lib/api/onboarding';

	let loading = true;
	let inviteCode = '';
	let pageError = '';
	let state: OnboardingState = { has_profile: false, telegram_linked: false, tasks: [], starter_templates: [] };

	$: steps = buildOnboardingSteps(state, false);

	onMount(async () => {
		inviteCode = readInviteCodeFromUrl();
		const sessionToken = readTokenFromUrl() || getStoredOnboardingToken();
		if (!sessionToken) {
			loading = false;
			return;
		}
		try {
			setStoredOnboardingToken(sessionToken);
			syncTokenInUrl(sessionToken);
			state = await getState(sessionToken);
			await goto(nextOnboardingPath(state));
		} catch {
			clearStoredOnboardingToken();
			loading = false;
		}
	});

	async function startSession() {
		pageError = '';
		if (!inviteCode.trim()) {
			pageError = 'Please enter your invite code to begin.';
			return;
		}
		loading = true;
		try {
			const session = await createSession(inviteCode.trim());
			setStoredOnboardingToken(session.token);
			syncTokenInUrl(session.token);
			state = await getState(session.token);
			await goto(nextOnboardingPath(state));
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not start onboarding.';
			loading = false;
		}
	}
</script>

{#if loading}
	<div class="loading">Opening your onboarding steps…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={0}
		title="Let's set up your first calm, clear plan."
		intro="Each step explains exactly what to do next. Required items are labeled, and the email field later in setup is optional."
	>
		<section class="panel active">
			<p class="label">Step 1 · Required</p>
			<h2>Start with your invite code</h2>
			<p>Type the invite code you received, then press the button to begin your guided setup.</p>

			<label>
				<span>Invite code *</span>
				<input bind:value={inviteCode} placeholder="Example: rahat-beta" />
			</label>

			{#if pageError}
				<p class="error-banner">{pageError}</p>
			{/if}

			<div class="actions">
				<button type="button" on:click={startSession}>Start onboarding</button>
			</div>
		</section>
	</OnboardingShell>
{/if}

<style>
	.loading {
		min-height: 100vh;
		display: grid;
		place-items: center;
		font-size: 1.1rem;
	}

	.panel {
		display: grid;
		gap: 1rem;
		border: 2px solid #d6e4ff;
	}

	label {
		display: grid;
		gap: 0.4rem;
		font-weight: 600;
	}

	span {
		font-size: 0.95rem;
	}

	input,
	button {
		font: inherit;
	}

	input {
		padding: 0.8rem 0.9rem;
		border-radius: 0.85rem;
		border: 1px solid #cbd5e1;
	}

	button {
		padding: 0.85rem 1.1rem;
		border-radius: 999px;
		border: none;
		background: #2a6df4;
		color: white;
		font-weight: 700;
		cursor: pointer;
	}

	.actions {
		display: flex;
	}

	.error-banner {
		color: #b42318;
		font-weight: 600;
		padding: 0.9rem 1rem;
		border-radius: 1rem;
		background: #fff1f0;
	}
</style>
