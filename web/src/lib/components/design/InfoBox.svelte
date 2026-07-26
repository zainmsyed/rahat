<script lang="ts">
	export let title = '';
</script>

<div class="info-box">
	<svg width="20" height="20" viewBox="0 0 24 24" fill="none" stroke="currentColor" stroke-width="2" stroke-linecap="round" stroke-linejoin="round">
		<circle cx="12" cy="12" r="10" />
		<line x1="12" y1="16" x2="12" y2="12" />
		<line x1="12" y1="8" x2="12.01" y2="8" />
	</svg>
	<div class="info-body">
		{#if title}
			<strong>{title}</strong>
		{/if}
		<slot />
	</div>
</div>

<style>
	.info-box {
		background: var(--bg-soft);
		border: 1px solid var(--line-soft);
		border-radius: var(--radius-lg);
		padding: 16px 18px;
		font-size: 13.5px;
		color: var(--ink-2);
		display: flex;
		gap: var(--space-3);
		align-items: flex-start;
	}

	.info-box svg {
		color: var(--primary-2);
		flex-shrink: 0;
		margin-top: 2px;
	}

	.info-body {
		display: grid;
		gap: var(--space-1);
	}

	.info-body strong {
		color: var(--ink);
		font-weight: 500;
	}
</style>
