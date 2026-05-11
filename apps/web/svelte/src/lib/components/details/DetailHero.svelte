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
		border: 1px solid rgb(255 246 229 / 11%);
		border-radius: 16px;
		background: rgb(24 26 23 / 78%);
		box-shadow:
			0 24px 56px rgb(0 0 0 / 32%),
			0 0 0 1px rgb(16 19 24 / 34%);
	}

	.detail-hero__art {
		position: absolute;
		inset: 0;
	}

	.detail-hero__backdrop {
		width: 100%;
		height: 100%;
		object-fit: cover;
		filter: saturate(0.56) contrast(0.93) brightness(0.9);
	}

	.detail-hero__shade {
		position: absolute;
		inset: 0;
		background:
			linear-gradient(90deg, rgb(16 16 14 / 84%) 0%, rgb(20 21 18 / 54%) 34%, rgb(16 17 15 / 60%) 100%),
			linear-gradient(180deg, rgb(14 14 12 / 7%), rgb(14 14 12 / 66%));
	}

	.detail-hero__content {
		position: relative;
		z-index: 1;
		display: grid;
		grid-template-columns: 190px minmax(0, 1fr);
		align-items: end;
		gap: 18px;
		padding: 24px 25px 24px;
	}

	.detail-hero__poster {
		aspect-ratio: 2 / 3;
		overflow: hidden;
		border: 1px solid rgb(255 246 229 / 16%);
		border-radius: 13px;
		background: rgb(24 26 23 / 82%);
		box-shadow: 0 16px 34px rgb(0 0 0 / 38%);
	}

	.detail-hero__poster img {
		display: block;
		width: 100%;
		height: 100%;
		object-fit: cover;
	}

	.detail-hero__copy {
		display: grid;
		gap: 8px;
		align-content: end;
	}

	.detail-hero__copy h1 {
		margin: 0;
		font-family: var(--lorivo-font-display);
		font-size: clamp(1.95rem, 2.45vw, 2.5rem);
		letter-spacing: -0.04em;
		line-height: 1.02;
	}

	.detail-hero__back {
		display: inline-flex;
		align-items: center;
		gap: 6px;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 86%, transparent);
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
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 88%, transparent);
		font-size: 0.95rem;
		font-weight: 580;
	}

	.detail-hero__overview {
		margin: 6px 0 0;
		max-width: 820px;
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		font-size: 0.91rem;
		line-height: 1.46;
	}

	.detail-hero__actions {
		margin-top: 8px;
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
	}

	.hero-progress {
		display: inline-flex;
		align-items: center;
		min-height: 31px;
		padding: 0 11px;
		border: 1px solid rgb(255 246 229 / 12%);
		border-radius: 999px;
		background: rgb(255 246 229 / 5%);
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		font-size: 0.82rem;
		font-weight: 630;
	}

	@media (max-width: 920px) {
		.detail-hero__content {
			grid-template-columns: 132px minmax(0, 1fr);
			padding: 18px 18px 16px;
		}

		.detail-hero__copy h1 {
			font-size: clamp(1.54rem, 6vw, 2.1rem);
		}
	}

	@media (max-width: 680px) {
		.detail-hero__content {
			grid-template-columns: 1fr;
			gap: 14px;
		}

		.detail-hero__poster {
			max-width: 168px;
		}
	}
</style>
