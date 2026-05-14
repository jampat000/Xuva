<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Folder, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from '$lib/components/shell/AppDrawer.svelte';
	import TopBar from '\$lib/Xuva/TopBar.svelte';
	import Logo from '\$lib/Xuva/Logo.svelte';

	let {
		children
	}: {
		children: Snippet;
	} = $props();

	let menuOpen = $state(false);
	let currentPath = $state('/');

	const activeRoute = $derived.by(() => {
		if (currentPath.startsWith('/movies')) return 'movies';
		if (currentPath.startsWith('/tv')) return 'tv';
		if (currentPath.startsWith('/setup') || currentPath === '/settings#libraries') return 'libraries';
		return 'home';
	});

	onMount(() => {
		currentPath = `${window.location.pathname}${window.location.hash || ''}`;
	});

	function closeMenu(): void {
		menuOpen = false;
		document.querySelector<HTMLElement>('[data-testid="media-menu-button"]')?.focus();
	}

	function toggleMenu(): void {
		menuOpen = !menuOpen;
		if (!menuOpen) document.querySelector<HTMLElement>('[data-testid="media-menu-button"]')?.focus();
	}

	function linkCurrent(id: string): 'page' | undefined {
		return activeRoute === id ? 'page' : undefined;
	}
</script>

<div class="xuva-app-shell" class:xuva-app-shell--drawer-open={menuOpen}>
	<AppDrawer
		open={menuOpen}
		label="Main navigation"
		testId="media-menu-drawer"
		onClose={closeMenu}
	>
		{#snippet brand()}
			<Logo />
		{/snippet}
		{#snippet main()}
			<a class="app-drawer__link" href="/" aria-current={linkCurrent('home')}>
				<Home size={19} />
				Home
			</a>
			<a class="app-drawer__link" href="/movies" aria-current={linkCurrent('movies')}>
				<Film size={19} />
				Movies
			</a>
			<a class="app-drawer__link" href="/tv" aria-current={linkCurrent('tv')}>
				<Tv size={19} />
				TV
			</a>
			<a class="app-drawer__link" href="/settings#libraries" aria-current={linkCurrent('libraries')}>
				<Folder size={19} />
				Libraries
			</a>
		{/snippet}
		{#snippet bottom()}
			<a class="app-drawer__link" href="/settings">
				<Settings size={19} />
				Settings
			</a>
		{/snippet}
	</AppDrawer>

	<div class="xuva-app-surface" data-testid="xuva-app-surface">
		<TopBar menuOpen={menuOpen} onMenuToggle={toggleMenu} onMenuClose={closeMenu} />
		{@render children()}
		<div class="h-16"></div>
	</div>
</div>

<style>
	.xuva-app-shell {
		min-height: 100dvh;
		overflow-x: hidden;
		background: #0b1120;
		color: white;
		font-family: var(--xuva-font-sans);
		-webkit-font-smoothing: antialiased;
	}

	.xuva-app-surface {
		min-height: 100dvh;
		background: #0b1120;
		width: 100%;
		margin-left: 0;
		transition:
			margin-left 260ms var(--xuva-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1)),
			width 260ms var(--xuva-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1));
		will-change: margin-left, width;
	}

	@media (min-width: 768px) {
		.xuva-app-shell--drawer-open .xuva-app-surface {
			width: calc(100% - var(--xuva-drawer-width, 320px));
			margin-left: var(--xuva-drawer-width, 320px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.xuva-app-surface {
			transition: none;
		}
	}
</style>
