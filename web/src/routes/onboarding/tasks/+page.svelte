<script lang="ts">
	import { goto } from '$app/navigation';
	import { onMount } from 'svelte';
	import OnboardingShell from '$lib/components/OnboardingShell.svelte';
	import TaskEditor from '$lib/components/TaskEditor.svelte';
	import {
		addStarterTask,
		buildOnboardingSteps,
		clearStoredOnboardingToken,
		createTask,
		deleteTask,
		emptyTaskDraft,
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
	let state: OnboardingState = { has_profile: false, telegram_linked: false, calendar_connected: false, tasks: [], starter_templates: [] };
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
		<section class="panel active">
			<div class="panel-header split">
				<div>
					<p class="label">Step 5 · Required</p>
					<h2>Add at least one task</h2>
					<p>Pick a starter idea or create your own. You only need one task to keep moving.</p>
				</div>
				<button type="button" class="ghost" on:click={beginNewTask}>Add a custom task</button>
			</div>

			<div class="starter-grid">
				{#each state.starter_templates as template}
					<article>
						<h3>{template.name}</h3>
						<p>{template.description}</p>
						<p class="meta">{template.duration_minutes} min · {formatTaskFrequency(template)}</p>
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
				<p>Required fields are marked in the form. If the task has separate steps, add them before you save.</p>
				<TaskEditor
					draft={taskDraft}
					saving={savingTask}
					submitLabel={editingTaskId ? 'Save changes' : 'Save custom task'}
					error={taskSaveError}
					on:save={saveTaskStep}
					on:cancel={beginNewTask}
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
								<p class="meta">{task.duration_minutes} min · {task.priority} priority · {task.time_of_day_preference}</p>
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

			{#if taskSaveError}
				<p class="error-banner">{taskSaveError}</p>
			{/if}

			<div class="actions between">
				<button type="button" class="ghost" on:click={() => goto('/onboarding/calendar')}>Back</button>
				<button type="button" on:click={() => goto('/onboarding/review')} disabled={state.tasks.length === 0}>Review my setup</button>
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

	.panel-header.split,
	.saved-task,
	.actions.between,
	.row-actions {
		display: flex;
		justify-content: space-between;
		gap: 1rem;
		align-items: start;
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

	button.danger {
		background: #b42318;
	}

	button:disabled {
		opacity: 0.7;
		cursor: wait;
	}

	.starter-grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(auto-fit, minmax(230px, 1fr));
	}

	.starter-grid article,
	.saved-task,
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
	.empty {
		color: #5d6b82;
	}

	.error-banner {
		color: #b42318;
		font-weight: 600;
		padding: 0.9rem 1rem;
		border-radius: 1rem;
		background: #fff1f0;
	}

	@media (max-width: 720px) {
		.panel-header.split,
		.saved-task,
		.actions.between,
		.row-actions {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
