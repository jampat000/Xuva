<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		children,
		segmented = false,
		ariaLabel = ''
	} = $props<{
		children?: Snippet;
		segmented?: boolean;
		ariaLabel?: string;
	}>();
</script>

<div class="filter-group" class:filter-group--segmented={segmented} role="group" aria-label={ariaLabel}>
	{@render children?.()}
</div>

<style>
	.filter-group {
		display: inline-flex;
		flex-wrap: wrap;
		gap: 6px;
		padding: 4px;
		border: 1px solid rgb(255 246 229 / 12%);
		border-radius: 12px;
		background: rgb(255 246 229 / 3%);
	}

	.filter-group--segmented {
		border-color: rgb(255 246 229 / 15%);
	}

	.filter-group :global(button) {
		min-height: 31px;
		padding: 0 11px;
		border: 1px solid transparent;
		border-radius: 9px;
		background: transparent;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 82%, transparent);
		font-size: 0.79rem;
		font-weight: 640;
		letter-spacing: 0.008em;
	}

	.filter-group :global(button.selected) {
		border-color: color-mix(in srgb, var(--vyrden-color-accent-teal) 36%, transparent);
		background:
			linear-gradient(180deg, rgb(255 246 229 / 7%), rgb(255 246 229 / 2%)),
			color-mix(in srgb, var(--vyrden-color-accent-teal) 14%, transparent);
		color: color-mix(in srgb, var(--vyrden-color-text) 95%, transparent);
	}
</style>
