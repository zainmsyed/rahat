<script lang="ts">
	import { onMount } from 'svelte';
	import LookaheadDay from '$lib/components/schedule/LookaheadDay.svelte';
	import { getLookaheadPlan, readLookaheadTokenFromUrl, type LookaheadResponse } from '$lib/api/lookahead';

	let loading = true;
	let pageError = '';
	let plan: LookaheadResponse | null = null;

	onMount(async () => {
		const token = readLookaheadTokenFromUrl();
		if (!token) {
			pageError = 'This lookahead link is missing its access token.';
			loading = false;
			return;
		}
		try {
			plan = await getLookaheadPlan(token);
		} catch (error) {
			pageError = error instanceof Error ? error.message : 'Could not load your lookahead.';
		} finally {
			loading = false;
		}
	});
</script>

<svelte:head>
	<title>Rahat lookahead</title>
	<meta name="description" content="Read-only Rahat schedule lookahead for today and tomorrow." />
</svelte:head>

<main class="page">
	<header class="hero">
		<p class="eyebrow">Rahat lookahead</p>
		<h1>Today and tomorrow, read-only.</h1>
		<p>
			This passive page shows what Rahat plans around your calendar. There are no edit,
			complete, or reschedule controls here.
		</p>
		{#if plan}
			<p class="user">For {plan.user.display_name || 'you'} · {plan.user.timezone}</p>
		{/if}
	</header>

	{#if loading}
		<section class="panel">Loading your lookahead…</section>
	{:else if pageError}
		<section class="panel error">
			<h2>We could not open this lookahead link.</h2>
			<p>{pageError}</p>
			<p>Ask Rahat for a fresh link if this one has expired.</p>
		</section>
	{:else if plan}
		<section class="days" aria-label="Today and tomorrow schedule">
			{#each plan.days as day}
				<LookaheadDay {day} />
			{/each}
		</section>
	{/if}
</main>

<style>
	:global(body) {
		margin: 0;
		font-family: Inter, system-ui, sans-serif;
		background: #f4f7fb;
		color: #14202c;
	}

	.page {
		max-width: 860px;
		margin: 0 auto;
		padding: 1rem;
		display: grid;
		gap: 1rem;
	}

	.hero,
	.panel {
		padding: 1.1rem;
		border-radius: 1.25rem;
		background: white;
		box-shadow: 0 12px 30px rgba(20, 32, 44, 0.08);
	}

	.hero h1,
	.hero p,
	.panel h2,
	.panel p {
		margin: 0;
	}

	.hero {
		display: grid;
		gap: 0.6rem;
	}

	.hero p,
	.panel p {
		line-height: 1.55;
	}

	.eyebrow {
		font-size: 0.8rem;
		font-weight: 800;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #4f46e5;
	}

	.user {
		color: #5d6b82;
		font-weight: 700;
	}

	.days {
		display: grid;
		gap: 1rem;
	}

	.error {
		border: 1px solid #fecaca;
		background: #fff7f7;
	}

	@media (min-width: 780px) {
		.page {
			padding: 1.5rem;
		}
	}
</style>
