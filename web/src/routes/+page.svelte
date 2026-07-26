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
	.page {
		max-width: var(--surface-max-width);
		margin: 0 auto;
		padding: var(--space-10) var(--space-5) var(--space-12);
	}

	.hero {
		padding: var(--space-6);
		border-radius: var(--radius-3xl);
		background: var(--paper);
		border: 1px solid var(--line);
		box-shadow: var(--shadow-md);
	}

	.eyebrow {
		margin: 0 0 var(--space-3);
		font-size: 11px;
		font-weight: 600;
		letter-spacing: 0.18em;
		text-transform: uppercase;
		color: var(--primary-2);
	}

	h1 {
		margin: 0;
		font-family: var(--font-display);
		font-size: clamp(2rem, 6vw, 3rem);
		font-weight: 400;
		color: var(--ink);
	}

	.lede {
		max-width: 42rem;
		font-size: 15.5px;
		color: var(--ink-2);
		line-height: 1.6;
	}

	.actions {
		display: flex;
		gap: var(--space-3);
		flex-wrap: wrap;
		margin-top: var(--space-5);
	}

	.actions a {
		padding: 14px 20px;
		border-radius: var(--radius-lg);
		text-decoration: none;
		font-weight: 600;
		background: var(--primary);
		color: white;
		transition: all 0.2s var(--ease-out);
	}

	.actions a:hover {
		background: var(--primary-2);
		transform: translateY(-1px);
	}

	.card-grid {
		display: grid;
		grid-template-columns: 1fr;
		gap: var(--space-4);
		margin-top: var(--space-6);
	}

	article {
		padding: var(--space-4);
		border-radius: var(--radius-xl);
		background: var(--paper);
		border: 1px solid var(--line);
		box-shadow: var(--shadow-sm);
	}

	article h2 {
		margin-top: 0;
		font-size: 16px;
		font-weight: 600;
		color: var(--ink);
	}

	article p {
		font-size: 14px;
		color: var(--ink-2);
	}
</style>
