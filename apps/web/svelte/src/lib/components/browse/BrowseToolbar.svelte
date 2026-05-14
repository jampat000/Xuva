<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		controls,
		actions,
		message = ''
	} = $props<{
		controls?: Snippet;
		actions?: Snippet;
		message?: string;
	}>();
</script>

<div class="browse-toolbar">
	<div class="browse-toolbar__controls" data-empty={!controls}>
		{#if controls}
			{@render controls()}
		{/if}
	</div>
	<div class="browse-toolbar__actions" data-empty={!actions}>
		{#if actions}
			{@render actions()}
		{/if}
	</div>
</div>
{#if message}
	<p class="toolbar-message">{message}</p>
{/if}

<style>
	.browse-toolbar {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		align-items: center;
		column-gap: 14px;
		row-gap: 10px;
		padding: 2px 0 0;
		min-height: 44px;
	}

	.browse-toolbar__controls {
		display: flex;
		flex-wrap: wrap;
		align-items: center;
		gap: 10px;
		min-height: 34px;
		min-width: 0;
	}

	.browse-toolbar__actions {
		display: flex;
		flex-wrap: wrap;
		justify-content: flex-end;
		gap: 9px;
		min-height: 34px;
		align-items: center;
	}

	.browse-toolbar__controls[data-empty='true']::before {
		content: '';
		display: block;
		width: 100%;
		height: 34px;
	}

	.browse-toolbar__actions[data-empty='true'] {
		display: none;
	}

	.toolbar-message {
		margin: 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 84%, transparent);
		font-size: 0.82rem;
	}

	@media (max-width: 760px) {
		.browse-toolbar {
			grid-template-columns: 1fr;
		}

		.browse-toolbar__controls {
			min-height: 0;
		}

		.browse-toolbar__controls[data-empty='true']::before {
			height: 0;
		}

		.browse-toolbar__actions {
			justify-content: flex-start;
			min-height: 0;
		}
	}
</style>
