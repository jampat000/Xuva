<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		brand,
		primary,
		secondary,
		libraries,
		profile
	} = $props<{
		brand?: Snippet;
		primary?: Snippet;
		secondary?: Snippet;
		libraries?: Snippet;
		profile?: Snippet;
	}>();
</script>

<div class="l-sidebar">
	<div class="l-sidebar__brand">
		{@render brand?.()}
	</div>

	<nav class="l-sidebar__nav" aria-label="Primary navigation">
		{@render primary?.()}
	</nav>

	{#if libraries}
		<section class="l-sidebar__libraries" aria-label="Libraries">
			{@render libraries()}
		</section>
	{/if}

	<div class="l-sidebar__spacer"></div>

	{#if secondary}
		<nav class="l-sidebar__nav" aria-label="Secondary navigation">
			{@render secondary()}
		</nav>
	{/if}

	{#if profile}
		<div class="l-sidebar__profile">
			{@render profile()}
		</div>
	{/if}
</div>

<style>
	.l-sidebar {
		display: flex;
		flex-direction: column;
		height: 100%;
		padding: 16px 14px 16px;
		gap: 11px;
	}

	.l-sidebar__brand {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 48px;
		margin: 10px 10px 18px;
	}

	.l-sidebar__nav,
	.l-sidebar__libraries {
		display: grid;
		gap: 5px;
	}

	.l-sidebar__libraries {
		padding-top: 12px;
		border-top: 1px solid var(--lorivo-color-border-soft);
	}

	.l-sidebar__spacer {
		flex: 1;
	}

	.l-sidebar__profile {
		padding-top: 10px;
		border-top: 1px solid var(--lorivo-color-border-soft);
	}

	@media (max-width: 900px) {
		.l-sidebar {
			display: grid;
			grid-template-columns: minmax(0, 1fr);
			padding: 10px 10px 12px;
			gap: 8px;
		}

		.l-sidebar__brand {
			min-height: 44px;
			margin: 6px 8px 10px;
		}

		.l-sidebar__nav {
			display: flex;
			gap: 7px;
			overflow-x: auto;
			padding-bottom: 2px;
		}

		.l-sidebar__nav :global(.sidebar-item) {
			flex: 0 0 auto;
		}

		.l-sidebar__libraries,
		.l-sidebar__profile {
			display: none;
		}
	}
</style>
