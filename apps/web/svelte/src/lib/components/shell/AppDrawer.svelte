<script lang="ts">
	import { tick } from 'svelte';
	import type { Snippet } from 'svelte';
	import { X } from 'lucide-svelte';

	let {
		open = false,
		label = 'Main navigation',
		testId = 'app-menu-drawer',
		closeLabel = 'Close menu',
		onClose = () => {},
		brand,
		main,
		bottom
	} = $props<{
		open?: boolean;
		label?: string;
		testId?: string;
		closeLabel?: string;
		onClose?: () => void;
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
			const target = event.target;
			if (!(target instanceof Element)) return;
			if (target.closest('[data-lorivo-drawer], [data-lorivo-menu-trigger]')) return;
			onClose();
		};

		const handleClick = (event: MouseEvent) => {
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

{#if open}
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
	aria-label={label}
	aria-hidden={!open}
	data-lorivo-drawer
	data-state={open ? 'open' : 'closed'}
	data-testid={testId}
>
	<div class="app-drawer__header">
		<div class="app-drawer__brand" data-testid="drawer-brand">
			{@render brand?.()}
		</div>
		<button type="button" class="app-drawer__close" aria-label={closeLabel} onclick={onClose}>
			<X size={18} />
		</button>
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
		--lorivo-drawer-width: 320px;
		--lorivo-drawer-ease: cubic-bezier(0.22, 1, 0.36, 1);
	}

	.app-drawer {
		position: fixed;
		inset: 0 auto 0 0;
		z-index: 80;
		display: grid;
		grid-template-rows: auto 1fr auto;
		width: min(var(--lorivo-drawer-width), 86vw);
		height: 100dvh;
		padding: 18px 14px 16px;
		border-right: 1px solid color-mix(in srgb, var(--lorivo-color-border-soft) 86%, white 10%);
		background:
			linear-gradient(180deg, rgb(255 255 255 / 6%), rgb(255 255 255 / 1%) 26%, transparent),
			radial-gradient(circle at 20% -12%, rgb(124 92 255 / 22%) 0%, rgb(124 92 255 / 0%) 38%),
			radial-gradient(circle at 84% 112%, rgb(88 201 176 / 12%) 0%, rgb(88 201 176 / 0%) 40%),
			color-mix(in srgb, var(--lorivo-color-bg-sidebar) 94%, black 6%);
		box-shadow:
			24px 0 54px rgb(0 0 0 / 34%),
			inset -1px 0 0 rgb(255 255 255 / 4%);
		transform: translateX(-102%);
		transition: transform 260ms var(--lorivo-drawer-ease), box-shadow 260ms var(--lorivo-drawer-ease);
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
		animation: drawer-backdrop-in 220ms var(--lorivo-drawer-ease);
	}

	.app-drawer__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 12px;
		min-height: 48px;
		padding: 0 4px 16px 6px;
	}

	.app-drawer__brand {
		display: flex;
		align-items: center;
		min-width: 0;
		opacity: 0;
		transform: translateX(-8px);
		transition:
			opacity 240ms var(--lorivo-drawer-ease),
			transform 240ms var(--lorivo-drawer-ease);
	}

	.app-drawer--open .app-drawer__brand {
		opacity: 1;
		transform: translateX(0);
	}

	.app-drawer__brand :global(.v-brand) {
		min-height: 40px;
		justify-content: flex-start;
	}

	.app-drawer__brand :global(.v-brand__wordmark) {
		font-size: 1.06rem;
	}

	.app-drawer__close {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 40px;
		height: 40px;
		border-radius: 12px;
		border: 1px solid rgb(255 255 255 / 14%);
		background: linear-gradient(180deg, rgb(255 255 255 / 8%), rgb(255 255 255 / 3%));
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		box-shadow: inset 0 1px 0 rgb(255 255 255 / 8%);
	}

	.app-drawer__close:hover,
	.app-drawer__close:focus-visible {
		border-color: rgb(255 255 255 / 28%);
		background: rgb(255 255 255 / 9%);
		color: var(--lorivo-color-text);
		outline: none;
	}

	.app-drawer__main,
	.app-drawer__bottom {
		display: grid;
		gap: 7px;
		align-content: start;
	}

	.app-drawer__bottom {
		padding-top: 14px;
		border-top: 1px solid var(--lorivo-color-border-soft);
	}

	:global(.app-drawer__link) {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 44px;
		padding: 0 13px;
		border: 1px solid transparent;
		border-radius: 12px;
		color: color-mix(in srgb, var(--lorivo-color-text) 78%, transparent);
		text-decoration: none;
		font-size: 0.96rem;
		font-weight: 650;
		transition:
			transform 180ms var(--lorivo-drawer-ease),
			background-color 180ms ease,
			border-color 180ms ease,
			color 180ms ease,
			box-shadow 180ms ease;
	}

	:global(.app-drawer__link:hover),
	:global(.app-drawer__link:focus-visible) {
		color: var(--lorivo-color-text);
		border-color: rgb(255 255 255 / 15%);
		background: rgb(255 255 255 / 6%);
		outline: none;
	}

	:global(.app-drawer__link[aria-current='page']) {
		color: var(--lorivo-color-text);
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
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
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
		color: color-mix(in srgb, currentColor 90%, var(--lorivo-color-accent-teal) 10%);
	}

	@media (min-width: 768px) {
		.app-drawer__backdrop {
			display: none;
		}
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
