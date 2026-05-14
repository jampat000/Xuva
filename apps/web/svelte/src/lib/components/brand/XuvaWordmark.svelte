<script lang="ts">
	import XuvaLogo from './XuvaLogo.svelte';

	let { wordmark = 'Xuva', showMark = true } = $props<{ wordmark?: string; showMark?: boolean }>();
	const suffix = $derived.by(() => (wordmark || 'Xuva').trim().slice(1).toLowerCase() || 'uva');
</script>

<div class="xuva-wordmark">
	{#if showMark}
		<XuvaLogo size={22} />
	{/if}
	<div class="xuva-wordmark__text">
		<svg class="xuva-wordmark__x" viewBox="0 0 72 100" aria-hidden="true">
			<defs>
				<linearGradient id="xuva-word-gradient" x1="0" y1="50" x2="72" y2="50" gradientUnits="userSpaceOnUse">
					<stop offset="0%" stop-color="var(--xuva-brand-stop-1, #A78BFA)" />
					<stop offset="55%" stop-color="var(--xuva-brand-stop-2, #7C3AED)" />
					<stop offset="100%" stop-color="var(--xuva-brand-stop-3, #DB2777)" />
				</linearGradient>
			</defs>
			<path d="M 4 96 L 20 96 L 68 4 L 52 4 Z" fill="url(#xuva-word-gradient)" />
			<path d="M 4 4 L 18 4 L 68 96 L 54 96 Z" fill="url(#xuva-word-gradient)" opacity="0.88" />
		</svg>
		<span class="xuva-wordmark__suffix">{suffix}</span>
	</div>
</div>

<style>
	.xuva-wordmark {
		display: inline-flex;
		align-items: center;
		gap: var(--xuva-brand-gap, 10px);
		width: var(--xuva-brand-width, 118px);
		min-height: var(--xuva-brand-min-height, 34px);
		flex: 0 0 auto;
	}

	.xuva-wordmark__text {
		display: inline-flex;
		align-items: baseline;
		gap: 2px;
	}

	.xuva-wordmark__x {
		width: 13px;
		height: 18px;
		display: block;
		flex: 0 0 auto;
	}

	.xuva-wordmark__suffix {
		color: var(--xuva-brand-wordmark-color, #ffffff);
		font-family: var(--xuva-font-sans);
		font-size: var(--xuva-brand-wordmark-size, 16px);
		line-height: 1;
		font-weight: 700;
		letter-spacing: 0.01em;
		text-transform: lowercase;
		text-shadow: var(--xuva-brand-wordmark-shadow, 0 1px 0 rgb(0 0 0 / 20%));
		white-space: nowrap;
	}
</style>

