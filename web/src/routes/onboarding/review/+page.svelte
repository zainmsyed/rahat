<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import {
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		finishOnboarding,
		formatTaskSummary,
		getState,
		getStoredOnboardingToken,
		readTokenFromUrl,
		setStoredOnboardingToken,
		syncTokenInUrl,
		type OnboardingFinishResult,
		type OnboardingState
	} from '$lib/api/onboarding';

	let loading = true;
	let finishing = false;
	let pageError = '';
	let finishResult: OnboardingFinishResult | null = null;
	let state: OnboardingState = { has_profile: false, tasks: [], starter_templates: [] };
	let sessionToken = '';

	$: steps = buildOnboardingSteps(state, !!sessionToken, finishResult !== null);

	onMount(async () => {
		sessionToken = readTokenFromUrl() || getStoredOnboardingToken();
		if (!sessionToken) {
			await goto('/onboarding');
			return;
		}
		setStoredOnboardingToken(sessionToken);
		syncTokenInUrl(sessionToken);
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
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not finish onboarding.';
		} finally {
			finishing = false;
		}
	}
</script>

{#if loading}
	<div class="loading">Loading your review step…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={3}
		finished={finishResult !== null}
		title="Review and finish."
		intro="Check your details below. When you press finish, Rahat will seed your first schedule and explain what happens next."
	>
		<section class="panel active">
			<p class="label">Step 4 · Required</p>
			<h2>Review and finish</h2>
			<p>Make sure your profile and tasks look right. You can still go back and edit anything before you finish.</p>

			<div class="review-grid">
				<article>
					<h3>Your profile</h3>
					<p><strong>Name:</strong> {state.user?.display_name || 'Not saved yet'}</p>
					<p><strong>Timezone:</strong> {state.user?.timezone}</p>
					<p><strong>Daily budget:</strong> {state.user?.daily_time_budget_minutes} minutes</p>
					<p><strong>Email:</strong> {state.user?.email || 'Skipped for now'}</p>
				</article>
				<article>
					<h3>Your tasks</h3>
					<p>{state.tasks.length} task(s) ready for Rahat to schedule.</p>
					<p>No raw IDs or technical setup is needed from you here.</p>
					{#if state.tasks.length > 0}
						<ul class="task-list">
							{#each state.tasks as task}
								<li>
									<strong>{task.name}</strong>
									<span>
										{formatTaskSummary(task)}
									</span>
								</li>
							{/each}
						</ul>
					{/if}
				</article>
			</div>

			{#if pageError}
				<p class="error-banner">{pageError}</p>
			{/if}

			<div class="actions between">
				<button type="button" class="ghost" on:click={() => goto('/onboarding/tasks')}>Back</button>
				<button type="button" on:click={finishSetup} disabled={state.tasks.length === 0 || finishing}>
					{finishing ? 'Seeding your first schedule…' : 'Finish onboarding'}
				</button>
			</div>
		</section>

		{#if finishResult}
			<section class="success">
				<h3>You are set up.</h3>
				{#each finishResult.summary as line}
					<p>{line}</p>
				{/each}

				<div class="review-grid">
					<article>
						<h4>First schedule</h4>
						<p><strong>Date:</strong> {finishResult.plan_date}</p>
						<p><strong>Scheduled now:</strong> {finishResult.scheduled_count}</p>
						<p><strong>Moved later:</strong> {finishResult.overflowed_count}</p>
						<p><strong>Skipped:</strong> {finishResult.skipped_count}</p>
					</article>
					<article>
						<h4>What happens next</h4>
						<p>Rahat has enough information to start planning your tasks.</p>
						<p>You can come back later to review your schedule and mark progress.</p>
					</article>
				</div>

				{#if finishResult.scheduled_items.length > 0}
					<h4>Today's planned items</h4>
					<ul class="schedule-list">
						{#each finishResult.scheduled_items as item}
							<li>
								<strong>{item.name}</strong>
								<span>{item.window}{item.ready_at ? ` · ready ${new Date(item.ready_at).toLocaleTimeString([], { hour: 'numeric', minute: '2-digit' })}` : ''}</span>
							</li>
						{/each}
					</ul>
				{/if}
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
	}

	.panel {
		display: grid;
		gap: 1rem;
		border: 2px solid #d6e4ff;
	}

	.review-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.review-grid article {
		padding: 1rem;
		border: 1px solid #dbe4ee;
		border-radius: 1rem;
		background: #fbfdff;
	}

	.actions.between {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
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

	.error-banner {
		color: #b42318;
		font-weight: 600;
		padding: 0.9rem 1rem;
		border-radius: 1rem;
		background: #fff1f0;
	}

	.task-list,
	.schedule-list {
		padding-left: 1.1rem;
	}

	.task-list li,
	.schedule-list li {
		margin-bottom: 0.55rem;
	}

	.task-list span,
	.schedule-list span {
		display: block;
		color: #5d6b82;
	}

	@media (max-width: 720px) {
		.review-grid {
			grid-template-columns: 1fr;
		}

		.actions.between {
			flex-direction: column;
		}
	}
</style>
