<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import TaskEditor from '$lib/components/TaskEditor.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import Tile from '$lib/components/design/Tile.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import {
		addStarterTask,
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		createTask,
		deleteTask,
		emptyTaskDraft,
		formatStepLabel,
		formatTaskFrequency,
		getState,
		getStoredOnboardingToken,
		readTokenFromUrl,
		setStoredOnboardingToken,
		syncTokenInUrl,
		updateTask,
		type OnboardingState,
		type OnboardingTask,
		type TaskDraft
	} from '$lib/api/onboarding';

	let loading = true;
	let savingTask = false;
	let taskSaveError = '';
	let addingStarterId = '';
	let editingTaskId = '';
	let state: OnboardingState = {
		has_profile: false,
		telegram_linked: false,
		calendar_connected: false,
		tasks: [],
		starter_templates: []
	};
	let sessionToken = '';
	let taskDraft: TaskDraft = emptyTaskDraft();

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
	});

	async function refreshState() {
		loading = true;
		try {
			state = await getState(sessionToken);
			if (!state.has_profile) {
				await goto('/onboarding/profile');
			}
		} catch {
			clearStoredOnboardingToken();
			await goto('/onboarding');
		} finally {
			loading = false;
		}
	}

	async function addStarter(templateId: string) {
		if (addingStarterId) return;
		taskSaveError = '';
		addingStarterId = templateId;
		try {
			await addStarterTask(sessionToken, templateId);
			await refreshState();
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
		} catch (error) {
			taskSaveError = error instanceof Error ? error.message : 'Could not remove that task.';
		}
	}

	function starterSubtitle(template: {
		duration_minutes: number;
		cadence_type: 'interval' | 'count';
		cadence_value: number;
		subtasks: { name: string; duration_minutes: number }[];
	}) {
		const parts = [`${template.duration_minutes} min`, formatTaskFrequency(template)];
		if (template.subtasks.length > 0) {
			parts.push(`${template.subtasks.length} step(s)`);
		}
		return parts.join(' · ');
	}
</script>

