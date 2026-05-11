<script lang="ts">
	import { onMount } from 'svelte';
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
		<aside
			class="server-shell__drawer"
			aria-label="Settings navigation drawer"
			data-testid="settings-menu-drawer"
		>
			<button
				type="button"
				class="server-shell__drawer-close"
				aria-label="Close settings menu"
				onclick={closeSettingsMenu}
			>
				<X size={18} />
			</button>
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

	.server-shell__backdrop {
		position: fixed;
		inset: 0;
		z-index: 70;
		border: 0;
		background: rgb(0 0 0 / 52%);
	}

	.server-shell__drawer {
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 80;
		width: min(320px, 86vw);
		overflow: auto;
		border-right: 1px solid var(--lorivo-color-border-soft);
		background:
			radial-gradient(circle at 14% -18%, rgb(88 201 176 / 12%) 0%, rgb(88 201 176 / 0%) 36%),
			radial-gradient(circle at 80% 108%, rgb(131 119 93 / 12%) 0%, rgb(131 119 93 / 0%) 42%),
			var(--lorivo-color-bg-sidebar);
		box-shadow: 22px 0 42px rgb(0 0 0 / 42%);
	}

	.server-shell__drawer-close {
		position: absolute;
		top: 18px;
		right: 14px;
		z-index: 2;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 38px;
		height: 38px;
		border-radius: 11px;
		border: 1px solid rgb(255 255 255 / 13%);
		background: rgb(255 255 255 / 4%);
		color: color-mix(in srgb, var(--lorivo-color-text) 86%, transparent);
	}

	.server-shell__drawer-close:hover,
	.server-shell__drawer-close:focus-visible {
		border-color: rgb(255 255 255 / 24%);
		background: rgb(255 255 255 / 8%);
		color: var(--lorivo-color-text);
		outline: none;
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
</style>
