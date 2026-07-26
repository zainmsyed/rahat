<script lang="ts">
	import Button from '$lib/components/design/Button.svelte';
</script>

<svelte:head>
	<title>Rahat</title>
	<meta name="description" content="Calm, adaptive routines for busy parents." />
</svelte:head>

<main class="stage">
	<div class="card hero-card">
		<p class="eyebrow">Welcome</p>
		<h1 class="display">Rahat</h1>
		<p class="lede">
			A calm, adaptive routine planner that fits around your calendar. Tune your routines once,
			then let Rahat keep the schedule current.
		</p>
		<div class="actions">
			<Button variant="primary" href="/tasks" fullWidth>Manage routines</Button>
		</div>
	</div>
</main>

<style>
	.hero-card {
		display: grid;
		gap: var(--space-4);
		text-align: center;
	}

	.hero-card .display {
		font-size: clamp(2.5rem, 10vw, 4rem);
	}

	.hero-card .lede {
		max-width: 36ch;
		margin: 0 auto;
	}

	.actions {
		margin-top: var(--space-2);
	}
</style>
