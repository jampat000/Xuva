<script lang="ts">
	import XuvaWordmark from '$lib/components/brand/XuvaWordmark.svelte';

	let { wordmark = 'Xuva', showBlurb = false, centered = false } = $props<{
		compact?: boolean;
		wordmark?: string;
		showBlurb?: boolean;
		centered?: boolean;
	}>();
</script>

<div class="v-brand" data-brand={wordmark} data-centered={centered ? 'true' : 'false'}>
	<XuvaWordmark {wordmark} />
	{#if showBlurb}
		<p class="v-brand__blurb">Your media, beautifully played.</p>
	{/if}
</div>

<style>
	.v-brand {
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		justify-content: flex-start;
		width: 100%;
		min-height: var(--xuva-brand-min-height, 28px);
		--xuva-brand-stop-1: #a78bfa;
		--xuva-brand-stop-2: #7c3aed;
		--xuva-brand-stop-3: #db2777;
		--xuva-brand-wordmark-color: #ffffff;
		--xuva-brand-wordmark-shadow: 0 1px 0 rgb(0 0 0 / 20%);
		padding: 0;
		gap: 8px;
		flex: 0 0 auto;
	}

	.v-brand__blurb {
		margin: 0;
		width: 100%;
		max-width: none;
		align-self: flex-start;
		display: block;
		text-align: left;
		font-size: 0.74rem;
		line-height: 1.35;
		letter-spacing: 0.01em;
		color: color-mix(in srgb, var(--xuva-color-text) 84%, transparent);
	}

	.v-brand[data-centered='true'] {
		align-items: center;
	}

	.v-brand[data-centered='true'] .v-brand__blurb {
		text-align: center;
	}
</style>
