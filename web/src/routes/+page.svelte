<script lang="ts">
	const apiBaseUrl = import.meta.env.VITE_API_BASE_URL ?? 'http://localhost:8080';
</script>

<svelte:head>
	<title>Rahat Workspace</title>
	<meta
		name="description"
		content="Rahat starter workspace with Go, SQLite, and SvelteKit foundations."
	/>
</svelte:head>

<div class="page">
	<section class="hero">
		<p class="eyebrow">Rahat</p>
		<h1>Workspace bootstrapped.</h1>
		<p class="lede">
			This placeholder shell confirms the SvelteKit frontend is wired and ready for later product
			stories.
		</p>
		<div class="actions">
			<a href="/tasks">Manage routines</a>
			<a href={`${apiBaseUrl}/healthz`} target="_blank" rel="noreferrer">API health endpoint</a>
			<a href={`${apiBaseUrl}/readyz`} target="_blank" rel="noreferrer">API readiness endpoint</a>
		</div>
	</section>

	<section class="card-grid">
		<article>
			<h2>Backend</h2>
			<p>Go HTTP server with JSON logging, health checks, and SQLite bootstrap.</p>
		</article>
		<article>
			<h2>Database</h2>
			<p>SQLite opens locally, enables WAL mode, and confirms connectivity on startup.</p>
		</article>
		<article>
			<h2>Frontend</h2>
			<p>SvelteKit provides the web shell for future onboarding and scheduling workflows.</p>
		</article>
	</section>
</div>

<style>
	:global(body) {
		margin: 0;
		font-family: Inter, system-ui, sans-serif;
		background: #f7f7fb;
		color: #18222f;
	}

	.page {
		max-width: 960px;
		margin: 0 auto;
		padding: 4rem 1.5rem 5rem;
	}

	.hero {
		padding: 2rem;
		border-radius: 1.5rem;
		background: linear-gradient(135deg, #ffffff, #eef4ff);
		box-shadow: 0 18px 45px rgba(24, 34, 47, 0.08);
	}

	.eyebrow {
		margin: 0 0 0.75rem;
		font-size: 0.875rem;
		font-weight: 700;
		letter-spacing: 0.14em;
		text-transform: uppercase;
		color: #2a6df4;
	}

	h1 {
		margin: 0;
		font-size: clamp(2.5rem, 6vw, 4rem);
	}

	.lede {
		max-width: 42rem;
		font-size: 1.125rem;
		line-height: 1.6;
	}

	.actions {
		display: flex;
		gap: 1rem;
		flex-wrap: wrap;
		margin-top: 1.5rem;
	}

	.actions a {
		padding: 0.85rem 1.15rem;
		border-radius: 999px;
		text-decoration: none;
		font-weight: 600;
		background: #18222f;
		color: white;
	}

	.card-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 1rem;
		margin-top: 2rem;
	}

	article {
		padding: 1.25rem;
		border-radius: 1rem;
		background: white;
		box-shadow: 0 10px 30px rgba(24, 34, 47, 0.08);
	}

	h2 {
		margin-top: 0;
	}
</style>
