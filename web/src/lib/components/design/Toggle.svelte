<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let id: string;
	export let checked = false;
	export let label = '';

	const dispatch = createEventDispatcher<{ change: { checked: boolean } }>();

	function toggle() {
		checked = !checked;
		dispatch('change', { checked });
		const button = document.getElementById(id);
		button?.dispatchEvent(new CustomEvent('change', { detail: { checked }, bubbles: true }));
	}
</script>

<div class="toggle">
	{#if label}
		<span class="toggle-label" id="{id}-label">{label}</span>
	{/if}
	<button
		{id}
		type="button"
		class="toggle-switch"
		role="switch"
		aria-checked={checked}
		aria-labelledby={label ? `${id}-label` : undefined}
		on:click={toggle}
	>
		<span class="toggle-thumb"></span>
	</button>
</div>

<style>
	.toggle {
		display: inline-flex;
		align-items: center;
		gap: var(--space-2);
	}

	.toggle-label {
		font-size: 13px;
		font-weight: 500;
		color: var(--ink-2);
	}

	.toggle-switch {
		width: 44px;
		height: 24px;
		border-radius: var(--radius-pill);
		background: var(--ink-4);
		border: none;
		padding: 0;
		position: relative;
		cursor: pointer;
		transition: background 0.2s var(--ease-out);
	}

	.toggle-switch[aria-checked='true'] {
		background: var(--primary);
	}

	.toggle-thumb {
		position: absolute;
		top: 2px;
		left: 2px;
		width: 20px;
		height: 20px;
		border-radius: 50%;
		background: var(--paper);
		box-shadow: 0 1px 3px rgba(31, 29, 26, 0.12);
		transition: transform 0.2s var(--ease-out);
	}

	.toggle-switch[aria-checked='true'] .toggle-thumb {
		transform: translateX(20px);
	}

	.toggle-switch:focus-visible {
		outline: none;
		box-shadow: 0 0 0 4px var(--primary-glow);
	}
</style>
