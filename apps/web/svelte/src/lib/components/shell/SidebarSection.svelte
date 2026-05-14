<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		actionLabel,
		actionAriaLabel,
		children
	} = $props<{
		title: string;
		actionLabel?: string;
		actionAriaLabel?: string;
		children?: Snippet;
	}>();
</script>

<section class="sidebar-section">
	<header class="sidebar-section__header">
		<h2>{title}</h2>
		{#if actionLabel}
			<button type="button" class="sidebar-section__action" aria-label={actionAriaLabel || actionLabel}>
				{actionLabel}
			</button>
		{/if}
	</header>
	<div class="sidebar-section__body">
		{@render children?.()}
	</div>
</section>

<style>
	.sidebar-section {
		display: grid;
		gap: 6px;
	}

	.sidebar-section__header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--xuva-space-2);
	}

	h2 {
		margin: 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 86%, transparent);
		font-size: 0.74rem;
		font-weight: 750;
		letter-spacing: 0.08em;
		text-transform: uppercase;
	}

	.sidebar-section__action {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 26px;
		height: 26px;
		border-radius: 8px;
		color: rgb(224 211 186 / 72%);
		font-size: 1.3rem;
		line-height: 1;
	}

	.sidebar-section__action:hover {
		color: var(--xuva-color-text);
		background: rgb(255 246 229 / 5%);
	}

	.sidebar-section__body {
		display: grid;
		gap: var(--xuva-space-1);
	}
</style>
