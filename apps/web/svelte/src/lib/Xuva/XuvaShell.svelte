<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from '$lib/components/shell/AppDrawer.svelte';
	import XuvaBrand from '$lib/components/shell/XuvaBrand.svelte';
	import TopBar from '\$lib/Xuva/TopBar.svelte';
	import { getAuthSession } from '$lib/api/auth';
	import { getClientBootstrap } from '$lib/api/auth';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { getSettings } from '$lib/api/operator';
	import { buildMediaNavItems, type MediaNavItem } from '$lib/navigation/media-nav';
	import { normalizeServerName } from '$lib/server-name';
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
	let serverGroupLabel = $state('Xuva');
	let libraries = $state<LibraryRecord[]>([]);
	let mediaNavItems = $state<MediaNavItem[]>(buildMediaNavItems());
	let canAccessSettings = $state(false);
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
		const handleServerNameChanged = (event: Event) => {
			const detail = (event as CustomEvent<{ serverName?: string }>).detail;
			serverGroupLabel = normalizeServerName(detail?.serverName);
		};
		window.addEventListener('resize', handleResize);
		window.addEventListener('xuva:server-name-changed', handleServerNameChanged);
		void loadLibraries();
		void loadServerName();
		void loadSessionAccess();
		return () => {
			window.removeEventListener('resize', handleResize);
			window.removeEventListener('xuva:server-name-changed', handleServerNameChanged);
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

<div class="xuva-app-shell" class:xuva-app-shell--drawer-open={drawerOpen}>
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
			{#if canAccessSettings}
				<a class="app-drawer__link" href="/settings">
					<Settings size={19} />
					Settings
				</a>
			{/if}
		{/snippet}
	</AppDrawer>

	<div class="xuva-app-surface" data-testid="xuva-app-surface">
		<TopBar
			menuOpen={drawerOpen}
			onMenuToggle={toggleMenu}
			onMenuClose={closeMenu}
			showSettingsShortcut={canAccessSettings}
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
			width: calc(100% - var(--xuva-drawer-width, 252px));
			margin-left: var(--xuva-drawer-width, 252px);
		}

		:global([data-testid='media-menu-drawer'][data-state='open']) ~ .xuva-app-surface {
			width: calc(100% - var(--xuva-drawer-width, 252px));
			margin-left: var(--xuva-drawer-width, 252px);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.xuva-app-surface {
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
