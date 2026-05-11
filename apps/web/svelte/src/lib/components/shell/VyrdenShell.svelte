<script lang="ts">
	import type { Snippet } from 'svelte';

	type Density = 'compact' | 'default' | 'comfortable' | 'expanded';

	let {
		density = 'default',
		sidebar,
		topbar,
		companion,
		children
	} = $props<{
		density?: Density;
		sidebar?: Snippet;
		topbar?: Snippet;
		companion?: Snippet;
		children?: Snippet;
	}>();
</script>

<div class="v-shell" data-density={density}>
	{#if sidebar}
		<aside class="v-shell__sidebar">
			{@render sidebar()}
		</aside>
	{/if}

	<div class="v-shell__main">
		<header class="v-shell__topbar">
			{@render topbar?.()}
		</header>

		<div class="v-shell__content" class:v-shell__content--with-companion={Boolean(companion)}>
			<main class="v-shell__primary">
				{@render children?.()}
			</main>

			{#if companion}
				<aside class="v-shell__companion">
					{@render companion()}
				</aside>
			{/if}
		</div>
	</div>
</div>

<style>
	.v-shell {
		display: grid;
		grid-template-columns: 222px minmax(0, 1fr);
		min-height: 100dvh;
		background:
			linear-gradient(180deg, rgb(255 246 229 / 2%), transparent 22%),
			var(--vyrden-color-bg-shell);
	}

	.v-shell__sidebar {
		border-right: 1px solid var(--vyrden-color-border-soft);
		background:
			radial-gradient(circle at 14% -18%, rgb(88 201 176 / 12%) 0%, rgb(88 201 176 / 0%) 36%),
			radial-gradient(circle at 80% 108%, rgb(131 119 93 / 12%) 0%, rgb(131 119 93 / 0%) 42%),
			var(--vyrden-color-bg-sidebar);
	}

	.v-shell__main {
		display: flex;
		flex-direction: column;
		min-width: 0;
		padding: 18px 28px 24px 42px;
	}

	.v-shell__topbar {
		display: flex;
		align-items: center;
		min-height: 0;
		margin-bottom: 0;
		padding-bottom: 12px;
	}

	.v-shell__content {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 20px;
		min-height: 0;
	}

	.v-shell__content--with-companion {
		grid-template-columns: minmax(0, 1fr) 284px;
	}

	.v-shell__primary {
		min-width: 0;
	}

	.v-shell__companion {
		min-width: 0;
	}

	@media (max-width: 900px) {
		.v-shell__content--with-companion {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 980px) {
		.v-shell {
			grid-template-columns: 1fr;
		}

		.v-shell__sidebar {
			border-right: 0;
			border-bottom: 1px solid var(--vyrden-color-border-soft);
		}

		.v-shell__main {
			padding: 14px 14px 20px;
		}

		.v-shell__topbar {
			padding-bottom: 8px;
		}
	}
</style>
