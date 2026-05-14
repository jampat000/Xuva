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
		padding: clamp(16px, 1.6vw, 22px) 0;
		border-top: 1px solid rgb(255 255 255 / 10%);
		background: transparent;
	}

	.detail-section__head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
	}

	.detail-section__head h2 {
		margin: 0;
		font-family: var(--xuva-font-display);
		font-size: 1.35rem;
		letter-spacing: 0;
	}

	.detail-section__head p {
		margin: 3px 0 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 84%, transparent);
		font-size: 0.86rem;
	}
</style>
