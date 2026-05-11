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
	<header class="mb-4 flex flex-col gap-2 sm:flex-row sm:items-end sm:justify-between">
		<div>
			<h2 class="text-lg font-semibold text-white">{title}</h2>
			{#if subtitle}
				<p class="mt-1 text-sm text-white/50">{subtitle}</p>
			{/if}
		</div>
	</header>
	<div class="grid grid-cols-[repeat(auto-fill,minmax(172px,1fr))] gap-x-2.5 gap-y-8 sm:grid-cols-[repeat(auto-fill,minmax(204px,1fr))] sm:gap-x-3 lg:gap-y-10">
		{@render children()}
	</div>
</section>
