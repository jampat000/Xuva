<script lang="ts">
	import type { Snippet } from 'svelte';
	import ArtworkFallback from './ArtworkFallback.svelte';

	let {
		badge = 'Featured',
		title,
		meta = '',
		description = '',
		imageUrl = '',
		featureImageUrl = '',
		showControls = true,
		onDismiss,
		actions
	} = $props<{
		badge?: string;
		title: string;
		meta?: string;
		description?: string;
		imageUrl?: string;
		featureImageUrl?: string;
		showControls?: boolean;
		onDismiss?: (() => void) | undefined;
		actions?: Snippet;
	}>();

	let showFallback = $state(false);
	$effect(() => {
		showFallback = !imageUrl;
	});

	function dismissHero(): void {
		onDismiss?.();
	}
</script>

<section class="hero-surface">
	<div class="hero-surface__art">
		{#if !showFallback}
			<img src={imageUrl} alt={`${title} hero`} loading="lazy" onerror={() => (showFallback = true)} />
		{:else}
			<ArtworkFallback variant="hero" {title} {meta} showCopy={false} />
		{/if}
		<div class="hero-surface__shade"></div>
	</div>
	{#if showControls}
		<button class="hero-surface__close" type="button" aria-label="Close feature" onclick={dismissHero}>
			&times;
		</button>
		<button class="hero-surface__nav" type="button" aria-label="Next feature">
			<svg viewBox="0 0 24 24" aria-hidden="true">
				<path
					d="m9 6 6 6-6 6"
					fill="none"
					stroke="currentColor"
					stroke-width="1.7"
					stroke-linecap="round"
					stroke-linejoin="round"
				/>
			</svg>
		</button>
		<div class="hero-surface__dots" aria-hidden="true">
			<span class="active"></span><span></span><span></span><span></span>
		</div>
	{/if}
	<div class="hero-surface__copy">
		<span class="hero-surface__badge">{badge}</span>
		<h1>{title}</h1>
		{#if meta}<p class="hero-surface__meta">{meta}</p>{/if}
		{#if description}<p class="hero-surface__description">{description}</p>{/if}
		{#if actions}
			<div class="hero-surface__actions">
				{@render actions()}
			</div>
		{/if}
	</div>
	<div class="hero-surface__feature" aria-hidden="true">
		{#if featureImageUrl}
			<img src={featureImageUrl} alt="" loading="lazy" />
		{:else if !showFallback}
			<img src={imageUrl} alt="" loading="lazy" />
		{:else}
			<ArtworkFallback variant="poster" {title} {meta} showCopy={false} />
		{/if}
		<div class="hero-surface__feature-shade"></div>
	</div>
</section>

<style>
	.hero-surface {
		position: relative;
		min-height: 418px;
		border: 1px solid rgb(255 246 229 / 10%);
		border-radius: 20px;
		overflow: hidden;
		background: #181915;
		box-shadow:
			0 28px 62px rgb(0 0 0 / 34%),
			0 0 0 1px rgb(16 19 24 / 34%);
	}

	.hero-surface__art {
		position: absolute;
		inset: 0;
	}

	img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		filter: saturate(0.44) contrast(0.93) brightness(0.89);
	}

	.hero-surface__shade {
		position: absolute;
		inset: 0;
		background:
			linear-gradient(
				90deg,
				rgb(12 13 11 / 87%) 0%,
				rgb(16 16 14 / 64%) 24%,
				rgb(20 20 18 / 24%) 56%,
				rgb(15 16 14 / 60%) 100%
			),
			radial-gradient(circle at 76% 18%, rgb(88 201 176 / 13%), transparent 42%),
			linear-gradient(180deg, rgb(16 15 12 / 5%), rgb(16 15 12 / 42%));
	}

	.hero-surface__copy {
		position: relative;
		z-index: 1;
		display: flex;
		flex-direction: column;
		align-items: flex-start;
		max-width: 610px;
		padding: 34px 34px 22px;
	}

	.hero-surface__feature {
		position: absolute;
		right: 26px;
		bottom: 22px;
		width: min(272px, 30%);
		aspect-ratio: 2 / 3;
		overflow: hidden;
		border: 1px solid rgb(255 246 229 / 16%);
		border-radius: 14px;
		background: rgb(22 24 21 / 84%);
		box-shadow: 0 18px 34px rgb(0 0 0 / 38%);
		z-index: 1;
	}

	.hero-surface__feature img {
		width: 100%;
		height: 100%;
		object-fit: cover;
		filter: saturate(0.64) contrast(0.94) brightness(0.86);
	}

	.hero-surface__feature-shade {
		position: absolute;
		inset: 0;
		background:
			linear-gradient(180deg, rgb(10 12 15 / 4%) 0%, rgb(10 12 15 / 14%) 48%, rgb(10 12 15 / 62%) 100%),
			radial-gradient(circle at 70% 18%, rgb(88 201 176 / 10%), transparent 42%);
	}

	.hero-surface__badge {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		min-height: 24px;
		padding: 0 11px;
		border: 1px solid rgb(39 211 189 / 20%);
		border-radius: 999px;
		background: rgb(88 201 176 / 12%);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 2%);
		color: #7ad5bf;
		font-size: 0.71rem;
		font-weight: 800;
		letter-spacing: 0.07em;
		line-height: 1;
		text-transform: uppercase;
	}

	h1 {
		margin: 10px 0 6px;
		font-family: var(--xuva-font-display);
		font-size: clamp(2.35rem, 2.7vw, 3.05rem);
		line-height: 0.98;
		letter-spacing: -0.05em;
		font-weight: 780;
		text-shadow: 0 14px 28px rgb(0 0 0 / 32%);
	}

	.hero-surface__meta {
		margin: 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 84%, transparent);
		font-size: 0.98rem;
		font-weight: 560;
		letter-spacing: 0.005em;
	}

	.hero-surface__description {
		margin: 10px 0 0;
		max-width: 520px;
		color: color-mix(in srgb, var(--xuva-color-text) 86%, transparent);
		font-size: 0.95rem;
		line-height: 1.5;
	}

	.hero-surface__actions {
		margin-top: 12px;
		display: flex;
		align-items: center;
		gap: 11px;
		flex-wrap: wrap;
	}

	.hero-surface__close,
	.hero-surface__nav {
		position: absolute;
		z-index: 2;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		color: var(--xuva-color-text-muted);
	}

	.hero-surface__close {
		right: 14px;
		top: 12px;
		width: 18px;
		height: 18px;
		font-size: 1.25rem;
		line-height: 1;
		cursor: pointer;
	}

	.hero-surface__nav {
		right: 22px;
		bottom: 112px;
		width: 20px;
		height: 20px;
		opacity: 0.92;
		cursor: pointer;
	}

	.hero-surface__nav svg {
		width: 18px;
		height: 18px;
	}

	.hero-surface__dots {
		position: absolute;
		bottom: 14px;
		left: 50%;
		display: inline-flex;
		gap: 10px;
		transform: translateX(-50%);
		z-index: 2;
	}

	.hero-surface__dots span {
		width: 7px;
		height: 7px;
		border-radius: 50%;
		background: color-mix(in srgb, var(--xuva-color-text-muted) 48%, transparent);
	}

	.hero-surface__dots span.active {
		background: var(--xuva-color-accent-teal);
	}

	@media (max-width: 920px) {
		.hero-surface {
			min-height: 350px;
		}

		.hero-surface__copy {
			max-width: 74%;
			padding: 22px 20px 16px;
		}

		.hero-surface__feature {
			right: 16px;
			bottom: 14px;
			width: min(170px, 28%);
		}
	}

	@media (max-width: 700px) {
		.hero-surface {
			min-height: 252px;
		}

		.hero-surface__copy {
			max-width: 100%;
			padding: 16px 14px 10px;
		}

		.hero-surface__feature {
			display: none;
		}

		h1 {
			margin-top: 8px;
			font-size: clamp(1.72rem, 7.3vw, 2.2rem);
		}

		.hero-surface__description {
			margin-top: 6px;
			font-size: 0.84rem;
			line-height: 1.36;
			display: -webkit-box;
			-webkit-line-clamp: 3;
			-webkit-box-orient: vertical;
			overflow: hidden;
		}

		.hero-surface__actions {
			margin-top: 8px;
			gap: 7px;
		}

		.hero-surface__dots {
			bottom: 9px;
		}
	}
</style>
