<script lang="ts">
	export let variant: 'primary' | 'secondary' | 'text' = 'primary';
	export let type: 'button' | 'submit' | 'reset' = 'button';
	export let disabled = false;
	export let fullWidth = false;
	export let href = '';
</script>

{#if href && !disabled}
	<a
		{href}
		class:btn-primary={variant === 'primary'}
		class:btn-secondary={variant === 'secondary'}
		class:btn-text={variant === 'text'}
		class:full-width={fullWidth}
		on:click
	>
		<slot />
	</a>
{:else}
	<button
		{type}
		{disabled}
		class:btn-primary={variant === 'primary'}
		class:btn-secondary={variant === 'secondary'}
		class:btn-text={variant === 'text'}
		class:full-width={fullWidth}
		on:click
	>
		<slot />
	</button>
{/if}

<style>
	button,
	a {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--space-2);
		padding: 15px 20px;
		border: none;
		border-radius: var(--radius-lg);
		font-size: 15px;
		font-weight: 600;
		cursor: pointer;
		font-family: inherit;
		letter-spacing: 0.01em;
		transition: all 0.2s var(--ease-out);
		min-width: 200px;
		text-decoration: none;
	}

	button:disabled {
		background: var(--ink-4);
		color: var(--paper);
		cursor: not-allowed;
		box-shadow: none;
	}

	.full-width {
		width: 100%;
	}

	.btn-primary {
		background: var(--primary);
		color: white;
		box-shadow: var(--shadow-button);
	}

	.btn-primary:hover:not(:disabled) {
		background: var(--primary-2);
		transform: translateY(-1px);
		box-shadow: var(--shadow-button-hover);
	}

	.btn-secondary {
		background: transparent;
		color: var(--ink-2);
		border: 1.5px solid var(--line);
	}

	.btn-secondary:hover:not(:disabled) {
		border-color: var(--ink-4);
	}

	.btn-text {
		background: none;
		color: var(--ink-3);
		padding: 8px 4px;
		min-width: auto;
		text-decoration: underline;
		text-decoration-color: var(--line);
		text-underline-offset: 3px;
	}

	.btn-text:hover:not(:disabled) {
		color: var(--ink-2);
	}
</style>
