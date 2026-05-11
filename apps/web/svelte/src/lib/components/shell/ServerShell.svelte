<script lang="ts">
	import type { Snippet } from 'svelte';
	import { Menu, X } from 'lucide-svelte';
	import AppTopbar from './AppTopbar.svelte';
	import ServerSidebar from './ServerSidebar.svelte';
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
</script>

<div data-shell="server">
	<LorivoShell density="default">
		{#snippet sidebar()}
			<ServerSidebar {active} {userDisplayName} {userRole} />
		{/snippet}

		{#snippet topbar()}
			<div class="server-shell__topbar-row">
				<button
					type="button"
					class="server-shell__menu-button"
					data-testid="settings-menu-button"
					aria-label={settingsMenuOpen ? 'Close settings menu' : 'Open settings menu'}
					aria-expanded={settingsMenuOpen}
					onclick={() => (settingsMenuOpen = !settingsMenuOpen)}
				>
					{#if settingsMenuOpen}
						<X size={19} />
					{:else}
						<Menu size={19} />
					{/if}
				</button>
				<AppTopbar bind:searchValue {userInitials} onProfileClick={() => (window.location.href = '/settings')} />
			</div>
		{/snippet}

		{@render children?.()}
	</LorivoShell>

	{#if settingsMenuOpen}
		<button
			type="button"
			class="server-shell__backdrop"
			aria-label="Close settings menu"
			onclick={closeSettingsMenu}
		></button>
		<aside class="server-shell__drawer" aria-label="Settings menu" data-testid="settings-menu-drawer">
			<div role="presentation" onclick={closeSettingsMenu}>
				<ServerSidebar {active} {userDisplayName} {userRole} />
			</div>
		</aside>
	{/if}
</div>

<style>
	.server-shell__topbar-row {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: center;
		gap: 10px;
		width: 100%;
	}

	.server-shell__menu-button {
		display: none;
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

	.server-shell__backdrop {
		position: fixed;
		inset: 0;
		z-index: 70;
		border: 0;
		background: rgb(0 0 0 / 52%);
	}

	.server-shell__drawer {
		position: fixed;
		inset: 10px auto 10px 10px;
		z-index: 80;
		width: min(312px, calc(100vw - 20px));
		overflow: auto;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: 18px;
		background:
			radial-gradient(circle at 14% -18%, rgb(88 201 176 / 12%) 0%, rgb(88 201 176 / 0%) 36%),
			radial-gradient(circle at 80% 108%, rgb(131 119 93 / 12%) 0%, rgb(131 119 93 / 0%) 42%),
			var(--lorivo-color-bg-sidebar);
		box-shadow: 22px 0 42px rgb(0 0 0 / 42%);
	}

	.server-shell__drawer :global(.v-sidebar) {
		display: flex;
		flex-direction: column;
		height: 100%;
		padding: 14px 12px;
	}

	.server-shell__drawer :global(.v-sidebar__nav) {
		display: grid;
		gap: 6px;
		overflow: visible;
	}

	.server-shell__drawer :global(.v-sidebar__brand) {
		display: flex;
		margin: 8px 8px 14px;
	}

	.server-shell__drawer :global(.v-sidebar__profile) {
		display: block;
	}

	@media (max-width: 980px) {
		.server-shell__menu-button {
			display: inline-flex;
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
	}
</style>
