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

<div class="v-empty">
	<div class="v-empty__icon" aria-hidden="true">
		{@render icon?.()}
	</div>
	<h3>{title}</h3>
	<p>{message}</p>
	{#if action}
		<div class="v-empty__action">{@render action()}</div>
	{/if}
</div>

<style>
	.v-empty {
		display: grid;
		justify-items: start;
		gap: var(--xuva-space-2);
		padding: 0.9rem 0;
		border-top: 1px solid rgb(255 255 255 / 10%);
	}

	.v-empty__icon {
		display: inline-flex;
		width: 1.5rem;
		height: 1.5rem;
		color: var(--xuva-color-text-soft);
		padding: 0.15rem;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 0.35rem;
		background: rgb(255 255 255 / 2%);
	}

	.v-empty__icon :global(svg) {
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
		color: var(--xuva-color-text-muted);
		font-size: 0.9rem;
		line-height: 1.42;
	}
</style>
