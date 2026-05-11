<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children } = $props<{ children?: Snippet }>();
</script>

<div class="browse-page">
	{@render children?.()}
</div>

<style>
	.browse-page {
		display: grid;
		gap: 18px;
		padding-inline: clamp(2px, 0.8vw, 12px);
		padding-bottom: var(--lorivo-space-8);
		min-width: 0;
	}
</style>
