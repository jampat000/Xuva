<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { logout } from '$lib/api/auth';
	import AppDrawer from './AppDrawer.svelte';
	import ProfileMenu from './ProfileMenu.svelte';
	import ServerSidebar from './ServerSidebar.svelte';
	import SettingsBrand from './SettingsBrand.svelte';
	import { isDesktopSidebarViewport } from '$lib/navigation/sidebar-preferences';

	let {
		active = 'dashboard',
		userDisplayName = '',
		userRole = '',
		userInitials = 'U',
		showStorage = true,
		serverGroupLabel = 'Xuva',
		canSignOut = false,
		children
	} = $props<{
		active?:
			| 'dashboard'
			| 'general'
			| 'libraries'
			| 'scanning'
			| 'metadata'
			| 'playback'
			| 'transcoding'
			| 'storage'
			| 'network'
			| 'pairing'
			| 'approved-devices'
			| 'discovery'
			| 'admin-access'
			| 'planned-tools'
			| 'about';
		userDisplayName?: string;
		userRole?: string;
		userInitials?: string;
		showStorage?: boolean;
		serverGroupLabel?: string;
		canSignOut?: boolean;
		children?: Snippet;
	}>();

	let menuOpen = $state(false);
	let desktopViewport = $state(false);
	const drawerOpen = $derived.by(() => (desktopViewport ? true : menuOpen));
	const drawerPersistent = $derived.by(() => desktopViewport);

	onMount(() => {
		syncViewportState();
		const handleResize = () => syncViewportState();
		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') closeMenu();
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

	function closeMenu(): void {
		if (desktopViewport) return;
		menuOpen = false;
		document.querySelector<HTMLElement>('[data-testid="settings-menu-button"]')?.focus();
	}

	function toggleMenu(): void {
		if (desktopViewport) return;
		menuOpen = !menuOpen;
		if (!menuOpen) document.querySelector<HTMLElement>('[data-testid="settings-menu-button"]')?.focus();
	}

	async function signOut(): Promise<void> {
		try {
			await logout();
		} finally {
			if (typeof window !== 'undefined') window.location.href = '/signin';
		}
	}
</script>

<div class="server-shell" class:server-shell--drawer-open={drawerPersistent} data-shell="server">
	<AppDrawer
		open={drawerOpen}
		label="Settings navigation"
		testId="settings-menu-drawer"
		drawerWidth="252px"
		showCloseButton={false}
		showBackdrop={!drawerPersistent}
		dismissOnInteractOutside={!drawerPersistent}
		closeOnNavigate={!drawerPersistent}
		onClose={closeMenu}
	>
		{#snippet brand()}
			<SettingsBrand />
		{/snippet}
		{#snippet main()}
			<ServerSidebar {active} {showStorage} {serverGroupLabel} showBrand={false} />
		{/snippet}
		{#snippet bottom()}
			<a href="/" class="app-drawer__link app-drawer__link--back">
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M14.5 5 8 12l6.5 7"
						fill="none"
						stroke="currentColor"
						stroke-width="1.7"
						stroke-linecap="round"
						stroke-linejoin="round"
					/>
				</svg>
				Back to Media
			</a>
		{/snippet}
	</AppDrawer>

	<div class="server-shell__surface" data-testid="settings-shell-surface">
		<header class="server-shell__topbar">
			<div class="server-shell__topbar-row">
				<div class="server-shell__nav-controls">
					<button
						type="button"
						data-testid="settings-menu-button"
						data-xuva-menu-trigger
						class="server-shell__utility-button"
						aria-label="Open settings navigation"
						aria-expanded={drawerOpen}
						onclick={toggleMenu}
					>
						<svg viewBox="0 0 24 24" aria-hidden="true">
							<path
								d="M5 7h14M5 12h14M5 17h14"
								fill="none"
								stroke="currentColor"
								stroke-linecap="round"
								stroke-width="1.8"
							/>
						</svg>
					</button>
				</div>
				<ProfileMenu
					initials={userInitials}
					name={userDisplayName || 'User'}
					role={userRole || 'User'}
					canSignOut={canSignOut}
					changePasswordHref="/settings#admin-access"
					onSignOut={signOut}
				/>
			</div>
		</header>

		<main class="server-shell__main">
			{@render children?.()}
		</main>
	</div>
</div>

<style>
	.server-shell {
		--xuva-settings-accent: #6a74d9;
		--xuva-settings-accent-soft: rgb(106 116 217 / 10%);
		--xuva-settings-accent-border: rgb(106 116 217 / 24%);
		--xuva-color-bg-shell: #e9edf3;
		--xuva-color-bg-sidebar: #d3dae5;
		--xuva-color-bg-panel: #f5f7fb;
		--xuva-color-bg-panel-elevated: #ffffff;
		--xuva-color-bg-soft: #e3e9f2;
		--xuva-color-border-soft: rgb(15 23 42 / 14%);
		--xuva-color-border-strong: rgb(99 102 241 / 28%);
		--xuva-color-text: #0f172a;
		--xuva-color-text-muted: #334155;
		--xuva-color-text-soft: #64748b;
		--profile-menu-trigger-border: rgb(15 23 42 / 14%);
		--profile-menu-trigger-bg: rgb(255 255 255 / 72%);
		--profile-menu-avatar-bg: linear-gradient(180deg, rgb(154 167 255 / 24%), rgb(130 145 193 / 22%));
		--profile-menu-avatar-color: #0f172a;
		--profile-menu-panel-bg: #ffffff;
		--profile-menu-panel-border: rgb(15 23 42 / 14%);
		--profile-menu-panel-text: #0f172a;
		--profile-menu-row-hover: rgb(99 102 241 / 10%);
		min-height: 100dvh;
		overflow-x: hidden;
		background:
			radial-gradient(circle at 16% -12%, rgb(255 255 255 / 82%) 0%, transparent 42%),
			radial-gradient(circle at 84% 2%, rgb(214 222 236 / 60%) 0%, transparent 34%),
			linear-gradient(180deg, #edf1f7 0%, #e5ebf3 42%, #dee5ef 100%);
	}

	.server-shell__surface {
		min-height: 100dvh;
		width: 100%;
		transition:
			margin-left 260ms var(--xuva-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1)),
			width 260ms var(--xuva-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1));
	}

	.server-shell__topbar {
		position: sticky;
		top: 0;
		z-index: 30;
		padding: 18px 28px 10px;
		background:
			linear-gradient(180deg, rgb(249 251 255 / 96%), rgb(242 246 252 / 84%) 72%, transparent),
			transparent;
	}

	.server-shell__topbar-row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		gap: 10px;
		width: 100%;
	}

	.server-shell__nav-controls {
		display: inline-flex;
		align-items: center;
		gap: 10px;
	}

	.server-shell__utility-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: 12px;
		border: 1px solid rgb(15 23 42 / 12%);
		background: rgb(255 255 255 / 72%);
		color: color-mix(in srgb, var(--xuva-color-text) 82%, transparent);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 62%);
		transition:
			border-color 180ms ease,
			background-color 180ms ease,
			color 180ms ease,
			box-shadow 180ms ease;
	}

	@media (min-width: 980px) {
		.server-shell__utility-button {
			display: none;
		}
	}

	.server-shell__utility-button:hover,
	.server-shell__utility-button:focus-visible {
		border-color: rgb(99 102 241 / 28%);
		background: rgb(255 255 255 / 94%);
		color: var(--xuva-color-text);
		outline: none;
	}

	.server-shell__utility-button svg {
		width: 18px;
		height: 18px;
	}

	.server-shell__main {
		min-height: calc(100dvh - 69px);
		padding: 20px 32px 28px;
	}

	.server-shell--drawer-open .server-shell__surface {
		width: calc(100% - var(--xuva-drawer-width, 252px));
		margin-left: var(--xuva-drawer-width, 252px);
	}

	:global([data-shell='server'] .app-drawer) {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 72%), rgb(255 255 255 / 26%) 34%, transparent),
			radial-gradient(circle at 20% -12%, rgb(255 255 255 / 42%) 0%, transparent 44%),
			radial-gradient(circle at 84% 112%, rgb(210 219 235 / 28%) 0%, transparent 44%),
			var(--xuva-color-bg-sidebar);
	}

	:global([data-shell='server'] .app-drawer__brand .v-brand) {
		--xuva-brand-wordmark-color: #0f172a;
		--xuva-brand-wordmark-shadow: none;
	}

	:global([data-shell='server'] .v-sidebar) {
		height: 100%;
		padding: 0;
		overflow-y: auto;
	}

	:global([data-shell='server'] .v-sidebar__brand) {
		justify-content: center;
		margin: 6px 8px 10px;
		min-height: 28px;
	}

	:global([data-shell='server'] .v-sidebar__brand .v-brand) {
		min-height: 28px;
		justify-content: center;
		--xuva-brand-wordmark-color: #0f172a;
		--xuva-brand-wordmark-shadow: none;
	}

	:global([data-shell='server'] .sidebar-item) {
		min-height: 48px;
		border-radius: 8px;
		font-size: 1rem;
	}

	:global([data-shell='server'] .sidebar-item:hover) {
		background: rgb(99 102 241 / 8%);
	}

	:global([data-shell='server'] .sidebar-item[data-active='true']) {
		border-color: color-mix(in srgb, var(--xuva-settings-accent-border) 72%, transparent);
		background: linear-gradient(90deg, rgb(154 167 255 / 9%), rgb(154 167 255 / 2%));
		box-shadow: inset 2px 0 0 rgb(154 167 255 / 58%);
	}

	:global([data-shell='server'] .sidebar-item[data-active='true'] .sidebar-item__icon) {
		color: var(--xuva-settings-accent);
	}

	@media (max-width: 979px) {
		.server-shell__topbar {
			padding: 14px 14px 8px;
		}

		.server-shell__main {
			padding: 14px 14px 20px;
		}

		.server-shell__surface {
			width: 100%;
			margin-left: 0;
		}

		:global([data-shell='server'] .v-sidebar) {
			display: flex;
			flex-direction: column;
			padding: 0;
			gap: 11px;
		}

		:global([data-shell='server'] .v-sidebar__brand) {
			margin: 6px 8px 10px;
			padding: 0;
			justify-content: center;
		}

		:global([data-shell='server'] .v-sidebar__nav) {
			display: grid;
			gap: 5px;
			overflow: visible;
			padding-bottom: 0;
		}
	}
</style>
