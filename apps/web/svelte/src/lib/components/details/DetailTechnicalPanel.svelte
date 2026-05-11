<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		children
	} = $props<{
		title: string;
		children?: Snippet;
	}>();
</script>

<article class="track-panel">
	<h3>{title}</h3>
	{@render children?.()}
</article>

<style>
	.track-panel {
		padding: 12px;
		border: 1px solid rgb(255 246 229 / 11%);
		border-radius: 12px;
		background: rgb(255 246 229 / 3%);
	}

	.track-panel h3 {
		margin: 0 0 10px;
		font-size: 0.95rem;
		font-weight: 680;
	}

	.track-panel :global(ul) {
		margin: 0;
		padding: 0;
		list-style: none;
		display: grid;
		gap: 8px;
	}

	.track-panel :global(li) {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 8px;
	}

	.track-panel :global(li strong) {
		font-size: 0.81rem;
		font-weight: 620;
	}

	.track-panel :global(li span) {
		display: inline-flex;
		align-items: center;
		min-height: 22px;
		padding: 0 8px;
		border: 1px solid rgb(255 246 229 / 14%);
		border-radius: 999px;
		background: rgb(255 246 229 / 4%);
		font-size: 0.72rem;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 88%, transparent);
	}
</style>
