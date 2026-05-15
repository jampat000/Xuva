<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from './AppDrawer.svelte';
	import XuvaBrand from './XuvaBrand.svelte';
	import TopBar from '\$lib/Xuva/TopBar.svelte';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { buildMediaNavItems, type MediaNavItem } from '$lib/navigation/media-nav';
	import {
		isDesktopSidebarViewport,
		readSidebarPinnedPreference,
		writeSidebarPinnedPreference
	} from '$lib/navigation/sidebar-preferences';

	type ActiveRoute =
		| 'home'
		| 'movies'
		| 'tv'
		| 'collections'
		| 'watchlist'
		| 'continue-watching'
		| 'recently-added';

	interface DrawerBottomItem {
		id: 'settings';
		label: string;
		href: string;
	}

	const drawerBottomItems: DrawerBottomItem[] = [{ id: 'settings', label: 'Settings', href: '/settings' }];

	let {
		active = 'home',
		searchValue = $bindable(''),
		userInitials = 'U',
		children,
		companion
	} = $props<{
		active?: ActiveRoute;
		searchValue?: string;
		userInitials?: string;
		children?: Snippet;
		companion?: Snippet;
	}>();

	let menuOpen = $state(false);
	let libraries = $state<LibraryRecord[]>([]);
	let mediaNavItems = $state<MediaNavItem[]>(buildMediaNavItems());
	let sidebarPinned = $state(false);
	let desktopViewport = $state(false);
	const sidebarPreferenceKey = 'xuva.sidebar.media.pinned.v1';
	const drawerOpen = $derived.by(() => (desktopViewport && sidebarPinned ? true : menuOpen));
	const drawerPersistent = $derived.by(() => desktopViewport && sidebarPinned);

	function closeMenu(): void {
		if (desktopViewport && sidebarPinned) {
			sidebarPinned = false;
			writeSidebarPinnedPreference(sidebarPreferenceKey, false);
		}
		menuOpen = false;
	}

	onMount(() => {
		syncViewportState();
		sidebarPinned = readSidebarPinnedPreference(sidebarPreferenceKey);
		const handleResize = () => syncViewportState();
		void loadLibraries();
		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				closeMenu();
			}
		};
		window.addEventListener('resize', handleResize);
		window.addEventListener('keydown', handleKeydown);
		return () => {
			window.removeEventListener('resize', handleResize);
			window.removeEventListener('keydown', handleKeydown);
		};
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

	function toggleMenu(): void {
		if (desktopViewport && sidebarPinned) {
			sidebarPinned = false;
			writeSidebarPinnedPreference(sidebarPreferenceKey, false);
			menuOpen = false;
			return;
		}
		menuOpen = !menuOpen;
	}

	function togglePinned(): void {
		if (!desktopViewport) return;
		sidebarPinned = !sidebarPinned;
		writeSidebarPinnedPreference(sidebarPreferenceKey, sidebarPinned);
		if (!sidebarPinned) menuOpen = false;
	}
</script>

<div class="media-shell" class:media-shell--drawer-open={drawerPersistent} data-shell="media">
	<AppDrawer
		open={drawerOpen}
		label="Main navigation"
		testId="media-menu-drawer"
		drawerWidth="252px"
		showBackdrop={!drawerPersistent}
		dismissOnInteractOutside={!drawerPersistent}
		closeOnNavigate={!drawerPersistent}
		onClose={closeMenu}
	>
		{#snippet brand()}
			<XuvaBrand />
		{/snippet}
		{#snippet main()}
			{#each mediaNavItems as item (item.id)}
				<a href={item.href} class="app-drawer__link" aria-current={active === item.id ? 'page' : undefined}>
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
			{#each drawerBottomItems as item (item.id)}
				<a href={item.href} class="app-drawer__link">
					<Settings size={19} />
					{item.label}
				</a>
			{/each}
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

	<div class="media-shell__surface" data-testid="media-shell-surface">
		<TopBar
			menuOpen={drawerOpen}
			onMenuToggle={toggleMenu}
			onMenuClose={closeMenu}
			bind:searchValue
		/>

		<div class="media-shell__content" class:media-shell__content--with-companion={Boolean(companion)}>
			<main class="media-shell__primary">
				{@render children?.()}
			</main>
			{#if companion}
				<aside class="media-shell__companion">
					{@render companion()}
				</aside>
			{/if}
		</div>
	</div>
</div>

<style>
	.media-shell {
		--xuva-drawer-width: 252px;
		min-height: 100dvh;
		overflow-x: hidden;
		padding: 16px 24px 32px;
		background:
			radial-gradient(circle at 22% -18%, rgb(124 92 255 / 16%) 0%, transparent 38%),
			radial-gradient(circle at 84% 3%, rgb(55 84 150 / 16%) 0%, transparent 32%),
			linear-gradient(180deg, rgb(255 255 255 / 3%), transparent 24%),
			var(--xuva-color-bg-shell);
	}

	.media-shell__surface {
		min-height: calc(100dvh - 48px);
		width: 100%;
		margin-left: 0;
		transition:
			margin-left 260ms var(--xuva-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1)),
			width 260ms var(--xuva-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1));
		will-change: margin-left, width;
	}

	@media (min-width: 768px) {
		.media-shell--drawer-open .media-shell__surface {
			width: calc(100% - var(--xuva-drawer-width, 252px));
			margin-left: var(--xuva-drawer-width, 252px);
		}
	}

	.media-shell__content {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 18px;
		min-height: 0;
		max-width: 1420px;
		margin: 0 auto;
		width: 100%;
	}

	.media-shell__content--with-companion {
		grid-template-columns: minmax(0, 1fr) 284px;
	}

	.media-shell__primary,
	.media-shell__companion {
		min-width: 0;
	}

	@media (max-width: 980px) {
		.media-shell {
			padding: 8px 12px 14px;
		}

		.media-shell__content--with-companion {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 620px) {
		.media-shell__surface {
			transition: none;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.media-shell__surface {
			transition: none;
		}
	}
</style>
