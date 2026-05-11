<script lang="ts">
	import type { Snippet } from 'svelte';
	import VyrdenSurface from './VyrdenSurface.svelte';

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

<VyrdenSurface tone="elevated">
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
</VyrdenSurface>

<style>
	.v-panel {
		display: grid;
		gap: var(--vyrden-space-4);
	}

	.v-panel__header {
		display: flex;
		justify-content: space-between;
		gap: var(--vyrden-space-4);
	}

	h2 {
		margin: 0;
		font-size: 1.1rem;
		font-weight: 690;
		letter-spacing: -0.012em;
	}

	p {
		margin: var(--vyrden-space-1) 0 0;
		color: var(--vyrden-color-text-muted);
		font-size: 0.9rem;
		line-height: 1.4;
	}
</style>
