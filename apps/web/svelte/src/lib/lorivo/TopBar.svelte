<script lang="ts">
	import { Menu, Search, Settings } from 'lucide-svelte';
	import Logo from '$lib/lorivo/Logo.svelte';

	let {
		menuOpen = false,
		onMenuToggle = () => {},
		onMenuClose = () => {}
	} = $props<{
		menuOpen?: boolean;
		onMenuToggle?: () => void;
		onMenuClose?: () => void;
	}>();
</script>

<header class="relative z-30 flex items-center justify-between gap-4 px-4 py-4 sm:px-6 lg:px-8">
	<div class="flex min-w-0 items-center gap-3">
		<button
			type="button"
			data-testid="media-menu-button"
			data-lorivo-menu-trigger
			class="inline-flex h-10 w-10 items-center justify-center rounded-lg border border-white/8 bg-[#111827]/72 text-white/85 transition hover:border-white/20 hover:bg-white/8"
			aria-label="Open menu"
			aria-expanded={menuOpen}
			onclick={onMenuToggle}
		>
			<Menu size={20} />
		</button>
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
	</div>
	<div class="flex max-w-2xl flex-1 justify-center">
		<div class="relative w-full max-w-[640px]">
			<Search size={16} class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-white/40" />
			<input
				type="text"
				placeholder="Search"
				class="h-10 w-full rounded-lg border border-white/6 bg-[#111827]/88 pl-11 pr-4 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-1 focus:ring-[#7C5CFF]/50"
			/>
		</div>
	</div>
	<div class="flex shrink-0 items-center gap-2">
		<a
			href="/settings"
			class="hidden h-10 w-10 items-center justify-center rounded-lg border border-white/8 bg-[#111827]/72 text-white/72 transition hover:border-white/20 hover:bg-white/8 hover:text-white sm:inline-flex"
			aria-label="Open Settings"
		>
			<Settings size={18} />
		</a>
		<div
			aria-label="Profile"
			class="grid h-9 w-9 place-items-center rounded-full border border-white/10 bg-[#22302c] text-xs font-semibold text-white/80"
		>
			L
		</div>
	</div>
</header>

<style>
	.topbar-brand {
		display: inline-flex;
		align-items: center;
		max-width: 190px;
		min-width: 0;
		overflow: hidden;
		text-decoration: none;
		opacity: 1;
		transform: translateX(0);
		transition:
			opacity 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1)),
			transform 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1)),
			max-width 260ms var(--lorivo-drawer-ease, cubic-bezier(0.22, 1, 0.36, 1));
		will-change: opacity, transform, max-width;
	}

	.topbar-brand--drawer-open {
		max-width: 0;
		opacity: 0;
		pointer-events: none;
		transform: translateX(-10px);
	}

	.topbar-brand :global(.v-brand) {
		min-height: 34px;
		justify-content: flex-start;
	}

	.topbar-brand :global(.v-brand__wordmark) {
		font-size: 1.04rem;
	}

	@media (prefers-reduced-motion: reduce) {
		.topbar-brand {
			transition: none;
		}
	}
</style>
