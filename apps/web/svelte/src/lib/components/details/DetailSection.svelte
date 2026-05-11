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
		gap: 13px;
		padding: 17px;
		border: 1px solid rgb(255 246 229 / 10%);
		border-radius: 14px;
		background:
			linear-gradient(180deg, rgb(255 246 229 / 3%), rgb(255 246 229 / 1%)),
			rgb(22 24 21 / 62%);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 3%);
	}

	.detail-section__head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
	}

	.detail-section__head h2 {
		margin: 0;
		font-family: var(--vyrden-font-display);
		font-size: 1.2rem;
		letter-spacing: -0.02em;
	}

	.detail-section__head p {
		margin: 3px 0 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 84%, transparent);
		font-size: 0.86rem;
	}
</style>
