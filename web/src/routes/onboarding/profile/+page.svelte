<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import Input from '$lib/components/design/Input.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import {
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		getState,
		getStoredOnboardingToken,
		readTokenFromUrl,
		saveProfile,
		setStoredOnboardingToken,
		syncTokenInUrl,
		type OnboardingProfile,
		type OnboardingState
	} from '$lib/api/onboarding';

	let loading = true;
	let savingProfile = false;
	let profileSaveError = '';
	let profileErrors = { display_name: '', timezone: '', daily_time_budget_minutes: '', email: '' };
	let state: OnboardingState = {
		has_profile: false,
		telegram_linked: false,
		calendar_connected: false,
		tasks: [],
		starter_templates: []
	};
	let sessionToken = '';
	let profileDraft: OnboardingProfile = {
		display_name: '',
		timezone: 'UTC',
		daily_time_budget_minutes: 45,
		email: ''
	};

	const budgetTicks = [15, 60, 120, 240, 480];

	$: steps = buildOnboardingSteps(state, !!sessionToken);

	onMount(async () => {
		const detectedTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
		if (detectedTimezone) {
			profileDraft = { ...profileDraft, timezone: detectedTimezone };
		}
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
			if (state.user) {
				profileDraft = { ...state.user };
			}
		} catch {
			clearStoredOnboardingToken();
			await goto('/onboarding');
		} finally {
			loading = false;
		}
	}

	function validateProfile() {
		profileErrors = { display_name: '', timezone: '', daily_time_budget_minutes: '', email: '' };
		if (!profileDraft.display_name.trim()) {
			profileErrors.display_name = 'Your name is required.';
		}
		if (!profileDraft.timezone.trim()) {
			profileErrors.timezone = 'Choose your timezone.';
		}
		if (
			profileDraft.daily_time_budget_minutes < 15 ||
			profileDraft.daily_time_budget_minutes > 480
		) {
			profileErrors.daily_time_budget_minutes = 'Use a number between 15 and 480 minutes.';
		}
		if (profileDraft.email && !/.+@.+\..+/.test(profileDraft.email)) {
			profileErrors.email = 'If you add an email, please use a valid one.';
		}
		return Object.values(profileErrors).every((message) => !message);
	}

	async function submitProfile() {
		profileSaveError = '';
		if (!validateProfile()) {
			return;
		}
		savingProfile = true;
		try {
			await saveProfile(sessionToken, profileDraft);
			await goto('/onboarding/telegram');
		} catch (error) {
			profileSaveError = error instanceof Error ? error.message : 'Could not save your profile.';
		} finally {
			savingProfile = false;
		}
	}
</script>

