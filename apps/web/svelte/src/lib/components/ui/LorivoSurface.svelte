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
		border-radius: var(--lorivo-radius-md);
	}

	.v-surface[data-tone='default'] {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 4%), transparent 18%),
			var(--lorivo-color-bg-panel);
	}

	.v-surface[data-tone='elevated'] {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 4%), transparent 18%),
			var(--lorivo-color-bg-panel-elevated);
	}

	.v-surface[data-tone='soft'] {
		background: rgb(255 255 255 / 3%);
	}

	.v-surface[data-padded='true'] {
		padding: var(--lorivo-space-4);
	}
</style>
