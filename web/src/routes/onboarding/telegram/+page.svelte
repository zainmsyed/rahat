<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount, onDestroy } from 'svelte';
	import QRCode from 'qrcode';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import ConnectTile from '$lib/components/design/ConnectTile.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import {
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		formatStepLabel,
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
	let state: OnboardingState = {
		has_profile: false,
		telegram_linked: false,
		calendar_connected: false,
		tasks: [],
		starter_templates: []
	};
	let sessionToken = '';
	let status: TelegramStatus = { available: false, linked: false };
	let pollTimer: ReturnType<typeof setInterval> | null = null;
	let qrCodeDataUrl = '';

	$: steps = buildOnboardingSteps(state, !!sessionToken);

	$: if (status.linked && pollTimer) {
		clearInterval(pollTimer);
		pollTimer = null;
	}

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
			if (status.deep_link) {
				qrCodeDataUrl = await QRCode.toDataURL(status.deep_link, { width: 200, margin: 2 });
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
			await goto('/onboarding/calendar');
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not skip Telegram setup.';
		} finally {
			skipping = false;
		}
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
		<div class="connection-page">
			<p class="step-label">{formatStepLabel(steps[2])}</p>
			<h2 class="page-title">Connect Telegram</h2>

			{#if !status.available}
				<InfoBox title="Telegram is not configured right now.">
					You can continue with email only. If you added an email address on the previous
					screen, Rahat will use it for recaps.
				</InfoBox>
				<div class="actions">
					<Button variant="secondary" on:click={() => goto('/onboarding/profile')}>Back</Button>
					<Button variant="primary" on:click={() => goto('/onboarding/calendar')} disabled={skipping}>
						Continue with email only
					</Button>
				</div>
			{:else if status.linked}
				<ConnectTile
					icon="✈️"
					name="Telegram"
					subtitle="Interactive reminders and check-ins"
					connected={true}
				/>
				<InfoBox title="Telegram is connected.">
					You should receive a welcome message in your chat with @{status.bot_username}.
				</InfoBox>
				<div class="actions">
					<Button variant="secondary" on:click={() => goto('/onboarding/profile')}>Back</Button>
					<Button variant="primary" on:click={() => goto('/onboarding/calendar')}>
						Continue to calendar
					</Button>
				</div>
			{:else}
				<ConnectTile
					icon="✈️"
					name="Telegram"
					subtitle="Interactive reminders and check-ins"
					connected={false}
					href={status.deep_link}
					target="_blank"
					rel="noopener noreferrer"
				/>

				<InfoBox title="How to connect">
					Tap the tile to open Telegram. Your code is already filled in — just send the
					message. Waiting for your message…
				</InfoBox>

				{#if status.code}
					<div class="code-box">
						<span class="code-label">Your code</span>
						<strong class="code">{status.code}</strong>
					</div>
				{/if}

				{#if qrCodeDataUrl}
					<div class="qr">
						<p>Or scan this QR code:</p>
						<img src={qrCodeDataUrl} alt="Telegram bot QR code" width="180" height="180" />
					</div>
				{/if}

				<div class="actions">
					<Button variant="secondary" on:click={() => goto('/onboarding/profile')}>Back</Button>
					<Button variant="text" on:click={skip} disabled={skipping}>
						{skipping ? 'Skipping…' : 'Skip and continue with email only'}
					</Button>
				</div>
			{/if}

			{#if pageError}
				<p class="error-banner" role="alert">{pageError}</p>
			{/if}
		</div>
	</OnboardingShell>
{/if}

<style>
	.loading {
		min-height: 100vh;
		display: grid;
		place-items: center;
		font-size: 1.1rem;
		color: var(--ink-2);
	}

	.connection-page {
		display: grid;
		gap: var(--space-5);
	}

	.step-label {
		font-size: 11px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--primary-2);
		font-weight: 600;
	}

	.page-title {
		font-family: var(--font-display);
		font-size: 28px;
		line-height: 1.1;
		font-weight: 400;
		color: var(--ink);
		margin: 0;
	}

	.code-box {
		display: flex;
		flex-direction: column;
		align-items: center;
		gap: var(--space-2);
		padding: var(--space-5);
		border-radius: var(--radius-lg);
		background: var(--bg-soft);
		border: 2px dashed var(--line);
	}

	.code-label {
		font-size: 12px;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.05em;
		color: var(--ink-3);
	}

	.code {
		font-size: 2rem;
		letter-spacing: 0.15em;
		color: var(--ink);
	}

	.qr {
		display: grid;
		place-items: center;
		gap: var(--space-2);
		font-size: 13px;
		color: var(--ink-3);
	}

	.actions {
		display: flex;
		justify-content: space-between;
		gap: var(--space-4);
		padding-top: var(--space-2);
		min-width: 0;
	}

	.error-banner {
		color: var(--rose);
		font-weight: 600;
		padding: var(--space-4);
		border-radius: var(--radius-lg);
		background: var(--rose-soft);
	}

	@media (max-width: 540px) {
		.actions {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