{#if loading}
	<div class="loading">Loading your profile step…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={1}
		title="Tell Rahat about you."
		intro="Save the basics first: your name, your timezone, how many minutes you can usually spend on tasks in a day, and an optional email for recaps."
	>
		<div class="profile-form">
			<div class="form-grid">
				<Input
					id="display_name"
					label="Name"
					placeholder="Your name"
					required
					bind:value={profileDraft.display_name}
					error={profileErrors.display_name}
				/>
				<Input
					id="timezone"
					label="Timezone"
					placeholder="Example: America/New_York"
					required
					bind:value={profileDraft.timezone}
					error={profileErrors.timezone}
				/>
			</div>

			<div class="slider-field">
				<label class="slider-label" for="daily-budget">
					Daily task-time budget <span aria-hidden="true">*</span>
				</label>
				<input
					id="daily-budget"
					type="range"
					min="15"
					max="480"
					step="15"
					bind:value={profileDraft.daily_time_budget_minutes}
					aria-describedby="budget-summary budget-hint"
				/>
				<div class="slider-ticks" aria-hidden="true">
					{#each budgetTicks as tick}
						<span>{tick}</span>
					{/each}
				</div>
				<div id="budget-summary" class="summary-box" role="status" aria-live="polite">
					<span class="budget-value">{profileDraft.daily_time_budget_minutes}</span>
					<span class="budget-unit">minutes per day</span>
					<span id="budget-hint" class="budget-hint">
						Friendly default: 45 minutes. Move the slider to change it.
					</span>
				</div>
				{#if profileErrors.daily_time_budget_minutes}
					<p class="error-text">{profileErrors.daily_time_budget_minutes}</p>
				{/if}
			</div>

			<Input
				id="email"
				label="Email for recaps (optional)"
				type="email"
				placeholder="Only if you want daily recaps later"
				bind:value={profileDraft.email}
				error={profileErrors.email}
			/>

			{#if profileSaveError}
				<p class="error-banner" role="alert">{profileSaveError}</p>
			{/if}

			<div class="actions">
				<Button variant="secondary" on:click={() => goto('/onboarding')}>Back</Button>
				<Button variant="primary" disabled={savingProfile} on:click={submitProfile}>
					{savingProfile ? 'Saving…' : 'Save and continue'}
				</Button>
			</div>
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

	.profile-form {
		display: grid;
		gap: var(--space-5);
	}

	.form-grid {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.slider-field {
		display: grid;
		gap: var(--space-3);
	}

	.slider-label {
		display: block;
		font-size: 13px;
		font-weight: 500;
		color: var(--ink-2);
	}

	.slider-label span {
		color: var(--rose);
	}

	input[type='range'] {
		-webkit-appearance: none;
		appearance: none;
		width: 100%;
		height: 8px;
		border-radius: var(--radius-pill);
		background: var(--primary-track);
		outline: none;
		cursor: pointer;
	}

	input[type='range']:focus-visible {
		box-shadow: 0 0 0 4px var(--primary-glow);
	}

	input[type='range']::-webkit-slider-thumb {
		-webkit-appearance: none;
		appearance: none;
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: var(--primary);
		border: 3px solid var(--paper);
		box-shadow: 0 2px 6px rgba(31, 29, 26, 0.15);
		margin-top: -8px;
		transition: transform 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
	}

	input[type='range']::-moz-range-thumb {
		width: 24px;
		height: 24px;
		border-radius: 50%;
		background: var(--primary);
		border: 3px solid var(--paper);
		box-shadow: 0 2px 6px rgba(31, 29, 26, 0.15);
		transition: transform 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
	}

	input[type='range']::-webkit-slider-thumb:hover {
		transform: scale(1.08);
		box-shadow: 0 0 0 6px var(--primary-soft);
	}

	input[type='range']::-moz-range-thumb:hover {
		transform: scale(1.08);
		box-shadow: 0 0 0 6px var(--primary-soft);
	}

	input[type='range']::-webkit-slider-runnable-track {
		width: 100%;
		height: 8px;
		border-radius: var(--radius-pill);
		background: var(--primary-track);
	}

	input[type='range']::-moz-range-track {
		width: 100%;
		height: 8px;
		border-radius: var(--radius-pill);
		background: var(--primary-track);
	}

	.slider-ticks {
		display: flex;
		justify-content: space-between;
		font-size: 12px;
		color: var(--ink-3);
		padding: 0 4px;
	}

	.summary-box {
		display: grid;
		gap: var(--space-1);
		justify-items: start;
		padding: var(--space-5);
		background: var(--primary-bg);
		border: 1px solid var(--primary-soft);
		border-radius: var(--radius-lg);
	}

	.budget-value {
		font-family: var(--font-display);
		font-size: 60px;
		line-height: 1;
		color: var(--primary-2);
	}

	.budget-unit {
		font-size: 15px;
		font-weight: 500;
		color: var(--ink-2);
	}

	.budget-hint {
		font-size: 13px;
		color: var(--ink-3);
	}

	.error-text {
		font-size: 13px;
		color: var(--rose);
		font-weight: 500;
	}

	.error-banner {
		color: var(--rose);
		font-weight: 600;
		padding: var(--space-4);
		border-radius: var(--radius-lg);
		background: var(--rose-soft);
	}

	.actions {
		display: flex;
		justify-content: space-between;
		gap: var(--space-4);
		padding-top: var(--space-2);
	}

	@media (max-width: 540px) {
		.form-grid {
			grid-template-columns: 1fr;
		}

		.actions {
			flex-direction: column;
		}
	}
</style>
