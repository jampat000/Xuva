<script lang="ts">
	import ArtworkFallback from './ArtworkFallback.svelte';

	let {
		title,
		meta = '',
		imageUrl = '',
		href
	} = $props<{
		title: string;
		meta?: string;
		imageUrl?: string;
		href?: string;
	}>();

	let showFallback = $state(false);
	$effect(() => {
		showFallback = !imageUrl;
	});
</script>

{#if href}
	<a class="poster-card" href={href}>
		<div class="poster-card__art">
			{#if !showFallback}
				<img src={imageUrl} alt={`${title} artwork`} loading="lazy" onerror={() => (showFallback = true)} />
			{:else}
				<ArtworkFallback variant="poster" {title} {meta} showCopy={false} />
			{/if}
			<div class="poster-card__overlay"></div>
			<div class="poster-card__copy">
				<h3>{title}</h3>
				{#if meta}<p>{meta}</p>{/if}
			</div>
		</div>
	</a>
{:else}
	<button class="poster-card" type="button" aria-label={`Open ${title}`}>
		<div class="poster-card__art">
			{#if !showFallback}
				<img src={imageUrl} alt="" loading="lazy" onerror={() => (showFallback = true)} />
			{:else}
				<ArtworkFallback variant="poster" {title} {meta} showCopy={false} />
			{/if}
			<div class="poster-card__overlay"></div>
			<div class="poster-card__copy">
				<h3>{title}</h3>
				{#if meta}<p>{meta}</p>{/if}
			</div>
		</div>
	</button>
{/if}

<style>
	.poster-card {
		display: block;
		width: 190px;
		padding: 0;
		border: 0;
		background: none;
		color: var(--lorivo-color-text);
		text-align: left;
		line-height: 1;
		text-decoration: none;
	}

	.poster-card__art {
		position: relative;
		aspect-ratio: 2 / 3;
		overflow: hidden;
		border: 1px solid rgb(255 255 255 / 12%);
		border-radius: 14px;
		background: rgb(17 24 39 / 84%);
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 6%),
			0 18px 34px rgb(0 0 0 / 28%),
			0 0 0 1px rgb(16 19 24 / 34%);
		transition:
			transform 180ms ease,
			border-color 180ms ease,
			box-shadow 180ms ease,
			filter 180ms ease;
	}

	img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 260ms ease;
	}

	.poster-card__overlay {
		position: absolute;
		inset: 0;
		background:
			linear-gradient(180deg, rgb(8 10 13 / 2%) 0%, rgb(10 13 17 / 12%) 42%, rgb(10 13 17 / 68%) 100%),
			radial-gradient(circle at 74% 20%, rgb(124 92 255 / 14%), transparent 44%);
	}

	.poster-card__copy {
		position: absolute;
		left: var(--lorivo-space-2);
		right: var(--lorivo-space-2);
		bottom: var(--lorivo-space-2);
		z-index: 1;
	}

	h3 {
		margin: 0;
		font-size: 0.98rem;
		font-weight: 660;
		line-height: 1.15;
		letter-spacing: -0.01em;
		text-shadow: 0 2px 10px rgb(0 0 0 / 48%);
	}

	p {
		margin: 6px 0 0;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 92%, transparent);
		font-size: 0.78rem;
		text-shadow: 0 1px 8px rgb(0 0 0 / 42%);
	}

	@media (hover: hover) and (pointer: fine) {
		.poster-card:hover .poster-card__art,
		.poster-card:focus-visible .poster-card__art {
			transform: translateY(-3px);
			border-color: color-mix(in srgb, var(--lorivo-color-accent-teal) 36%, rgb(255 246 229 / 22%));
			box-shadow:
				inset 0 1px 0 rgb(255 255 255 / 9%),
				0 24px 42px rgb(0 0 0 / 36%),
				0 0 26px rgb(124 92 255 / 18%),
				0 0 0 1px rgb(18 22 28 / 44%);
		}

		.poster-card:hover img,
		.poster-card:focus-visible img {
			transform: scale(1.035);
		}
	}
</style>
