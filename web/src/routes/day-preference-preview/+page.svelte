<script lang="ts">
	import TaskEditor from '$lib/components/TaskEditor.svelte';
	import { emptyTaskDraft, type TaskDraft } from '$lib/api/onboarding';

	let cardsDraft: TaskDraft = emptyTaskDraft();
	let segmentedDraft: TaskDraft = emptyTaskDraft();
</script>

<svelte:head>
	<title>Day preference preview · Rahat</title>
</svelte:head>

<div class="page">
	<header class="intro">
		<p class="eyebrow">Story 037 preview</p>
		<h1>Choose the days that work.</h1>
		<p>Try both versions below. Weekend selection automatically changes the cadence to weekly planning capped at two times.</p>
	</header>

	<section class="preview-card">
		<div class="section-heading">
			<p class="eyebrow">Onboarding</p>
			<h2>Guided radio cards</h2>
			<p>Uses the roomier version with plain-language hints.</p>
		</div>
		<TaskEditor draft={cardsDraft} dayPickerVariant="cards" submitLabel="Preview only" on:save={(event) => (cardsDraft = event.detail.draft)} on:cancel={() => undefined} />
	</section>

	<section class="preview-card">
		<div class="section-heading">
			<p class="eyebrow">Task management</p>
			<h2>Compact segmented control</h2>
			<p>Uses the compact version for returning users.</p>
		</div>
		<TaskEditor draft={segmentedDraft} dayPickerVariant="segmented" submitLabel="Preview only" on:save={(event) => (segmentedDraft = event.detail.draft)} on:cancel={() => undefined} />
	</section>
</div>

<style>
	.page {
		max-width: 980px;
		margin: 0 auto;
		padding: var(--space-8) var(--space-5) var(--space-12);
		display: grid;
		gap: var(--space-8);
	}

	.intro {
		max-width: 680px;
		display: grid;
		gap: var(--space-3);
	}

	.eyebrow {
		margin: 0;
		font-size: 12px;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: var(--primary-2);
	}

	h1, h2, p { margin: 0; }
	h1 { font-family: var(--font-display); font-size: clamp(32px, 5vw, 52px); line-height: 1.05; }
	h2 { font-family: var(--font-display); font-size: 28px; }
	.intro > p:last-child, .section-heading > p:last-child { color: var(--ink-2); line-height: 1.6; }

	.preview-card {
		min-width: 0;
		padding: var(--space-6);
		background: var(--paper);
		border: 1px solid var(--line-soft);
		border-radius: var(--radius-xl);
		display: grid;
		gap: var(--space-6);
		box-shadow: var(--shadow-sm);
	}

	.section-heading { display: grid; gap: var(--space-2); }

	@media (max-width: 540px) {
		.page { padding: var(--space-6) var(--space-4) var(--space-8); gap: var(--space-6); }
		.preview-card { padding: var(--space-4); }
	}
</style>
