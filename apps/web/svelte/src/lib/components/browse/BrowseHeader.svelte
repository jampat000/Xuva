<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		title,
		subtitle,
		chips
	} = $props<{
		title: string;
		subtitle: string;
		chips?: Snippet;
	}>();
</script>

<section class="browse-head">
	<div>
		<h1>{title}</h1>
		<p>{subtitle}</p>
	</div>
	{#if chips}
		<div class="browse-head__chips" role="list">
			{@render chips()}
		</div>
	{/if}
</section>

<style>
	.browse-head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 20px;
		padding-top: 6px;
	}

	.browse-head h1 {
		margin: 0;
		font-family: var(--xuva-font-display);
		font-size: clamp(2rem, 2.2vw + 1rem, 3.6rem);
		letter-spacing: 0;
		line-height: 1.02;
		text-shadow: 0 10px 28px rgb(0 0 0 / 48%);
	}

	.browse-head p {
		margin: 8px 0 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 86%, transparent);
		font-size: 1rem;
		line-height: 1.4;
	}

	.browse-head__chips {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 8px;
	}

	@media (max-width: 760px) {
		.browse-head {
			flex-direction: column;
			align-items: flex-start;
		}

		.browse-head__chips {
			justify-content: flex-start;
		}
	}
</style>
