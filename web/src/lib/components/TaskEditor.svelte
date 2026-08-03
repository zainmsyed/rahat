<script lang="ts">
	import { createEventDispatcher } from 'svelte';
	import Input from '$lib/components/design/Input.svelte';
	import Button from '$lib/components/design/Button.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import DayPreferencePicker, { type DayPreferencePickerVariant } from '$lib/components/design/DayPreferencePicker.svelte';
	import type { TaskDraft } from '$lib/api/onboarding';

	export let draft: TaskDraft;
	export let saving = false;
	export let submitLabel = 'Save task';
	export let error = '';
	export let dayPickerVariant: DayPreferencePickerVariant = 'cards';

	const dispatch = createEventDispatcher<{
		save: { draft: TaskDraft };
		cancel: void;
	}>();

	let workingDraft: TaskDraft = cloneDraft(draft);

	$: workingDraft = cloneDraft(draft);
	$: if (workingDraft.day_preference === 'weekend' && (workingDraft.cadence_type !== 'count' || workingDraft.cadence_value > 2)) {
		workingDraft = { ...workingDraft, cadence_type: 'count', cadence_value: Math.min(Math.max(workingDraft.cadence_value || 1, 1), 2) };
	}

	function cloneDraft(value: TaskDraft): TaskDraft {
		return {
			...value,
			day_preference: value.day_preference ?? 'any',
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
	<div class="row two">
		<Input
			id="task-name"
			label="Task name"
			placeholder="Example: Wipe down the kitchen"
			required
			bind:value={workingDraft.name}
		/>
		<Input
			id="task-duration"
			label="How many minutes does this usually take?"
			type="number"
			min={1}
			max={240}
			required
			bind:value={workingDraft.duration_minutes}
		/>
	</div>

	<label class="field" for="task-description">
		<span class="field-label">Helpful note <span class="optional">(optional)</span></span>
		<textarea
			id="task-description"
			bind:value={workingDraft.description}
			rows="3"
			placeholder="Add any reminder that would help later."
		></textarea>
	</label>

	<div class="row three">
		<label class="field" for="task-cadence-type">
			<span class="field-label">How often should Rahat plan this?</span>
			<select id="task-cadence-type" bind:value={workingDraft.cadence_type}>
				<option value="interval">Every few days</option>
				<option value="count">A few times each week</option>
			</select>
		</label>
		<Input
			id="task-cadence-value"
			label="How many?"
			type="number"
			min={1}
			max={31}
			required
			bind:value={workingDraft.cadence_value}
		/>
		<label class="field" for="task-priority">
			<span class="field-label">Priority</span>
			<select id="task-priority" bind:value={workingDraft.priority}>
				<option value="high">High</option>
				<option value="medium">Medium</option>
				<option value="low">Low</option>
			</select>
		</label>
	</div>

	<label class="field" for="task-time">
		<span class="field-label">Best time of day</span>
		<select id="task-time" bind:value={workingDraft.time_of_day_preference}>
			<option value="morning">Morning</option>
			<option value="afternoon">Afternoon</option>
			<option value="evening">Evening</option>
			<option value="any">Any time</option>
		</select>
	</label>

	<div class="field">
		<span class="field-label">Which days work best for this task?</span>
		<DayPreferencePicker variant={dayPickerVariant} bind:value={workingDraft.day_preference} />
	</div>
	{#if workingDraft.day_preference === 'weekend'}
		<InfoBox title="Weekend cadence updated">
			Weekend tasks are planned per week — up to 2 times, once per weekend day.
		</InfoBox>
	{/if}

	<section class="subtasks" aria-labelledby="subtasks-title">
		<div class="subtasks-header">
			<div>
				<h3 id="subtasks-title">Optional smaller steps</h3>
				<p class="subtasks-lede">If this task works better as steps, add them here.</p>
			</div>
			<Button variant="secondary" on:click={addSubtask}>Add a step</Button>
		</div>

		{#if workingDraft.subtasks.length === 0}
			<InfoBox>No steps added. That is okay for a simple one-step task.</InfoBox>
		{/if}

		{#each workingDraft.subtasks as subtask, index}
			<div class="subtask-card">
				<div class="subtask-headline">
					<h4>Step {index + 1}</h4>
					<Button variant="text" on:click={() => removeSubtask(index)}>Remove</Button>
				</div>
				<div class="row three">
					<Input
						id="subtask-{index}-name"
						label="Step name"
						placeholder="Example: Move to dryer"
						required
						bind:value={subtask.name}
					/>
					<Input
						id="subtask-{index}-duration"
						label="Minutes"
						type="number"
						min={1}
						max={240}
						required
						bind:value={subtask.duration_minutes}
					/>
					<label class="field" for="subtask-{index}-time">
						<span class="field-label">Best time</span>
						<select id="subtask-{index}-time" bind:value={subtask.time_of_day_preference}>
							<option value="morning">Morning</option>
							<option value="afternoon">Afternoon</option>
							<option value="evening">Evening</option>
							<option value="any">Any time</option>
						</select>
					</label>
				</div>
				<Input
					id="subtask-{index}-gap"
					label="Wait this many minutes after the previous step (optional)"
					type="number"
					min={0}
					max={1440}
					bind:value={subtask.min_gap_after_previous_minutes}
				/>
			</div>
		{/each}
	</section>

	{#if error}
		<p class="error-banner" role="alert">{error}</p>
	{/if}

	<div class="actions">
		<Button variant="secondary" on:click={() => dispatch('cancel')}>Cancel</Button>
		<Button variant="primary" disabled={saving} on:click={submit}>
			{saving ? 'Saving…' : submitLabel}
		</Button>
	</div>
</div>

<style>
	.editor {
		display: grid;
		gap: var(--space-5);
	}

	.row {
		display: grid;
		gap: var(--space-4);
		grid-template-columns: 1fr;
	}

	.row.two {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.row.three {
		grid-template-columns: repeat(3, minmax(0, 1fr));
	}

	.field {
		display: grid;
		gap: var(--space-2);
	}

	.field-label {
		display: block;
		font-size: 13px;
		font-weight: 500;
		color: var(--ink-2);
	}

	.field-label .optional {
		color: var(--ink-3);
		font-weight: 400;
	}

	textarea,
	select {
		width: 100%;
		padding: 14px 16px;
		background: var(--paper);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-md);
		font-size: 16px;
		color: var(--ink);
		font-family: inherit;
		transition: border-color 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
	}

	textarea::placeholder,
	select::placeholder {
		color: var(--ink-4);
	}

	textarea:focus,
	select:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 4px var(--primary-glow);
	}

	textarea {
		resize: vertical;
		min-height: 80px;
	}

	select {
		appearance: none;
		-webkit-appearance: none;
		background-image: url("data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='16' height='16' viewBox='0 0 24 24' fill='none' stroke='%238a8278' stroke-width='2' stroke-linecap='round' stroke-linejoin='round'%3E%3Cpath d='m6 9 6 6 6-6'/%3E%3C/svg%3E");
		background-repeat: no-repeat;
		background-position: right 14px center;
		padding-right: 40px;
		cursor: pointer;
	}

	select option {
		background: var(--paper);
		color: var(--ink);
	}

	select option:checked,
	select option:hover {
		background: var(--primary-bg);
		color: var(--primary-2);
	}

	.subtasks {
		display: grid;
		gap: var(--space-4);
		padding: var(--space-5);
		background: var(--bg-soft);
		border: 1px solid var(--line-soft);
		border-radius: var(--radius-lg);
	}

	.subtasks-header,
	.subtask-headline {
		display: flex;
		justify-content: space-between;
		align-items: center;
		gap: var(--space-4);
		min-width: 0;
	}

	.subtasks-header h3,
	.subtask-headline h4 {
		font-size: 15px;
		font-weight: 500;
		color: var(--ink);
		margin: 0;
	}

	.subtasks-lede {
		font-size: 13px;
		color: var(--ink-3);
		margin: var(--space-1) 0 0;
	}

	.subtask-card {
		display: grid;
		gap: var(--space-4);
		padding: var(--space-4);
		background: var(--paper);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-lg);
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
		.row.two,
		.row.three {
			grid-template-columns: 1fr;
		}

		.subtasks-header,
		.subtask-headline,
		.actions {
			flex-direction: column;
			align-items: stretch;
		}
	}
</style>
