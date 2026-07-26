<script lang="ts">
	export let steps: { id: number; title: string; required: boolean; description: string; complete: boolean }[] = [];
	export let currentStep = 0;
	export let finished = false;

	$: total = steps.length;
	$: progress = finished ? 100 : total > 1 ? Math.round((currentStep / (total - 1)) * 100) : 0;
</script>

<div
	class="progress-track"
	role="progressbar"
	aria-valuenow={progress}
	aria-valuemin={0}
	aria-valuemax={100}
	aria-label="Onboarding progress"
>
	<div class="progress-fill" style="width: {progress}%"></div>
</div>

<style>
	.progress-track {
		height: 3px;
		background: rgba(122, 155, 118, 0.15);
		border-radius: 2px;
		overflow: hidden;
		margin-bottom: var(--space-8);
	}

	.progress-fill {
		height: 100%;
		background: var(--primary);
		border-radius: 2px;
		transition: width 0.5s var(--ease-out);
	}
</style>
