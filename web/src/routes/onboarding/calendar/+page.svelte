<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
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
		<section class="panel active">
			<p class="label">{formatStepLabel(steps[3])}</p>
			<h2>Connect Google Calendar</h2>

			{#if !status.available}
				<div class="fallback">
					<p><strong>Google Calendar is not available right now.</strong></p>
					<p>
						The server is not configured for Google OAuth, so this step is skipped. Calendar
						connection is completely optional — continue with your setup.
					</p>
					<button type="button" on:click={() => goto('/onboarding/tasks')}>Continue to tasks</button>
				</div>
			{:else if status.connected}
				<div class="success-banner">
					<h3>Google Calendar is connected.</h3>
					<p>
						Rahat can read your calendar to avoid scheduling tasks over meetings and other busy
						times.
					</p>
					<div class="row-actions">
						<button type="button" class="ghost" on:click={disconnect} disabled={disconnecting}>
							{disconnecting ? 'Disconnecting…' : 'Disconnect calendar'}
						</button>
						<button type="button" on:click={() => goto('/onboarding/tasks')}>Continue to tasks</button>
					</div>
				</div>
			{:else}
				<div class="connection">
					<p>
						Connecting is quick and read-only. Rahat asks only to view your calendar events so it
						can leave those slots free when it plans your tasks.
					</p>
					<ul>
						<li>Rahat cannot create, edit, or delete events.</li>
						<li>You can disconnect at any time.</li>
						<li>Skipping this step will not break anything.</li>
					</ul>

					{#if status.auth_url}
						<a class="connect-button" href={status.auth_url}>Connect Google Calendar</a>
					{/if}

					<div class="fallback inline">
						<p>Prefer not to connect your calendar?</p>
						<button type="button" class="ghost" on:click={() => goto('/onboarding/tasks')}>
							Skip and continue
						</button>
					</div>
				</div>
			{/if}

			{#if pageError}
				<p class="error-banner">{pageError}</p>
			{/if}

			<div class="actions between">
				<button type="button" class="ghost" on:click={() => goto('/onboarding/telegram')}>Back</button>
				{#if status.connected}
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

	.fallback.inline {
		background: transparent;
		border: none;
		padding: 0;
	}

	.connect-button {
		display: inline-block;
		padding: 0.85rem 1.1rem;
		border-radius: 999px;
		background: #2a6df4;
		color: white;
		font-weight: 700;
		text-decoration: none;
		text-align: center;
	}

	.row-actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
	}

	ul {
		margin: 0;
		padding-left: 1.25rem;
	}

	li {
		margin-bottom: 0.4rem;
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
		.actions.between,
		.row-actions {
			flex-direction: column;
		}
	}
</style>
