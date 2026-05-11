<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		description = '',
		children
	} = $props<{
		title: string;
		description?: string;
		children?: Snippet;
	}>();
</script>

<section class="settings-section">
	<header>
		<h3>{title}</h3>
		{#if description}
			<p>{description}</p>
		{/if}
	</header>
	<div class="settings-section__body">
		{@render children?.()}
	</div>
</section>

<style>
	.settings-section {
		display: grid;
		gap: var(--vyrden-space-3);
		padding: var(--vyrden-space-4);
		border: 1px solid var(--vyrden-color-border-soft);
		border-radius: var(--vyrden-radius-md);
		background: rgb(255 255 255 / 2%);
	}

	h3 {
		margin: 0;
		font-size: 1rem;
	}

	p {
		margin: var(--vyrden-space-1) 0 0;
		font-size: 0.84rem;
		color: var(--vyrden-color-text-muted);
	}

	.settings-section__body {
		display: grid;
		gap: var(--vyrden-space-3);
	}
</style>
