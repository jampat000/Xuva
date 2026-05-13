<script lang="ts">
	import type { Snippet } from 'svelte';

	type Tone = 'default' | 'elevated' | 'soft';

	let { tone = 'default', padded = true, children } = $props<{
		tone?: Tone;
		padded?: boolean;
		children?: Snippet;
	}>();
</script>

<div class="v-surface" data-tone={tone} data-padded={padded}>
	{@render children?.()}
</div>

<style>
	.v-surface {
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: 8px;
	}

	.v-surface[data-tone='default'] {
		background: color-mix(in srgb, var(--lorivo-color-bg-panel) 96%, #101827 4%);
	}

	.v-surface[data-tone='elevated'] {
		background: color-mix(in srgb, var(--lorivo-color-bg-panel-elevated) 97%, #101827 3%);
	}

	.v-surface[data-tone='soft'] {
		background: rgb(255 255 255 / 2.5%);
	}

	.v-surface[data-padded='true'] {
		padding: 0.85rem 0.95rem;
	}
</style>
