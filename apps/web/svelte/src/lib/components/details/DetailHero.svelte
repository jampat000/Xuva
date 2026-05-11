<script lang="ts">
	import type { Snippet } from 'svelte';
	import ArtworkFallback from '$lib/components/media/ArtworkFallback.svelte';

	let {
		title,
		meta = '',
		overview = '',
		backHref,
		backLabel,
		backdropUrl = '',
		posterUrl = '',
		actions,
		progressLabel = ''
	} = $props<{
		title: string;
		meta?: string;
		overview?: string;
		backHref: string;
		backLabel: string;
		backdropUrl?: string;
		posterUrl?: string;
		actions?: Snippet;
		progressLabel?: string;
	}>();

	let backdropFailed = $state(false);
	let posterFailed = $state(false);

	$effect(() => {
		backdropFailed = !backdropUrl;
	});

	$effect(() => {
		posterFailed = !posterUrl;
	});
</script>

<section class="detail-hero">
	<div class="detail-hero__art">
		{#if !backdropFailed}
			<img
				class="detail-hero__backdrop"
				src={backdropUrl}
				alt={`${title} backdrop`}
				loading="eager"
				onerror={() => (backdropFailed = true)}
			/>
		{:else}
			<ArtworkFallback variant="hero" {title} meta={meta} showCopy={false} />
		{/if}
		<div class="detail-hero__shade"></div>
	</div>
	<div class="detail-hero__content">
		<div class="detail-hero__poster">
			{#if !posterFailed}
				<img
					src={posterUrl}
					alt={`${title} poster`}
					loading="lazy"
					onerror={() => (posterFailed = true)}
				/>
			{:else}
				<ArtworkFallback variant="poster" {title} meta={meta} />
			{/if}
		</div>
		<div class="detail-hero__copy">
			<a class="detail-hero__back" href={backHref}>{backLabel}</a>
			<h1>{title}</h1>
			{#if meta}<p class="detail-hero__meta">{meta}</p>{/if}
			{#if overview}<p class="detail-hero__overview">{overview}</p>{/if}
			{#if actions || progressLabel}
				<div class="detail-hero__actions">
					{#if actions}
						{@render actions()}
					{/if}
					{#if progressLabel}
						<span class="hero-progress">{progressLabel}</span>
					{/if}
				</div>
			{/if}
		</div>
	</div>
</section>

<style>
	.detail-hero {
		position: relative;
		overflow: hidden;
		min-height: clamp(430px, 48vw, 560px);
		border: 0;
		border-radius: 24px;
		background: rgb(11 17 32 / 88%);
		box-shadow:
			0 28px 70px rgb(0 0 0 / 42%),
			0 0 0 1px rgb(255 255 255 / 7%) inset;
	}

	.detail-hero__art {
		position: absolute;
		inset: 0;
	}

	.detail-hero__backdrop {
		width: 100%;
		height: 100%;
		object-fit: cover;
		filter: saturate(0.82) contrast(1.04) brightness(0.9);
	}

	.detail-hero__shade {
		position: absolute;
		inset: 0;
		background:
			linear-gradient(90deg, rgb(11 17 32 / 92%) 0%, rgb(11 17 32 / 62%) 40%, rgb(11 17 32 / 74%) 100%),
			linear-gradient(180deg, rgb(11 17 32 / 10%), rgb(11 17 32 / 72%)),
			radial-gradient(circle at 78% 22%, rgb(124 92 255 / 18%), transparent 38%);
	}

	.detail-hero__content {
		position: relative;
		z-index: 1;
		display: grid;
		grid-template-columns: minmax(0, 1fr) 280px;
		align-items: center;
		gap: clamp(24px, 5vw, 64px);
		min-height: inherit;
		padding: clamp(32px, 4.4vw, 72px);
	}

	.detail-hero__poster {
		grid-column: 2;
		grid-row: 1;
		aspect-ratio: 2 / 3;
		overflow: hidden;
		border: 1px solid rgb(255 255 255 / 12%);
		border-radius: 16px;
		background: rgb(17 24 39 / 82%);
		box-shadow: 0 24px 56px rgb(0 0 0 / 48%);
	}

	.detail-hero__poster img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.detail-hero__copy {
		grid-column: 1;
		grid-row: 1;
		display: grid;
		gap: 12px;
		align-content: center;
		max-width: 720px;
	}

	.detail-hero__copy h1 {
		margin: 0;
		font-family: var(--lorivo-font-display);
		font-size: clamp(3rem, 5.6vw, 5.4rem);
		letter-spacing: 0;
		line-height: 1.02;
		text-shadow: 0 14px 34px rgb(0 0 0 / 58%);
	}

	.detail-hero__back {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 90%, transparent);
		text-decoration: none;
		font-size: 0.82rem;
		font-weight: 620;
		padding: 0;
	}

	.detail-hero__back::before {
		content: '‹';
		font-size: 1rem;
		line-height: 1;
	}

	.detail-hero__meta {
		margin: 0;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 90%, transparent);
		font-size: 1rem;
		font-weight: 580;
	}

	.detail-hero__overview {
		margin: 8px 0 0;
		max-width: 680px;
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		font-size: 1rem;
		line-height: 1.58;
	}

	.detail-hero__actions {
		margin-top: 16px;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
	}

	.hero-progress {
		display: inline-flex;
		align-items: center;
		min-height: 34px;
		padding: 0 13px;
		border: 1px solid rgb(255 255 255 / 14%);
		border-radius: 999px;
		background: rgb(255 255 255 / 7%);
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		font-size: 0.82rem;
		font-weight: 630;
	}

	@media (max-width: 920px) {
		.detail-hero__content {
			grid-template-columns: minmax(0, 1fr) 210px;
			padding: 28px;
		}

		.detail-hero__copy h1 {
			font-size: clamp(2.4rem, 6vw, 4rem);
		}
	}

	@media (max-width: 680px) {
		.detail-hero__content {
			grid-template-columns: 1fr;
			gap: 14px;
			min-height: 0;
		}

		.detail-hero__poster {
			grid-column: 1;
			grid-row: 2;
			max-width: 168px;
		}

		.detail-hero__copy {
			grid-column: 1;
			grid-row: 1;
		}
	}
</style>
