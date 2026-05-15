<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from './AppDrawer.svelte';
	import XuvaBrand from './XuvaBrand.svelte';
	import TopBar from '\$lib/Xuva/TopBar.svelte';
	import { getAuthSession, getClientBootstrap } from '$lib/api/auth';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { getSettings } from '$lib/api/operator';
	import { buildMediaNavItems, type MediaNavItem } from '$lib/navigation/media-nav';
	import { normalizeServerName } from '$lib/server-name';
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
	let serverGroupLabel = $state('Xuva');
	let canAccessSettings = $state(false);
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
		const handleServerNameChanged = (event: Event) => {
			const detail = (event as CustomEvent<{ serverName?: string }>).detail;
			serverGroupLabel = normalizeServerName(detail?.serverName);
		};
		void loadLibraries();
		void loadServerName();
		void loadSessionAccess();
		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				closeMenu();
			}
		};
		window.addEventListener('resize', handleResize);
		window.addEventListener('xuva:server-name-changed', handleServerNameChanged);
		window.addEventListener('keydown', handleKeydown);
		return () => {
			window.removeEventListener('resize', handleResize);
			window.removeEventListener('xuva:server-name-changed', handleServerNameChanged);
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

	async function loadServerName(): Promise<void> {
		let configuredName = '';
		try {
			configuredName = await loadServerNameFromBootstrap();
			if (!configuredName) {
				const payload = await getClientBootstrap().catch(() => null);
				const server = (payload as { server?: { name?: string } } | null)?.server;
				configuredName = asText(server?.name);
			}
			if (!configuredName) {
				const settingsPayload = await getSettings().catch(() => null);
				configuredName = asText(settingsPayload?.config?.serverName);
			}
			if (!configuredName && typeof document !== 'undefined') {
				const title = asText(document.title);
				configuredName = title.endsWith('· Xuva') ? asText(title.replace(/\s*·\s*Xuva$/, '')) : '';
			}
		} catch {
			configuredName = '';
		}
		serverGroupLabel = normalizeServerName(configuredName);
	}

	async function loadServerNameFromBootstrap(): Promise<string> {
		if (typeof window === 'undefined') return '';
		try {
			const response = await window.fetch('/api/client/bootstrap', { credentials: 'include' });
			if (!response.ok) return '';
			const payload = (await response.json()) as { server?: { name?: string } };
			return asText(payload?.server?.name);
		} catch {
			return '';
		}
	}

	async function loadSessionAccess(): Promise<void> {
		try {
			const session = await getAuthSession().catch(() => null);
			if (!session) {
				if (typeof window !== 'undefined') window.location.replace('/signin');
				return;
			}
			if (session?.authDisabled) {
				canAccessSettings = true;
				return;
			}
			const role = asText(session?.user?.role).toLowerCase();
			if (!role) {
				if (typeof window !== 'undefined') window.location.replace('/signin');
				return;
			}
			canAccessSettings = role === 'admin';
		} catch {
			canAccessSettings = false;
			if (typeof window !== 'undefined') window.location.replace('/signin');
		}
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
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

<div class="media-shell" class:media-shell--drawer-open={drawerOpen} data-shell="media">
	<AppDrawer
		open={drawerOpen}
		label="Main navigation"
		testId="media-menu-drawer"
		drawerWidth="252px"
		showCloseButton={false}
		showPinButton={desktopViewport}
		pinLabel={sidebarPinned ? 'Unpin sidebar' : 'Pin sidebar'}
		pinned={sidebarPinned}
		showBackdrop={!drawerPersistent}
		dismissOnInteractOutside={!drawerPersistent}
		closeOnNavigate={!drawerPersistent}
		onClose={closeMenu}
		onPinToggle={togglePinned}
	>
		{#snippet brand()}
			<div class="media-brand">
				<XuvaBrand showBlurb={true} />
			</div>
		{/snippet}
		{#snippet main()}
			<p class="media-nav-group">{serverGroupLabel}</p>
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
			{#if canAccessSettings}
				{#each drawerBottomItems as item (item.id)}
					<a href={item.href} class="app-drawer__link">
						<Settings size={19} />
						{item.label}
					</a>
				{/each}
			{/if}
		{/snippet}
	</AppDrawer>

	<div class="media-shell__surface" data-testid="media-shell-surface">
		<TopBar
			menuOpen={drawerOpen}
			onMenuToggle={toggleMenu}
			onMenuClose={closeMenu}
			showSettingsShortcut={canAccessSettings}
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

		:global([data-testid='media-menu-drawer'][data-state='open']) ~ .media-shell__surface {
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

	.media-nav-group {
		margin: 6px 10px 2px;
		color: color-mix(in srgb, var(--xuva-color-text-soft) 84%, transparent);
		font-size: 0.69rem;
		font-weight: 750;
		letter-spacing: 0.04em;
	}

	.media-brand {
		display: flex;
		align-items: center;
		justify-content: flex-start;
		gap: 0;
		width: 100%;
		min-width: 0;
		min-height: 28px;
	}
</style>
