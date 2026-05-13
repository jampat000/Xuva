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

<section class="relative px-4 pt-8 sm:px-6 sm:pt-9 lg:px-8 lg:pt-10">
	<div class="rounded-xl border border-white/10 bg-[#101827]/88 p-5 sm:p-7">
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
