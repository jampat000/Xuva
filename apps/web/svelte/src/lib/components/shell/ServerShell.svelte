<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Menu } from 'lucide-svelte';
	import { logout } from '$lib/api/auth';
	import AppDrawer from './AppDrawer.svelte';
	import XuvaShell from './XuvaShell.svelte';
	import ProfileMenu from './ProfileMenu.svelte';
	import ServerSidebar from './ServerSidebar.svelte';
	import SettingsBrand from './SettingsBrand.svelte';
	import SettingsNav from './SettingsNav.svelte';

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

	let settingsMenuOpen = $state(false);

	function closeSettingsMenu(): void {
		settingsMenuOpen = false;
	}

	async function signOut(): Promise<void> {
		try {
			await logout();
		} finally {
			if (typeof window !== 'undefined') window.location.href = '/signin';
		}
	}

	onMount(() => {
		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				closeSettingsMenu();
			}
		};
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});
</script>

<div class="server-shell" data-shell="server">
	<AppDrawer
		open={settingsMenuOpen}
		label="Settings navigation"
		testId="settings-menu-drawer"
		drawerWidth="252px"
		onClose={closeSettingsMenu}
	>
		{#snippet brand()}
			<SettingsBrand />
		{/snippet}
		{#snippet main()}
			<SettingsNav {active} {showStorage} {serverGroupLabel} />
		{/snippet}
		{#snippet bottom()}
			<SettingsNav section="secondary" />
		{/snippet}
	</AppDrawer>

	<div class="server-shell__surface" data-testid="settings-shell-surface">
		<XuvaShell density="default">
			{#snippet sidebar()}
				<div class="server-shell__sidebar-panel" data-testid="settings-mode-sidebar">
					<ServerSidebar {active} {showStorage} {serverGroupLabel} />
				</div>
			{/snippet}

			{#snippet topbar()}
				<div class="server-shell__topbar-row">
					<button
						type="button"
						class="server-shell__menu-button"
						data-testid="settings-menu-button"
						data-xuva-menu-trigger
						aria-label="Open menu"
						aria-expanded={settingsMenuOpen}
						onclick={() => (settingsMenuOpen = !settingsMenuOpen)}
					>
						<Menu size={19} />
					</button>
					<div class="server-shell__topbar-spacer" aria-hidden="true"></div>
					<ProfileMenu
						initials={userInitials}
						name={userDisplayName || 'User'}
						role={userRole || 'User'}
						canSignOut={canSignOut}
						changePasswordHref="/settings#admin-access"
						onSignOut={signOut}
					/>
				</div>
			{/snippet}

			{@render children?.()}
		</XuvaShell>
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
	}

	.server-shell__topbar-row {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 10px;
		width: 100%;
	}

	.server-shell__topbar-spacer {
		min-width: 0;
	}

	.server-shell__menu-button {
		display: inline-flex;
		width: 40px;
		height: 40px;
		border-radius: 8px;
		border: 1px solid rgb(15 23 42 / 14%);
		background: rgb(255 255 255 / 72%);
		color: color-mix(in srgb, var(--xuva-color-text) 84%, transparent);
		align-items: center;
		justify-content: center;
	}

	:global([data-shell='server'] .v-shell) {
		display: block;
		min-height: 100dvh;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 44%), transparent 24%),
			linear-gradient(180deg, #f0f4f9 0%, #e7edf5 100%);
	}

	:global([data-shell='server'] .v-shell__sidebar) {
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 42;
		width: 252px;
		height: 100dvh;
		border-right-color: rgb(154 167 255 / 12%);
		background:
			linear-gradient(180deg, rgb(255 255 255 / 60%), rgb(255 255 255 / 20%) 34%, transparent),
			radial-gradient(circle at 20% -12%, rgb(255 255 255 / 40%) 0%, transparent 46%),
			radial-gradient(circle at 88% 112%, rgb(206 216 233 / 26%) 0%, transparent 44%),
			var(--xuva-color-bg-sidebar);
		box-shadow:
			inset -1px 0 0 rgb(255 255 255 / 62%),
			12px 0 34px rgb(15 23 42 / 10%);
	}

	.server-shell__sidebar-panel {
		height: 100%;
	}

	:global([data-shell='server'] .v-shell__topbar) {
		position: sticky;
		top: 0;
		z-index: 30;
		background:
			linear-gradient(180deg, rgb(249 251 255 / 96%), rgb(242 246 252 / 84%) 72%, transparent),
			transparent;
		border-bottom: 1px solid rgb(15 23 42 / 12%);
	}

	:global([data-shell='server'] .v-shell__main) {
		width: calc(100% - 252px);
		min-height: 100dvh;
		margin-left: 252px;
		padding: 20px 32px 28px;
	}

	:global([data-shell='server'] .v-shell__main h1),
	:global([data-shell='server'] .v-shell__main h2),
	:global([data-shell='server'] .v-shell__main h3),
	:global([data-shell='server'] .v-shell__main h4) {
		color: var(--xuva-color-text) !important;
	}

	:global([data-shell='server'] .v-sidebar) {
		padding: 16px 10px 16px;
		overflow-y: auto;
	}

	:global([data-shell='server'] .v-sidebar__brand) {
		justify-content: center;
		margin: 1px 8px 9px;
		min-height: 34px;
	}

	:global([data-shell='server'] .v-sidebar__brand .v-brand) {
		min-height: 34px;
		justify-content: center;
	}

	:global([data-shell='server'] .v-brand) {
		--xuva-brand-wordmark-color: #0f172a;
		--xuva-brand-wordmark-shadow: none;
		--xuva-brand-mark-primary: #6a5ce8;
		--xuva-brand-mark-secondary: #0f172a;
		--xuva-brand-mark-shadow: drop-shadow(0 6px 10px rgb(15 23 42 / 14%));
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

	:global([data-shell='server'] .app-drawer) {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 72%), rgb(255 255 255 / 26%) 34%, transparent),
			radial-gradient(circle at 20% -12%, rgb(255 255 255 / 42%) 0%, transparent 44%),
			radial-gradient(circle at 84% 112%, rgb(210 219 235 / 28%) 0%, transparent 44%),
			var(--xuva-color-bg-sidebar);
	}

	@media (max-width: 980px) {
		:global([data-shell='server'] .v-shell) {
			display: grid;
			grid-template-columns: 1fr;
		}

		:global([data-shell='server'] .v-shell__sidebar) {
			display: none;
		}

		:global([data-shell='server'] .v-shell__main) {
			width: 100%;
			margin-left: 0;
			padding: 14px 14px 20px;
		}
	}

	@media (min-width: 981px) {
		.server-shell__topbar-row {
			grid-template-columns: minmax(0, 1fr) auto;
		}

		.server-shell__menu-button {
			display: none;
		}

		:global([data-shell='server'] .app-drawer) {
			display: none;
		}
	}
</style>
