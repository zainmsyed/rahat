<script lang="ts">
	import { onMount } from 'svelte';
	import LookaheadDay from '$lib/components/schedule/LookaheadDay.svelte';
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import {
		getLookaheadPlan,
		readLookaheadTokenFromUrl,
		type LookaheadResponse
	} from '$lib/api/lookahead';

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
	<header class="hero card">
		<p class="eyebrow">Rahat lookahead</p>
		<h1 class="display">Today and tomorrow, read-only.</h1>
		<p class="lede">
			This passive page shows what Rahat plans around your calendar. There are no edit, complete,
			or reschedule controls here.
		</p>
		{#if plan}
			<p class="user">For {plan.user.display_name || 'you'} · {plan.user.timezone}</p>
		{/if}
	</header>

	{#if loading}
		<InfoBox title="Loading your lookahead…">One moment while we fetch the schedule.</InfoBox>
	{:else if pageError}
		<InfoBox title="We could not open this lookahead link.">
			{pageError} Ask Rahat for a fresh link if this one has expired.
		</InfoBox>
	{:else if plan}
		<section class="days" aria-label="Today and tomorrow schedule">
			{#each plan.days as day}
				<LookaheadDay {day} />
			{/each}
		</section>
	{/if}
</main>

<style>
	.page {
		max-width: 860px;
		margin: 0 auto;
		padding: var(--space-8) var(--space-5) var(--space-12);
		display: grid;
		gap: var(--space-5);
	}

	.hero {
		display: grid;
		gap: var(--space-3);
	}

	.hero .display {
		font-size: clamp(1.75rem, 5vw, 2.5rem);
	}

	.hero .lede {
		max-width: 54ch;
	}

	.user {
		margin-top: var(--space-2);
		font-size: 13px;
		font-weight: 500;
		color: var(--ink-3);
	}

	.days {
		display: grid;
		gap: var(--space-5);
	}

	@media (max-width: 540px) {
		.page {
			padding: var(--space-5) var(--space-4);
		}
	}
</style>
