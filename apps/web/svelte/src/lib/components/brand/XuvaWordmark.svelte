<script lang="ts">
	import XuvaLogo from './XuvaLogo.svelte';

	let { wordmark = 'Xuva', showMark = true } = $props<{ wordmark?: string; showMark?: boolean }>();
	const brandText = $derived.by(() => (wordmark || 'Xuva').trim() || 'Xuva');
	const leading = $derived.by(() => brandText.slice(0, 1) || 'X');
	const trailing = $derived.by(() => brandText.slice(1) || 'uva');
</script>

<div class="xuva-wordmark">
	{#if showMark}
		<XuvaLogo size={28} />
	{/if}
	<div class="xuva-wordmark__text">
		<span class="xuva-wordmark__leading">{leading}</span>
		<span class="xuva-wordmark__trailing">{trailing}</span>
	</div>
</div>

<style>
	.xuva-wordmark {
		display: inline-flex;
		align-items: center;
		gap: var(--xuva-brand-gap, 8px);
		width: var(--xuva-brand-width, 116px);
		min-height: var(--xuva-brand-min-height, 28px);
		flex: 0 0 auto;
	}

	.xuva-wordmark__text {
		display: inline-flex;
		align-items: baseline;
		gap: 0.24px;
		line-height: 1;
	}

	.xuva-wordmark__leading,
	.xuva-wordmark__trailing {
		font-family:
			'Space Grotesk',
			'Inter',
			ui-sans-serif,
			system-ui,
			-apple-system,
			sans-serif;
		font-size: var(--xuva-brand-wordmark-size, 26px);
		line-height: 1;
		font-weight: 500;
		letter-spacing: -0.035em;
		white-space: nowrap;
	}

	.xuva-wordmark__leading {
		background:
			linear-gradient(
				90deg,
				var(--xuva-brand-stop-1, #a78bfa) 0%,
				var(--xuva-brand-stop-2, #7c3aed) 55%,
				var(--xuva-brand-stop-3, #db2777) 100%
			);
		-webkit-background-clip: text;
		background-clip: text;
		color: transparent;
	}

	.xuva-wordmark__trailing {
		color: var(--xuva-brand-wordmark-color, #ffffff);
		text-shadow: var(--xuva-brand-wordmark-shadow, 0 1px 0 rgb(0 0 0 / 20%));
	}
</style>

