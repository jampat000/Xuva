<script lang="ts">
	import { Info, Play } from 'lucide-svelte';

	let {
		heroPoster = '',
		heroBackdrop = '',
		title = 'Add your first library',
		meta = 'Setup',
		description = 'Choose a Movies or TV folder and Lorivo will build your personal streaming home.',
		progress = 0,
		progressLabel = '',
		playHref = '',
		detailHref = ''
	}: {
		heroPoster?: string;
		heroBackdrop?: string;
		title?: string;
		meta?: string;
		description?: string;
		progress?: number;
		progressLabel?: string;
		playHref?: string;
		detailHref?: string;
	} = $props();

	const clampedProgress = $derived(Math.max(0, Math.min(100, Number(progress) || 0)));
</script>

<section class="relative mx-4 mt-4 h-[540px] min-h-[520px] overflow-hidden rounded-2xl sm:mx-6 lg:mx-8 lg:h-[560px]">
	{#if heroBackdrop}
		<img
			src={heroBackdrop}
			alt=""
			aria-hidden="true"
			class="absolute inset-0 h-full w-full object-cover brightness-75"
		/>
	{:else}
		<div class="absolute inset-0 bg-[radial-gradient(circle_at_70%_25%,rgba(124,92,255,0.22),transparent_34%),linear-gradient(135deg,#111827_0%,#0B1120_62%)]"></div>
	{/if}
	<div class="absolute inset-0 bg-gradient-to-r from-[#0B1120] via-[#0B1120]/70 to-[#0B1120]/30"></div>
	<div class="absolute inset-0 bg-gradient-to-t from-[#0B1120] via-transparent to-transparent"></div>
	<div class="relative flex h-full items-center justify-between px-6 sm:px-10 lg:px-12 xl:px-16">
		<div class="max-w-[560px] lg:max-w-[520px] xl:max-w-[600px]">
			<h1 class="text-5xl font-bold leading-tight text-white [text-shadow:0_4px_28px_rgba(0,0,0,0.72)] sm:text-6xl xl:text-7xl">{title}</h1>
			{#if meta}
				<p class="mt-4 text-base text-white/60">{meta}</p>
			{/if}
			<p class="mt-5 text-base leading-relaxed text-white/70">
				{description}
			</p>
			{#if playHref || detailHref || clampedProgress > 0}
				<div class="mt-8 flex flex-wrap items-center gap-3 sm:gap-4">
					{#if playHref}
						<a href={playHref} class="inline-flex min-h-[60px] items-center gap-3 rounded-xl !bg-[#7C5CFF] px-8 py-4 text-base font-semibold text-white shadow-xl shadow-[#7C5CFF]/40 transition duration-200 hover:-translate-y-0.5 hover:!bg-[#6a4af0] hover:brightness-110 active:translate-y-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#7C5CFF]/70 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120] sm:px-10">
							<Play size={18} class="fill-white text-white" /> Resume
						</a>
					{/if}
					{#if detailHref}
						<a href={detailHref} class="inline-flex min-h-[60px] items-center gap-3 rounded-xl !border !border-white/25 !bg-white/10 px-8 py-4 text-base font-semibold text-white shadow-lg shadow-black/20 backdrop-blur transition duration-200 hover:-translate-y-0.5 hover:!border-white/40 hover:!bg-white/15 active:translate-y-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/35 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120] sm:px-10">
							<Info size={18} /> Details
						</a>
					{/if}
					{#if clampedProgress > 0}
						<div class="ml-4">
							<p class="text-sm text-white/60">{progressLabel || `${clampedProgress}% watched`}</p>
							<div class="mt-2 h-1.5 w-40 overflow-hidden rounded-full bg-white/10">
								<div class="h-full bg-[#7C5CFF]" style="width: {clampedProgress}%"></div>
							</div>
						</div>
					{/if}
				</div>
			{/if}
		</div>
		{#if heroPoster}
			<div class="hidden md:block">
				<img
					src={heroPoster}
					alt={`${title} poster`}
					class="h-[400px] w-[276px] rounded-xl object-cover shadow-2xl shadow-black/60 ring-1 ring-white/10 xl:h-[440px] xl:w-[304px]"
				/>
			</div>
		{:else}
			<div class="hidden md:block">
				<div class="grid h-[400px] w-[276px] place-items-center rounded-xl border border-white/10 bg-[#111827]/70 shadow-2xl shadow-black/60 xl:h-[440px] xl:w-[304px]">
					<div class="h-20 w-20 rounded-full bg-[#7C5CFF]/25"></div>
				</div>
			</div>
		{/if}
	</div>
</section>
