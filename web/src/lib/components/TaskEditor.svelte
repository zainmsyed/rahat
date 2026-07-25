<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import type { TaskDraft } from '$lib/api/onboarding';

	export let draft: TaskDraft;
	export let saving = false;
	export let submitLabel = 'Save task';
	export let error = '';

	const dispatch = createEventDispatcher<{
		save: { draft: TaskDraft };
		cancel: void;
	}>();

	let workingDraft: TaskDraft = cloneDraft(draft);

	$: workingDraft = cloneDraft(draft);

	function cloneDraft(value: TaskDraft): TaskDraft {
		return {
			...value,
			subtasks: value.subtasks.map((subtask) => ({ ...subtask }))
		};
	}

	function addSubtask() {
		workingDraft = {
			...workingDraft,
			subtasks: [
				...workingDraft.subtasks,
				{
					name: '',
					duration_minutes: 10,
					time_of_day_preference: 'morning',
					min_gap_after_previous_minutes: 0
				}
			]
		};
	}

	function removeSubtask(index: number) {
		workingDraft = {
			...workingDraft,
			subtasks: workingDraft.subtasks.filter((_, itemIndex) => itemIndex !== index)
		};
	}

	function submit() {
		dispatch('save', { draft: cloneDraft(workingDraft) });
	}
</script>

<div class="editor">
	<div class="grid">
		<label>
			<span>Task name *</span>
			<input bind:value={workingDraft.name} placeholder="Example: Wipe down the kitchen" />
		</label>

		<label>
			<span>How many minutes does this usually take? *</span>
			<input bind:value={workingDraft.duration_minutes} type="number" min="1" max="240" />
		</label>
	</div>

	<label>
		<span>Helpful note (optional)</span>
		<textarea bind:value={workingDraft.description} rows="3" placeholder="Add any reminder that would help later."></textarea>
	</label>

	<div class="grid three">
		<label>
			<span>How often should Rahat plan this? *</span>
			<select bind:value={workingDraft.cadence_type}>
				<option value="interval">Every few days</option>
				<option value="count">A few times each week</option>
			</select>
		</label>

		<label>
			<span>How many? *</span>
			<input bind:value={workingDraft.cadence_value} type="number" min="1" max="31" />
		</label>

		<label>
			<span>Priority *</span>
			<select bind:value={workingDraft.priority}>
				<option value="high">High</option>
				<option value="medium">Medium</option>
				<option value="low">Low</option>
			</select>
		</label>
	</div>

	<label>
		<span>Best time of day *</span>
		<select bind:value={workingDraft.time_of_day_preference}>
			<option value="morning">Morning</option>
			<option value="afternoon">Afternoon</option>
			<option value="evening">Evening</option>
			<option value="any">Any time</option>
		</select>
	</label>

	<section class="subtasks">
		<div class="subtask-header">
			<div>
				<h3>Optional smaller steps</h3>
				<p>If this task works better as steps, add them here.</p>
			</div>
			<button type="button" on:click={addSubtask}>Add a step</button>
		</div>

		{#if workingDraft.subtasks.length === 0}
			<p class="empty">No steps added. That is okay for a simple one-step task.</p>
		{/if}

		{#each workingDraft.subtasks as subtask, index}
			<div class="subtask-card">
				<div class="subtask-headline">
					<h4>Step {index + 1}</h4>
					<button type="button" class="ghost" on:click={() => removeSubtask(index)}>Remove</button>
				</div>
				<div class="grid three">
					<label>
						<span>Step name *</span>
						<input bind:value={subtask.name} placeholder="Example: Move to dryer" />
					</label>
					<label>
						<span>Minutes *</span>
						<input bind:value={subtask.duration_minutes} type="number" min="1" max="240" />
					</label>
					<label>
						<span>Best time *</span>
						<select bind:value={subtask.time_of_day_preference}>
							<option value="morning">Morning</option>
							<option value="afternoon">Afternoon</option>
							<option value="evening">Evening</option>
							<option value="any">Any time</option>
						</select>
					</label>
				</div>
				<label>
					<span>Wait this many minutes after the previous step (optional)</span>
					<input bind:value={subtask.min_gap_after_previous_minutes} type="number" min="0" max="1440" />
				</label>
			</div>
		{/each}
	</section>

	{#if error}
		<p class="error">{error}</p>
	{/if}

	<div class="actions">
		<button type="button" class="ghost" on:click={() => dispatch('cancel')}>Cancel</button>
		<button type="button" on:click={submit} disabled={saving}>{saving ? 'Saving…' : submitLabel}</button>
	</div>
</div>

<style>
	.editor {
		display: grid;
		gap: 1rem;
	}

	.grid {
		display: grid;
		gap: 1rem;
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.grid.three {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}

	label {
		display: grid;
		gap: 8px;
		font-weight: 600;
	}

	span {
		font-size: 0.95rem;
	}

	input,
	textarea,
	select {
		padding: 12px 16px;
		border-radius: 12px;
		border: 1.5px solid var(--rahat-line, #e3dccc);
		font: inherit;
		color: var(--rahat-ink, #1f1d1a);
		background: var(--rahat-paper, #ffffff);
	}

	input:focus,
	textarea:focus,
	select:focus {
		outline: none;
		border-color: var(--rahat-primary, #7a9b76);
		box-shadow: 0 0 0 4px var(--rahat-primary-glow, rgba(122, 155, 118, 0.22));
	}

	textarea {
		resize: vertical;
	}

	.subtasks {
		padding: 1rem;
		border-radius: 1rem;
		background: var(--rahat-surface-soft, #f4f0e6);
		border: 1px solid var(--rahat-line-soft, #ebe5d8);
	}

	.subtask-header,
	.subtask-headline,
	.actions {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: 1rem;
	}

	h3,
	h4,
	p {
		margin: 0;
	}

	.subtask-card {
		margin-top: 1rem;
		padding-top: 1rem;
		border-top: 1px solid var(--rahat-line, #e3dccc);
		display: grid;
		gap: 12px;
	}

	.empty {
		margin-top: 0.75rem;
		color: var(--rahat-ink-muted, #8a8278);
	}

	button {
		padding: 12px 16px;
		border: none;
		border-radius: 999px;
		background: var(--rahat-primary, #7a9b76);
		color: var(--rahat-paper, #ffffff);
		font: inherit;
		font-weight: 700;
		cursor: pointer;
	}

	button.ghost {
		background: var(--rahat-paper, #ffffff);
		color: var(--rahat-ink-secondary, #4a4640);
		border: 1.5px solid var(--rahat-line, #e3dccc);
	}

	button:disabled {
		opacity: 0.65;
		cursor: wait;
	}

	.error {
		color: var(--rahat-rose, #b87a7a);
		font-weight: 600;
	}

	@media (max-width: 720px) {
		.grid,
		.grid.three {
			grid-template-columns: 1fr;
		}

		.subtask-header,
		.subtask-headline,
		.actions {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
