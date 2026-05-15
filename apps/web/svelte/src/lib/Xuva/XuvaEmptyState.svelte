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
	<div class="empty-state__content">
		{#if visual}
			<div class="empty-state__visual" aria-hidden="true">
				{@render visual()}
			</div>
		{/if}
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
		padding: 1rem 0;
	}

	.empty-state__content {
		display: flex;
		align-items: flex-start;
		gap: 1rem;
	}

	.empty-state__visual {
		display: grid;
		flex: 0 0 auto;
		width: 2.5rem;
		height: 2.5rem;
		place-items: center;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 0.4rem;
		background: rgb(255 255 255 / 2%);
		color: rgb(255 255 255 / 82%);
		font-size: 0.95rem;
		font-weight: 800;
	}

	.empty-state__visual :global(svg) {
		width: 1.35rem;
		height: 1.35rem;
	}

	.empty-state__copy {
		min-width: 0;
		max-width: 48rem;
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
		font-size: clamp(1.2rem, 2vw, 1.85rem);
		font-weight: 760;
		line-height: 1.15;
		letter-spacing: 0;
	}

	p {
		margin: 0.55rem 0 0;
		color: rgb(255 255 255 / 66%);
		font-size: 0.94rem;
		line-height: 1.5;
	}

	.empty-state__body {
		margin-top: 1rem;
		padding-top: 0.9rem;
	}

	.empty-state__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 0.75rem;
		margin-top: 1rem;
	}

	.empty-state--compact {
		padding: 0.8rem 0;
	}

	.empty-state--compact .empty-state__content {
		align-items: center;
		gap: 0.85rem;
	}

	.empty-state--compact .empty-state__visual {
		width: 2.35rem;
		height: 2.35rem;
		border-radius: 0.4rem;
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
			padding: 0.85rem 0;
		}

		.empty-state__content {
			flex-direction: column;
		}

		.empty-state--compact .empty-state__content {
			flex-direction: row;
		}
	}
</style>
