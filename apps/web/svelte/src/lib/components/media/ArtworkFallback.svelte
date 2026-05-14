<script lang="ts">
	type ArtworkVariant = 'poster' | 'landscape' | 'hero';

	let {
		variant = 'poster',
		title,
		meta = '',
		showCopy = true
	} = $props<{
		variant?: ArtworkVariant;
		title: string;
		meta?: string;
		showCopy?: boolean;
	}>();

	const seed = $derived.by(() => hashString(`${title}-${meta}-${variant}`));
	const tiltA = $derived.by(() => 8 + (seed % 19));
	const tiltB = $derived.by(() => 118 + (seed % 37));
	const orbX = $derived.by(() => 58 + (seed % 28));
	const orbY = $derived.by(() => 18 + (seed % 22));
	const orbR = $derived.by(() => 22 + (seed % 16));
	const tone = $derived.by(() => ['#7c5cff', '#8eb6ff', '#d79056', '#b794de', '#5fa0e8'][seed % 5]);

	function hashString(value: string): number {
		let hash = 0;
		for (let index = 0; index < value.length; index += 1) {
			hash = (hash << 5) - hash + value.charCodeAt(index);
			hash |= 0;
		}
		return Math.abs(hash);
	}
</script>

<div
	class="artwork-fallback"
	data-variant={variant}
	style={`--fallback-tilt-a:${tiltA}deg;--fallback-tilt-b:${tiltB}deg;--fallback-orb-x:${orbX}%;--fallback-orb-y:${orbY}%;--fallback-orb-r:${orbR}%;--fallback-tone:${tone};`}
>
	<div class="artwork-fallback__ambient"></div>
	<div class="artwork-fallback__line"></div>
	{#if showCopy}
		<div class="artwork-fallback__copy">
			<span class="artwork-fallback__mark" aria-hidden="true"></span>
			<strong>{title}</strong>
			{#if meta}
				<p>{meta}</p>
			{/if}
		</div>
	{:else}
		<span class="artwork-fallback__mark artwork-fallback__mark--solo" aria-hidden="true"></span>
	{/if}
</div>

<style>
	.artwork-fallback {
		position: relative;
		display: flex;
		align-items: flex-end;
		overflow: hidden;
		width: 100%;
		height: 100%;
		padding: var(--xuva-space-3);
		color: var(--xuva-color-text);
		background:
			linear-gradient(160deg, rgb(255 255 255 / 8%), rgb(255 255 255 / 0%) 58%),
			linear-gradient(180deg, #1f2937, #111827 57%, #0b1120 100%);
	}

	.artwork-fallback__ambient {
		position: absolute;
		inset: 0;
		background:
			radial-gradient(
				circle at var(--fallback-orb-x) var(--fallback-orb-y),
				color-mix(in srgb, var(--fallback-tone) 46%, white 8%) 0%,
				rgb(124 92 255 / 0%) var(--fallback-orb-r)
			),
			radial-gradient(circle at 14% 78%, rgb(124 92 255 / 16%) 0%, rgb(124 92 255 / 0%) 42%),
			linear-gradient(var(--fallback-tilt-a), rgb(255 255 255 / 4%), transparent 50%),
			repeating-linear-gradient(
				var(--fallback-tilt-b),
				rgb(255 255 255 / 0%) 0 18px,
				rgb(255 255 255 / 3%) 18px 20px
			);
		opacity: 0.52;
	}

	.artwork-fallback__line {
		position: absolute;
		left: var(--xuva-space-3);
		right: var(--xuva-space-3);
		top: var(--xuva-space-3);
		height: 2px;
		background: linear-gradient(
			90deg,
			color-mix(in srgb, var(--fallback-tone) 84%, white 16%),
			transparent 72%
		);
		opacity: 0.48;
	}

	.artwork-fallback__copy {
		position: relative;
		z-index: 1;
		display: grid;
		gap: var(--xuva-space-1);
	}

	.artwork-fallback__mark {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 1.5rem;
		height: 1.5rem;
		border-radius: 7px;
		border: 1px solid rgb(124 92 255 / 42%);
		background:
			linear-gradient(
				145deg,
				color-mix(in srgb, var(--fallback-tone) 52%, rgb(255 255 255 / 20%)),
				color-mix(in srgb, var(--fallback-tone) 30%, rgb(18 21 27 / 20%))
			),
			linear-gradient(180deg, rgb(255 255 255 / 6%), rgb(255 255 255 / 0%));
		box-shadow:
			inset 0 0 0 1px rgb(255 255 255 / 4%),
			0 6px 14px rgb(0 0 0 / 24%);
	}

	.artwork-fallback__mark::before {
		content: '';
		display: block;
		width: 0.58rem;
		height: 0.58rem;
		border-radius: 3px;
		border: 1px solid rgb(255 246 229 / 48%);
		transform: rotate(45deg);
	}

	.artwork-fallback__mark--solo {
		position: absolute;
		left: var(--xuva-space-3);
		bottom: var(--xuva-space-3);
		z-index: 1;
	}

	strong {
		font-size: 0.92rem;
		font-weight: 660;
		letter-spacing: -0.008em;
	}

	p {
		margin: 0;
		color: var(--xuva-color-text-muted);
		font-size: 0.74rem;
	}

	.artwork-fallback[data-variant='hero'] strong {
		font-size: clamp(1.55rem, 2.4vw, 2.25rem);
	}

	.artwork-fallback[data-variant='hero'] p {
		font-size: 0.9rem;
	}

	.artwork-fallback[data-variant='landscape'] strong {
		font-size: 1rem;
	}
</style>
