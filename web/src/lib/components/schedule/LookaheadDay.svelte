<script lang="ts">
	import { formatReadyTime, windows, type LookaheadDay } from '$lib/api/lookahead';

	export let day: LookaheadDay;
</script>

<article class="day-card">
	<header>
		<p class="eyebrow">{day.label}</p>
		<h2>{day.date}</h2>
	</header>

	{#if day.small_task_only_reason}
		<p class="notice">{day.small_task_only_reason}</p>
	{/if}

	<div class="windows">
		{#each windows as window}
			<section class="window">
				<div class="window-header">
					<h3>{window}</h3>
					<p>{day.window_budgets_minutes[window] ?? 0} min open</p>
				</div>

				{#if day.blocked_windows[window]?.length}
					<div class="blocked">
						<strong>Calendar limits this window</strong>
						<ul>
							{#each day.blocked_windows[window] as reason}
								<li>{reason}</li>
							{/each}
						</ul>
					</div>
				{/if}

				{#if day.windows[window]?.length}
					<ul class="items">
						{#each day.windows[window] as item}
							<li>
								<div>
									<strong>{item.name}</strong>
									<span>{item.duration_minutes} min{item.ready_at ? ` · ready ${formatReadyTime(item.ready_at)}` : ''}</span>
								</div>
							</li>
						{/each}
					</ul>
				{:else}
					<p class="empty">No tasks scheduled for this window.</p>
				{/if}
			</section>
		{/each}
	</div>

	{#if day.omitted_items.length}
		<section class="omitted">
			<h3>Tasks not shown in the plan</h3>
			<p>Rahat is being conservative because of your calendar or time budget.</p>
			<ul>
				{#each day.omitted_items as item}
					<li>
						<strong>{item.name}</strong>
						<span>{item.window}: {item.reason}</span>
					</li>
				{/each}
			</ul>
		</section>
	{/if}
</article>

<style>
	.day-card {
		display: grid;
		gap: 1rem;
		padding: 1rem;
		border-radius: 1.25rem;
		background: white;
		box-shadow: 0 12px 30px rgba(20, 32, 44, 0.08);
	}

	header h2,
	header p,
	.window-header h3,
	.window-header p,
	.empty,
	.notice,
	.omitted p {
		margin: 0;
	}

	.eyebrow {
		font-size: 0.8rem;
		font-weight: 800;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #4f46e5;
	}

	.notice,
	.blocked,
	.omitted {
		padding: 0.85rem;
		border-radius: 1rem;
		background: #fff7ed;
		border: 1px solid #fed7aa;
	}

	.windows {
		display: grid;
		gap: 0.9rem;
	}

	.window {
		display: grid;
		gap: 0.75rem;
		padding: 0.9rem;
		border: 1px solid #dbe4ee;
		border-radius: 1rem;
		background: #fbfdff;
	}

	.window-header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 1rem;
	}

	.window-header h3 {
		text-transform: capitalize;
	}

	.window-header p,
	.empty,
	.items span,
	.omitted span {
		color: #5d6b82;
	}

	ul {
		margin: 0;
		padding-left: 1.1rem;
	}

	.items {
		padding: 0;
		list-style: none;
		display: grid;
		gap: 0.65rem;
	}

	.items li,
	.omitted li {
		line-height: 1.45;
	}

	.items span,
	.omitted span {
		display: block;
	}
</style>
