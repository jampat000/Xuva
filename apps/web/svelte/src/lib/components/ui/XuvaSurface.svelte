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
		border: 1px solid var(--xuva-color-border-soft);
		border-radius: 4px;
		box-shadow: none;
	}

	.v-surface[data-tone='default'] {
		background: color-mix(in srgb, var(--xuva-color-bg-panel) 98%, #0f172a 2%);
	}

	.v-surface[data-tone='elevated'] {
		background: color-mix(in srgb, var(--xuva-color-bg-panel-elevated) 98%, #0f172a 2%);
	}

	.v-surface[data-tone='soft'] {
		background: rgb(255 255 255 / 1.5%);
	}

	.v-surface[data-padded='true'] {
		padding: 0.85rem 0.95rem;
	}
</style>
