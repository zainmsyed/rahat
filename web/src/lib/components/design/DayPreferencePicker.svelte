<script context="module" lang="ts">
	export type DayPreference = 'any' | 'weekday' | 'weekend';
	export type DayPreferencePickerVariant = 'cards' | 'segmented';
</script>

<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let value: DayPreference = 'any';
	export let variant: DayPreferencePickerVariant = 'cards';

	const dispatch = createEventDispatcher<{ value: DayPreference }>();

	const options: Array<{ value: DayPreference; name: string; hint: string; compact: string }> = [
		{ value: 'any', name: 'Any day is fine', hint: 'Rahat will plan it whenever it fits best.', compact: 'Any day' },
		{ value: 'weekday', name: 'Weekdays only', hint: 'Planned Monday to Friday, never on weekends.', compact: 'Weekdays' },
		{ value: 'weekend', name: 'Weekends only', hint: 'Planned on Saturday and Sunday.', compact: 'Weekends' }
	];

	function select(next: DayPreference) {
		value = next;
		dispatch('value', next);
	}
</script>

<div class="picker" class:segmented={variant === 'segmented'} aria-label="Which days work best for this task?">
	{#if variant === 'cards'}
		<div class="radio-cards">
			{#each options as option}
				<button
					type="button"
					class="radio-card"
					class:selected={value === option.value}
					aria-pressed={value === option.value}
					on:click={() => select(option.value)}
				>
					<span class="radio-dot" aria-hidden="true"></span>
					<span class="option-copy">
						<span class="option-name">{option.name}</span>
						<span class="option-hint">{option.hint}</span>
					</span>
				</button>
			{/each}
		</div>
	{:else}
		<div class="segmented-control">
			{#each options as option}
				<button
					type="button"
					class="segment"
					class:selected={value === option.value}
					aria-pressed={value === option.value}
					on:click={() => select(option.value)}
				>
					{option.compact}
				</button>
			{/each}
		</div>
	{/if}
</div>

<style>
	.picker { display: grid; gap: var(--space-2); min-width: 0; }
	.radio-cards { display: grid; gap: var(--space-2); }
	.radio-card {
		appearance: none;
		width: 100%;
		display: flex;
		align-items: center;
		gap: var(--space-3);
		padding: var(--space-3) var(--space-4);
		text-align: left;
		font: inherit;
		color: var(--ink);
		background: var(--paper);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-md);
		cursor: pointer;
		transition: border-color 0.2s var(--ease-out), background 0.2s var(--ease-out), box-shadow 0.2s var(--ease-out);
	}
	.radio-card:hover, .radio-card.selected { border-color: var(--primary); }
	.radio-card.selected { background: var(--primary-bg); box-shadow: 0 0 0 4px var(--primary-glow); }
	.radio-dot {
		width: 18px;
		height: 18px;
		flex: 0 0 18px;
		border: 2px solid var(--line);
		border-radius: 50%;
		display: grid;
		place-items: center;
	}
	.radio-card.selected .radio-dot { border-color: var(--primary); }
	.radio-card.selected .radio-dot::after { content: ''; width: 8px; height: 8px; border-radius: 50%; background: var(--primary); }
	.option-copy { display: grid; gap: var(--space-1); min-width: 0; }
	.option-name { font-size: 15px; font-weight: 600; }
	.option-hint { font-size: 13px; color: var(--ink-3); overflow-wrap: anywhere; }
	.segmented-control {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: var(--space-1);
		padding: var(--space-1);
		background: var(--bg-soft);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-md);
	}
	.segment {
		min-width: 0;
		padding: var(--space-2) var(--space-1);
		border: 0;
		border-radius: var(--radius-md);
		font: inherit;
		font-size: 13px;
		color: var(--ink-2);
		background: transparent;
		cursor: pointer;
	}
	.segment:hover { color: var(--ink); }
	.segment.selected { color: var(--ink); background: var(--paper); font-weight: 600; box-shadow: var(--shadow-sm); }
</style>
