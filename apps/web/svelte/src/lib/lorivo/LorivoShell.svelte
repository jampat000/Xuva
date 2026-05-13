<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Folder, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from '$lib/components/shell/AppDrawer.svelte';
	import TopBar from '$lib/lorivo/TopBar.svelte';
	import Logo from '$lib/lorivo/Logo.svelte';

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

<div class="lorivo-app-shell" class:lorivo-app-shell--drawer-open={menuOpen}>
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

	<div class="lorivo-app-surface" data-testid="lorivo-app-surface">
		<TopBar menuOpen={menuOpen} onMenuToggle={toggleMenu} onMenuClose={closeMenu} />
		{@render children()}
		<div class="h-16"></div>
	</div>
</div>

<style>
	.lorivo-app-shell {
		min-height: 100dvh;
		overflow-x: hidden;
		background: #0b1120;
		color: white;
		font-family: var(--font-sans, ui-sans-serif, system-ui, sans-serif);
		-webkit-font-smoothing: antialiased;
	}

	.lorivo-app-surface {
		min-height: 100dvh;
		background: #0b1120;
		width: 100%;
		margin-left: 0;
		transition:
			margin-left 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1)),
			width 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1));
		will-change: margin-left, width;
	}

	@media (min-width: 768px) {
		.lorivo-app-shell--drawer-open .lorivo-app-surface {
			width: calc(100% - var(--lorivo-drawer-width, 320px));
			margin-left: var(--lorivo-drawer-width, 320px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.lorivo-app-surface {
			transition: none;
		}
	}
</style>
