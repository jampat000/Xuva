<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		message,
		icon,
		action
	} = $props<{
		title: string;
		message: string;
		icon?: Snippet;
		action?: Snippet;
	}>();
</script>

<div class="l-empty">
	<div class="l-empty__icon" aria-hidden="true">
		{@render icon?.()}
	</div>
	<h3>{title}</h3>
	<p>{message}</p>
	{#if action}
		<div class="l-empty__action">{@render action()}</div>
	{/if}
</div>

<style>
	.l-empty {
		display: grid;
		justify-items: start;
		gap: var(--lorivo-space-2);
		padding: var(--lorivo-space-6);
		border: 1px dashed var(--lorivo-color-border-strong);
		border-radius: var(--lorivo-radius-lg);
		background: linear-gradient(180deg, rgb(255 246 229 / 4%), rgb(255 246 229 / 1%));
	}

	.l-empty__icon {
		display: inline-flex;
		width: 1.5rem;
		height: 1.5rem;
		color: var(--lorivo-color-text-soft);
	}

	.l-empty__icon :global(svg) {
		width: 100%;
		height: 100%;
	}

	h3 {
		margin: 0;
		font-size: 1.02rem;
		letter-spacing: -0.01em;
	}

	p {
		margin: 0;
		color: var(--lorivo-color-text-muted);
		font-size: 0.9rem;
		line-height: 1.42;
	}
</style>
