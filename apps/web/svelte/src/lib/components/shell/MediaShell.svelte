<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Folder, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from './AppDrawer.svelte';
	import LorivoBrand from './LorivoBrand.svelte';
	import LorivoSearch from '../ui/LorivoSearch.svelte';

	type ActiveRoute =
		| 'home'
		| 'movies'
		| 'tv'
		| 'collections'
		| 'watchlist'
		| 'continue-watching'
		| 'recently-added'
		| 'setup';

	interface NavItem {
		id: ActiveRoute | 'settings';
		label: string;
		href: string;
	}

	const mediaNavItems: NavItem[] = [
		{ id: 'home', label: 'Home', href: '/' },
		{ id: 'movies', label: 'Movies', href: '/movies' },
		{ id: 'tv', label: 'TV', href: '/tv' },
		{ id: 'setup', label: 'Libraries', href: '/settings#library' }
	];

	const drawerBottomItems: NavItem[] = [{ id: 'settings', label: 'Settings', href: '/settings' }];

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

	function closeMenu(): void {
		menuOpen = false;
	}

	function openProfileMenu(): void {
		menuOpen = true;
	}

	onMount(() => {
		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') {
				closeMenu();
			}
		};
		window.addEventListener('keydown', handleKeydown);
		return () => window.removeEventListener('keydown', handleKeydown);
	});
</script>

<div class="media-shell" class:media-shell--drawer-open={menuOpen} data-shell="media">
	<AppDrawer open={menuOpen} label="Main navigation" testId="media-menu-drawer" onClose={closeMenu}>
		{#snippet brand()}
			<LorivoBrand />
		{/snippet}
		{#snippet main()}
			{#each mediaNavItems as item (item.id)}
				<a href={item.href} class="app-drawer__link" aria-current={active === item.id ? 'page' : undefined}>
					{#if item.id === 'home'}
						<Home size={19} />
					{:else if item.id === 'movies'}
						<Film size={19} />
					{:else if item.id === 'tv'}
						<Tv size={19} />
					{:else}
						<Folder size={19} />
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
		{/snippet}
	</AppDrawer>

	<div class="media-shell__surface" data-testid="media-shell-surface">
		<header class="media-shell__topbar">
			<div class="media-shell__topbar-left">
				<button
					class="menu-button"
					type="button"
					data-testid="media-menu-button"
					data-lorivo-menu-trigger
					aria-label="Open menu"
					aria-expanded={menuOpen}
					onclick={() => (menuOpen = !menuOpen)}
				>
					<span></span>
					<span></span>
					<span></span>
				</button>
				<a class="brand-link" href="/" aria-label="Go to Home">
					<LorivoBrand />
				</a>
			</div>
			<div class="media-shell__topbar-search">
				<LorivoSearch bind:value={searchValue} />
			</div>
			<div class="media-shell__topbar-actions">
				<button class="profile-button" type="button" aria-label="Open profile menu" onclick={openProfileMenu}>
					<span class="profile-button__avatar">{userInitials}</span>
				</button>
			</div>
		</header>

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
		min-height: 100dvh;
		overflow-x: hidden;
		padding: 16px 24px 32px;
		background:
			radial-gradient(circle at 22% -18%, rgb(124 92 255 / 16%) 0%, transparent 38%),
			radial-gradient(circle at 84% 3%, rgb(55 84 150 / 16%) 0%, transparent 32%),
			linear-gradient(180deg, rgb(255 255 255 / 3%), transparent 24%),
			var(--lorivo-color-bg-shell);
	}

	.media-shell__surface {
		min-height: calc(100dvh - 48px);
		transform: translateX(0);
		transition: transform 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1));
		will-change: transform;
	}

	@media (min-width: 768px) {
		.media-shell--drawer-open .media-shell__surface {
			transform: translateX(var(--lorivo-drawer-width, 320px));
		}
	}

	.media-shell__topbar {
		position: sticky;
		top: 0;
		z-index: 24;
		display: grid;
		grid-template-columns: auto minmax(0, 1fr) auto;
		align-items: center;
		gap: 10px;
		padding: 0 0 22px;
		background:
			linear-gradient(180deg, rgb(11 17 32 / 96%), rgb(11 17 32 / 86%) 72%, transparent),
			transparent;
		border-bottom: 0;
	}

	.media-shell__topbar-left {
		display: flex;
		align-items: center;
		gap: 7px;
		min-width: 0;
	}

	.brand-link {
		display: inline-flex;
		align-items: center;
		text-decoration: none;
	}

	.brand-link :global(.v-brand) {
		min-height: 34px;
		justify-content: flex-start;
	}

	.menu-button {
		display: inline-flex;
		flex-direction: column;
		justify-content: center;
		gap: 3px;
		width: 32px;
		height: 32px;
		border-radius: 9px;
		border: 1px solid rgb(255 255 255 / 16%);
		background: linear-gradient(180deg, rgb(31 41 55 / 82%), rgb(17 24 39 / 76%));
	}

	.menu-button span {
		display: block;
		width: 15px;
		height: 2px;
		margin: 0 auto;
		border-radius: 999px;
		background: color-mix(in srgb, var(--lorivo-color-text) 86%, transparent);
	}

	.media-shell__topbar-search {
		min-width: 0;
	}

	.media-shell__topbar-search :global(.v-search) {
		width: min(100%, 560px);
		margin: 0 auto;
	}

	.media-shell__topbar-actions {
		display: inline-flex;
		align-items: center;
		justify-content: flex-end;
	}

	.profile-button {
		display: grid;
		place-items: center;
		width: 34px;
		height: 34px;
		border-radius: 999px;
		border: 1px solid rgb(255 255 255 / 16%);
		background: linear-gradient(180deg, rgb(31 41 55 / 84%), rgb(17 24 39 / 80%));
		padding: 0;
	}

	.profile-button__avatar {
		display: inline-grid;
		place-items: center;
		width: 26px;
		height: 26px;
		border-radius: 999px;
		font-size: 0.72rem;
		font-weight: 700;
		background: linear-gradient(180deg, rgb(124 92 255 / 36%), rgb(31 41 55 / 66%));
		color: #f4f1ea;
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

		.media-shell__topbar {
			gap: 8px;
			padding-bottom: 9px;
		}

		.media-shell__topbar-search :global(.v-search) {
			width: 100%;
		}

		.media-shell__content--with-companion {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 620px) {
		.media-shell__topbar {
			grid-template-columns: auto minmax(0, 1fr) auto;
			grid-template-areas:
				'left left actions'
				'search search search';
			row-gap: 8px;
		}

		.media-shell__topbar-left {
			grid-area: left;
			min-width: 0;
		}

		.media-shell__topbar-actions {
			grid-area: actions;
		}

		.media-shell__topbar-search {
			grid-area: search;
		}

		.brand-link :global(.v-brand__wordmark) {
			font-size: 1.08rem;
		}

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
