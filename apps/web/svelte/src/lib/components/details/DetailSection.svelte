<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		subtitle = '',
		children
	} = $props<{
		title: string;
		subtitle?: string;
		children?: Snippet;
	}>();
</script>

<section class="detail-section">
	<header class="detail-section__head">
		<div>
			<h2>{title}</h2>
			{#if subtitle}<p>{subtitle}</p>{/if}
		</div>
	</header>
	{@render children?.()}
</section>

<style>
	.detail-section {
		display: grid;
		gap: 16px;
		padding: clamp(18px, 2vw, 26px);
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 18px;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), rgb(255 255 255 / 1%)),
			rgb(17 24 39 / 66%);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 3%);
		backdrop-filter: blur(12px);
	}

	.detail-section__head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
	}

	.detail-section__head h2 {
		margin: 0;
		font-family: var(--lorivo-font-display);
		font-size: 1.35rem;
		letter-spacing: 0;
	}

	.detail-section__head p {
		margin: 3px 0 0;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 84%, transparent);
		font-size: 0.86rem;
	}
</style>
