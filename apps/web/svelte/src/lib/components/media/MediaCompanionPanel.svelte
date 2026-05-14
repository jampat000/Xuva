<script lang="ts">
	import type { Snippet } from 'svelte';
	import XuvaSurface from '../ui/XuvaSurface.svelte';

	let {
		title,
		linkLabel,
		linkHref = '#',
		children
	} = $props<{
		title: string;
		linkLabel?: string;
		linkHref?: string;
		children?: Snippet;
	}>();
</script>

<XuvaSurface tone="elevated">
	<section class="media-companion">
		<header>
			<h3>{title}</h3>
			{#if linkLabel}
				<a href={linkHref}>{linkLabel}</a>
			{/if}
		</header>
		<div class="media-companion__body">
			{@render children?.()}
		</div>
	</section>
</XuvaSurface>

<style>
	.media-companion {
		display: grid;
		gap: var(--xuva-space-3);
	}

	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--xuva-space-2);
	}

	h3 {
		margin: 0;
		font-size: 1rem;
		font-weight: 670;
	}

	a {
		color: var(--xuva-color-text-muted);
		font-size: 0.82rem;
		text-decoration: none;
	}

	.media-companion__body {
		display: grid;
		gap: var(--xuva-space-2);
	}
</style>
