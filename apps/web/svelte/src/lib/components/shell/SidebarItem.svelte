<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		active = false,
		href = '#',
		variant = 'default',
		icon,
		trailing
	} = $props<{
		label: string;
		active?: boolean;
		href?: string;
		variant?: 'default' | 'back';
		icon?: Snippet;
		trailing?: Snippet;
	}>();
</script>

<a
	class="sidebar-item"
	class:sidebar-item--back={variant === 'back'}
	data-active={active}
	href={href}
	aria-current={active ? 'page' : undefined}
>
	<span class="sidebar-item__icon">
		{@render icon?.()}
	</span>
	<span class="sidebar-item__label">{label}</span>
	{#if trailing}
		<span class="sidebar-item__trailing">
			{@render trailing()}
		</span>
	{/if}
</a>

<style>
	.sidebar-item {
		display: flex;
		align-items: center;
		gap: 12px;
		min-height: 41px;
		padding: 0 13px;
		border: 1px solid transparent;
		border-radius: 6px;
		color: color-mix(in srgb, var(--xuva-color-text) 76%, transparent);
		text-decoration: none;
		transition:
			background-color 150ms ease,
			border-color 150ms ease,
			color 150ms ease,
			transform 150ms ease;
	}

	.sidebar-item:hover {
		color: var(--xuva-color-text);
		background: rgb(255 249 236 / 4%);
		transform: none;
	}

	.sidebar-item[data-active='true'] {
		color: var(--xuva-color-text);
		border-color: rgb(255 246 229 / 12%);
		background: color-mix(in srgb, var(--xuva-color-accent-teal) 6%, transparent);
		box-shadow: none;
	}

	.sidebar-item[data-active='true'] .sidebar-item__icon {
		color: var(--xuva-color-accent-teal);
	}

	.sidebar-item--back {
		min-height: 48px;
		border-color: rgb(255 255 255 / 9%);
		background: rgb(154 167 255 / 4%);
		color: color-mix(in srgb, var(--xuva-color-text) 88%, transparent);
	}

	.sidebar-item--back:hover,
	.sidebar-item--back:focus-visible {
		border-color: rgb(154 167 255 / 28%);
		background:
			linear-gradient(90deg, rgb(154 167 255 / 12%), rgb(255 255 255 / 4%)),
			rgb(255 255 255 / 4%);
	}

	.sidebar-item--back .sidebar-item__icon {
		color: color-mix(in srgb, var(--xuva-color-text) 82%, var(--xuva-settings-accent, #9aa7ff) 18%);
	}

	.sidebar-item__icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 21px;
		height: 21px;
		color: color-mix(in srgb, var(--xuva-color-text-soft) 86%, transparent);
	}

	.sidebar-item__icon :global(svg) {
		width: 100%;
		height: 100%;
	}

	.sidebar-item__label {
		flex: 1;
		font-size: 0.95rem;
		font-weight: 620;
		letter-spacing: 0.005em;
		line-height: 1.2;
	}

	.sidebar-item__trailing {
		font-size: 0.76rem;
		color: var(--xuva-color-text-soft);
	}
</style>
