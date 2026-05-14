<script lang="ts">
	let {
		img,
		title,
		meta = '',
		href
	}: {
		img: string;
		title: string;
		meta?: string;
		href: string;
	} = $props();
</script>

<a {href} class="group block min-w-0 cursor-pointer transition duration-200 hover:-translate-y-1">
	<div class="aspect-[2/3] overflow-hidden rounded-md bg-[#1F2937] shadow-lg shadow-black/20">
		<img
			src={img}
			alt={title}
			class="h-full w-full object-cover transition group-hover:brightness-110"
		/>
	</div>
	<p class="mt-3 truncate text-base font-medium text-white">{title}</p>
	{#if meta}
		<p class="mt-1 truncate text-sm text-white/50">{meta}</p>
	{/if}
</a>
