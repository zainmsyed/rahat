<script lang="ts">
	import { createEventDispatcher } from 'svelte';

	export let icon = '';
	export let name = '';
	export let subtitle = '';
	export let connected = false;
	export let href = '';
	export let target = '';
	export let rel = '';

	const dispatch = createEventDispatcher<{ click: void }>();

	$: statusText = connected ? 'Connected' : 'Not connected';
</script>

{#if href}
	<a
		class="connect-tile"
		class:connected
		{href}
		{target}
		{rel}
		on:click={() => dispatch('click')}
	>
		{#if icon}
			<div class="connect-icon">{icon}</div>
		{/if}
		<div class="connect-body">
			<div class="connect-name">{name}</div>
			{#if subtitle}
				<div class="connect-sub">{subtitle}</div>
			{/if}
			<div class="connect-status">
				<span class="status-dot" class:connected></span>
				<span class="status-text">{statusText}</span>
			</div>
		</div>
	</a>
{:else}
	<button
		type="button"
		class="connect-tile"
		class:connected
		on:click={() => dispatch('click')}
	>
		{#if icon}
			<div class="connect-icon">{icon}</div>
		{/if}
		<div class="connect-body">
			<div class="connect-name">{name}</div>
			{#if subtitle}
				<div class="connect-sub">{subtitle}</div>
			{/if}
			<div class="connect-status">
				<span class="status-dot" class:connected></span>
				<span class="status-text">{statusText}</span>
			</div>
		</div>
	</button>
{/if}

<style>
	.connect-tile,
	.connect-tile:visited {
		display: flex;
		align-items: center;
		gap: var(--space-3);
		width: 100%;
		padding: 16px 18px;
		background: var(--paper);
		border: 1.5px solid var(--line);
		border-radius: var(--radius-lg);
		cursor: pointer;
		transition: all 0.2s var(--ease-out);
		font-family: inherit;
		text-align: left;
		text-decoration: none;
		color: inherit;
	}

	.connect-tile:hover {
		border-color: var(--ink-4);
	}

	.connect-tile.connected {
		border-color: var(--primary);
		background: var(--primary-bg);
		box-shadow: 0 0 0 4px var(--primary-soft);
	}

	.connect-tile.connected .connect-icon {
		background: var(--primary);
		color: white;
	}

	.connect-icon {
		width: 42px;
		height: 42px;
		border-radius: var(--radius-md);
		background: var(--bg-soft);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--ink-2);
		flex-shrink: 0;
		font-size: 20px;
		transition: all 0.2s var(--ease-out);
	}

	.connect-body {
		flex: 1;
		min-width: 0;
	}

	.connect-name {
		font-size: 15px;
		font-weight: 500;
		color: var(--ink);
	}

	.connect-sub {
		font-size: 13px;
		color: var(--ink-3);
		margin-top: 2px;
	}

	.connect-status {
		display: flex;
		align-items: center;
		gap: var(--space-2);
		margin-top: var(--space-2);
	}

	.status-dot {
		width: 8px;
		height: 8px;
		border-radius: 50%;
		background: var(--ink-4);
	}

	.status-dot.connected {
		background: var(--primary);
	}

	.status-text {
		font-size: 12px;
		font-weight: 500;
		color: var(--ink-3);
	}

	.connect-tile.connected .status-text {
		color: var(--primary-2);
	}
</style>
