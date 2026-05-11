<script lang="ts">
	let {
		img,
		title,
		ep
	}: {
		img: string;
		title: string;
		ep?: string;
	} = $props();
</script>

<div class="group w-[172px] flex-shrink-0 cursor-pointer transition duration-200 hover:-translate-y-1 sm:w-[204px]">
	<div class="aspect-[2/3] overflow-hidden rounded-md bg-[#1F2937]">
		<img
			src={img}
			alt={title}
			class="h-full w-full object-cover transition group-hover:brightness-110"
		/>
	</div>
	{#if ep}
		<p class="mt-2 text-sm font-medium text-white/70">{ep}</p>
	{/if}
</div>
