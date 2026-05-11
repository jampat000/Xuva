<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';

	let {
		title,
		linkLabel,
		linkHref = '#',
		trackGap = 12,
		children
	} = $props<{
		title: string;
		linkLabel?: string;
		linkHref?: string;
		trackGap?: number;
		children?: Snippet;
	}>();

	let trackElement = $state<HTMLDivElement | null>(null);
	let canScrollPrev = $state(false);
	let canScrollNext = $state(false);

	function updateScrollState(): void {
		if (!trackElement) return;
		const maxScrollLeft = trackElement.scrollWidth - trackElement.clientWidth;
		canScrollPrev = trackElement.scrollLeft > 4;
		canScrollNext = maxScrollLeft - trackElement.scrollLeft > 4;
	}

	function scrollTrack(direction: 'prev' | 'next'): void {
		if (!trackElement) return;
		const distance = Math.max(220, Math.round(trackElement.clientWidth * 0.78));
		trackElement.scrollBy({
			left: direction === 'next' ? distance : -distance,
			behavior: 'smooth'
		});
	}

	onMount(() => {
		if (!trackElement) return;
		updateScrollState();
		const resizeObserver = new ResizeObserver(() => updateScrollState());
		resizeObserver.observe(trackElement);
		const onScroll = () => updateScrollState();
		trackElement.addEventListener('scroll', onScroll, { passive: true });
		return () => {
			resizeObserver.disconnect();
			trackElement?.removeEventListener('scroll', onScroll);
		};
	});
</script>

<section class="media-row" style={`--row-track-gap:${trackGap}px;`}>
	<header class="media-row__header">
		<h2>{title}</h2>
		{#if linkLabel}
			<a href={linkHref}>{linkLabel}</a>
		{/if}
	</header>
	<div class="media-row__track-wrap" data-show-controls={canScrollPrev || canScrollNext}>
		{#if canScrollPrev}
			<button
				type="button"
				class="media-row__arrow media-row__arrow--prev"
				aria-label="Scroll media row left"
				onclick={() => scrollTrack('prev')}
			>
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="m14.5 6-6 6 6 6" />
				</svg>
			</button>
		{/if}
		<div class="media-row__track" bind:this={trackElement}>
			{@render children?.()}
		</div>
		{#if canScrollNext}
			<button
				type="button"
				class="media-row__arrow media-row__arrow--next"
				aria-label="Scroll media row right"
				onclick={() => scrollTrack('next')}
			>
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path d="m9.5 6 6 6-6 6" />
				</svg>
			</button>
		{/if}
	</div>
</section>

<style>
	.media-row {
		display: grid;
		gap: 14px;
	}

	.media-row__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--lorivo-space-3);
	}

	h2 {
		margin: 0;
		font-family: var(--lorivo-font-display);
		font-size: 1.12rem;
		line-height: 1.03;
		letter-spacing: -0.021em;
		font-weight: 680;
	}

	a {
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 78%, transparent);
		font-size: 0.79rem;
		font-weight: 620;
		text-decoration: none;
	}

	.media-row__track-wrap {
		position: relative;
		padding-inline: 1px;
		border-radius: 14px;
	}

	.media-row__track {
		display: flex;
		gap: var(--row-track-gap);
		overflow-x: auto;
		overflow-y: hidden;
		padding: 1px;
		scroll-snap-type: x proximity;
		flex-wrap: nowrap;
		scrollbar-width: none;
		-ms-overflow-style: none;
		mask-image: linear-gradient(90deg, transparent 0, #000 22px, #000 calc(100% - 22px), transparent 100%);
	}

	.media-row__track::-webkit-scrollbar {
		display: none;
	}

	.media-row__track > :global(*) {
		scroll-snap-align: start;
		flex: 0 0 auto;
	}

	.media-row__track-wrap::before,
	.media-row__track-wrap::after {
		content: '';
		position: absolute;
		top: 0;
		bottom: 0;
		width: 30px;
		z-index: 1;
		pointer-events: none;
		opacity: 0;
		transition: opacity 180ms ease;
	}

	.media-row__track-wrap::before {
		left: 0;
		background: linear-gradient(90deg, rgb(14 17 23 / 72%), transparent);
	}

	.media-row__track-wrap::after {
		right: 0;
		background: linear-gradient(270deg, rgb(14 17 23 / 72%), transparent);
	}

	.media-row__track-wrap[data-show-controls='true']::before,
	.media-row__track-wrap[data-show-controls='true']::after {
		opacity: 1;
	}

	.media-row__arrow {
		position: absolute;
		top: 50%;
		transform: translateY(-50%);
		z-index: 2;
		width: 42px;
		height: 42px;
		border-radius: 999px;
		border: 1px solid rgb(255 255 255 / 18%);
		background:
			linear-gradient(180deg, rgb(31 41 55 / 90%), rgb(11 17 32 / 82%)),
			rgb(11 17 32 / 86%);
		color: color-mix(in srgb, var(--lorivo-color-text) 96%, transparent);
		line-height: 1;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		box-shadow:
			0 12px 28px rgb(0 0 0 / 34%),
			0 0 18px rgb(124 92 255 / 14%);
		backdrop-filter: blur(8px);
		opacity: 0.86;
		transition:
			transform 180ms ease,
			opacity 180ms ease,
			border-color 180ms ease,
			background 180ms ease,
			box-shadow 180ms ease;
	}

	.media-row__arrow svg {
		width: 24px;
		height: 24px;
		fill: none;
		stroke: currentColor;
		stroke-width: 2.35;
		stroke-linecap: round;
		stroke-linejoin: round;
	}

	.media-row__arrow--prev {
		left: 6px;
	}

	.media-row__arrow--next {
		right: 6px;
	}

	.media-row__arrow:hover,
	.media-row__arrow:focus-visible {
		transform: translateY(-50%) scale(1.04);
		opacity: 1;
		border-color: color-mix(in srgb, var(--lorivo-color-accent-teal) 54%, white 10%);
		background:
			linear-gradient(180deg, rgb(124 92 255 / 82%), rgb(83 60 202 / 82%)),
			rgb(11 17 32 / 88%);
		box-shadow:
			0 14px 32px rgb(0 0 0 / 38%),
			0 0 22px rgb(124 92 255 / 28%);
	}

	@media (max-width: 620px) {
		.media-row__arrow {
			width: 38px;
			height: 38px;
		}
	}
</style>
