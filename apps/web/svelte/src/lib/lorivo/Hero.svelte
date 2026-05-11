<script lang="ts">
	import { Info, Play } from 'lucide-svelte';

	let {
		heroPoster,
		heroBackdrop
	}: {
		heroPoster: string;
		heroBackdrop: string;
	} = $props();
</script>

<section class="relative mx-4 mt-4 h-[540px] min-h-[520px] overflow-hidden rounded-2xl sm:mx-6 lg:mx-8 lg:h-[560px]">
	<!-- Poster used as the hero backdrop, scaled & faded -->
	<img
		src={heroBackdrop}
		alt=""
		aria-hidden="true"
		class="absolute inset-0 h-full w-full object-cover brightness-75"
	/>
	<div class="absolute inset-0 bg-gradient-to-r from-[#0B1120] via-[#0B1120]/70 to-[#0B1120]/30"></div>
	<div class="absolute inset-0 bg-gradient-to-t from-[#0B1120] via-transparent to-transparent"></div>
	<div class="relative flex h-full items-center justify-between px-6 sm:px-10 lg:px-12 xl:px-16">
		<div class="max-w-[560px] lg:max-w-[520px] xl:max-w-[600px]">
			<h1 class="text-5xl font-bold leading-tight text-white [text-shadow:0_4px_28px_rgba(0,0,0,0.72)] sm:text-6xl xl:text-7xl">Dune: Part Two</h1>
			<p class="mt-4 text-base text-white/60">2024 &nbsp;•&nbsp; Movie &nbsp;•&nbsp; 2h 46m</p>
			<p class="mt-5 text-base leading-relaxed text-white/70">
				Paul Atreides unites with the Fremen and wages war against the conspirators
				who destroyed his family, torn between love and the prophecy that could
				ignite a galaxy-spanning holy war.
			</p>
			<div class="mt-8 flex flex-wrap items-center gap-3 sm:gap-4">
				<button class="inline-flex min-h-[60px] items-center gap-3 rounded-xl !bg-[#7C5CFF] px-8 py-4 text-base font-semibold text-white shadow-xl shadow-[#7C5CFF]/40 transition duration-200 hover:-translate-y-0.5 hover:!bg-[#6a4af0] hover:brightness-110 active:translate-y-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#7C5CFF]/70 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120] sm:px-10">
					<Play size={18} class="fill-white text-white" /> Resume
				</button>
				<button class="inline-flex min-h-[60px] items-center gap-3 rounded-xl !border !border-white/25 !bg-white/10 px-8 py-4 text-base font-semibold text-white shadow-lg shadow-black/20 backdrop-blur transition duration-200 hover:-translate-y-0.5 hover:!border-white/40 hover:!bg-white/15 active:translate-y-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-white/35 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120] sm:px-10">
					<Info size={18} /> Details
				</button>
				<div class="ml-4">
					<p class="text-sm text-white/60">42% watched</p>
					<div class="mt-2 h-1.5 w-40 overflow-hidden rounded-full bg-white/10">
						<div class="h-full w-[42%] bg-[#7C5CFF]"></div>
					</div>
				</div>
			</div>
		</div>
		<div class="hidden md:block">
			<img
				src={heroPoster}
				alt="Dune: Part Two poster"
				class="h-[400px] w-[276px] rounded-xl object-cover shadow-2xl shadow-black/60 ring-1 ring-white/10 xl:h-[440px] xl:w-[304px]"
			/>
		</div>
	</div>
</section>
