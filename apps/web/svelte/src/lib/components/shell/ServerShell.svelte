<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { ArrowLeft, Database, Info, Menu, Play, RefreshCw, Server, Tags } from 'lucide-svelte';
	import AppTopbar from './AppTopbar.svelte';
	import AppDrawer from './AppDrawer.svelte';
	import LorivoBrand from './LorivoBrand.svelte';
	import LorivoShell from './LorivoShell.svelte';

	let {
		active = 'library',
		searchValue = $bindable(''),
		userDisplayName = 'Local User',
		userRole = 'Local Account',
		userInitials = 'U',
		children
	} = $props<{
		active?: 'library' | 'scanning' | 'metadata' | 'playback' | 'server' | 'about';
		searchValue?: string;
		userDisplayName?: string;
		userRole?: string;
		userInitials?: string;
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

<div class="server-shell" class:server-shell--drawer-open={settingsMenuOpen} data-shell="server">
	<AppDrawer
		open={settingsMenuOpen}
		label="Settings navigation"
		testId="settings-menu-drawer"
		onClose={closeSettingsMenu}
	>
		{#snippet brand()}
			<LorivoBrand />
		{/snippet}
		{#snippet main()}
			<a class="app-drawer__link" href="/settings#library" aria-current={active === 'library' ? 'page' : undefined}>
				<Database size={19} />
				Library
			</a>
			<a class="app-drawer__link" href="/settings#scanning" aria-current={active === 'scanning' ? 'page' : undefined}>
				<RefreshCw size={19} />
				Scanning
			</a>
			<a class="app-drawer__link" href="/settings#metadata" aria-current={active === 'metadata' ? 'page' : undefined}>
				<Tags size={19} />
				Metadata
			</a>
			<a class="app-drawer__link" href="/settings#playback" aria-current={active === 'playback' ? 'page' : undefined}>
				<Play size={19} />
				Playback
			</a>
			<a class="app-drawer__link" href="/settings#server" aria-current={active === 'server' ? 'page' : undefined}>
				<Server size={19} />
				Server
			</a>
			<a class="app-drawer__link" href="/settings#about" aria-current={active === 'about' ? 'page' : undefined}>
				<Info size={19} />
				About
			</a>
		{/snippet}
		{#snippet bottom()}
			<a class="app-drawer__link" href="/">
				<ArrowLeft size={19} />
				Back to Media
			</a>
		{/snippet}
	</AppDrawer>

	<div class="server-shell__surface" data-testid="settings-shell-surface">
		<LorivoShell density="default">
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
					<AppTopbar bind:searchValue {userInitials} onProfileClick={() => (window.location.href = '/settings')} />
				</div>
			{/snippet}

			{@render children?.()}
		</LorivoShell>
	</div>
</div>

<style>
	.server-shell {
		min-height: 100dvh;
		overflow-x: hidden;
		background: var(--lorivo-color-bg-shell);
	}

	.server-shell__surface {
		min-height: 100dvh;
		width: 100%;
		margin-left: 0;
		transition:
			margin-left 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1)),
			width 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1));
		will-change: margin-left, width;
	}

	@media (min-width: 768px) {
		.server-shell--drawer-open .server-shell__surface {
			width: calc(100% - var(--lorivo-drawer-width, 320px));
			margin-left: var(--lorivo-drawer-width, 320px);
		}
	}

	.server-shell__topbar-row {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: 10px;
		width: 100%;
	}

	.server-shell__menu-button {
		display: inline-flex;
		width: 40px;
		height: 40px;
		border-radius: 12px;
		border: 1px solid rgb(255 255 255 / 13%);
		background: linear-gradient(180deg, rgb(31 41 55 / 82%), rgb(17 24 39 / 76%));
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		align-items: center;
		justify-content: center;
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 9%),
			0 12px 28px rgb(0 0 0 / 22%);
	}

	:global([data-shell='server'] .v-shell) {
		grid-template-columns: minmax(0, 1fr);
	}

	:global([data-shell='server'] .v-shell__sidebar) {
		display: none;
	}

	:global([data-shell='server'] .v-shell__topbar) {
		position: sticky;
		top: 0;
		z-index: 30;
		background:
			linear-gradient(180deg, rgb(11 17 32 / 94%), rgb(11 17 32 / 82%) 72%, transparent),
			transparent;
	}

	@media (max-width: 980px) {
		:global([data-shell='server'] .v-shell__main) {
			padding: 14px 14px 20px;
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.server-shell__surface {
			transition: none;
		}
	}
</style>
