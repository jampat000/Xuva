<script lang="ts">
	import type { Snippet } from 'svelte';
	import LorivoSurface from './LorivoSurface.svelte';

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

<LorivoSurface tone="elevated">
	<section class="l-panel">
		<header class="l-panel__header">
			<div>
				<h2>{title}</h2>
				{#if subtitle}
					<p>{subtitle}</p>
				{/if}
			</div>
			{#if actions}
				<div class="l-panel__actions">{@render actions()}</div>
			{/if}
		</header>
		<div class="l-panel__body">
			{@render children?.()}
		</div>
	</section>
</LorivoSurface>

<style>
	.l-panel {
		display: grid;
		gap: var(--lorivo-space-4);
	}

	.l-panel__header {
		display: flex;
		justify-content: space-between;
		gap: var(--lorivo-space-4);
	}

	h2 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 690;
		letter-spacing: -0.012em;
	}

	p {
		margin: var(--lorivo-space-1) 0 0;
		color: var(--lorivo-color-text-muted);
		font-size: 0.9rem;
		line-height: 1.4;
	}
</style>
