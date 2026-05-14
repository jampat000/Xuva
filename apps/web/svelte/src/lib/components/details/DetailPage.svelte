<script lang="ts">
	import type { Snippet } from 'svelte';

	let { children } = $props<{ children?: Snippet }>();
</script>

<div class="detail-page">
	{@render children?.()}
</div>

<style>
	.detail-page {
		display: grid;
		gap: 18px;
		padding-bottom: var(--xuva-space-8);
		min-width: 0;
	}
</style>
