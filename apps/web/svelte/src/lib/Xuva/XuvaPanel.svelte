<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		subtitle = '',
		children
	}: {
		title: string;
		subtitle?: string;
		children?: Snippet;
	} = $props();
</script>

<section class="v-panel-shell">
	<div class="v-panel">
		<header class="v-panel__header">
			<h2 class="text-lg font-semibold text-white">{title}</h2>
		{#if subtitle}
				<p class="mt-2 max-w-2xl text-sm leading-relaxed text-white/60">{subtitle}</p>
			{/if}
		</header>
		{#if children}
			<div class="v-panel__body">
				{@render children()}
			</div>
		{/if}
	</div>
</section>

<style>
	.v-panel-shell {
		position: relative;
		padding: 1.15rem 1rem 0;
	}

	.v-panel {
		padding: 0.95rem 0 0.35rem;
	}

	.v-panel__header {
		display: grid;
		gap: 0;
		padding: 0 0.15rem 0.7rem;
	}

	.v-panel__body {
		padding-top: 0.75rem;
	}

	@media (min-width: 640px) {
		.v-panel-shell {
			padding-left: 1.5rem;
			padding-right: 1.5rem;
			padding-top: 1.4rem;
		}
	}

	@media (min-width: 1024px) {
		.v-panel-shell {
			padding-left: 2rem;
			padding-right: 2rem;
			padding-top: 1.5rem;
		}
	}
</style>
