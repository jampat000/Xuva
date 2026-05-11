<script lang="ts">
	import { onMount } from 'svelte';
	import { Menu, Search, Settings, X } from 'lucide-svelte';
	import Logo from '$lib/lorivo/Logo.svelte';

	const mediaNavItems = [
		{ id: 'home', label: 'Home', href: '/' },
		{ id: 'movies', label: 'Movies', href: '/movies' },
		{ id: 'tv', label: 'TV', href: '/tv' }
	];

	let menuOpen = $state(false);
	let currentPath = $state('/');

	const activeRoute = $derived.by(() => {
		if (currentPath.startsWith('/movies')) return 'movies';
		if (currentPath.startsWith('/tv')) return 'tv';
		return 'home';
	});

	onMount(() => {
		currentPath = window.location.pathname || '/';
		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				closeMenu();
			}
		};
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});

	function closeMenu(): void {
		menuOpen = false;
	}
</script>

<header class="relative z-30 flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
	<div class="flex min-w-0 items-center gap-3">
		<button
			type="button"
			data-testid="media-menu-button"
			class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-[#111827]/80 text-white/85 shadow-lg shadow-black/20 transition hover:border-white/25 hover:bg-white/10"
			aria-label={menuOpen ? 'Close menu' : 'Open menu'}
			aria-expanded={menuOpen}
			onclick={() => (menuOpen = !menuOpen)}
		>
			{#if menuOpen}
				<X size={20} />
			{:else}
				<Menu size={20} />
			{/if}
		</button>
		<a href="/" class="shrink-0" aria-label="Go to Home" onclick={closeMenu}>
			<Logo />
		</a>
	</div>
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
	<div class="flex shrink-0 items-center gap-2">
		<a
			href="/settings"
			class="hidden h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-[#111827]/80 text-white/72 transition hover:border-white/25 hover:bg-white/10 hover:text-white sm:inline-flex"
			aria-label="Open Settings"
		>
			<Settings size={18} />
		</a>
		<div
			aria-label="Profile"
			class="grid h-9 w-9 place-items-center rounded-full border border-white/10 bg-[#22302c] text-xs font-semibold text-white/80"
		>
			L
		</div>
	</div>
</header>

{#if menuOpen}
	<button
		type="button"
		class="fixed inset-0 z-40 bg-black/55"
		aria-label="Close menu"
		onclick={closeMenu}
	></button>
	<aside
		class="fixed left-3 top-3 z-50 w-[min(320px,calc(100vw-24px))] rounded-2xl border border-white/10 bg-[#111827] p-4 shadow-2xl shadow-black/45"
		data-testid="media-menu-drawer"
	>
		<div class="mb-4 flex items-center justify-between gap-3">
			<Logo />
			<button
				type="button"
				class="inline-flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-white/[0.04] text-white/85"
				aria-label="Close menu"
				onclick={closeMenu}
			>
				<X size={20} />
			</button>
		</div>
		<nav class="grid gap-2" aria-label="Mobile media navigation">
			{#each mediaNavItems as item (item.id)}
				<a
					href={item.href}
					aria-current={activeRoute === item.id ? 'page' : undefined}
					class={`rounded-xl border px-4 py-3 text-base font-semibold ${
						activeRoute === item.id
							? 'border-[#7C5CFF]/45 bg-[#7C5CFF]/18 text-white'
							: 'border-white/10 bg-white/[0.03] text-white/72'
					}`}
					onclick={closeMenu}
				>
					{item.label}
				</a>
			{/each}
		</nav>
		<div class="mt-3 border-t border-white/10 pt-3">
			<a
				href="/settings"
				class="flex items-center gap-3 rounded-xl border border-white/10 bg-white/[0.03] px-4 py-3 text-base font-semibold text-white/80"
				onclick={closeMenu}
			>
				<Settings size={18} />
				Settings
			</a>
		</div>
	</aside>
{/if}
