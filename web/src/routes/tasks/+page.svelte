<script lang="ts">
	import { onMount } from 'svelte';
	import TaskEditor from '$lib/components/TaskEditor.svelte';
	import TaskGroup from '$lib/components/tasks/TaskGroup.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import { emptyTaskDraft, type TaskDraft } from '$lib/api/onboarding';
	import {
		createTask,
		listTasks,
		removeTask,
		setTaskPaused,
		toDraft,
		updateTask,
		type ManagedTask
	} from '$lib/api/tasks';

	let tasks: ManagedTask[] = [];
	let loading = true;
	let saving = false;
	let error = '';
	let editorError = '';
	let mode: 'list' | 'create' | 'edit' = 'list';
	let draft: TaskDraft = emptyTaskDraft();
	let editingTask: ManagedTask | null = null;
	let confirmAction: { type: 'pause' | 'resume' | 'remove'; task: ManagedTask } | null = null;

	$: activeTasks = tasks.filter((task) => !task.archived_at && !task.is_paused);
	$: pausedTasks = tasks.filter((task) => !task.archived_at && task.is_paused);
	$: removedTasks = tasks.filter((task) => task.archived_at);

	onMount(loadTasks);

	async function loadTasks() {
		loading = true;
		error = '';
		try {
			tasks = await listTasks();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not load routines.';
		} finally {
			loading = false;
		}
	}

	function startCreate() {
		draft = emptyTaskDraft();
		editingTask = null;
		editorError = '';
		mode = 'create';
	}

	function startEdit(task: ManagedTask) {
		draft = toDraft(task);
		editingTask = task;
		editorError = '';
		mode = 'edit';
	}

	async function save(event: CustomEvent<{ draft: TaskDraft }>) {
		saving = true;
		editorError = '';
		try {
			if (mode === 'edit' && editingTask) {
				await updateTask(editingTask.id, event.detail.draft);
			} else {
				await createTask(event.detail.draft);
			}
			mode = 'list';
			await loadTasks();
		} catch (err) {
			editorError = err instanceof Error ? err.message : 'Could not save this routine.';
		} finally {
			saving = false;
		}
	}

	async function confirmChange() {
		if (!confirmAction) return;
		saving = true;
		error = '';
		try {
			if (confirmAction.type === 'pause') await setTaskPaused(confirmAction.task.id, true);
			if (confirmAction.type === 'resume') await setTaskPaused(confirmAction.task.id, false);
			if (confirmAction.type === 'remove') await removeTask(confirmAction.task.id);
			confirmAction = null;
			await loadTasks();
		} catch (err) {
			error = err instanceof Error ? err.message : 'Could not update this routine.';
		} finally {
			saving = false;
		}
	}
</script>

<svelte:head>
	<title>Rahat routines</title>
</svelte:head>

