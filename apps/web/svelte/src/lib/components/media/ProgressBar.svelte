<script lang="ts">
	let { value = 0, label = 'Progress' } = $props<{ value?: number; label?: string }>();
	const normalized = $derived(Math.max(0, Math.min(100, value)));
</script>

<div class="progress-wrap" role="progressbar" aria-label={label} aria-valuemin={0} aria-valuemax={100} aria-valuenow={Math.round(normalized)}>
	<span style={`width:${normalized}%`}></span>
</div>

<style>
	.progress-wrap {
		height: 0.42rem;
		width: 100%;
		border-radius: var(--vyrden-radius-pill);
		background: var(--vyrden-track);
		overflow: hidden;
	}

	.progress-wrap span {
		display: block;
		height: 100%;
		background: var(--vyrden-progress);
		box-shadow: 0 0 12px rgb(124 92 255 / 42%);
	}
</style>
