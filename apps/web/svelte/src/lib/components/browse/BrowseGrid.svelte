<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		minColumnWidth = 176,
		maxColumnWidth = 186
	} = $props<{
		children?: Snippet;
		minColumnWidth?: number;
		maxColumnWidth?: number;
	}>();
</script>

<div
	class="browse-grid"
	style={`--browse-grid-min:${Math.max(120, Math.round(minColumnWidth))}px;--browse-grid-max:${Math.max(130, Math.round(maxColumnWidth))}px;`}
	role="list"
>
	{@render children?.()}
</div>

<style>
	.browse-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(var(--browse-grid-min), var(--browse-grid-max)));
		justify-content: start;
		gap: 28px 18px;
		min-width: 0;
	}

	.browse-grid :global(.poster-card) {
		width: 100%;
	}

	@media (max-width: 760px) {
		.browse-grid {
			grid-template-columns: repeat(2, minmax(152px, 1fr));
			gap: 16px 12px;
		}
	}
</style>
