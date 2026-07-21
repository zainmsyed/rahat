<script lang="ts">
	import { onMount } from 'svelte';
	import OnboardingStepper from '$lib/components/OnboardingStepper.svelte';
	import TaskEditor from '$lib/components/TaskEditor.svelte';
	import {
		addStarterTask,
		createSession,
		createTask,
		deleteTask,
		emptyTaskDraft,
		finishOnboarding,
		getState,
		saveProfile,
		updateTask,
		type OnboardingFinishResult,
		type OnboardingProfile,
		type OnboardingState,
		type OnboardingTask,
		type TaskDraft
	} from '$lib/api/onboarding';

	type Step = {
		id: number;
		title: string;
		required: boolean;
		description: string;
		complete: boolean;
	};

	const storageKey = 'rahat-onboarding-token';

	let loading = true;
	let sessionToken = '';
	let inviteCode = '';
	let currentStep = 0;
	let state: OnboardingState = { has_profile: false, tasks: [], starter_templates: [] };
	let profileDraft: OnboardingProfile = {
		display_name: '',
		timezone: 'UTC',
		daily_time_budget_minutes: 45,
		email: ''
	};
	let profileErrors = { display_name: '', timezone: '', daily_time_budget_minutes: '', email: '' };
	let pageError = '';
	let profileSaveError = '';
	let taskSaveError = '';
	let savingProfile = false;
	let savingTask = false;
	let finishing = false;
	let addingStarterId = '';
	let finishResult: OnboardingFinishResult | null = null;
	let editingTaskId = '';
	let taskDraft: TaskDraft = emptyTaskDraft();

	onMount(async () => {
		const detectedTimezone = Intl.DateTimeFormat().resolvedOptions().timeZone;
		if (detectedTimezone) {
			profileDraft = { ...profileDraft, timezone: detectedTimezone };
		}

		const url = new URL(window.location.href);
		inviteCode = url.searchParams.get('invite') ?? '';
		const tokenFromUrl = url.searchParams.get('token');
		const tokenFromStorage = localStorage.getItem(storageKey);
		sessionToken = tokenFromUrl ?? tokenFromStorage ?? '';

		if (sessionToken) {
			await refreshState();
		} else {
			loading = false;
		}
	});

	$: steps = buildSteps(state, !!sessionToken);

	function buildSteps(currentState: OnboardingState, hasSession: boolean): Step[] {
		return [
			{
				id: 0,
				title: 'Start with your invite code',
				required: true,
				description: 'Enter the invite code you were given so Rahat can open your guided setup.',
				complete: hasSession
			},
			{
				id: 1,
				title: 'Tell Rahat about you',
				required: true,
				description:
					'Add your name, timezone, daily task-time budget, and an optional email for recaps.',
				complete: currentState.has_profile
			},
			{
				id: 2,
				title: 'Pick at least one task',
				required: true,
				description:
					'Choose from the starter ideas or add your own task and steps in plain language.',
				complete: currentState.tasks.length > 0
			},
			{
				id: 3,
				title: 'Review and finish',
				required: true,
				description:
					'Confirm everything looks right, let Rahat seed your first schedule, and read what happens next.',
				complete: finishResult !== null
			}
		];
	}

	function setToken(token: string) {
		sessionToken = token;
		localStorage.setItem(storageKey, token);
		const url = new URL(window.location.href);
		url.searchParams.set('token', token);
		url.searchParams.delete('invite');
		window.history.replaceState({}, '', url);
	}

	function clearSession(message: string) {
		sessionToken = '';
		localStorage.removeItem(storageKey);
		finishResult = null;
		state = { has_profile: false, tasks: [], starter_templates: [] };
		pageError = message;
		currentStep = 0;
	}

	async function refreshState() {
		loading = true;
		pageError = '';
		try {
			state = await getState(sessionToken);
			if (state.user) {
				profileDraft = { ...state.user };
			}
			if (finishResult === null) {
				currentStep = state.tasks.length > 0 ? 3 : state.has_profile ? 2 : 1;
			}
		} catch (error) {
			clearSession('Your onboarding link expired. Please start again with your invite code.');
		} finally {
			loading = false;
		}
	}

	async function startSession() {
		pageError = '';
		if (!inviteCode.trim()) {
			pageError = 'Please enter your invite code to begin.';
			return;
		}
		loading = true;
		try {
			const session = await createSession(inviteCode.trim());
			setToken(session.token);
			await refreshState();
			currentStep = 1;
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not start onboarding.';
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

	async function saveProfileStep() {
		profileSaveError = '';
		if (!validateProfile()) {
			return;
		}
		savingProfile = true;
		try {
			await saveProfile(sessionToken, profileDraft);
			await refreshState();
			currentStep = 2;
		} catch (error) {
			profileSaveError = error instanceof Error ? error.message : 'Could not save your profile.';
		} finally {
			savingProfile = false;
		}
	}

	async function addStarter(templateId: string) {
		taskSaveError = '';
		addingStarterId = templateId;
		try {
			await addStarterTask(sessionToken, templateId);
			await refreshState();
			currentStep = 3;
		} catch (error) {
			taskSaveError = error instanceof Error ? error.message : 'Could not add that starter task.';
		} finally {
			addingStarterId = '';
		}
	}

	function beginNewTask() {
		editingTaskId = '';
		taskSaveError = '';
		taskDraft = emptyTaskDraft();
	}

	function beginEditTask(task: OnboardingTask) {
		editingTaskId = task.id;
		taskSaveError = '';
		taskDraft = {
			name: task.name,
			description: task.description,
			duration_minutes: task.duration_minutes,
			cadence_type: task.cadence_type,
			cadence_value: task.cadence_value,
			priority: task.priority,
			time_of_day_preference: task.time_of_day_preference,
			subtasks: task.subtasks.map((subtask) => ({ ...subtask }))
		};
	}

	async function saveTaskStep(event: CustomEvent<{ draft: TaskDraft }>) {
		taskSaveError = '';
		savingTask = true;
		try {
			if (editingTaskId) {
				await updateTask(sessionToken, editingTaskId, event.detail.draft);
			} else {
				await createTask(sessionToken, event.detail.draft);
			}
			editingTaskId = '';
			taskDraft = emptyTaskDraft();
			await refreshState();
			currentStep = 3;
		} catch (error) {
			taskSaveError = error instanceof Error ? error.message : 'Could not save that task.';
		} finally {
			savingTask = false;
		}
	}

	async function removeTask(taskId: string) {
		taskSaveError = '';
		try {
			await deleteTask(sessionToken, taskId);
			if (editingTaskId === taskId) {
				editingTaskId = '';
				taskDraft = emptyTaskDraft();
			}
			await refreshState();
			if (state.tasks.length === 0) {
				currentStep = 2;
			}
		} catch (error) {
			taskSaveError = error instanceof Error ? error.message : 'Could not remove that task.';
		}
	}

	async function finishSetup() {
		pageError = '';
		finishing = true;
		try {
			finishResult = await finishOnboarding(sessionToken);
			currentStep = 3;
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not finish onboarding.';
		} finally {
			finishing = false;
		}
	}
</script>

<svelte:head>
	<title>Rahat onboarding</title>
	<meta
		name="description"
		content="Guided Rahat onboarding for profile setup, starter tasks, and first schedule seeding."
	/>
</svelte:head>

{#if loading}
	<div class="loading">Loading your onboarding steps…</div>
{:else}
	<div class="page">
		<OnboardingStepper {steps} {currentStep} finished={finishResult !== null} />

		<main class="content">
			<header class="hero">
				<p class="eyebrow">Welcome to Rahat</p>
				<h1>Let's set up your first calm, clear plan.</h1>
				<p>
					Each step below explains exactly what to do next. Required items are labeled. The email field is optional.
				</p>
			</header>

			{#if pageError}
				<p class="error-banner">{pageError}</p>
			{/if}

			<section class="panel" class:active={currentStep === 0}>
				<div class="panel-header">
					<div>
						<p class="label">Step 1 · Required</p>
						<h2>Start with your invite code</h2>
						<p>Type the invite code you received, then press the button to begin your guided setup.</p>
					</div>
				</div>
				<label>
					<span>Invite code *</span>
					<input bind:value={inviteCode} placeholder="Example: rahat-beta" />
				</label>
				<div class="actions">
					<button type="button" on:click={startSession}>Start onboarding</button>
				</div>
			</section>

			<section class="panel" class:active={currentStep === 1}>
				<div class="panel-header">
					<div>
						<p class="label">Step 2 · Required</p>
						<h2>Tell Rahat about you</h2>
						<p>
							Save the basics first: your name, your timezone, how many minutes you can usually spend on tasks in a day, and an optional email for recaps.
						</p>
					</div>
				</div>
				<div class="grid two">
					<label>
						<span>Name *</span>
						<input bind:value={profileDraft.display_name} placeholder="What should Rahat call you?" />
						{#if profileErrors.display_name}
							<small class="field-error">{profileErrors.display_name}</small>
						{/if}
					</label>
					<label>
						<span>Timezone *</span>
						<input bind:value={profileDraft.timezone} placeholder="Example: America/New_York" />
						{#if profileErrors.timezone}
							<small class="field-error">{profileErrors.timezone}</small>
						{/if}
					</label>
				</div>
				<div class="grid two">
					<label>
						<span>Daily task-time budget in minutes *</span>
						<input bind:value={profileDraft.daily_time_budget_minutes} type="number" min="15" max="480" />
						<small>Friendly default: 45 minutes.</small>
						{#if profileErrors.daily_time_budget_minutes}
							<small class="field-error">{profileErrors.daily_time_budget_minutes}</small>
						{/if}
					</label>
					<label>
						<span>Email for recaps (optional)</span>
						<input bind:value={profileDraft.email} type="email" placeholder="Only if you want daily recaps later" />
						<small>You can leave this blank.</small>
						{#if profileErrors.email}
							<small class="field-error">{profileErrors.email}</small>
						{/if}
					</label>
				</div>
				{#if profileSaveError}
					<p class="error-banner">{profileSaveError}</p>
				{/if}
				<div class="actions between">
					<button type="button" class="ghost" on:click={() => (currentStep = 0)}>Back</button>
					<button type="button" on:click={saveProfileStep} disabled={savingProfile}>
						{savingProfile ? 'Saving…' : 'Save and continue'}
					</button>
				</div>
			</section>

			<section class="panel" class:active={currentStep === 2 || currentStep === 3}>
				<div class="panel-header split">
					<div>
						<p class="label">Step 3 · Required</p>
						<h2>Add at least one task</h2>
						<p>
							Start with a ready-made task or make your own. You can also add smaller steps if a task is easier to do in parts.
						</p>
					</div>
					<button type="button" class="ghost" on:click={beginNewTask}>Add a custom task</button>
				</div>

				<div class="starter-grid">
					{#each state.starter_templates as template}
						<article>
							<h3>{template.name}</h3>
							<p>{template.description}</p>
							<p class="meta">
								{template.duration_minutes} min · {template.cadence_type === 'interval' ? `Every ${template.cadence_value} day(s)` : `${template.cadence_value} time(s) each week`}
							</p>
							{#if template.subtasks.length > 0}
								<ul>
									{#each template.subtasks as subtask}
										<li>{subtask.name} · {subtask.duration_minutes} min</li>
									{/each}
								</ul>
							{/if}
							<button type="button" on:click={() => addStarter(template.id)} disabled={addingStarterId === template.id}>
								{addingStarterId === template.id ? 'Adding…' : 'Add this starter task'}
							</button>
						</article>
					{/each}
				</div>

				<div class="editor-panel">
					<h3>{editingTaskId ? 'Edit this task' : 'Create a custom task'}</h3>
					<p>
						Required fields are marked in the form. If the task has separate steps, add them before you save.
					</p>
					<TaskEditor
						draft={taskDraft}
						saving={savingTask}
						submitLabel={editingTaskId ? 'Save changes' : 'Save custom task'}
						error={taskSaveError}
						on:save={saveTaskStep}
						on:cancel={() => {
							editingTaskId = '';
							taskSaveError = '';
							taskDraft = emptyTaskDraft();
						}}
					/>
				</div>

				<div class="saved-tasks">
					<h3>Your saved tasks</h3>
					{#if state.tasks.length === 0}
						<p class="empty">No tasks yet. Add at least one starter or custom task to continue.</p>
					{:else}
						{#each state.tasks as task}
							<article class="saved-task">
								<div>
									<h4>{task.name}</h4>
									<p>{task.description || 'No extra note added.'}</p>
									<p class="meta">
										{task.duration_minutes} min · {task.priority} priority · {task.time_of_day_preference}
									</p>
									{#if task.subtasks.length > 0}
										<ul>
											{#each task.subtasks as subtask}
												<li>{subtask.name} · {subtask.duration_minutes} min</li>
											{/each}
										</ul>
									{/if}
								</div>
								<div class="row-actions">
									<button type="button" class="ghost" on:click={() => beginEditTask(task)}>Edit</button>
									<button type="button" class="danger" on:click={() => removeTask(task.id)}>Remove</button>
								</div>
							</article>
						{/each}
					{/if}
				</div>

				<div class="actions between">
					<button type="button" class="ghost" on:click={() => (currentStep = 1)}>Back</button>
					<button type="button" on:click={() => (currentStep = 3)} disabled={state.tasks.length === 0}>
						Review my setup
					</button>
				</div>
			</section>

			<section class="panel" class:active={currentStep === 3}>
				<div class="panel-header">
					<div>
						<p class="label">Step 4 · Required</p>
						<h2>Review and finish</h2>
						<p>
							Check your details below. When you press finish, Rahat will seed your first schedule and show what happens next.
						</p>
					</div>
				</div>

				<div class="review-grid">
					<article>
						<h3>Your profile</h3>
						<p><strong>Name:</strong> {profileDraft.display_name || 'Not saved yet'}</p>
						<p><strong>Timezone:</strong> {profileDraft.timezone}</p>
						<p><strong>Daily budget:</strong> {profileDraft.daily_time_budget_minutes} minutes</p>
						<p><strong>Email:</strong> {profileDraft.email || 'Skipped for now'}</p>
					</article>
					<article>
						<h3>Your tasks</h3>
						<p>{state.tasks.length} task(s) ready for Rahat to schedule.</p>
						<p>You can still go back and edit anything before you finish.</p>
					</article>
				</div>

				<div class="actions between">
					<button type="button" class="ghost" on:click={() => (currentStep = 2)}>Back</button>
					<button type="button" on:click={finishSetup} disabled={state.tasks.length === 0 || finishing}>
						{finishing ? 'Seeding your first schedule…' : 'Finish onboarding'}
					</button>
				</div>

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
			</section>
		</main>
	</div>
{/if}

<style>
	:global(body) {
		margin: 0;
		font-family: Inter, system-ui, sans-serif;
		background: #f4f7fb;
		color: #14202c;
	}

	.loading {
		min-height: 100vh;
		display: grid;
		place-items: center;
		font-size: 1.1rem;
	}

	.page {
		max-width: 1200px;
		margin: 0 auto;
		padding: 1.5rem;
		display: grid;
		grid-template-columns: 320px minmax(0, 1fr);
		gap: 1.5rem;
	}

	.content {
		display: grid;
		gap: 1.25rem;
	}

	.hero,
	.panel,
	.success {
		padding: 1.4rem;
		border-radius: 1.4rem;
		background: white;
		box-shadow: 0 16px 40px rgba(20, 32, 44, 0.08);
	}

	.hero h1,
	.panel h2,
	.panel h3,
	.panel h4,
	.success h3,
	.success h4 {
		margin: 0;
	}

	.hero p,
	.panel p,
	.success p,
	.panel li,
	.success li {
		line-height: 1.55;
	}

	.eyebrow,
	.label {
		margin: 0 0 0.4rem;
		font-size: 0.85rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #4f46e5;
	}

	.panel {
		display: grid;
		gap: 1rem;
		border: 2px solid transparent;
	}

	.panel.active {
		border-color: #d6e4ff;
	}

	.panel-header.split,
	.saved-task,
	.actions.between,
	.row-actions {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		align-items: start;
	}

	label {
		display: grid;
		gap: 0.4rem;
		font-weight: 600;
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

	button.danger {
		background: #b42318;
	}

	button:disabled {
		opacity: 0.7;
		cursor: wait;
	}

	.grid.two,
	.review-grid,
	.starter-grid {
		display: grid;
		gap: 1rem;
	}

	.grid.two,
	.review-grid {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.starter-grid {
		grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
	}

	.starter-grid article,
	.saved-task,
	.review-grid article,
	.editor-panel {
		padding: 1rem;
		border: 1px solid #dbe4ee;
		border-radius: 1rem;
		background: #fbfdff;
	}

	.saved-tasks {
		display: grid;
		gap: 0.8rem;
	}

	.meta,
	small,
	.empty,
	.schedule-list span {
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

	.schedule-list {
		padding-left: 1.1rem;
	}

	.schedule-list li {
		margin-bottom: 0.55rem;
	}

	.schedule-list span {
		display: block;
	}

	@media (max-width: 960px) {
		.page {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 720px) {
		.grid.two,
		.review-grid {
			grid-template-columns: 1fr;
		}

		.panel-header.split,
		.saved-task,
		.actions.between,
		.row-actions {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
