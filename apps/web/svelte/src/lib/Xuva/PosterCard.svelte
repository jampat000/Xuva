<script lang="ts">
	let {
		img,
		title,
		ep
	}: {
		img?: string;
		title: string;
		ep?: string;
	} = $props();
</script>

<div class="group w-[172px] flex-shrink-0 cursor-pointer transition duration-200 hover:-translate-y-1 sm:w-[204px]">
	<div class="aspect-[2/3] overflow-hidden rounded-md bg-[#1F2937]">
		{#if img}
			<img
				src={img}
				alt={title}
				class="h-full w-full object-cover transition group-hover:brightness-110"
			/>
		{:else}
			<div class="flex h-full w-full items-end bg-[radial-gradient(circle_at_50%_24%,rgba(124,92,255,0.22),transparent_34%),linear-gradient(145deg,#1F2937,#0B1120)] p-4">
				<p class="text-sm font-semibold text-white/80">{title}</p>
			</div>
		{/if}
	</div>
	{#if ep}
		<p class="mt-2 text-sm font-medium text-white/70">{ep}</p>
	{/if}
</div>
