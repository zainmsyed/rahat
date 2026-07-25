<script lang="ts">
	import { onMount } from 'svelte';
	import TaskEditor from '$lib/components/TaskEditor.svelte';
	import TaskGroup from '$lib/components/tasks/TaskGroup.svelte';
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
		<a class="wordmark" href="/">Rah<span>at</span></a>
		<a class="nav" href="/lookahead">Lookahead</a>
	</header>

	<section class="hero">
		<p class="eyebrow">Routine care</p>
		<h1>Keep your routines current.</h1>
		<p>
			Add, tune, pause, or remove routines here. Completed history stays intact; future schedule previews use only active routines.
		</p>
		<button on:click={startCreate}>Add a routine</button>
	</section>

	{#if error}
		<p class="notice error">{error}</p>
	{/if}

	{#if mode !== 'list'}
		<section class="panel">
			<p class="eyebrow">{mode === 'edit' ? 'Edit routine' : 'New routine'}</p>
			<TaskEditor {draft} {saving} submitLabel={mode === 'edit' ? 'Save changes' : 'Create routine'} error={editorError} on:save={save} on:cancel={() => (mode = 'list')} />
		</section>
	{:else if loading}
		<p class="notice">Loading routines…</p>
	{:else}
		<TaskGroup title="Active" tasks={activeTasks} empty="No active routines yet." onEdit={startEdit} onAction={(type, task) => (confirmAction = { type, task })} />
		<TaskGroup title="Paused" tasks={pausedTasks} empty="Paused routines will not create new schedule items." onEdit={startEdit} onAction={(type, task) => (confirmAction = { type, task })} />
		<TaskGroup title="Removed" tasks={removedTasks} empty="Removed routines appear here after removal." onEdit={startEdit} onAction={(type, task) => (confirmAction = { type, task })} />
	{/if}
</div>

{#if confirmAction}
	<div class="modal-backdrop" role="presentation">
		<div class="modal" role="dialog" aria-modal="true" aria-labelledby="confirm-title">
			<p class="eyebrow">Confirm</p>
			<h2 id="confirm-title">{confirmAction.type === 'remove' ? 'Remove this routine?' : confirmAction.type === 'pause' ? 'Pause this routine?' : 'Resume this routine?'}</h2>
			<p>
				{#if confirmAction.type === 'remove'}
					Rahat will hide it from active planning and stop creating future schedule items. Completed occurrences and event history are preserved.
				{:else if confirmAction.type === 'pause'}
					Rahat will stop creating new schedule items for this routine until you resume it. Existing completed history is not changed.
				{:else}
					Rahat will include this routine in the next preview and future schedule runs again.
				{/if}
			</p>
			<div class="actions">
				<button class="ghost" on:click={() => (confirmAction = null)}>Cancel</button>
				<button class:danger={confirmAction.type === 'remove'} on:click={confirmChange} disabled={saving}>{saving ? 'Working…' : 'Confirm'}</button>
			</div>
		</div>
	</div>
{/if}

<style>
	:global(body) { margin: 0; font-family: Outfit, system-ui, sans-serif; background: #faf7f2; color: #1f1d1a; }
	.page { max-width: 980px; margin: 0 auto; padding: 28px 20px 80px; }
	.topbar, .actions { display: flex; align-items: center; justify-content: space-between; gap: 12px; }
	.wordmark { font-family: 'DM Serif Display', Georgia, serif; font-size: 24px; color: #1f1d1a; text-decoration: none; }
	.wordmark span, .eyebrow { color: #5a7a56; }
	.nav { color: #4a4640; }
	.hero, .panel, .modal { background: #fff; border: 1px solid #e3dccc; border-radius: 20px; box-shadow: 0 2px 8px rgba(31,29,26,.05), 0 12px 32px -12px rgba(31,29,26,.08); }
	.hero { margin: 24px 0; padding: 36px; display: grid; gap: 16px; }
	.panel { padding: 24px; }
	.eyebrow { margin: 0; font-size: 11px; letter-spacing: .18em; text-transform: uppercase; font-weight: 700; }
	h1 { margin: 0; font-family: 'DM Serif Display', Georgia, serif; font-size: clamp(34px, 6vw, 52px); line-height: 1.05; font-weight: 400; }
	p { color: #4a4640; line-height: 1.6; }
	button { border: 0; border-radius: 999px; background: #7a9b76; color: #fff; padding: 13px 18px; font: inherit; font-weight: 700; cursor: pointer; }
	button:hover { background: #5a7a56; }
	button.ghost { background: transparent; color: #4a4640; border: 1.5px solid #e3dccc; }
	button.danger { background: #b87a7a; }
	button:disabled { opacity: .65; cursor: wait; }
	.notice { padding: 16px 18px; background: #f4f0e6; border: 1px solid #ebe5d8; border-radius: 12px; }
	.error { color: #8f3d3d; }
	.modal-backdrop { position: fixed; inset: 0; background: rgba(31,29,26,.28); display: grid; place-items: center; padding: 20px; }
	.modal { max-width: 440px; padding: 28px; }
	@media (max-width: 700px) { .hero { padding: 24px; } }
</style>
