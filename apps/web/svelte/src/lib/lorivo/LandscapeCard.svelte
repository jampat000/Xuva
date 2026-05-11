<script lang="ts">
	import { Play } from 'lucide-svelte';

	type ContinueWatchingItem = {
		title: string;
		sub: string;
		progress: number;
		img: string;
	};

	let { item }: { item: ContinueWatchingItem } = $props();
</script>

<div class="group w-[280px] flex-shrink-0 cursor-pointer transition duration-200 hover:-translate-y-1 sm:w-[320px]">
	<div class="relative aspect-video overflow-hidden rounded-lg bg-[#1F2937]">
		<img src={item.img} alt={item.title} class="h-full w-full object-cover brightness-90 transition group-hover:brightness-100" />
		<div class="absolute inset-0 grid place-items-center">
			<div class="grid h-12 w-12 place-items-center rounded-full bg-black/55 shadow-lg shadow-black/30 backdrop-blur transition duration-200 group-hover:bg-[#7C5CFF] group-hover:shadow-[#7C5CFF]/25">
				<Play size={18} class="fill-white text-white" />
			</div>
		</div>
		<div class="absolute inset-x-0 bottom-0 h-2 bg-white/14">
			<div class="h-full bg-[#7C5CFF] shadow-[0_0_12px_rgba(124,92,255,0.55)]" style="width: {item.progress}%"></div>
		</div>
	</div>
	<div class="mt-3">
		<p class="truncate text-base font-medium text-white">{item.title}</p>
		<p class="text-sm text-white/50">{item.sub}</p>
	</div>
</div>
