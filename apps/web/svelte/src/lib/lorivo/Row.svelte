<script lang="ts">
	import type { Snippet } from 'svelte';
	import { ChevronRight } from 'lucide-svelte';

	let {
		title,
		children
	}: {
		title: string;
		children: Snippet;
	} = $props();
</script>

<section class="relative px-4 pt-9 sm:px-6 sm:pt-10 lg:px-8 lg:pt-11">
	<h2 class="mb-3.5 text-lg font-semibold text-white">{title}</h2>
	<div class="relative">
		<div class="flex gap-2.5 overflow-x-auto pb-3 sm:gap-3 [scrollbar-width:none] [&::-webkit-scrollbar]:hidden">
			{@render children()}
		</div>
		<button
			aria-label="Scroll right"
			class="absolute -right-3 top-[42%] hidden h-12 w-12 -translate-y-1/2 items-center justify-center rounded-full !border !border-white/10 !bg-black/55 text-white/85 opacity-90 shadow-lg shadow-black/35 backdrop-blur transition duration-200 hover:scale-105 hover:!border-[#7C5CFF]/45 hover:!bg-[#7C5CFF]/70 hover:text-white hover:opacity-100 hover:shadow-[#7C5CFF]/25 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#7C5CFF]/70 md:flex"
		>
			<ChevronRight size={24} strokeWidth={2.6} />
		</button>
	</div>
</section>
