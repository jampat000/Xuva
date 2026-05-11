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
		gap: 24px;
		padding-inline: clamp(0px, 1vw, 16px);
		padding-bottom: var(--vyrden-space-8);
		min-width: 0;
	}
</style>
