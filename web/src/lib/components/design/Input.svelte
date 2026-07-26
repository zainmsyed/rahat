<script lang="ts">
	export let id: string;
	export let label: string;
	export let value: string | number = '';
	export let placeholder = '';
	export let type: 'text' | 'email' | 'password' | 'number' = 'text';
	export let required = false;
	export let error = '';
	export let invalid = false;
	export let min: number | string = '';
	export let max: number | string = '';
</script>

<div class="input-wrapper">
	<label class="label" for={id}>{label}{required ? ' *' : ''}</label>
	<input
		id={id}
		name={id}
		type={type}
		placeholder={placeholder}
		required={required}
		min={min}
		max={max}
		bind:value
		class:error={error !== '' || invalid}
	/>
	{#if error}
		<p class="error-text">{error}</p>
	{/if}
</div>

<style>
	.input-wrapper {
		display: grid;
		gap: var(--space-2);
	}

	input {
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

	input::placeholder {
		color: var(--ink-4);
	}

	input:focus {
		outline: none;
		border-color: var(--primary);
		box-shadow: 0 0 0 4px var(--primary-glow);
	}

	input.error {
		border-color: var(--rose);
	}

	.error-text {
		font-size: 13px;
		color: var(--rose);
		font-weight: 500;
	}
</style>
