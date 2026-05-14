<script lang="ts">
	import XuvaWordmark from '$lib/components/brand/XuvaWordmark.svelte';

	let { wordmark = 'Xuva' } = $props<{ compact?: boolean; wordmark?: string }>();
</script>

<div class="v-brand" data-brand={wordmark}>
	<XuvaWordmark {wordmark} />
</div>

<style>
	.v-brand {
		display: flex;
		align-items: center;
		justify-content: flex-start;
		width: var(--xuva-brand-width, 118px);
		min-height: var(--xuva-brand-min-height, 34px);
		padding: 0;
		flex: 0 0 auto;
	}
</style>
