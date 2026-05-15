<script lang="ts">
	import { tick } from 'svelte';
	import type { Snippet } from 'svelte';
	import { X } from 'lucide-svelte';

	let {
		open = false,
		label = 'Main navigation',
		testId = 'app-menu-drawer',
		closeLabel = 'Close menu',
		showCloseButton = true,
		showPinButton = false,
		pinLabel = 'Pin sidebar',
		pinned = false,
		drawerWidth = '252px',
		showBackdrop = true,
		dismissOnInteractOutside = true,
		closeOnNavigate = true,
		onClose = () => {},
		onPinToggle = () => {},
		brand,
		main,
		bottom
	} = $props<{
		open?: boolean;
		label?: string;
		testId?: string;
		closeLabel?: string;
		showCloseButton?: boolean;
		showPinButton?: boolean;
		pinLabel?: string;
		pinned?: boolean;
		drawerWidth?: string;
		showBackdrop?: boolean;
		dismissOnInteractOutside?: boolean;
		closeOnNavigate?: boolean;
		onClose?: () => void;
		onPinToggle?: () => void;
		brand?: Snippet;
		main?: Snippet;
		bottom?: Snippet;
	}>();

	let drawerElement: HTMLElement | null = $state(null);

	$effect(() => {
		if (!open) return;

		const handleKeydown = (event: KeyboardEvent) => {
			if (event.key === 'Escape') onClose();
		};

		const handlePointerDown = (event: PointerEvent) => {
			if (!dismissOnInteractOutside) return;
			const target = event.target;
			if (!(target instanceof Element)) return;
			if (target.closest('[data-xuva-drawer], [data-xuva-menu-trigger]')) return;
			onClose();
		};

		const handleClick = (event: MouseEvent) => {
			if (!closeOnNavigate) return;
			const target = event.target;
			if (!(target instanceof Element)) return;
			if (drawerElement?.contains(target) && target.closest('a')) onClose();
		};

		window.addEventListener('keydown', handleKeydown);
		window.addEventListener('pointerdown', handlePointerDown);
		window.addEventListener('click', handleClick);
		void tick().then(() => {
			const firstLink = drawerElement?.querySelector<HTMLElement>('a, button');
			firstLink?.focus();
		});

		return () => {
			window.removeEventListener('keydown', handleKeydown);
			window.removeEventListener('pointerdown', handlePointerDown);
			window.removeEventListener('click', handleClick);
		};
	});
</script>

