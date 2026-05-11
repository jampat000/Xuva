<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		linkLabel,
		linkHref = '#',
		children
	} = $props<{
		title: string;
		linkLabel?: string;
		linkHref?: string;
		children?: Snippet;
	}>();
</script>

<section class="activity-shell">
	<header>
		<h3>{title}</h3>
		{#if linkLabel}
			<a href={linkHref}>{linkLabel}</a>
		{/if}
	</header>
	<div class="activity-shell__body">
		{@render children?.()}
	</div>
</section>

<style>
	.activity-shell {
		display: grid;
		gap: var(--vyrden-space-3);
	}

	header {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: var(--vyrden-space-2);
	}

	h3 {
		margin: 0;
		font-size: 0.98rem;
		letter-spacing: -0.01em;
	}

	a {
		font-size: 0.82rem;
		color: var(--vyrden-color-text-muted);
	}

	.activity-shell__body {
		display: grid;
		gap: var(--vyrden-space-2);
	}
</style>
