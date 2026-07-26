<script lang="ts">
	import Toggle from './Toggle.svelte';

	export let id = 'toggle';
	export let label = '';
	export let checked = false;
	export let onChange: (event: CustomEvent<{ checked: boolean }>) => void = () => {};
</script>

<Toggle {id} {label} {checked} on:change={onChange} />
