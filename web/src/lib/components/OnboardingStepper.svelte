<script lang="ts">
	export let steps: {
		id: number;
		title: string;
		required: boolean;
		description: string;
		complete: boolean;
	}[] = [];
	export let currentStep = 0;
	export let finished = false;
</script>

<aside class="stepper" aria-label="Onboarding progress">
	<p class="eyebrow">Step-by-step help</p>
	<h2>Your setup path</h2>
	<ol>
		{#each steps as step}
			<li class:active={currentStep === step.id} class:complete={step.complete || finished}>
				<div class="row">
					<span class="number">{step.id + 1}</span>
					<div>
						<p class="title">{step.title}</p>
						<p class="meta">{step.required ? 'Required' : 'Optional or review'}</p>
					</div>
				</div>
				<p class="description">{step.description}</p>
			</li>
		{/each}
	</ol>
</aside>

<style>
	.stepper {
		position: sticky;
		top: 1.5rem;
		padding: 1.25rem;
		border-radius: 1.25rem;
		background: #ffffff;
		box-shadow: 0 14px 35px rgba(24, 34, 47, 0.08);
	}

	.eyebrow {
		margin: 0 0 0.5rem;
		font-size: 0.8rem;
		font-weight: 700;
		letter-spacing: 0.08em;
		text-transform: uppercase;
		color: #5d6b82;
	}

	h2 {
		margin: 0 0 1rem;
		font-size: 1.3rem;
	}

	ol {
		margin: 0;
		padding: 0;
		list-style: none;
		display: grid;
		gap: 0.85rem;
	}

	li {
		padding: 0.9rem;
		border-radius: 1rem;
		border: 1px solid #d9e1ee;
		background: #f8faff;
	}

	li.active {
		border-color: #2a6df4;
		background: #eef4ff;
	}

	li.complete .number {
		background: #1f9d5c;
	}

	.row {
		display: flex;
		gap: 0.75rem;
		align-items: center;
	}

	.number {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		border-radius: 999px;
		background: #2a6df4;
		color: white;
		font-weight: 700;
	}

	.title,
	.meta,
	.description {
		margin: 0;
	}

	.title {
		font-weight: 700;
	}

	.meta {
		font-size: 0.85rem;
		color: #5d6b82;
	}

	.description {
		margin-top: 0.65rem;
		font-size: 0.94rem;
		line-height: 1.45;
		color: #334155;
	}
</style>
