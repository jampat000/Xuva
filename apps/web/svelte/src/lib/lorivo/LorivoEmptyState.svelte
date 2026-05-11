<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description,
		eyebrow = '',
		compact = false,
		visual,
		primaryAction,
		secondaryAction,
		children
	}: {
		title: string;
		description: string;
		eyebrow?: string;
		compact?: boolean;
		visual?: Snippet;
		primaryAction?: Snippet;
		secondaryAction?: Snippet;
		children?: Snippet;
	} = $props();
</script>

<section class={compact ? 'empty-state empty-state--compact' : 'empty-state'}>
	<div class="empty-state__glow" aria-hidden="true"></div>
	<div class="empty-state__content">
		<div class="empty-state__visual" aria-hidden="true">
			{#if visual}
				{@render visual()}
			{:else}
				<span>L</span>
			{/if}
		</div>
		<div class="empty-state__copy">
			{#if eyebrow}
				<p class="empty-state__eyebrow">{eyebrow}</p>
			{/if}
			<h2>{title}</h2>
			<p>{description}</p>
		</div>
	</div>

	{#if children}
		<div class="empty-state__body">
			{@render children()}
		</div>
	{/if}

	{#if primaryAction || secondaryAction}
		<div class="empty-state__actions">
			{@render primaryAction?.()}
			{@render secondaryAction?.()}
		</div>
	{/if}
</section>

<style>
	.empty-state {
		position: relative;
		overflow: hidden;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 1rem;
		background:
			linear-gradient(135deg, rgb(255 255 255 / 7%), rgb(255 255 255 / 2%)),
			rgb(17 24 39 / 72%);
		box-shadow: 0 24px 60px rgb(0 0 0 / 24%);
		backdrop-filter: blur(16px);
		padding: 1.5rem;
	}

	.empty-state__glow {
		position: absolute;
		inset: auto -12% -42% auto;
		width: min(22rem, 70vw);
		aspect-ratio: 1;
		border-radius: 999px;
		background: radial-gradient(circle, rgb(124 92 255 / 22%), transparent 68%);
		pointer-events: none;
	}

	.empty-state__content {
		position: relative;
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.empty-state__visual {
		display: grid;
		flex: 0 0 auto;
		width: 3.25rem;
		height: 3.25rem;
		place-items: center;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 1rem;
		background: rgb(11 17 32 / 72%);
		color: rgb(255 255 255 / 82%);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 8%);
		font-size: 1rem;
		font-weight: 800;
	}

	.empty-state__visual :global(svg) {
		width: 1.35rem;
		height: 1.35rem;
	}

	.empty-state__copy {
		min-width: 0;
		max-width: 44rem;
	}

	.empty-state__eyebrow {
		margin: 0 0 0.45rem;
		color: rgb(255 255 255 / 45%);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	h2 {
		margin: 0;
		color: white;
		font-size: clamp(1.35rem, 2.4vw, 2.15rem);
		font-weight: 760;
		line-height: 1.1;
		letter-spacing: 0;
	}

	p {
		margin: 0.7rem 0 0;
		color: rgb(255 255 255 / 66%);
		font-size: 0.98rem;
		line-height: 1.55;
	}

	.empty-state__body {
		position: relative;
		margin-top: 1.25rem;
	}

	.empty-state__actions {
		position: relative;
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-top: 1.4rem;
	}

	.empty-state--compact {
		padding: 1rem;
		box-shadow: 0 14px 36px rgb(0 0 0 / 16%);
	}

	.empty-state--compact .empty-state__content {
		align-items: center;
		gap: 0.85rem;
	}

	.empty-state--compact .empty-state__visual {
		width: 2.35rem;
		height: 2.35rem;
		border-radius: 0.75rem;
		font-size: 0.8rem;
	}

	.empty-state--compact h2 {
		font-size: 1rem;
		line-height: 1.25;
	}

	.empty-state--compact p {
		margin-top: 0.3rem;
		font-size: 0.85rem;
		line-height: 1.45;
	}

	.empty-state--compact .empty-state__actions {
		margin-top: 1rem;
	}

	@media (max-width: 520px) {
		.empty-state {
			padding: 1.1rem;
		}

		.empty-state__content {
			flex-direction: column;
		}

		.empty-state--compact .empty-state__content {
			flex-direction: row;
		}
	}
</style>
