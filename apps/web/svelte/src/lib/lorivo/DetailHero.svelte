<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		meta,
		overview,
		backHref,
		backLabel,
		backdropUrl,
		posterUrl,
		progress = 0,
		progressLabel = '',
		actions
	}: {
		title: string;
		meta: string;
		overview: string;
		backHref: string;
		backLabel: string;
		backdropUrl: string;
		posterUrl: string;
		progress?: number;
		progressLabel?: string;
		actions?: Snippet;
	} = $props();

	const boundedProgress = $derived(Math.max(0, Math.min(100, Math.round(Number(progress || 0)))));
</script>

<section class="relative mx-4 mt-4 min-h-[520px] overflow-hidden rounded-2xl sm:mx-6 lg:mx-8 lg:min-h-[560px]">
	{#if backdropUrl}
		<img src={backdropUrl} alt="" aria-hidden="true" class="absolute inset-0 h-full w-full object-cover brightness-75" />
	{/if}
	<div class="absolute inset-0 bg-gradient-to-r from-[#0B1120] via-[#0B1120]/75 to-[#0B1120]/35"></div>
	<div class="absolute inset-0 bg-gradient-to-t from-[#0B1120] via-transparent to-transparent"></div>
	<div class="relative flex min-h-[520px] items-center justify-between px-6 py-10 sm:px-10 lg:min-h-[560px] lg:px-12 xl:px-16">
		<div class="max-w-[560px] lg:max-w-[520px] xl:max-w-[600px]">
			<a href={backHref} class="mb-6 inline-flex text-sm font-medium text-white/60 transition hover:text-white">
				{backLabel}
			</a>
			<h1 class="text-5xl font-bold leading-tight text-white [text-shadow:0_4px_28px_rgba(0,0,0,0.72)] sm:text-6xl xl:text-7xl">{title}</h1>
			{#if meta}
				<p class="mt-4 text-base text-white/60">{meta}</p>
			{/if}
			<p class="mt-5 text-base leading-relaxed text-white/70">{overview}</p>
			<div class="mt-8 flex flex-wrap items-center gap-3 sm:gap-4">
				{#if actions}
					{@render actions()}
				{/if}
				{#if progressLabel}
					<div class="ml-4">
						<p class="text-sm text-white/60">{progressLabel}</p>
						<div class="mt-2 h-1.5 w-40 overflow-hidden rounded-full bg-white/10">
							<div class="h-full bg-[#7C5CFF]" style="width: {boundedProgress}%"></div>
						</div>
					</div>
				{/if}
			</div>
		</div>
		{#if posterUrl}
			<div class="hidden md:block">
				<img
					src={posterUrl}
					alt={`${title} poster`}
					class="h-[400px] w-[276px] rounded-xl object-cover shadow-2xl shadow-black/60 ring-1 ring-white/10 xl:h-[440px] xl:w-[304px]"
				/>
			</div>
		{/if}
	</div>
</section>
