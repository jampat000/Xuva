<script lang="ts">
	import { onMount } from 'svelte';
	import { Menu, Search, Settings } from 'lucide-svelte';
	import Logo from '\$lib/Xuva/Logo.svelte';
	import ProfileMenu from '$lib/components/shell/ProfileMenu.svelte';
	import { getAuthSession, logout } from '$lib/api/auth';
	import { ApiClientError } from '$lib/api/client';

	let {
		menuOpen = false,
		onMenuToggle = () => {},
		onMenuClose = () => {},
		showSettingsShortcut = true,
		searchValue = $bindable(''),
		searchPlaceholder = 'Search',
		avatarInitialsOverride = '',
		avatarNameOverride = ''
	} = $props<{
		menuOpen?: boolean;
		onMenuToggle?: () => void;
		onMenuClose?: () => void;
		showSettingsShortcut?: boolean;
		searchValue?: string;
		searchPlaceholder?: string;
		avatarInitialsOverride?: string;
		avatarNameOverride?: string;
	}>();

	let avatarInitials = $state('U');
	let avatarName = $state('');
	let avatarRole = $state('User');
	let canSignOut = $state(false);

	onMount(() => {
		if (avatarInitialsOverride || avatarNameOverride) return;
		void loadProfile();
	});

	$effect(() => {
		if (avatarNameOverride) avatarName = avatarNameOverride;
		if (avatarInitialsOverride) avatarInitials = avatarInitialsOverride;
	});

	async function loadProfile(): Promise<void> {
		try {
			const session = await getAuthSession().catch((error: unknown) => {
				if (isApiStatus(error, 401)) return null;
				throw error;
			});
			const userName = asText(session?.user?.displayName) || asText(session?.user?.username);
			if (userName) {
				avatarName = userName;
				avatarInitials = initialsForName(userName);
				avatarRole = roleLabel(session?.user?.role);
				canSignOut = true;
				return;
			}
			if (session?.authDisabled) {
				avatarName = 'Local access';
				avatarInitials = 'L';
				avatarRole = 'No sign-in mode';
				canSignOut = false;
				return;
			}
			avatarName = 'Signed out';
			avatarInitials = 'SO';
			avatarRole = 'Sign in required';
			canSignOut = false;
		} catch {
			// Keep the media top bar usable even if session lookup fails.
		}
	}

	function initialsForName(name: string): string {
		const parts = name
			.split(/\s+/)
			.filter(Boolean)
			.slice(0, 2)
			.map((part) => part[0]?.toUpperCase() || '');
		return parts.join('') || 'U';
	}

	function isApiStatus(error: unknown, expectedStatus: number): boolean {
		if (error instanceof ApiClientError) return error.status === expectedStatus;
		if (typeof error !== 'object' || !error) return false;
		return Number((error as { status?: unknown }).status) === expectedStatus;
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function roleLabel(role: unknown): string {
		const normalized = asText(role).toLowerCase();
		if (normalized === 'admin') return 'Admin';
		if (normalized === 'standard') return 'User';
		return normalized ? normalized : 'User';
	}

	async function signOut(): Promise<void> {
		try {
			await logout();
		} finally {
			if (typeof window !== 'undefined') window.location.href = '/signin';
		}
	}
</script>

<header class="relative z-30 flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
	<div class="topbar-brand-rail min-w-0">
		{#if !menuOpen}
			<button
				type="button"
				data-testid="media-menu-button"
				data-xuva-menu-trigger
				class="topbar-menu-button inline-flex h-10 w-10 items-center justify-center rounded-lg border border-white/8 bg-[#111827]/72 text-white/85 transition hover:border-white/20 hover:bg-white/8"
				aria-label="Open menu"
				aria-expanded={menuOpen}
				onclick={onMenuToggle}
			>
				<Menu size={20} />
			</button>
		{:else}
			<span class="topbar-menu-spacer" aria-hidden="true"></span>
		{/if}
		<a
			href="/"
			class="topbar-brand"
			class:topbar-brand--drawer-open={menuOpen}
			aria-label="Go to Home"
			aria-hidden={menuOpen}
			tabindex={menuOpen ? -1 : undefined}
			data-testid="topbar-brand"
			onclick={onMenuClose}
		>
			<Logo />
		</a>
		<span class="topbar-brand-spacer" aria-hidden="true"></span>
	</div>
	<div class="flex max-w-2xl flex-1 justify-center">
		<div class="relative w-full max-w-[640px]">
			<Search size={16} class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-white/40" />
			<input
				type="text"
				placeholder={searchPlaceholder}
				bind:value={searchValue}
				class="h-10 w-full rounded-lg border border-white/6 bg-[#111827]/88 pl-11 pr-4 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-1 focus:ring-[#7C5CFF]/50"
			/>
		</div>
	</div>
	<div class="flex shrink-0 items-center gap-2">
		{#if showSettingsShortcut}
			<a
				href="/settings"
				class="hidden h-10 w-10 items-center justify-center rounded-lg border border-white/8 bg-[#111827]/72 text-white/72 transition hover:border-white/20 hover:bg-white/8 hover:text-white sm:inline-flex"
				aria-label="Open Settings"
			>
				<Settings size={18} />
			</a>
		{/if}
		<ProfileMenu
			initials={avatarInitials}
			name={avatarName || 'User'}
			role={avatarRole}
			canSignOut={canSignOut}
			changePasswordHref="/settings#admin-access"
			onSignOut={signOut}
		/>
	</div>
</header>

<style>
	.topbar-brand-rail {
		display: grid;
		grid-template-columns: 40px minmax(118px, auto) 40px;
		align-items: center;
		column-gap: 8px;
		margin-left: 14px;
	}

	.topbar-brand {
		display: inline-flex;
		align-items: center;
		width: 118px;
		min-width: 0;
		overflow: hidden;
		text-decoration: none;
		opacity: 1;
		transform: none;
		justify-self: center;
	}

	.topbar-brand--drawer-open {
		opacity: 0;
		pointer-events: none;
		transform: none;
	}

	.topbar-brand :global(.v-brand) {
		min-height: 34px;
		justify-content: flex-start;
	}

	.topbar-brand-spacer {
		display: inline-flex;
		width: 40px;
		height: 40px;
	}

	.topbar-menu-spacer {
		display: inline-flex;
		width: 40px;
		height: 40px;
	}
</style>
