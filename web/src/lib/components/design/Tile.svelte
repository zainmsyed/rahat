<script lang="ts">
	export let title: string;
	export let subtitle = '';
	export let selected = false;
	export let icon = '';
</script>

<button type="button" class="tile" class:sel={selected} on:click>
	{#if icon}
		<div class="tile-icon">{icon}</div>
	{/if}
	<div class="tile-body">
		<div class="tile-title">{title}</div>
		{#if subtitle}
			<div class="tile-sub">{subtitle}</div>
		{/if}
	</div>
	<div class="tile-check" class:sel={selected}>
		{#if selected}✓{/if}
	</div>
</button>

<style>
	.tile {
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
	}

	.tile:hover {
		border-color: var(--ink-4);
	}

	.tile.sel {
		border-color: var(--primary);
		background: var(--primary-bg);
		box-shadow: 0 0 0 4px var(--primary-soft);
	}

	.tile.sel .tile-icon {
		background: var(--primary);
		color: white;
	}

	.tile-icon {
		width: 38px;
		height: 38px;
		border-radius: var(--radius-md);
		background: var(--bg-soft);
		display: flex;
		align-items: center;
		justify-content: center;
		color: var(--ink-2);
		flex-shrink: 0;
		font-size: 18px;
		transition: all 0.2s var(--ease-out);
	}

	.tile-body {
		flex: 1;
	}

	.tile-title {
		font-size: 15px;
		font-weight: 500;
		color: var(--ink);
	}

	.tile-sub {
		font-size: 13px;
		color: var(--ink-3);
	}

	.tile-check {
		width: 22px;
		height: 22px;
		border-radius: 50%;
		border: 1.5px solid var(--line);
		display: flex;
		align-items: center;
		justify-content: center;
		background: var(--paper);
		flex-shrink: 0;
		font-size: 12px;
		color: transparent;
		transition: all 0.2s var(--ease-out);
	}

	.tile-check.sel {
		background: var(--primary);
		border-color: var(--primary);
		color: white;
	}
</style>