{#if open && showBackdrop}
	<button
		type="button"
		class="app-drawer__backdrop"
		aria-label={closeLabel}
		onclick={onClose}
	></button>
{/if}

<aside
	bind:this={drawerElement}
	class="app-drawer"
	class:app-drawer--open={open}
	style={`--xuva-drawer-width: ${drawerWidth};`}
	aria-label={label}
	aria-hidden={!open}
	data-xuva-drawer
	data-state={open ? 'open' : 'closed'}
	data-testid={testId}
>
	<div class="app-drawer__header">
		<div class="app-drawer__brand" data-testid="drawer-brand">
			{@render brand?.()}
		</div>
		{#if showCloseButton}
			<button type="button" class="app-drawer__close" aria-label={closeLabel} onclick={onClose}>
				<X size={14} />
			</button>
		{:else if showPinButton}
			<button
				type="button"
				class="app-drawer__pin"
				aria-label={pinLabel}
				aria-pressed={pinned}
				data-pinned={pinned ? 'true' : 'false'}
				onclick={onPinToggle}
			>
				<svg viewBox="0 0 24 24" aria-hidden="true">
					<path
						d="M8 4.8h8l-1.4 4.1 2.7 2.8v1.1H12.8v5.8l-1.6.8v-6.6H6.7v-1.1l2.7-2.8Z"
						fill="none"
						stroke="currentColor"
						stroke-linejoin="round"
						stroke-width="1.55"
					/>
				</svg>
			</button>
			<span class="app-drawer__header-spacer" aria-hidden="true"></span>
		{:else}
			<span class="app-drawer__header-spacer" aria-hidden="true"></span>
		{/if}
	</div>

	<nav class="app-drawer__main" aria-label={label}>
		{@render main?.()}
	</nav>

	{#if bottom}
		<nav class="app-drawer__bottom" aria-label="Drawer actions">
			{@render bottom()}
		</nav>
	{/if}
</aside>

<style>
	:global(:root) {
		--xuva-drawer-width: 320px;
		--xuva-drawer-ease: cubic-bezier(0.22, 1, 0.36, 1);
	}

	.app-drawer {
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 80;
		display: grid;
		grid-template-rows: auto 1fr auto;
		width: min(var(--xuva-drawer-width), 86vw);
		height: 100dvh;
		padding: 16px 14px 16px;
		border-right: 1px solid color-mix(in srgb, var(--xuva-color-border-soft) 86%, white 10%);
		background:
			linear-gradient(180deg, rgb(255 255 255 / 6%), rgb(255 255 255 / 1%) 26%, transparent),
			radial-gradient(circle at 20% -12%, rgb(124 92 255 / 22%) 0%, rgb(124 92 255 / 0%) 38%),
			radial-gradient(circle at 84% 112%, rgb(88 201 176 / 12%) 0%, rgb(88 201 176 / 0%) 40%),
			color-mix(in srgb, var(--xuva-color-bg-sidebar) 94%, black 6%);
		box-shadow:
			24px 0 54px rgb(0 0 0 / 34%),
			inset -1px 0 0 rgb(255 255 255 / 4%);
		transform: translateX(-102%);
		transition: transform 260ms var(--xuva-drawer-ease), box-shadow 260ms var(--xuva-drawer-ease);
		pointer-events: none;
	}

	.app-drawer--open {
		transform: translateX(0);
		pointer-events: auto;
	}

	.app-drawer__backdrop {
		position: fixed;
		inset: 0;
		z-index: 70;
		border: 0;
		background: rgb(0 0 0 / 58%);
		animation: drawer-backdrop-in 220ms var(--xuva-drawer-ease);
	}

	.app-drawer__header {
		position: relative;
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		justify-items: stretch;
		align-items: start;
		gap: 8px;
		min-height: 34px;
		padding: 0;
		margin-bottom: 6px;
	}

	.app-drawer__brand {
		display: flex;
		align-items: flex-start;
		justify-content: flex-start;
		justify-self: stretch;
		width: 100%;
		min-width: 0;
		min-height: 28px;
		margin-left: 0;
		margin-top: 0;
		opacity: 0;
		transform: none;
		transition: opacity 240ms var(--xuva-drawer-ease);
	}

	.app-drawer--open .app-drawer__brand {
		opacity: 1;
		transform: none;
	}

	.app-drawer__brand :global(.v-brand) {
		min-height: 28px;
		width: 100%;
		max-width: 100%;
		justify-content: flex-start;
		align-items: flex-start;
	}

	.app-drawer__brand :global(.xuva-wordmark) {
		width: 100%;
		justify-content: flex-start;
	}

	.app-drawer__header-spacer {
		display: block;
		height: 0;
		line-height: 0;
		font-size: 0;
	}

	.app-drawer__close {
		position: absolute;
		top: 1px;
		right: 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 10px;
		border: 1px solid rgb(255 255 255 / 14%);
		background: linear-gradient(180deg, rgb(255 255 255 / 8%), rgb(255 255 255 / 3%));
		color: color-mix(in srgb, var(--xuva-color-text) 88%, transparent);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 8%);
	}

	.app-drawer__pin {
		position: absolute;
		top: 1px;
		right: 0;
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 32px;
		height: 32px;
		border-radius: 10px;
		border: 1px solid rgb(255 255 255 / 14%);
		background: linear-gradient(180deg, rgb(255 255 255 / 8%), rgb(255 255 255 / 3%));
		color: color-mix(in srgb, var(--xuva-color-text) 88%, transparent);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 8%);
	}

	.app-drawer__pin svg {
		width: 14px;
		height: 14px;
	}

	.app-drawer__close:hover,
	.app-drawer__close:focus-visible,
	.app-drawer__pin:hover,
	.app-drawer__pin:focus-visible {
		border-color: rgb(255 255 255 / 28%);
		background: rgb(255 255 255 / 9%);
		color: var(--xuva-color-text);
		outline: none;
	}

	.app-drawer__pin[data-pinned='true'] {
		border-color: rgb(124 92 255 / 52%);
		background:
			linear-gradient(180deg, rgb(124 92 255 / 30%), rgb(124 92 255 / 14%)),
			rgb(255 255 255 / 8%);
		color: color-mix(in srgb, var(--xuva-color-text) 92%, white 8%);
	}

	.app-drawer__main,
	.app-drawer__bottom {
		display: grid;
		gap: 7px;
		align-content: start;
	}

	.app-drawer__main {
		min-height: 0;
		overflow-y: auto;
		padding-right: 2px;
	}

	.app-drawer__bottom {
		padding-top: 14px;
		border-top: 1px solid var(--xuva-color-border-soft);
	}

	:global(.app-drawer__link) {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 44px;
		padding: 0 13px;
		border: 1px solid transparent;
		border-radius: 12px;
		color: color-mix(in srgb, var(--xuva-color-text) 78%, transparent);
		text-decoration: none;
		font-size: 0.96rem;
		font-weight: 650;
		transition:
			transform 180ms var(--xuva-drawer-ease),
			background-color 180ms ease,
			border-color 180ms ease,
			color 180ms ease,
			box-shadow 180ms ease;
	}

	:global(.app-drawer__link:hover),
	:global(.app-drawer__link:focus-visible) {
		color: var(--xuva-color-text);
		border-color: rgb(255 255 255 / 15%);
		background: rgb(255 255 255 / 6%);
		outline: none;
	}

	:global(.app-drawer__link[aria-current='page']) {
		color: var(--xuva-color-text);
		border-color: rgb(124 92 255 / 42%);
		background:
			linear-gradient(90deg, rgb(124 92 255 / 22%), rgb(124 92 255 / 5%)),
			rgb(255 255 255 / 3%);
		box-shadow:
			inset 3px 0 0 rgb(124 92 255 / 78%),
			0 12px 30px rgb(124 92 255 / 10%);
	}

	:global(.app-drawer__link--back) {
		min-height: 48px;
		border-color: rgb(255 255 255 / 14%);
		background:
			linear-gradient(90deg, rgb(255 255 255 / 8%), rgb(255 255 255 / 3%)),
			rgb(154 167 255 / 4%);
		color: color-mix(in srgb, var(--xuva-color-text) 88%, transparent);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 7%);
	}

	:global(.app-drawer__link--back:hover),
	:global(.app-drawer__link--back:focus-visible) {
		border-color: rgb(154 167 255 / 32%);
		background:
			linear-gradient(90deg, rgb(154 167 255 / 13%), rgb(255 255 255 / 4%)),
			rgb(255 255 255 / 4%);
	}

	:global(.app-drawer__link svg) {
		width: 19px;
		height: 19px;
		flex: 0 0 auto;
		color: color-mix(in srgb, currentColor 90%, var(--xuva-color-accent-teal) 10%);
	}

	@media (max-width: 767px) {
		.app-drawer {
			width: min(340px, 86vw);
			box-shadow: 28px 0 70px rgb(0 0 0 / 54%);
		}
	}

	@media (prefers-reduced-motion: reduce) {
		.app-drawer,
		.app-drawer__backdrop,
		.app-drawer__brand,
		:global(.app-drawer__link) {
			animation: none;
			transition: none;
		}
	}

	@keyframes drawer-backdrop-in {
		from {
			opacity: 0;
		}
		to {
			opacity: 1;
		}
	}
</style>
