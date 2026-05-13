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
		padding: 1.6rem 1rem 0;
	}

	.v-panel {
		border-top: 1px solid rgb(255 255 255 / 9%);
		border-bottom: 1px solid rgb(255 255 255 / 7%);
		background: rgb(16 24 39 / 34%);
		padding: 1.15rem 0;
	}

	.v-panel__header {
		display: grid;
		gap: 0;
		padding: 0 0.15rem 0.95rem;
		border-bottom: 1px solid rgb(255 255 255 / 7%);
	}

	.v-panel__body {
		padding-top: 1rem;
	}

	@media (min-width: 640px) {
		.v-panel-shell {
			padding-left: 1.5rem;
			padding-right: 1.5rem;
			padding-top: 1.9rem;
		}
	}

	@media (min-width: 1024px) {
		.v-panel-shell {
			padding-left: 2rem;
			padding-right: 2rem;
			padding-top: 2rem;
		}
	}
</style>
