<script lang="ts">
	import { onMount } from 'svelte';
	import type { Snippet } from 'svelte';
	import { Film, Folder, Home, Settings, Tv } from 'lucide-svelte';
	import AppDrawer from './AppDrawer.svelte';
	import XuvaBrand from './XuvaBrand.svelte';
	import TopBar from '\$lib/Xuva/TopBar.svelte';

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
		{ id: 'setup', label: 'Libraries', href: '/settings#libraries' }
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
	<AppDrawer
		open={menuOpen}
		label="Main navigation"
		testId="media-menu-drawer"
		drawerWidth="252px"
		onClose={closeMenu}
	>
		{#snippet brand()}
			<XuvaBrand />
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
		<TopBar
			{menuOpen}
			onMenuToggle={() => (menuOpen = !menuOpen)}
			onMenuClose={closeMenu}
			showSettingsShortcut={false}
			bind:searchValue
			avatarInitialsOverride={userInitials}
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
</style>
