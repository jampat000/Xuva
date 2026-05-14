<script lang="ts">
	import type { Snippet } from 'svelte';
	import XuvaSurface from './XuvaSurface.svelte';

	let {
		title,
		subtitle = '',
		actions,
		children
	} = $props<{
		title: string;
		subtitle?: string;
		actions?: Snippet;
		children?: Snippet;
	}>();
</script>

<XuvaSurface tone="elevated">
	<section class="v-panel">
		<header class="v-panel__header">
			<div>
				<h2>{title}</h2>
				{#if subtitle}
					<p>{subtitle}</p>
				{/if}
			</div>
			{#if actions}
				<div class="v-panel__actions">{@render actions()}</div>
			{/if}
		</header>
		<div class="v-panel__body">
			{@render children?.()}
		</div>
	</section>
</XuvaSurface>

<style>
	.v-panel {
		display: grid;
		gap: var(--xuva-space-3);
	}

	.v-panel__header {
		display: flex;
		justify-content: space-between;
		gap: var(--xuva-space-4);
		padding-bottom: 0.55rem;
		border-bottom: 1px solid var(--xuva-color-border-soft);
	}

	h2 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 690;
		letter-spacing: -0.012em;
	}

	p {
		margin: var(--xuva-space-1) 0 0;
		color: var(--xuva-color-text-muted);
		font-size: 0.9rem;
		line-height: 1.4;
	}

	.v-panel__body {
		padding-top: 0.2rem;
	}
</style>
