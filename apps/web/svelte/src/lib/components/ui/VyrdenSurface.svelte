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
		border: 1px solid var(--vyrden-color-border-soft);
		border-radius: var(--vyrden-radius-lg);
		box-shadow: var(--vyrden-shadow-soft);
	}

	.v-surface[data-tone='default'] {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), transparent 22%),
			var(--vyrden-color-bg-panel);
	}

	.v-surface[data-tone='elevated'] {
		background:
			linear-gradient(180deg, rgb(255 255 255 / 6%), transparent 24%),
			var(--vyrden-color-bg-panel-elevated);
	}

	.v-surface[data-tone='soft'] {
		background: rgb(255 255 255 / 5%);
	}

	.v-surface[data-padded='true'] {
		padding: calc(var(--vyrden-space-4) + 1px);
	}
</style>
