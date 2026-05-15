<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from '$lib/components/shell/AppDrawer.svelte';
	import TopBar from '\$lib/Xuva/TopBar.svelte';
	import Logo from '\$lib/Xuva/Logo.svelte';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { buildMediaNavItems, type MediaNavItem } from '$lib/navigation/media-nav';
	import {
		isDesktopSidebarViewport,
		readSidebarPinnedPreference,
		writeSidebarPinnedPreference
	} from '$lib/navigation/sidebar-preferences';

	let {
		children
	}: {
		children: Snippet;
	} = $props();

	let menuOpen = $state(false);
	let currentPath = $state('/');
	let libraries = $state<LibraryRecord[]>([]);
	let mediaNavItems = $state<MediaNavItem[]>(buildMediaNavItems());
	let sidebarPinned = $state(false);
	let desktopViewport = $state(false);
	const sidebarPreferenceKey = 'xuva.sidebar.media.pinned.v1';
	const drawerOpen = $derived.by(() => (desktopViewport && sidebarPinned ? true : menuOpen));
	const drawerPersistent = $derived.by(() => desktopViewport && sidebarPinned);

	const activeRoute = $derived.by(() => {
		if (currentPath.startsWith('/movies')) return 'movies';
		if (currentPath.startsWith('/tv')) return 'tv';
		return 'home';
	});

	onMount(() => {
		currentPath = `${window.location.pathname}${window.location.hash || ''}`;
		syncViewportState();
		sidebarPinned = readSidebarPinnedPreference(sidebarPreferenceKey);
		const handleResize = () => syncViewportState();
		window.addEventListener('resize', handleResize);
		void loadLibraries();
		return () => window.removeEventListener('resize', handleResize);
	});

	function syncViewportState(): void {
		desktopViewport = isDesktopSidebarViewport();
		if (!desktopViewport) menuOpen = false;
	}

	async function loadLibraries(): Promise<void> {
		try {
			const payload = await getLibraries();
			libraries = payload.libraries || [];
			mediaNavItems = buildMediaNavItems(libraries);
		} catch {
			mediaNavItems = buildMediaNavItems();
		}
	}

	function closeMenu(): void {
		if (desktopViewport && sidebarPinned) {
			sidebarPinned = false;
			writeSidebarPinnedPreference(sidebarPreferenceKey, false);
		}
		menuOpen = false;
		document.querySelector<HTMLElement>('[data-testid="media-menu-button"]')?.focus();
	}

	function toggleMenu(): void {
		if (desktopViewport && sidebarPinned) {
			sidebarPinned = false;
			writeSidebarPinnedPreference(sidebarPreferenceKey, false);
			menuOpen = false;
			return;
		}
		menuOpen = !menuOpen;
		if (!menuOpen) document.querySelector<HTMLElement>('[data-testid="media-menu-button"]')?.focus();
	}

	function togglePinned(): void {
		if (!desktopViewport) return;
		sidebarPinned = !sidebarPinned;
		writeSidebarPinnedPreference(sidebarPreferenceKey, sidebarPinned);
		if (!sidebarPinned) menuOpen = false;
	}

	function linkCurrent(id: string): 'page' | undefined {
		return activeRoute === id ? 'page' : undefined;
	}
</script>

<div class="xuva-app-shell" class:xuva-app-shell--drawer-open={drawerPersistent}>
	<AppDrawer
		open={drawerOpen}
		label="Main navigation"
		testId="media-menu-drawer"
		showBackdrop={!drawerPersistent}
		dismissOnInteractOutside={!drawerPersistent}
		closeOnNavigate={!drawerPersistent}
		onClose={closeMenu}
	>
		{#snippet brand()}
			<Logo />
		{/snippet}
		{#snippet main()}
			{#each mediaNavItems as item (item.id)}
				<a class="app-drawer__link" href={item.href} aria-current={linkCurrent(item.id)}>
					{#if item.id === 'home'}
						<Home size={19} />
					{:else if item.id === 'movies'}
						<Film size={19} />
					{:else}
						<Tv size={19} />
					{/if}
					{item.label}
				</a>
			{/each}
		{/snippet}
		{#snippet bottom()}
			<a class="app-drawer__link" href="/settings">
				<Settings size={19} />
				Settings
			</a>
			{#if desktopViewport}
				<button
					type="button"
					class="app-drawer__link"
					aria-pressed={sidebarPinned}
					aria-label={sidebarPinned ? 'Unpin sidebar' : 'Pin sidebar'}
					onclick={togglePinned}
				>
					<svg viewBox="0 0 24 24" aria-hidden="true">
						<path
							d="M8 4.8h8l-1.4 4.1 2.7 2.8v1.1H12.8v5.8l-1.6.8v-6.6H6.7v-1.1l2.7-2.8Z"
							fill="none"
							stroke="currentColor"
							stroke-linejoin="round"
							stroke-width="1.55"
						/>
					</svg>
					{sidebarPinned ? 'Unpin sidebar' : 'Pin sidebar'}
				</button>
			{/if}
		{/snippet}
	</AppDrawer>

	<div class="xuva-app-surface" data-testid="xuva-app-surface">
		<TopBar
			menuOpen={drawerOpen}
			onMenuToggle={toggleMenu}
			onMenuClose={closeMenu}
		/>
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
