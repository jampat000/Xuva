<script lang="ts">
	import type { Snippet } from 'svelte';

	let {
		label,
		description = '',
		children
	} = $props<{
		label: string;
		description?: string;
		children?: Snippet;
	}>();
</script>

<label class="form-row">
	<span class="form-row__label">{label}</span>
	{#if description}
		<span class="form-row__description">{description}</span>
	{/if}
	<div class="form-row__control">
		{@render children?.()}
	</div>
</label>

<style>
	.form-row {
		display: grid;
		gap: var(--lorivo-space-2);
	}

	.form-row__label {
		font-size: 0.86rem;
		font-weight: 620;
		color: var(--lorivo-color-text);
	}

	.form-row__description {
		font-size: 0.78rem;
		color: var(--lorivo-color-text-soft);
	}

	.form-row__control :global(input),
	.form-row__control :global(select),
	.form-row__control :global(textarea) {
		width: 100%;
		min-height: var(--lorivo-control-height-md);
		padding: 0 var(--lorivo-space-3);
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-sm);
		background: rgb(255 255 255 / 3%);
		color: var(--lorivo-color-text);
	}
</style>