<div class="page">
	<header class="topbar">
		<a class="wordmark" href="/">Rahat<span>.</span></a>
	</header>

	<section class="hero card">
		<p class="eyebrow">Routine care</p>
		<h1 class="display">Keep your routines current.</h1>
		<p class="lede">
			Add, tune, pause, or remove routines here. Completed history stays intact; future schedule
			previews use only active routines.
		</p>
		<Button variant="primary" on:click={startCreate}>Add a routine</Button>
	</section>

	{#if error}
		<p class="error-banner" role="alert">{error}</p>
	{/if}

	{#if mode !== 'list'}
		<section class="editor card">
			<p class="eyebrow">{mode === 'edit' ? 'Edit routine' : 'New routine'}</p>
			<TaskEditor
				{draft}
				{saving}
				submitLabel={mode === 'edit' ? 'Save changes' : 'Create routine'}
				error={editorError}
				on:save={save}
				on:cancel={() => (mode = 'list')}
			/>
		</section>
	{:else if loading}
		<InfoBox title="Loading routines…">One moment while we fetch your routines.</InfoBox>
	{:else}
		<TaskGroup
			title="Active"
			tasks={activeTasks}
			empty="No active routines yet."
			onEdit={startEdit}
			onAction={(type, task) => (confirmAction = { type, task })}
		/>
		<TaskGroup
			title="Paused"
			tasks={pausedTasks}
			empty="Paused routines will not create new schedule items."
			onEdit={startEdit}
			onAction={(type, task) => (confirmAction = { type, task })}
		/>
		<TaskGroup
			title="Removed"
			tasks={removedTasks}
			empty="Removed routines appear here after removal."
			onEdit={startEdit}
			onAction={(type, task) => (confirmAction = { type, task })}
		/>
	{/if}
</div>

{#if confirmAction}
	<div class="modal-backdrop" role="presentation">
		<div class="modal card" role="dialog" aria-modal="true" aria-labelledby="confirm-title">
			<p class="eyebrow">Confirm</p>
			<h2 id="confirm-title" class="display-sm">
				{confirmAction.type === 'remove'
					? 'Remove this routine?'
					: confirmAction.type === 'pause'
						? 'Pause this routine?'
						: 'Resume this routine?'}
			</h2>
			<p class="lede">
				{#if confirmAction.type === 'remove'}
					Rahat will hide it from active planning and stop creating future schedule items.
					Completed occurrences and event history are preserved.
				{:else if confirmAction.type === 'pause'}
					Rahat will stop creating new schedule items for this routine until you resume it.
					Existing completed history is not changed.
				{:else}
					Rahat will include this routine in the next preview and future schedule runs again.
				{/if}
			</p>
			<div class="modal-actions">
				<Button variant="secondary" on:click={() => (confirmAction = null)}>Cancel</Button>
				<Button
					variant={confirmAction.type === 'remove' ? 'text' : 'primary'}
					disabled={saving}
					on:click={confirmChange}
				>
					{saving ? 'Working…' : 'Confirm'}
				</Button>
			</div>
		</div>
	</div>
{/if}

<style>
	.page {
		max-width: 980px;
		margin: 0 auto;
		padding: var(--space-8) var(--space-5) var(--space-12);
	}

	.topbar {
		display: flex;
		align-items: center;
		justify-content: space-between;
		margin-bottom: var(--space-6);
	}

	.wordmark {
		font-family: var(--font-display);
		font-size: 24px;
		color: var(--ink);
		text-decoration: none;
	}

	.wordmark span {
		color: var(--primary-2);
	}

	.card {
		background: var(--paper);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-2xl);
		box-shadow: var(--shadow-md);
	}

	.hero {
		margin: 0 0 var(--space-6);
		padding: var(--space-8);
		display: grid;
		gap: var(--space-4);
	}

	.editor {
		padding: var(--space-6);
	}

	.eyebrow {
		margin: 0;
		font-size: 11px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		font-weight: 700;
		color: var(--primary-2);
	}

	.display {
		margin: 0;
		font-family: var(--font-display);
		font-size: clamp(32px, 6vw, 40px);
		line-height: 1.05;
		font-weight: 400;
		color: var(--ink);
	}

	.display-sm {
		margin: 0;
		font-family: var(--font-display);
		font-size: 28px;
		line-height: 1.1;
		font-weight: 400;
		color: var(--ink);
	}

	.lede {
		margin: 0;
		font-size: 15.5px;
		line-height: 1.6;
		color: var(--ink-2);
	}

	.error-banner {
		padding: var(--space-4) var(--space-5);
		background: var(--rose-soft);
		border: 1px solid var(--rose);
		border-radius: var(--radius-lg);
		color: var(--rose);
		margin: 0 0 var(--space-6);
	}

	.modal-backdrop {
		position: fixed;
		inset: 0;
		background: rgba(31, 29, 26, 0.28);
		display: grid;
		place-items: center;
		padding: var(--space-5);
		z-index: 100;
	}

	.modal {
		width: 100%;
		max-width: 440px;
		padding: var(--space-8);
		display: grid;
		gap: var(--space-4);
	}

	.modal-actions {
		display: flex;
		align-items: center;
		justify-content: flex-end;
		gap: var(--space-3);
		margin-top: var(--space-2);
	}

	@media (max-width: 700px) {
		.hero {
			padding: var(--space-6);
		}
	}
</style>