{#if loading}
	<div class="loading">Loading your task step…</div>
{:else}
	<OnboardingShell
		{steps}
		currentStep={4}
		title="Add at least one task."
		intro="Start with a ready-made task or make your own. If a task works better in steps, you can break it into smaller parts."
	>
		<div class="tasks-page">
			<div class="section-header">
				<div>
					<p class="step-label">{formatStepLabel(steps[4])}</p>
					<h2 class="section-title">Add at least one task</h2>
					<p class="section-lede">
						Pick a starter idea or create your own. You only need one task to keep moving.
					</p>
				</div>
				<Button variant="secondary" on:click={beginNewTask}>Add a custom task</Button>
			</div>

			<section class="starters" aria-labelledby="starters-title">
				<h3 id="starters-title" class="subsection-title">Starter ideas</h3>
				{#if state.starter_templates.length === 0}
					<p class="empty">No starter ideas available right now.</p>
				{:else}
					<div class="starter-grid">
						{#each state.starter_templates as template}
							<Tile
								title={template.name}
								subtitle={starterSubtitle(template)}
								icon="✨"
								selected={addingStarterId === template.id}
								on:click={() => addStarter(template.id)}
							/>
						{/each}
					</div>
				{/if}
			</section>

			<section class="editor-section" aria-labelledby="editor-title">
				<h3 id="editor-title" class="subsection-title">
					{editingTaskId ? 'Edit this task' : 'Create a custom task'}
				</h3>
				<p class="hint">
					Required fields are marked in the form. If the task has separate steps, add them
					before you save.
				</p>
				<TaskEditor
					draft={taskDraft}
					saving={savingTask}
					submitLabel={editingTaskId ? 'Save changes' : 'Save custom task'}
					error={taskSaveError}
					on:save={saveTaskStep}
					on:cancel={beginNewTask}
				/>
			</section>

			<section class="saved-tasks" aria-labelledby="saved-title">
				<h3 id="saved-title" class="subsection-title">Your saved tasks</h3>
				{#if state.tasks.length === 0}
					<InfoBox>No tasks yet. Add at least one starter or custom task to continue.</InfoBox>
				{:else}
					<ul class="saved-list">
						{#each state.tasks as task}
							<li class="saved-task">
								<div class="saved-task-body">
									<h4>{task.name}</h4>
									<p class="task-note">
										{task.description || 'No extra note added.'}
									</p>
									<p class="task-meta">
										{task.duration_minutes} min · {task.priority} priority · {task.time_of_day_preference}
										{#if task.subtasks.length > 0}
											· {task.subtasks.length} step(s)
										{/if}
									</p>
									{#if task.subtasks.length > 0}
										<ul class="subtask-list">
											{#each task.subtasks as subtask}
												<li>{subtask.name} · {subtask.duration_minutes} min</li>
											{/each}
										</ul>
									{/if}
								</div>
								<div class="saved-task-actions">
									<Button variant="secondary" on:click={() => beginEditTask(task)}>Edit</Button>
									<Button variant="text" on:click={() => removeTask(task.id)}>Remove</Button>
								</div>
							</li>
						{/each}
					</ul>
				{/if}
			</section>

			{#if taskSaveError}
				<p class="error-banner" role="alert">{taskSaveError}</p>
			{/if}

			<div class="actions">
				<Button variant="secondary" on:click={() => goto('/onboarding/calendar')}>Back</Button>
				<Button
					variant="primary"
					disabled={state.tasks.length === 0}
					on:click={() => goto('/onboarding/review')}
				>
					Review my setup
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

	.tasks-page {
		display: grid;
		gap: var(--space-6);
	}

	.section-header {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-4);
		min-width: 0;
	}

	.section-header > div {
		min-width: 0;
	}

	.step-label {
		font-size: 11px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--primary-2);
		font-weight: 600;
		margin-bottom: var(--space-2);
	}

	.section-title {
		font-family: var(--font-display);
		font-size: 28px;
		line-height: 1.1;
		font-weight: 400;
		color: var(--ink);
		margin-bottom: var(--space-2);
	}

	.section-lede {
		font-size: 15.5px;
		color: var(--ink-2);
		line-height: 1.6;
	}

	.subsection-title {
		font-size: 15px;
		font-weight: 500;
		color: var(--ink);
		margin-bottom: var(--space-3);
	}

	.hint,
	.empty {
		font-size: 14px;
		color: var(--ink-3);
		margin-bottom: var(--space-4);
	}

	.starter-grid {
		display: grid;
		gap: var(--space-3);
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
	}

	.editor-section,
	.saved-tasks {
		display: grid;
	}

	.saved-list {
		list-style: none;
		display: grid;
		gap: var(--space-3);
		padding: 0;
		margin: 0;
	}

	.saved-task {
		display: flex;
		flex-direction: column;
		gap: var(--space-4);
		min-width: 0;
		padding: var(--space-4);
		background: var(--paper);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-lg);
	}

	.saved-task-body {
		min-width: 0;
	}

	.saved-task-body h4 {
		font-size: 15px;
		font-weight: 500;
		color: var(--ink);
		margin: 0 0 var(--space-1);
	}

	.task-note {
		font-size: 14px;
		color: var(--ink-2);
		margin: 0 0 var(--space-1);
		word-wrap: break-word;
	}

	.task-meta {
		font-size: 13px;
		color: var(--ink-3);
		margin: 0 0 var(--space-2);
	}

	.subtask-list {
		list-style: disc;
		padding-left: var(--space-5);
		margin: 0;
		font-size: 13px;
		color: var(--ink-3);
	}

	.subtask-list li {
		margin-bottom: var(--space-1);
	}

	.saved-task-actions {
		display: flex;
		justify-content: flex-end;
		gap: var(--space-3);
		min-width: 0;
	}

	.saved-task-actions :global(button) {
		min-width: auto;
		padding: 10px 16px;
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
		min-width: 0;
	}

	@media (max-width: 540px) {
		.section-header,
		.saved-task,
		.actions {
			flex-direction: column;
			align-items: stretch;
		}

		.saved-task-actions {
			width: 100%;
		}
	}
</style>
