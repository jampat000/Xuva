<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		active = false,
		href = '#',
		icon,
		trailing
	} = $props<{
		label: string;
		active?: boolean;
		href?: string;
		icon?: Snippet;
		trailing?: Snippet;
	}>();
</script>

<a class="sidebar-item" data-active={active} href={href} aria-current={active ? 'page' : undefined}>
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
		border-radius: 11px;
		color: color-mix(in srgb, var(--vyrden-color-text) 76%, transparent);
		text-decoration: none;
		transition:
			background-color 150ms ease,
			border-color 150ms ease,
			color 150ms ease,
			transform 150ms ease;
	}

	.sidebar-item:hover {
		color: var(--vyrden-color-text);
		background: rgb(255 249 236 / 5%);
		transform: translateY(-1px);
	}

	.sidebar-item[data-active='true'] {
		color: var(--vyrden-color-text);
		border-color: rgb(255 246 229 / 16%);
		background:
			linear-gradient(180deg, rgb(255 246 229 / 9%), rgb(255 246 229 / 4%)),
			color-mix(in srgb, var(--vyrden-color-accent-teal) 10%, transparent);
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 7%),
			0 8px 20px rgb(0 0 0 / 20%);
	}

	.sidebar-item[data-active='true'] .sidebar-item__icon {
		color: var(--vyrden-color-accent-teal);
	}

	.sidebar-item__icon {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 21px;
		height: 21px;
		color: color-mix(in srgb, var(--vyrden-color-text-soft) 86%, transparent);
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
		color: var(--vyrden-color-text-soft);
	}
</style>
