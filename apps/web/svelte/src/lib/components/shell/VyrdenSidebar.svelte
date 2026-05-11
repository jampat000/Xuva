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

<div class="v-sidebar">
	<div class="v-sidebar__brand">
		{@render brand?.()}
	</div>

	<nav class="v-sidebar__nav" aria-label="Primary navigation">
		{@render primary?.()}
	</nav>

	{#if libraries}
		<section class="v-sidebar__libraries" aria-label="Libraries">
			{@render libraries()}
		</section>
	{/if}

	<div class="v-sidebar__spacer"></div>

	{#if secondary}
		<nav class="v-sidebar__nav" aria-label="Secondary navigation">
			{@render secondary()}
		</nav>
	{/if}

	{#if profile}
		<div class="v-sidebar__profile">
			{@render profile()}
		</div>
	{/if}
</div>

<style>
	.v-sidebar {
		display: flex;
		flex-direction: column;
		height: 100%;
		padding: 16px 14px 16px;
		gap: 11px;
	}

	.v-sidebar__brand {
		display: flex;
		align-items: center;
		justify-content: center;
		min-height: 48px;
		margin: 10px 10px 18px;
	}

	.v-sidebar__nav,
	.v-sidebar__libraries {
		display: grid;
		gap: 5px;
	}

	.v-sidebar__libraries {
		padding-top: 12px;
		border-top: 1px solid var(--vyrden-color-border-soft);
	}

	.v-sidebar__spacer {
		flex: 1;
	}

	.v-sidebar__profile {
		padding-top: 10px;
		border-top: 1px solid var(--vyrden-color-border-soft);
	}

	@media (max-width: 900px) {
		.v-sidebar {
			display: grid;
			grid-template-columns: minmax(0, 1fr);
			padding: 10px 10px 12px;
			gap: 8px;
		}

		.v-sidebar__brand {
			min-height: 44px;
			margin: 6px 8px 10px;
		}

		.v-sidebar__nav {
			display: flex;
			gap: 7px;
			overflow-x: auto;
			padding-bottom: 2px;
		}

		.v-sidebar__nav :global(.sidebar-item) {
			flex: 0 0 auto;
		}

		.v-sidebar__libraries,
		.v-sidebar__profile {
			display: none;
		}
	}
</style>
