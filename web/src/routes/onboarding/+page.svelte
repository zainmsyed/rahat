<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import Input from '$lib/components/design/Input.svelte';
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
	let submitting = false;
	let inviteCode = '';
	let pageError = '';
	let state: OnboardingState = {
		has_profile: false,
		telegram_linked: false,
		calendar_connected: false,
		tasks: [],
		starter_templates: []
	};

	$: steps = buildOnboardingSteps(state, false);

	onMount(async () => {
		inviteCode = readInviteCodeFromUrl();
		const sessionToken = readTokenFromUrl() || getStoredOnboardingToken();
		if (!sessionToken) {
			if (inviteCode.trim()) {
				await startSession();
				return;
			}
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
		submitting = true;
		try {
			const session = await createSession(inviteCode.trim());
			setStoredOnboardingToken(session.token);
			syncTokenInUrl(session.token);
			state = await getState(session.token);
			await goto(nextOnboardingPath(state));
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not start onboarding.';
		} finally {
			submitting = false;
		}
	}

	function handleSubmit(event: Event) {
		event.preventDefault();
		startSession();
	}
</script>

{#if loading}
	<div class="loading">Opening your onboarding steps…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={0}
		title="Start with your invite code"
		intro="If you opened a setup link, Rahat starts automatically. Otherwise, type the invite code you received and press the button."
	>
		<form class="invite-form" on:submit={handleSubmit}>
			<Input
				id="inviteCode"
				label="Invite code"
				placeholder="Example: rahat-beta"
				bind:value={inviteCode}
				required
			/>

			{#if pageError}
				<div class="error-banner" role="alert">{pageError}</div>
			{/if}

			<Button type="submit" variant="primary" fullWidth disabled={submitting}>
				{submitting ? 'Starting…' : 'Start onboarding'}
			</Button>
		</form>

		<div class="help">
			<InfoBox title="Where do I get a code?">
				Your onboarding operator sent it to you. It usually looks like <strong>rahat-beta</strong>.
			</InfoBox>
		</div>
	</OnboardingShell>
{/if}

<style>
	.loading {
		min-height: 100vh;
		display: grid;
		place-items: center;
		font-size: 15.5px;
		color: var(--ink-2);
	}

	.invite-form {
		display: grid;
		gap: var(--space-5);
	}

	.error-banner {
		padding: var(--space-3) var(--space-4);
		background: var(--rose-soft);
		color: var(--rose);
		border-radius: var(--radius-lg);
		font-size: 13.5px;
		font-weight: 500;
	}

	.help {
		margin-top: var(--space-6);
	}
</style>
