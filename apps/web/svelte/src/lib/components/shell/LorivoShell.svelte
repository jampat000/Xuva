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

<div class="l-shell" data-density={density}>
	{#if sidebar}
		<aside class="l-shell__sidebar">
			{@render sidebar()}
		</aside>
	{/if}

	<div class="l-shell__main">
		<header class="l-shell__topbar">
			{@render topbar?.()}
		</header>

		<div class="l-shell__content" class:l-shell__content--with-companion={Boolean(companion)}>
			<main class="l-shell__primary">
				{@render children?.()}
			</main>

			{#if companion}
				<aside class="l-shell__companion">
					{@render companion()}
				</aside>
			{/if}
		</div>
	</div>
</div>

<style>
	.l-shell {
		display: grid;
		grid-template-columns: 222px minmax(0, 1fr);
		min-height: 100dvh;
		background:
			linear-gradient(180deg, rgb(255 246 229 / 2%), transparent 22%),
			var(--lorivo-color-bg-shell);
	}

	.l-shell__sidebar {
		border-right: 1px solid var(--lorivo-color-border-soft);
		background:
			radial-gradient(circle at 14% -18%, rgb(88 201 176 / 12%) 0%, rgb(88 201 176 / 0%) 36%),
			radial-gradient(circle at 80% 108%, rgb(131 119 93 / 12%) 0%, rgb(131 119 93 / 0%) 42%),
			var(--lorivo-color-bg-sidebar);
	}

	.l-shell__main {
		display: flex;
		flex-direction: column;
		min-width: 0;
		padding: 18px 28px 24px 42px;
	}

	.l-shell__topbar {
		display: flex;
		align-items: center;
		min-height: 0;
		margin-bottom: 0;
		padding-bottom: 12px;
	}

	.l-shell__content {
		display: grid;
		grid-template-columns: minmax(0, 1fr);
		gap: 20px;
		min-height: 0;
	}

	.l-shell__content--with-companion {
		grid-template-columns: minmax(0, 1fr) 284px;
	}

	.l-shell__primary {
		min-width: 0;
	}

	.l-shell__companion {
		min-width: 0;
	}

	@media (max-width: 900px) {
		.l-shell__content--with-companion {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 980px) {
		.l-shell {
			grid-template-columns: 1fr;
		}

		.l-shell__sidebar {
			border-right: 0;
			border-bottom: 1px solid var(--lorivo-color-border-soft);
		}

		.l-shell__main {
			padding: 14px 14px 20px;
		}

		.l-shell__topbar {
			padding-bottom: 8px;
		}
	}
</style>
