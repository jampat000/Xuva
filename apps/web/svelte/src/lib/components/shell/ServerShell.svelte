<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Menu } from 'lucide-svelte';
	import AppDrawer from './AppDrawer.svelte';
	import LorivoShell from './LorivoShell.svelte';
	import ServerSidebar from './ServerSidebar.svelte';
	import SettingsBrand from './SettingsBrand.svelte';
	import SettingsNav from './SettingsNav.svelte';

	let {
		active = 'dashboard',
		userDisplayName = 'Local User',
		userRole = 'Local Account',
		userInitials = 'U',
		showStorage = true,
		children
	} = $props<{
		active?: 'dashboard' | 'library' | 'scanning' | 'metadata' | 'playback' | 'storage' | 'access' | 'about';
		userDisplayName?: string;
		userRole?: string;
		userInitials?: string;
		showStorage?: boolean;
		children?: Snippet;
	}>();

	let settingsMenuOpen = $state(false);

	function closeSettingsMenu(): void {
		settingsMenuOpen = false;
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
		onClose={closeSettingsMenu}
	>
		{#snippet brand()}
			<SettingsBrand />
		{/snippet}
		{#snippet main()}
			<SettingsNav {active} {showStorage} />
		{/snippet}
		{#snippet bottom()}
			<SettingsNav section="secondary" />
		{/snippet}
	</AppDrawer>

	<div class="server-shell__surface" data-testid="settings-shell-surface">
		<LorivoShell density="default">
			{#snippet sidebar()}
				<div class="server-shell__sidebar-panel" data-testid="settings-mode-sidebar">
					<ServerSidebar {active} {userDisplayName} {userRole} {showStorage} />
				</div>
			{/snippet}

			{#snippet topbar()}
				<div class="server-shell__topbar-row">
					<button
						type="button"
						class="server-shell__menu-button"
						data-testid="settings-menu-button"
						data-lorivo-menu-trigger
						aria-label="Open menu"
						aria-expanded={settingsMenuOpen}
						onclick={() => (settingsMenuOpen = !settingsMenuOpen)}
					>
						<Menu size={19} />
					</button>
					<div class="server-shell__topbar-spacer" aria-hidden="true"></div>
					<button class="server-shell__profile-button" type="button" aria-label="Open Settings" onclick={() => (window.location.href = '/settings')}>
						<span>{userInitials}</span>
					</button>
				</div>
			{/snippet}

			{@render children?.()}
		</LorivoShell>
	</div>
</div>

<style>
	.server-shell {
		--lorivo-settings-accent: #9aa7ff;
		--lorivo-settings-accent-soft: rgb(154 167 255 / 10%);
		--lorivo-settings-accent-border: rgb(154 167 255 / 22%);
		min-height: 100dvh;
		overflow-x: hidden;
		background:
			radial-gradient(circle at 12% -14%, rgb(124 92 255 / 8%) 0%, transparent 34%),
			radial-gradient(circle at 92% 2%, rgb(62 78 120 / 10%) 0%, transparent 28%),
			var(--lorivo-color-bg-shell);
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
		border: 1px solid rgb(255 255 255 / 10%);
		background: rgb(17 24 39 / 78%);
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		align-items: center;
		justify-content: center;
	}

	.server-shell__profile-button {
		display: grid;
		place-items: center;
		width: 40px;
		height: 40px;
		padding: 0;
		border: 1px solid rgb(255 255 255 / 8%);
		border-radius: 999px;
		background: rgb(18 25 38 / 68%);
		color: color-mix(in srgb, var(--lorivo-color-text) 94%, transparent);
	}

	.server-shell__profile-button span {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 30px;
		height: 30px;
		border-radius: 999px;
		background: linear-gradient(180deg, rgb(154 167 255 / 18%), rgb(38 40 37 / 28%));
		color: #f4f1ea;
		font-size: 0.72rem;
		font-weight: 700;
	}

	:global([data-shell='server'] .v-shell) {
		display: block;
		min-height: 100dvh;
		background:
			linear-gradient(180deg, rgb(124 92 255 / 3%), transparent 24%),
			var(--lorivo-color-bg-shell);
	}

	:global([data-shell='server'] .v-shell__sidebar) {
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 42;
		width: 304px;
		height: 100dvh;
		border-right-color: rgb(154 167 255 / 12%);
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), rgb(255 255 255 / 1%) 36%, transparent),
			radial-gradient(circle at 16% -12%, rgb(124 92 255 / 12%) 0%, transparent 39%),
			radial-gradient(circle at 84% 112%, rgb(80 111 160 / 9%) 0%, transparent 42%),
			color-mix(in srgb, var(--lorivo-color-bg-sidebar) 95%, #101827 5%);
		box-shadow: inset -1px 0 0 rgb(255 255 255 / 3%);
	}

	.server-shell__sidebar-panel {
		height: 100%;
	}

	:global([data-shell='server'] .v-shell__topbar) {
		position: sticky;
		top: 0;
		z-index: 30;
		background:
			linear-gradient(180deg, rgb(11 17 32 / 94%), rgb(11 17 32 / 82%) 72%, transparent),
			transparent;
	}

	:global([data-shell='server'] .v-shell__main) {
		width: calc(100% - 304px);
		min-height: 100dvh;
		margin-left: 304px;
		padding: 20px 32px 28px;
	}

	:global([data-shell='server'] .v-sidebar) {
		padding: 18px 14px 18px;
	}

	:global([data-shell='server'] .v-sidebar__brand) {
		justify-content: flex-start;
		margin: 2px 8px 14px;
		min-height: 48px;
	}

	:global([data-shell='server'] .sidebar-item) {
		min-height: 48px;
		border-radius: 8px;
		font-size: 1rem;
	}

	:global([data-shell='server'] .sidebar-item:hover) {
		background: rgb(154 167 255 / 6%);
	}

	:global([data-shell='server'] .sidebar-item[data-active='true']) {
		border-color: color-mix(in srgb, var(--lorivo-settings-accent-border) 72%, transparent);
		background: linear-gradient(90deg, rgb(154 167 255 / 9%), rgb(154 167 255 / 2%));
		box-shadow: inset 2px 0 0 rgb(154 167 255 / 58%);
	}

	:global([data-shell='server'] .sidebar-item[data-active='true'] .sidebar-item__icon) {
		color: var(--lorivo-settings-accent);
	}

	:global([data-shell='server'] .app-drawer) {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), rgb(255 255 255 / 1%) 30%, transparent),
			radial-gradient(circle at 20% -12%, rgb(124 92 255 / 12%) 0%, transparent 38%),
			radial-gradient(circle at 84% 112%, rgb(80 111 160 / 10%) 0%, transparent 40%),
			color-mix(in srgb, var(--lorivo-color-bg-sidebar) 95%, #101827 5%);
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
