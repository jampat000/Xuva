<script lang="ts">
	import { ChevronDown, Search } from 'lucide-svelte';
	import Logo from '$lib/lorivo/Logo.svelte';
</script>

<header class="flex items-center justify-between gap-6 px-8 py-4">
	<Logo />
	<div class="flex max-w-2xl flex-1 justify-center">
		<div class="relative w-full max-w-[640px]">
			<Search size={16} class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-white/40" />
			<input
				type="text"
				placeholder="Search"
				class="h-10 w-full rounded-full border border-white/5 bg-[#111827] pl-11 pr-4 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-1 focus:ring-[#7C5CFF]/50"
			/>
		</div>
	</div>
	<div class="flex items-center gap-2">
		<img
			src="https://i.pravatar.cc/80?img=12"
			alt="Profile"
			class="h-9 w-9 rounded-full object-cover"
		/>
		<ChevronDown size={16} class="text-white/60" />
	</div>
</header>
