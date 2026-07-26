<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import ConnectTile from '$lib/components/design/ConnectTile.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import {
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		disconnectCalendar,
		formatStepLabel,
		getCalendarStatus,
		getState,
		getStoredOnboardingToken,
		readTokenFromUrl,
		setStoredOnboardingToken,
		syncTokenInUrl,
		type CalendarStatus,
		type OnboardingState
	} from '$lib/api/onboarding';

	let loading = true;
	let disconnecting = false;
	let pageError = '';
	let state: OnboardingState = {
		has_profile: false,
		telegram_linked: false,
		calendar_connected: false,
		tasks: [],
		starter_templates: []
	};
	let sessionToken = '';
	let status: CalendarStatus = { available: false, connected: false };

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
		await loadCalendarStatus();
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

	async function loadCalendarStatus() {
		try {
			status = await getCalendarStatus(sessionToken);
			if (status.connected) {
				state.calendar_connected = true;
			}
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not load calendar status.';
		}
	}

	async function disconnect() {
		pageError = '';
		disconnecting = true;
		try {
			await disconnectCalendar(sessionToken);
			state.calendar_connected = false;
			await loadCalendarStatus();
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not disconnect calendar.';
		} finally {
			disconnecting = false;
		}
	}
</script>

{#if loading}
	<div class="loading">Loading calendar connection…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={3}
		title="Connect Google Calendar for smarter planning."
		intro="This step is optional. If you connect your calendar, Rahat can read your busy times and plan around them — nothing is ever written to your calendar."
	>
		<div class="connection-page">
			<p class="step-label">{formatStepLabel(steps[3])}</p>
			<h2 class="page-title">Connect Google Calendar</h2>

			{#if !status.available}
				<InfoBox title="Google Calendar is not available right now.">
					The server is not configured for Google OAuth, so this step is skipped. Calendar
					connection is completely optional — continue with your setup.
				</InfoBox>
				<div class="actions">
					<Button variant="secondary" on:click={() => goto('/onboarding/telegram')}>Back</Button>
					<Button variant="primary" on:click={() => goto('/onboarding/tasks')}>
						Continue to tasks
					</Button>
				</div>
			{:else if status.connected}
				<ConnectTile
					icon="📅"
					name="Google Calendar"
					subtitle="Read-only access to your busy times"
					connected={true}
				/>
				<InfoBox title="Google Calendar is connected.">
					Rahat can read your calendar to avoid scheduling tasks over meetings and other busy
					times.
				</InfoBox>
				<div class="actions">
					<Button variant="secondary" on:click={() => goto('/onboarding/telegram')}>Back</Button>
					<div class="row-actions">
						<Button variant="text" on:click={disconnect} disabled={disconnecting}>
							{disconnecting ? 'Disconnecting…' : 'Disconnect calendar'}
						</Button>
						<Button variant="primary" on:click={() => goto('/onboarding/tasks')}>
							Continue to tasks
						</Button>
					</div>
				</div>
			{:else}
				<ConnectTile
					icon="📅"
					name="Google Calendar"
					subtitle="Read-only access to your busy times"
					connected={false}
					href={status.auth_url}
					target="_blank"
					rel="noopener noreferrer"
				/>

				<InfoBox title="This connection is read-only.">
					<ul class="info-list">
						<li>Rahat cannot create, edit, or delete events.</li>
						<li>You can disconnect at any time.</li>
						<li>Skipping this step will not break anything.</li>
					</ul>
				</InfoBox>

				<div class="actions">
					<Button variant="secondary" on:click={() => goto('/onboarding/telegram')}>Back</Button>
					<Button variant="primary" on:click={() => goto('/onboarding/tasks')}>
						Skip and continue
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

	.info-list {
		margin: 0;
		padding-left: var(--space-5);
	}

	.info-list li {
		margin-bottom: var(--space-1);
	}

	.actions {
		display: flex;
		justify-content: space-between;
		gap: var(--space-4);
		padding-top: var(--space-2);
		min-width: 0;
	}

	.row-actions {
		display: flex;
		gap: var(--space-3);
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
		.actions,
		.row-actions {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
