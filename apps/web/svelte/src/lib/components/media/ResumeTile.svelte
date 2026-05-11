<script lang="ts">
	import ArtworkFallback from './ArtworkFallback.svelte';
	import ProgressBar from './ProgressBar.svelte';

	let {
		title,
		subtitle = '',
		meta = '',
		progress = 0,
		imageUrl = '',
		href
	} = $props<{
		title: string;
		subtitle?: string;
		meta?: string;
		progress?: number;
		imageUrl?: string;
		href?: string;
	}>();

	let showFallback = $state(false);
	$effect(() => {
		showFallback = !imageUrl;
	});

	const secondaryMeta = $derived.by(() => {
		const subtitleValue = subtitle.trim().toLowerCase();
		const metaValue = meta.trim();
		if (!metaValue) return '';
		if (!subtitleValue) return metaValue;
		const normalizedMeta = metaValue.toLowerCase();
		if (
			normalizedMeta === subtitleValue ||
			normalizedMeta.startsWith(`${subtitleValue} -`) ||
			normalizedMeta.startsWith(`${subtitleValue} ·`)
		) {
			return '';
		}
		return metaValue;
	});
	const subline = $derived.by(() => asText(subtitle) || asText(secondaryMeta));
	const progressLine = $derived.by(() => {
		const metaValue = asText(meta);
		if (!metaValue) return progress > 0 ? `Resume from ${Math.round(progress)}%` : '';
		return /resume from/i.test(metaValue) ? metaValue : progress > 0 ? `Resume from ${Math.round(progress)}%` : '';
	});

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}
</script>

{#if href}
	<a class="resume-tile" href={href}>
		<div class="resume-tile__art">
			{#if !showFallback}
				<img src={imageUrl} alt={`${title} artwork`} loading="lazy" onerror={() => (showFallback = true)} />
			{:else}
				<ArtworkFallback variant="landscape" title={title} meta={subtitle || meta} showCopy={false} />
			{/if}
			<div class="resume-tile__shade"></div>
			<div class="resume-tile__copy">
				<h3>{title}</h3>
				{#if subline}<p>{subline}</p>{/if}
				{#if progressLine}<p class="resume-tile__progress-copy">{progressLine}</p>{/if}
			</div>
			<div class="resume-tile__progress">
				<ProgressBar value={progress} label={`Resume progress for ${title}`} />
			</div>
		</div>
	</a>
{:else}
	<button class="resume-tile" type="button" aria-label={`Resume ${title}`}>
		<div class="resume-tile__art">
			{#if !showFallback}
				<img src={imageUrl} alt="" loading="lazy" onerror={() => (showFallback = true)} />
			{:else}
				<ArtworkFallback variant="landscape" title={title} meta={subtitle || meta} showCopy={false} />
			{/if}
			<div class="resume-tile__shade"></div>
			<div class="resume-tile__copy">
				<h3>{title}</h3>
				{#if subline}<p>{subline}</p>{/if}
				{#if progressLine}<p class="resume-tile__progress-copy">{progressLine}</p>{/if}
			</div>
			<div class="resume-tile__progress">
				<ProgressBar value={progress} label={`Resume progress for ${title}`} />
			</div>
		</div>
	</button>
{/if}

<style>
	.resume-tile {
		width: 316px;
		display: block;
		padding: 0;
		border: 0;
		background: none;
		text-align: left;
		line-height: 1;
		color: var(--lorivo-color-text);
		text-decoration: none;
	}

	.resume-tile__art {
		position: relative;
		aspect-ratio: 16 / 9;
		overflow: hidden;
		border: 1px solid rgb(255 255 255 / 12%);
		border-radius: 14px;
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 8%),
			0 16px 32px rgb(0 0 0 / 28%),
			0 0 0 1px rgb(16 19 24 / 44%);
		transition:
			transform 180ms ease,
			border-color 180ms ease,
			box-shadow 180ms ease;
	}

	img {
		position: absolute;
		inset: 0;
		width: 100%;
		height: 100%;
		object-fit: cover;
		transition: transform 240ms ease;
	}

	.resume-tile__shade {
		position: absolute;
		inset: 0;
		background:
			linear-gradient(180deg, rgb(8 10 13 / 2%) 0%, rgb(10 13 17 / 12%) 44%, rgb(10 13 17 / 62%) 100%),
			linear-gradient(90deg, rgb(8 10 13 / 7%) 0%, rgb(8 10 13 / 2%) 28%, rgb(8 10 13 / 14%) 100%);
	}

	.resume-tile__copy {
		position: absolute;
		left: var(--lorivo-space-3);
		right: var(--lorivo-space-3);
		bottom: calc(var(--lorivo-space-2) + 0.9rem);
		z-index: 1;
	}

	h3 {
		margin: 0;
		font-size: 0.98rem;
		font-weight: 660;
		letter-spacing: -0.01em;
		display: -webkit-box;
		-webkit-line-clamp: 1;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	p {
		margin: 0;
		font-size: 0.75rem;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 92%, transparent);
		margin-top: 4px;
		display: -webkit-box;
		-webkit-line-clamp: 1;
		-webkit-box-orient: vertical;
		overflow: hidden;
	}

	.resume-tile__progress-copy {
		margin: 4px 0 0;
		font-size: 0.72rem;
		font-weight: 560;
		color: color-mix(in srgb, var(--lorivo-color-text) 82%, transparent);
	}

	.resume-tile__progress {
		position: absolute;
		left: var(--lorivo-space-3);
		right: var(--lorivo-space-3);
		bottom: var(--lorivo-space-2);
		z-index: 1;
	}

	@media (hover: hover) and (pointer: fine) {
		.resume-tile:hover .resume-tile__art,
		.resume-tile:focus-visible .resume-tile__art {
			transform: translateY(-3px);
			border-color: color-mix(in srgb, var(--lorivo-color-accent-teal) 44%, rgb(255 255 255 / 20%));
			box-shadow:
				inset 0 1px 0 rgb(255 255 255 / 10%),
				0 22px 40px rgb(0 0 0 / 34%),
				0 0 0 1px rgb(18 22 28 / 44%);
		}

		.resume-tile:hover img,
		.resume-tile:focus-visible img {
			transform: scale(1.03);
		}
	}
</style>
