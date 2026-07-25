<script lang="ts">
	import { formatTaskSummary } from '$lib/api/onboarding';
	import type { ManagedTask } from '$lib/api/tasks';

	export let title: string;
	export let tasks: ManagedTask[] = [];
	export let empty: string;
	export let onEdit: (task: ManagedTask) => void;
	export let onAction: (type: 'pause' | 'resume' | 'remove', task: ManagedTask) => void;
</script>

<section class="group">
	<div class="group-head">
		<h2>{title}</h2>
		<span>{tasks.length}</span>
	</div>
	{#if tasks.length === 0}
		<p class="empty">{empty}</p>
	{:else}
		<div class="cards">
			{#each tasks as task}
				<article class:paused={task.is_paused} class:removed={task.archived_at}>
					<p class="status">{task.archived_at ? 'Removed' : task.is_paused ? 'Paused' : 'Active'}</p>
					<h3>{task.name}</h3>
					<p class="summary">{formatTaskSummary(task)}</p>
					{#if task.description}<p>{task.description}</p>{/if}
					{#if task.subtasks.length > 0}
						<ol>
							{#each task.subtasks as subtask}
								<li>{subtask.name} · {subtask.duration_minutes} min</li>
							{/each}
						</ol>
					{/if}
					<div class="actions">
						{#if !task.archived_at}
							<button class="ghost" on:click={() => onEdit(task)}>Edit</button>
							<button class="ghost" on:click={() => onAction(task.is_paused ? 'resume' : 'pause', task)}>{task.is_paused ? 'Resume' : 'Pause'}</button>
							<button class="ghost danger" on:click={() => onAction('remove', task)}>Remove</button>
						{:else}
							<span class="read-only">Preserved for history</span>
						{/if}
					</div>
				</article>
			{/each}
		</div>
	{/if}
</section>

<style>
	.group { margin-top: 32px; }
	.group-head { display: flex; align-items: baseline; justify-content: space-between; border-bottom: 1px solid var(--rahat-line); margin-bottom: 16px; }
	h2 { font-family: 'DM Serif Display', Georgia, serif; font-weight: 400; font-size: 32px; margin: 0 0 12px; }
	.group-head span { color: var(--rahat-ink-muted); font-size: 13px; }
	.cards { display: grid; gap: 16px; }
	article { background: var(--rahat-paper); border: 1px solid var(--rahat-line); border-radius: 16px; padding: 20px; box-shadow: var(--rahat-shadow-sm); }
	article.paused, article.removed { background: var(--rahat-surface-soft); }
	article.removed { opacity: .72; }
	.status { margin: 0 0 8px; color: var(--rahat-primary-deep); font-size: 11px; letter-spacing: .14em; text-transform: uppercase; font-weight: 700; }
	h3 { margin: 0; font-size: 20px; }
	p { color: var(--rahat-ink-secondary); line-height: 1.55; }
	.summary { color: var(--rahat-ink-muted); margin-top: 8px; }
	ol { color: var(--rahat-ink-secondary); }
	.actions { display: flex; gap: 12px; flex-wrap: wrap; margin-top: 16px; }
	button { border-radius: 999px; padding: 12px 16px; font: inherit; font-weight: 700; cursor: pointer; }
	button.ghost { background: transparent; color: var(--rahat-ink-secondary); border: 1.5px solid var(--rahat-line); }
	button.danger { color: var(--rahat-rose); }
	.empty, .read-only { color: var(--rahat-ink-muted); }
</style>
