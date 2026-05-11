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

<section class="relative px-4 pt-9 sm:px-6 sm:pt-10 lg:px-8 lg:pt-11">
	<div class="rounded-2xl border border-white/10 bg-white/[0.04] p-6 shadow-lg shadow-black/20 backdrop-blur sm:p-8">
		<h2 class="text-lg font-semibold text-white">{title}</h2>
		{#if subtitle}
			<p class="mt-2 max-w-2xl text-sm leading-relaxed text-white/60">{subtitle}</p>
		{/if}
		{#if children}
			<div class="mt-5">
				{@render children()}
			</div>
		{/if}
	</div>
</section>
