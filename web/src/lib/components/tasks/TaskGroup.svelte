<script lang="ts">
	import { formatTaskSummary } from '$lib/api/onboarding';
	import Button from '$lib/components/design/Button.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import Toggle from '$lib/components/design/Toggle.svelte';
	import type { ManagedTask } from '$lib/api/tasks';

	export let title: string;
	export let tasks: ManagedTask[] = [];
	export let empty: string;
	export let onEdit: (task: ManagedTask) => void;
	export let onAction: (type: 'pause' | 'resume' | 'remove', task: ManagedTask) => void;
</script>

<section class="group" aria-labelledby="group-title-{title}">
	<div class="group-head">
		<h2 id="group-title-{title}">{title}</h2>
		<span class="count">{tasks.length}</span>
	</div>

	{#if tasks.length === 0}
		<InfoBox>{empty}</InfoBox>
	{:else}
		<ul class="cards">
			{#each tasks as task}
				<li class="card" class:removed={task.archived_at}>
					<div class="card-main">
						<div class="card-meta">
							<span class="status">{task.archived_at ? 'Removed' : task.is_paused ? 'Paused' : 'Active'}</span>
						</div>
						<h3 class="card-title">{task.name}</h3>
						<p class="summary">{formatTaskSummary(task)}</p>
						{#if task.description}
							<p class="description">{task.description}</p>
						{/if}
						{#if task.subtasks.length > 0}
							<ul class="subtask-list">
								{#each task.subtasks as subtask}
									<li>{subtask.name} · {subtask.duration_minutes} min</li>
								{/each}
							</ul>
						{/if}
					</div>

					<div class="card-actions">
						{#if !task.archived_at}
							<Toggle
								id="toggle-{task.id}"
								label={task.is_paused ? 'Paused' : 'Active'}
								checked={!task.is_paused}
								on:change={(event) =>
									onAction(event.detail.checked ? 'resume' : 'pause', task)}
							/>
							<div class="secondary-actions">
								<Button variant="text" on:click={() => onEdit(task)}>Edit</Button>
								<span class="danger-action">
									<Button variant="text" on:click={() => onAction('remove', task)}>
										Remove
									</Button>
								</span>
							</div>
						{:else}
							<span class="read-only">Preserved for history</span>
						{/if}
					</div>
				</li>
			{/each}
		</ul>
	{/if}
</section>

<style>
	.group {
		margin-top: var(--space-8);
	}

	.group-head {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-3);
		border-bottom: 1px solid var(--line);
		margin-bottom: var(--space-4);
		padding-bottom: var(--space-3);
		min-width: 0;
	}

	.group-head h2 {
		font-family: var(--font-display);
		font-weight: 400;
		font-size: 28px;
		line-height: 1.1;
		color: var(--ink);
		margin: 0;
	}

	.count {
		font-size: 13px;
		font-weight: 500;
		color: var(--ink-3);
		background: var(--bg-soft);
		padding: var(--space-1) var(--space-3);
		border-radius: var(--radius-pill);
	}

	.cards {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: var(--space-3);
	}

	.card {
		display: flex;
		justify-content: space-between;
		align-items: flex-start;
		gap: var(--space-4);
		min-width: 0;
		padding: var(--space-4);
		background: var(--paper);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-lg);
		box-shadow: var(--shadow-sm);
	}

	.card.removed {
		background: var(--bg-soft);
		opacity: 0.72;
	}

	.card-main {
		flex: 1;
		min-width: 0;
	}

	.card-meta {
		margin-bottom: var(--space-2);
	}

	.status {
		font-size: 11px;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		font-weight: 600;
		color: var(--primary-2);
	}

	.card-title {
		font-family: var(--font-display);
		font-size: 22px;
		line-height: 1.2;
		font-weight: 400;
		color: var(--ink);
		margin: 0 0 var(--space-2);
	}

	.summary {
		font-size: 13px;
		color: var(--ink-3);
		margin: 0 0 var(--space-2);
	}

	.description {
		font-size: 14px;
		color: var(--ink-2);
		margin: 0 0 var(--space-3);
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

	.card-actions {
		display: flex;
		flex-direction: column;
		align-items: flex-end;
		gap: var(--space-3);
		flex-shrink: 0;
		min-width: 0;
	}

	.secondary-actions {
		display: flex;
		gap: var(--space-2);
	}

	.danger-action :global(.btn-text) {
		color: var(--rose);
		text-decoration-color: var(--rose-soft);
	}

	.danger-action :global(.btn-text:hover:not(:disabled)) {
		color: #9e5e5e;
	}

	.read-only {
		font-size: 13px;
		color: var(--ink-3);
	}

	@media (max-width: 540px) {
		.card {
			flex-direction: column;
			align-items: stretch;
		}

		.card-actions {
			flex-direction: row;
			justify-content: space-between;
			align-items: center;
		}
	}
</style>
