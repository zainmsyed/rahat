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
	.group { margin-top: 28px; }
	.group-head { display: flex; align-items: baseline; justify-content: space-between; border-bottom: 1px solid #e3dccc; margin-bottom: 14px; }
	h2 { font-family: 'DM Serif Display', Georgia, serif; font-weight: 400; font-size: 30px; margin: 0 0 10px; }
	.group-head span { color: #8a8278; font-size: 13px; }
	.cards { display: grid; gap: 14px; }
	article { background: #fff; border: 1px solid #e3dccc; border-radius: 16px; padding: 20px; box-shadow: 0 1px 2px rgba(31,29,26,.04), 0 4px 12px -4px rgba(31,29,26,.06); }
	article.paused { background: #fffaf0; }
	article.removed { opacity: .72; background: #f4f0e6; }
	.status { margin: 0 0 6px; color: #5a7a56; font-size: 11px; letter-spacing: .14em; text-transform: uppercase; font-weight: 700; }
	h3 { margin: 0; font-size: 20px; }
	p { color: #4a4640; line-height: 1.55; }
	.summary { color: #8a8278; margin-top: 6px; }
	ol { color: #4a4640; }
	.actions { display: flex; gap: 10px; flex-wrap: wrap; margin-top: 16px; }
	button { border-radius: 999px; padding: 10px 14px; font: inherit; font-weight: 700; cursor: pointer; }
	button.ghost { background: transparent; color: #4a4640; border: 1.5px solid #e3dccc; }
	button.danger { color: #8f3d3d; }
	.empty, .read-only { color: #8a8278; }
</style>
