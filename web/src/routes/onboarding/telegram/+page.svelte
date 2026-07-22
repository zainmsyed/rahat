<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import {
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		getState,
		getStoredOnboardingToken,
		getTelegramStatus,
		readTokenFromUrl,
		setStoredOnboardingToken,
		skipTelegram,
		syncTokenInUrl,
		type OnboardingState,
		type TelegramStatus
	} from '$lib/api/onboarding';

	let loading = true;
	let skipping = false;
	let pageError = '';
	let state: OnboardingState = { has_profile: false, telegram_linked: false, tasks: [], starter_templates: [] };
	let sessionToken = '';
	let status: TelegramStatus = { available: false, linked: false };
	let pollTimer: ReturnType<typeof setInterval> | null = null;

	$: steps = buildOnboardingSteps(state, !!sessionToken);

	onMount(async () => {
		sessionToken = readTokenFromUrl() || getStoredOnboardingToken();
		if (!sessionToken) {
			await goto('/onboarding');
			return;
		}
		setStoredOnboardingToken(sessionToken);
		syncTokenInUrl(sessionToken);
		await refreshState();
		if (!state.has_profile) {
			await goto('/onboarding/profile');
			return;
		}
		await loadTelegramStatus();
		startPolling();
	});

	onDestroy(() => {
		if (pollTimer) {
			clearInterval(pollTimer);
		}
	});

	async function refreshState() {
		loading = true;
		try {
			state = await getState(sessionToken);
		} catch {
			clearStoredOnboardingToken();
			await goto('/onboarding');
		} finally {
			loading = false;
		}
	}

	async function loadTelegramStatus() {
		try {
			status = await getTelegramStatus(sessionToken);
			if (status.linked) {
				state.telegram_linked = true;
			}
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not load Telegram status.';
		}
	}

	function startPolling() {
		if (pollTimer) {
			clearInterval(pollTimer);
		}
		pollTimer = setInterval(async () => {
			if (status.linked) {
				return;
			}
			await loadTelegramStatus();
			if (status.linked) {
				state.telegram_linked = true;
			}
		}, 2500);
	}

	async function skip() {
		pageError = '';
		skipping = true;
		try {
			await skipTelegram(sessionToken);
			await goto('/onboarding/tasks');
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not skip Telegram setup.';
		} finally {
			skipping = false;
		}
	}

	function qrCodeUrl(deepLink: string): string {
		return `https://api.qrserver.com/v1/create-qr-code/?size=200x200&data=${encodeURIComponent(deepLink)}`;
	}
</script>

{#if loading}
	<div class="loading">Loading Telegram connection…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={2}
		title="Connect Telegram for interactive reminders."
		intro="Telegram is the best way to get quick check-ins and reminders. One tap opens the bot, then send your short code to link this account."
	>
		<section class="panel active">
			<p class="label">Step 3 · Recommended</p>
			<h2>Connect Telegram</h2>

			{#if !status.available}
				<div class="fallback">
					<p><strong>Telegram is not configured right now.</strong></p>
					<p>
						You can continue with email only. If you added an email address on the previous screen,
						Rahat will use it for recaps.
					</p>
					<button type="button" on:click={() => goto('/onboarding/tasks')} disabled={skipping}>
						Continue with email only
					</button>
				</div>
			{:else if status.linked}
				<div class="success-banner">
					<h3>Telegram is connected.</h3>
					<p>You should receive a welcome message in your chat with @{status.bot_username}.</p>
					<button type="button" on:click={() => goto('/onboarding/tasks')}>Continue to tasks</button>
				</div>
			{:else}
				<div class="connection">
					<p>Tap the button to open Telegram. Your code is already filled in — just send the message.</p>

					{#if status.deep_link}
						<a class="deep-link" href={status.deep_link} target="_blank" rel="noopener noreferrer">
							Open @{status.bot_username} in Telegram
						</a>
					{/if}

					{#if status.code}
						<div class="code-box">
							<span class="code-label">Your code</span>
							<strong class="code">{status.code}</strong>
						</div>
					{/if}

					{#if status.deep_link}
						<div class="qr">
							<p>Or scan this QR code:</p>
							<img src={qrCodeUrl(status.deep_link)} alt="Telegram bot QR code" width="200" height="200" />
						</div>
					{/if}

					<p class="waiting">Waiting for your message…</p>
				</div>

				<div class="fallback">
					<p>Prefer not to use Telegram?</p>
					<button type="button" class="ghost" on:click={skip} disabled={skipping}>
						{skipping ? 'Skipping…' : 'Skip and continue with email only'}
					</button>
				</div>
			{/if}

			{#if pageError}
				<p class="error-banner">{pageError}</p>
			{/if}

			<div class="actions between">
				<button type="button" class="ghost" on:click={() => goto('/onboarding/profile')}>Back</button>
				{#if status.linked}
					<button type="button" on:click={() => goto('/onboarding/tasks')}>Continue to tasks</button>
				{/if}
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

	.connection,
	.fallback,
	.success-banner {
		display: grid;
		gap: 1rem;
		padding: 1rem;
		border-radius: 1rem;
		background: #fbfdff;
		border: 1px solid #dbe4ee;
	}

	.success-banner {
		background: #f0fdf4;
		border-color: #bbf7d0;
	}

	.deep-link {
		display: inline-block;
		padding: 0.85rem 1.1rem;
		border-radius: 999px;
		background: #2a6df4;
		color: white;
		font-weight: 700;
		text-decoration: none;
		text-align: center;
	}

	.code-box {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: 0.4rem;
		padding: 1rem;
		border-radius: 1rem;
		background: white;
		border: 2px dashed #cbd5e1;
	}

	.code-label {
		font-size: 0.85rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: #5d6b82;
	}

	.code {
		font-size: 2rem;
		letter-spacing: 0.15em;
		color: #14202c;
	}

	.qr {
		display: grid;
		place-items: center;
		gap: 0.5rem;
	}

	.qr p {
		margin: 0;
		color: #5d6b82;
	}

	.waiting {
		margin: 0;
		color: #5d6b82;
		font-style: italic;
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

	button:disabled {
		opacity: 0.7;
		cursor: wait;
	}

	.actions.between {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
	}

	.error-banner {
		color: #b42318;
		font-weight: 600;
		padding: 0.9rem 1rem;
		border-radius: 1rem;
		background: #fff1f0;
	}

	@media (max-width: 720px) {
		.actions.between {
			flex-direction: column;
		}
	}
</style>
