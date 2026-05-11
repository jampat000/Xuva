<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		subtitle = '',
		children
	}: {
		title: string;
		subtitle?: string;
		children: Snippet;
	} = $props();
</script>

<section class="relative px-4 pt-9 sm:px-6 sm:pt-10 lg:px-8 lg:pt-11">
	<header class="mb-4">
		<h2 class="text-lg font-semibold text-white">{title}</h2>
		{#if subtitle}
			<p class="mt-1 text-sm text-white/50">{subtitle}</p>
		{/if}
	</header>
	<div class="rounded-2xl border border-white/10 bg-white/[0.04] p-4 shadow-lg shadow-black/20 backdrop-blur sm:p-5 lg:p-6">
		{@render children()}
	</div>
</section>
