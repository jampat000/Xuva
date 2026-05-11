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
		border-radius: var(--lorivo-radius-lg);
		box-shadow: var(--lorivo-shadow-soft);
	}

	.v-surface[data-tone='default'] {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), transparent 22%),
			var(--lorivo-color-bg-panel);
	}

	.v-surface[data-tone='elevated'] {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 6%), transparent 24%),
			var(--lorivo-color-bg-panel-elevated);
	}

	.v-surface[data-tone='soft'] {
		background: rgb(255 255 255 / 5%);
	}

	.v-surface[data-padded='true'] {
		padding: calc(var(--lorivo-space-4) + 1px);
	}
</style>
