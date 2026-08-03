<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import {
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		finishOnboarding,
		formatStepLabel,
		formatTaskSummary,
		getState,
		getStoredOnboardingToken,
		readTokenFromUrl,
		type OnboardingFinishResult,
		type OnboardingState
	} from '$lib/api/onboarding';

	let loading = true;
	let finishing = false;
	let pageError = '';
	let finishResult: OnboardingFinishResult | null = null;
	let state: OnboardingState = {
		has_profile: false,
		telegram_linked: false,
		calendar_connected: false,
		tasks: [],
		starter_templates: []
	};
	let sessionToken = '';

	$: steps = buildOnboardingSteps(state, !!sessionToken, finishResult !== null);

	onMount(async () => {
		sessionToken = readTokenFromUrl() || getStoredOnboardingToken();
		if (!sessionToken) {
			await goto('/onboarding');
			return;
		}
		await refreshState();
	});

	async function refreshState() {
		loading = true;
		try {
			state = await getState(sessionToken);
			if (!state.has_profile) {
				await goto('/onboarding/profile');
				return;
			}
			if (state.tasks.length === 0) {
				await goto('/onboarding/tasks');
			}
		} catch {
			clearStoredOnboardingToken();
			await goto('/onboarding');
		} finally {
			loading = false;
		}
	}

	async function finishSetup() {
		pageError = '';
		finishing = true;
		try {
			finishResult = await finishOnboarding(sessionToken);
			clearStoredOnboardingToken();
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not finish onboarding.';
		} finally {
			finishing = false;
		}
	}

	async function continueAfterFinish() {
		await goto('/');
	}
</script>

