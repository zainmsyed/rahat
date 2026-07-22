<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
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
	let state: OnboardingState = { has_profile: false, telegram_linked: false, tasks: [], starter_templates: [] };
	let sessionToken = '';
	let profileDraft: OnboardingProfile = {
		display_name: '',
		timezone: 'UTC',
		daily_time_budget_minutes: 45,
		email: ''
	};

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
		if (profileDraft.daily_time_budget_minutes < 15 || profileDraft.daily_time_budget_minutes > 480) {
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
		<section class="panel active">
			<p class="label">Step 2 · Required</p>
			<h2>Tell Rahat about you</h2>
			<p>Everything on this screen is straightforward. The email field is optional, so you can leave it blank.</p>

			<div class="grid two">
				<label>
					<span>Name *</span>
					<input bind:value={profileDraft.display_name} placeholder="What should Rahat call you?" />
					{#if profileErrors.display_name}<small class="field-error">{profileErrors.display_name}</small>{/if}
				</label>
				<label>
					<span>Timezone *</span>
					<input bind:value={profileDraft.timezone} placeholder="Example: America/New_York" />
					{#if profileErrors.timezone}<small class="field-error">{profileErrors.timezone}</small>{/if}
				</label>
			</div>

			<div class="grid two">
				<label>
					<span>Daily task-time budget in minutes *</span>
					<input bind:value={profileDraft.daily_time_budget_minutes} type="number" min="15" max="480" />
					<small>Friendly default: 45 minutes.</small>
					{#if profileErrors.daily_time_budget_minutes}<small class="field-error">{profileErrors.daily_time_budget_minutes}</small>{/if}
				</label>
				<label>
					<span>Email for recaps (optional)</span>
					<input bind:value={profileDraft.email} type="email" placeholder="Only if you want daily recaps later" />
					<small>You can leave this blank.</small>
					{#if profileErrors.email}<small class="field-error">{profileErrors.email}</small>{/if}
				</label>
			</div>

			{#if profileSaveError}
				<p class="error-banner">{profileSaveError}</p>
			{/if}

			<div class="actions between">
				<button type="button" class="ghost" on:click={() => goto('/onboarding')}>Back</button>
				<button type="button" on:click={submitProfile} disabled={savingProfile}>
					{savingProfile ? 'Saving…' : 'Save and continue'}
				</button>
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

	.grid.two {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(2, minmax(0, 1fr));
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

	small {
		color: #5d6b82;
	}

	.field-error,
	.error-banner {
		color: #b42318;
		font-weight: 600;
	}

	.error-banner {
		padding: 0.9rem 1rem;
		border-radius: 1rem;
		background: #fff1f0;
	}

	@media (max-width: 720px) {
		.grid.two {
			grid-template-columns: 1fr;
		}

		.actions.between {
			flex-direction: column;
		}
	}
</style>
