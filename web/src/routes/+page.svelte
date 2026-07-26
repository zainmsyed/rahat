<script lang="ts">
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
			<a class="cta" href="/tasks">Manage routines</a>
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

	.cta {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 100%;
		padding: 15px 20px;
		border: none;
		border-radius: var(--radius-lg);
		background: var(--primary);
		color: white;
		font-size: 15px;
		font-weight: 600;
		letter-spacing: 0.01em;
		text-decoration: none;
		box-shadow: var(--shadow-button);
		transition: all 0.2s var(--ease-out);
		cursor: pointer;
	}

	.cta:hover {
		background: var(--primary-2);
		transform: translateY(-1px);
		box-shadow: var(--shadow-button-hover);
	}
</style>
