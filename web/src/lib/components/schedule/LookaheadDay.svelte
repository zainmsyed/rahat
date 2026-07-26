<script lang="ts">
	import InfoBox from '$lib/components/design/InfoBox.svelte';
	import { formatReadyTime, windows, type LookaheadDay } from '$lib/api/lookahead';

	export let day: LookaheadDay;
</script>

<article class="day-card card">
	<header class="day-header">
		<p class="eyebrow">{day.label}</p>
		<h2 class="display-sm">{day.date}</h2>
	</header>

	{#if day.small_task_only_reason}
		<InfoBox title="Small tasks only">{day.small_task_only_reason}</InfoBox>
	{/if}

	<div class="windows">
		{#each windows as window}
			<section class="window">
				<div class="window-header">
					<h3>{window}</h3>
					<span class="budget">{day.window_budgets_minutes[window] ?? 0} min open</span>
				</div>

				{#if day.blocked_windows[window]?.length}
					<InfoBox title="Calendar limits this window">
						<ul class="reasons">
							{#each day.blocked_windows[window] as reason}
								<li>{reason}</li>
							{/each}
						</ul>
					</InfoBox>
				{/if}

				{#if day.windows[window]?.length}
					<ul class="items">
						{#each day.windows[window] as item}
							<li>
								<strong>{item.name}</strong>
								<span>
									{item.duration_minutes} min
									{#if item.ready_at}
										· ready {formatReadyTime(item.ready_at)}
									{/if}
								</span>
							</li>
						{/each}
					</ul>
				{:else}
					<InfoBox>No tasks scheduled for this window.</InfoBox>
				{/if}
			</section>
		{/each}
	</div>

	{#if day.omitted_items.length}
		<section class="omitted">
			<InfoBox title="Tasks not shown in the plan">
				<p class="omitted-lede">
					Rahat is being conservative because of your calendar or time budget.
				</p>
				<ul class="omitted-items">
					{#each day.omitted_items as item}
						<li>
							<strong>{item.name}</strong>
							<span>{item.window}: {item.reason}</span>
						</li>
					{/each}
				</ul>
			</InfoBox>
		</section>
	{/if}
</article>

<style>
	.day-card {
		display: grid;
		gap: var(--space-5);
	}

	.day-header {
		display: grid;
		gap: var(--space-2);
	}

	.day-header .display-sm {
		font-family: var(--font-display);
		font-size: 26px;
		line-height: 1.1;
		font-weight: 400;
		color: var(--ink);
	}

	.windows {
		display: grid;
		gap: var(--space-4);
	}

	.window {
		display: grid;
		gap: var(--space-3);
		padding: var(--space-4);
		background: var(--bg-soft);
		border: 1px solid var(--line-soft);
		border-radius: var(--radius-xl);
	}

	.window-header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: var(--space-3);
	}

	.window-header h3 {
		font-size: 15px;
		font-weight: 600;
		text-transform: capitalize;
		color: var(--ink);
	}

	.budget {
		font-size: 13px;
		font-weight: 500;
		color: var(--ink-3);
	}

	.reasons {
		margin: 0;
		padding-left: var(--space-5);
		color: var(--ink-2);
	}

	.items {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: var(--space-3);
	}

	.items li {
		display: grid;
		gap: var(--space-1);
		padding: var(--space-3);
		background: var(--paper);
		border: 1px solid var(--line);
		border-radius: var(--radius-lg);
	}

	.items strong {
		font-weight: 600;
		color: var(--ink);
	}

	.items span {
		font-size: 13px;
		color: var(--ink-3);
	}

	.omitted-lede {
		color: var(--ink-2);
	}

	.omitted-items {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: var(--space-2);
	}

	.omitted-items li {
		display: grid;
		gap: var(--space-1);
	}

	.omitted-items strong {
		color: var(--ink);
	}

	.omitted-items span {
		font-size: 13px;
		color: var(--ink-3);
	}
</style>