{#if loading}
	<div class="loading">Loading your review step…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={5}
		finished={finishResult !== null}
		title="Review and finish."
		intro="Check your details below. When you press finish, Rahat will seed your first schedule and explain what happens next."
	>
		<section class="review-content" aria-labelledby="review-heading">
			<div class="section-intro">
				<p class="eyebrow">{formatStepLabel(steps[5])}</p>
				<h2 id="review-heading">Everything looks ready.</h2>
				<p>Review the essentials below. You can go back and edit anything before you finish.</p>
			</div>

			<div class="review-grid">
				<article class="summary-card">
					<div class="card-heading">
						<div>
							<p class="card-kicker">Profile</p>
							<h3>Your details</h3>
						</div>
						<span class="status-pill">Ready</span>
					</div>
					<dl class="detail-list">
						<div><dt>Name</dt><dd>{state.user?.display_name || 'Not saved yet'}</dd></div>
						<div><dt>Timezone</dt><dd>{state.user?.timezone || 'Not set'}</dd></div>
						<div><dt>Daily budget</dt><dd>{state.user?.daily_time_budget_minutes ?? '—'} minutes</dd></div>
						<div><dt>Email</dt><dd>{state.user?.email || 'Skipped for now'}</dd></div>
					</dl>
				</article>

				<article class="summary-card">
					<div class="card-heading">
						<div>
							<p class="card-kicker">Tasks</p>
							<h3>Your routines</h3>
						</div>
						<span class="status-pill">{state.tasks.length} ready</span>
					</div>
					{#if state.tasks.length > 0}
						<ul class="item-list">
							{#each state.tasks as task}
								<li>
									<strong>{task.name}</strong>
									<span>{formatTaskSummary(task)}</span>
								</li>
							{/each}
						</ul>
					{:else}
						<p class="muted">No routines have been added yet.</p>
					{/if}
				</article>

				<article class="summary-card">
					<div class="card-heading">
						<div>
							<p class="card-kicker">Calendar</p>
							<h3>Google Calendar</h3>
						</div>
						<span class:connected={state.calendar_connected} class="status-pill">
							{state.calendar_connected ? 'Connected' : 'Optional'}
						</span>
					</div>
					<p class="card-copy">
						{state.calendar_connected
							? 'Rahat can plan around the busy times on your calendar.'
							: 'You can connect a calendar later whenever you are ready.'}
					</p>
				</article>
			</div>

			{#if pageError}
				<div role="alert">
					<InfoBox title="We could not finish setup">{pageError}</InfoBox>
				</div>
			{/if}

			<div class="actions">
				<Button variant="secondary" on:click={() => goto('/onboarding/tasks')}>Back</Button>
				<Button
					variant="primary"
					disabled={state.tasks.length === 0 || finishing}
					on:click={finishSetup}
				>
					{finishing ? 'Seeding your first schedule…' : 'Finish onboarding'}
				</Button>
			</div>
		</section>

		{#if finishResult}
			<section class="success-card" aria-labelledby="success-heading">
				<p class="eyebrow">All set</p>
				<h2 id="success-heading">Your first schedule is ready.</h2>
				<div class="success-summary">
					{#each finishResult.summary as line}
						<p>{line}</p>
					{/each}
				</div>

				<div class="success-grid">
					<div class="summary-card compact">
						<p class="card-kicker">First schedule</p>
						<h3>{finishResult.plan_date}</h3>
						<dl class="detail-list">
							<div><dt>Scheduled now</dt><dd>{finishResult.scheduled_count}</dd></div>
							<div><dt>Moved later</dt><dd>{finishResult.overflowed_count}</dd></div>
							<div><dt>Skipped</dt><dd>{finishResult.skipped_count}</dd></div>
						</dl>
					</div>
					<div class="summary-card compact">
						<p class="card-kicker">What happens next</p>
						<h3>Keep building momentum.</h3>
						<p class="card-copy">Rahat will help you plan your routines. You can review your schedule and mark progress whenever you return.</p>
					</div>
				</div>

				{#if finishResult.scheduled_items.length > 0}
					<div class="schedule-block">
						<h3>Today's planned items</h3>
						<ul class="item-list">
							{#each finishResult.scheduled_items as item}
								<li>
									<strong>{item.name}</strong>
									<span>{item.window}{item.ready_at ? ` · ready ${new Date(item.ready_at).toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })}` : ''}</span>
								</li>
							{/each}
						</ul>
					</div>
				{/if}

				<p class="delivery-note">
					{#if finishResult.telegram_delivered}
						A summary was also sent to your linked Telegram chat.
				{:else if state.telegram_linked}
						We could not deliver the Telegram summary right now, but your schedule is saved. Send <code>/edit</code> in Telegram whenever you need routine settings.
				{:else}
						Telegram was not linked, so no message was sent. You can still use Rahat in the browser.
				{/if}
				</p>
				<div class="success-actions">
					<Button variant="primary" on:click={continueAfterFinish}>Continue to Rahat</Button>
				</div>
			</section>
		{/if}
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

	.review-content,
	.success-card {
		display: grid;
		gap: var(--space-6);
		min-width: 0;
	}

	.section-intro {
		display: grid;
		gap: var(--space-2);
	}

	.eyebrow,
	.card-kicker {
		font-size: 11px;
		letter-spacing: 0.16em;
		text-transform: uppercase;
		color: var(--primary-2);
		font-weight: 600;
	}

	.section-intro h2,
	.success-card h2 {
		font-family: var(--font-display);
		font-size: 28px;
		line-height: 1.15;
		font-weight: 400;
		color: var(--ink);
	}

	.section-intro > p:last-child,
	.card-copy,
	.muted,
	.success-summary p,
	.delivery-note {
		color: var(--ink-2);
		font-size: 14px;
		line-height: 1.55;
	}

	.review-grid,
	.success-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--space-4);
		min-width: 0;
	}

	.summary-card {
		display: grid;
		align-content: start;
		gap: var(--space-4);
		min-width: 0;
		padding: var(--space-5);
		background: var(--primary-bg);
		border: 1px solid var(--line-soft);
		border-radius: var(--radius-xl);
	}

	.summary-card.compact {
		background: var(--paper);
		border-color: var(--line);
	}

	.card-heading {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--space-3);
		min-width: 0;
	}

	.card-heading h3,
	.success-card h3,
	.schedule-block h3 {
		font-size: 17px;
		line-height: 1.3;
		font-weight: 600;
		color: var(--ink);
		margin-top: var(--space-1);
	}

	.status-pill {
		flex-shrink: 0;
		padding: 4px 9px;
		border-radius: var(--radius-pill);
		background: var(--secondary);
		color: var(--ink-2);
		font-size: 11px;
		font-weight: 600;
		white-space: nowrap;
	}

	.status-pill.connected {
		background: var(--primary-soft);
		color: var(--primary-2);
	}

	.detail-list {
		display: grid;
		gap: var(--space-3);
		min-width: 0;
	}

	.detail-list div {
		display: grid;
		grid-template-columns: minmax(72px, 0.7fr) minmax(0, 1.3fr);
		gap: var(--space-3);
		min-width: 0;
	}

	dt {
		color: var(--ink-3);
		font-size: 13px;
	}

	dd {
		color: var(--ink);
		font-size: 13px;
		font-weight: 500;
		min-width: 0;
		overflow-wrap: anywhere;
	}

	.item-list {
		display: grid;
		gap: var(--space-3);
		list-style: none;
		min-width: 0;
	}

	.item-list li {
		display: grid;
		gap: var(--space-1);
		min-width: 0;
		padding-bottom: var(--space-3);
		border-bottom: 1px solid var(--line-soft);
	}

	.item-list li:last-child {
		padding-bottom: 0;
		border-bottom: 0;
	}

	.item-list strong {
		color: var(--ink);
		font-size: 14px;
		overflow-wrap: anywhere;
	}

	.item-list span {
		color: var(--ink-3);
		font-size: 13px;
	}

	.actions {
		display: flex;
		justify-content: space-between;
		gap: var(--space-4);
		min-width: 0;
	}

	.success-card {
		padding-top: var(--space-6);
		border-top: 1px solid var(--line);
	}

	.success-summary {
		display: grid;
		gap: var(--space-2);
		padding: var(--space-4);
		background: var(--primary-bg);
		border: 1px solid var(--primary-soft);
		border-radius: var(--radius-lg);
	}

	.schedule-block {
		display: grid;
		gap: var(--space-3);
	}

	.delivery-note {
		padding-top: var(--space-2);
	}

	.success-actions {
		display: flex;
		justify-content: flex-end;
		padding-top: var(--space-2);
	}

	code {
		font-family: var(--font-mono);
		font-size: 0.9em;
		color: var(--primary-3);
	}

	@media (max-width: 540px) {
		.review-grid,
		.success-grid {
			grid-template-columns: 1fr;
		}

		.actions {
			flex-direction: column-reverse;
		}

		.actions :global(button) {
			width: 100%;
		}
	}
</style>
