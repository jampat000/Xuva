<script lang="ts">
	let { label } = $props<{ label: string }>();
</script>

<span class="summary-chip" role="listitem">{label}</span>

<style>
	.summary-chip {
		display: inline-flex;
		align-items: center;
		min-height: 31px;
		padding: 0 11px;
		border: 1px solid rgb(255 246 229 / 11%);
		border-radius: 999px;
		background: rgb(255 246 229 / 4%);
		color: color-mix(in srgb, var(--vyrden-color-text) 90%, transparent);
		font-size: 0.8rem;
		font-weight: 650;
		letter-spacing: 0.008em;
	}
</style>
