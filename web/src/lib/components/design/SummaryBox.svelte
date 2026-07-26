<script lang="ts">
	export let id: string;
	export let value: number | string;
	export let unit: string;
	export let hint: string;
</script>

<div {id} class="summary-box" role="status" aria-live="polite">
	<span class="summary-value">{value}</span>
	<span class="summary-unit">{unit}</span>
	<span id="{id}-hint" class="summary-hint">{hint}</span>
</div>

<style>
	.summary-box {
		display: grid;
		gap: var(--space-1);
		justify-items: start;
		padding: var(--space-5);
		background: var(--primary-bg);
		border: 1px solid var(--primary-soft);
		border-radius: var(--radius-lg);
	}

	.summary-value {
		font-family: var(--font-display);
		font-size: 60px;
		line-height: 1;
		color: var(--primary-2);
	}

	.summary-unit {
		font-size: 15px;
		font-weight: 500;
		color: var(--ink-2);
	}

	.summary-hint {
		font-size: 13px;
		color: var(--ink-3);
	}
</style>
