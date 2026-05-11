<script lang="ts">
	import type { Snippet } from 'svelte';
	import VyrdenBrand from './VyrdenBrand.svelte';
	import VyrdenSearch from '../ui/VyrdenSearch.svelte';

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
		id: ActiveRoute | 'admin' | 'settings';
		label: string;
		href: string;
	}

	const mediaNavItems: NavItem[] = [
		{ id: 'home', label: 'Home', href: '/' },
		{ id: 'movies', label: 'Movies', href: '/movies' },
		{ id: 'tv', label: 'TV Shows', href: '/tv' },
		{ id: 'continue-watching', label: 'Continue Watching', href: '/continue-watching' },
		{ id: 'recently-added', label: 'Recently Added', href: '/recently-added' },
		{ id: 'watchlist', label: 'Watchlist', href: '/watchlist' },
		{ id: 'collections', label: 'Collections', href: '/collections' }
	];

	const manageNavItems: NavItem[] = [
		{ id: 'admin', label: 'Manage Server', href: '/admin' },
		{ id: 'settings', label: 'Server Settings', href: '/settings' }
	];

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
</script>

<div class="media-shell" data-shell="media">
	<header class="media-shell__topbar">
		<div class="media-shell__topbar-left">
			<button
				class="menu-button"
				type="button"
				aria-label={menuOpen ? 'Close navigation menu' : 'Open navigation menu'}
				aria-expanded={menuOpen}
				onclick={() => (menuOpen = !menuOpen)}
			>
				<span></span>
				<span></span>
				<span></span>
			</button>
			<a class="brand-link" href="/" aria-label="Go to Home">
				<VyrdenBrand />
			</a>
		</div>
		<div class="media-shell__topbar-search">
			<VyrdenSearch bind:value={searchValue} />
		</div>
		<div class="media-shell__topbar-actions">
			<button class="profile-button" type="button" aria-label="Open profile menu" onclick={openProfileMenu}>
				<span class="profile-button__avatar">{userInitials}</span>
			</button>
		</div>
	</header>

	{#if menuOpen}
		<button
			type="button"
			class="media-shell__backdrop"
			aria-label="Close navigation menu"
			onclick={closeMenu}
		></button>
	{/if}

	<aside class="media-shell__drawer" class:media-shell__drawer--open={menuOpen}>
		<div class="media-shell__drawer-brand">
			<VyrdenBrand />
		</div>
		<nav class="media-shell__drawer-nav" aria-label="Media navigation">
			{#each mediaNavItems as item (item.id)}
				<a href={item.href} class:active={active === item.id} onclick={closeMenu}>{item.label}</a>
			{/each}
		</nav>
		<nav class="media-shell__drawer-manage" aria-label="Server management">
			{#each manageNavItems as item (item.id)}
				<a href={item.href} onclick={closeMenu}>{item.label}</a>
			{/each}
		</nav>
	</aside>

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

<style>
	.media-shell {
		min-height: 100dvh;
		padding: 16px 24px 32px;
		background:
			radial-gradient(circle at 22% -18%, rgb(124 92 255 / 16%) 0%, transparent 38%),
			radial-gradient(circle at 84% 3%, rgb(55 84 150 / 16%) 0%, transparent 32%),
			linear-gradient(180deg, rgb(255 255 255 / 3%), transparent 24%),
			var(--vyrden-color-bg-shell);
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
		background: color-mix(in srgb, var(--vyrden-color-text) 86%, transparent);
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

	.media-shell__backdrop {
		position: fixed;
		inset: 0;
		border: 0;
		background: rgb(0 0 0 / 52%);
		z-index: 30;
	}

	.media-shell__drawer {
		position: fixed;
		top: 0;
		left: 0;
		z-index: 32;
		width: min(280px, calc(100vw - 28px));
		height: 100dvh;
		padding: 16px 12px;
		display: grid;
		grid-template-rows: auto 1fr auto;
		gap: 14px;
		border-right: 1px solid var(--vyrden-color-border-soft);
		background:
			radial-gradient(circle at 14% -18%, rgb(124 92 255 / 16%) 0%, rgb(124 92 255 / 0%) 36%),
			radial-gradient(circle at 80% 108%, rgb(55 84 150 / 14%) 0%, rgb(55 84 150 / 0%) 42%),
			var(--vyrden-color-bg-sidebar);
		box-shadow: 18px 0 36px rgb(0 0 0 / 36%);
		transform: translateX(-103%);
		transition: transform 180ms ease;
	}

	.media-shell__drawer--open {
		transform: translateX(0);
	}

	.media-shell__drawer-brand :global(.v-brand) {
		justify-content: flex-start;
		padding-left: 6px;
	}

	.media-shell__drawer-nav,
	.media-shell__drawer-manage {
		display: grid;
		gap: 6px;
		align-content: start;
	}

	.media-shell__drawer a {
		display: inline-flex;
		align-items: center;
		min-height: 38px;
		padding: 0 12px;
		border-radius: 10px;
		font-size: 0.92rem;
		font-weight: 620;
		color: color-mix(in srgb, var(--vyrden-color-text) 90%, transparent);
		text-decoration: none;
		border: 1px solid transparent;
	}

	.media-shell__drawer-nav a::before {
		content: '';
		display: inline-block;
		width: 5px;
		height: 5px;
		margin-right: 8px;
		border-radius: 999px;
		background: color-mix(in srgb, var(--vyrden-color-text-muted) 60%, transparent);
	}

	.media-shell__drawer a.active {
		border-color: color-mix(in srgb, var(--vyrden-color-accent-teal) 35%, transparent);
		background: linear-gradient(90deg, rgb(124 92 255 / 18%), rgb(124 92 255 / 3%));
	}

	.media-shell__drawer a.active::before {
		background: color-mix(in srgb, var(--vyrden-color-accent-teal) 75%, white 25%);
	}

	.media-shell__drawer-manage {
		padding-top: 10px;
		border-top: 1px solid var(--vyrden-color-border-soft);
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

		.media-shell__drawer {
			width: min(300px, calc(100vw - 20px));
		}
	}
</style>
