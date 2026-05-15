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
		width: var(--xuva-brand-width, 116px);
		min-height: var(--xuva-brand-min-height, 28px);
		--xuva-brand-stop-1: #a78bfa;
		--xuva-brand-stop-2: #7c3aed;
		--xuva-brand-stop-3: #db2777;
		--xuva-brand-wordmark-color: #ffffff;
		--xuva-brand-wordmark-shadow: 0 1px 0 rgb(0 0 0 / 20%);
		padding: 0;
		flex: 0 0 auto;
	}
</style>
