<script lang="ts">
	import type { Snippet } from 'svelte';

	type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
	type Size = 'sm' | 'md' | 'lg';

	let {
		variant = 'secondary',
		size = 'md',
		href,
		type = 'button',
		disabled = false,
		ariaLabel,
		onclick,
		leading,
		trailing,
		children
	} = $props<{
		variant?: Variant;
		size?: Size;
		href?: string;
		type?: 'button' | 'submit' | 'reset';
		disabled?: boolean;
		ariaLabel?: string;
		onclick?: ((event: MouseEvent) => void) | undefined;
		leading?: Snippet;
		trailing?: Snippet;
		children?: Snippet;
	}>();
</script>

{#if href}
	<a class="v-button" data-variant={variant} data-size={size} href={href} aria-label={ariaLabel} {onclick}>
		{#if leading}<span class="v-button__slot">{@render leading()}</span>{/if}
		<span class="v-button__label">{@render children?.()}</span>
		{#if trailing}<span class="v-button__slot">{@render trailing()}</span>{/if}
	</a>
{:else}
	<button
		class="v-button"
		data-variant={variant}
		data-size={size}
		{type}
		{disabled}
		aria-label={ariaLabel}
		{onclick}
	>
		{#if leading}<span class="v-button__slot">{@render leading()}</span>{/if}
		<span class="v-button__label">{@render children?.()}</span>
		{#if trailing}<span class="v-button__slot">{@render trailing()}</span>{/if}
	</button>
{/if}

<style>
	.v-button {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		gap: var(--xuva-space-2);
		padding: 0 var(--xuva-space-4);
		border: 1px solid transparent;
		border-radius: var(--xuva-radius-md);
		font-weight: 650;
		letter-spacing: 0.006em;
		text-decoration: none;
		transition:
			border-color 140ms ease,
			background-color 140ms ease,
			color 140ms ease,
			transform 140ms ease,
			box-shadow 140ms ease;
	}

	.v-button[data-size='sm'] {
		min-height: var(--xuva-control-height-sm);
		font-size: 0.85rem;
	}

	.v-button[data-size='md'] {
		min-height: var(--xuva-control-height-md);
		font-size: 0.95rem;
	}

	.v-button[data-size='lg'] {
		min-height: var(--xuva-control-height-lg);
		font-size: 1rem;
	}

	.v-button[data-variant='primary'] {
		background: linear-gradient(135deg, #8d6cff, #7c5cff 58%, #6a4af0);
		color: #ffffff;
		border-color: rgb(155 124 255 / 46%);
		box-shadow:
			0 14px 34px rgb(124 92 255 / 28%),
			0 0 0 1px rgb(255 255 255 / 7%) inset;
	}

	.v-button[data-variant='secondary'] {
		background: rgb(255 255 255 / 6%);
		border-color: rgb(255 255 255 / 18%);
		color: var(--xuva-color-text);
		backdrop-filter: blur(12px);
	}

	.v-button[data-variant='ghost'] {
		color: var(--xuva-color-text-muted);
		background: rgb(255 255 255 / 3%);
	}

	.v-button[data-variant='danger'] {
		background: rgb(220 139 131 / 15%);
		border-color: rgb(220 139 131 / 28%);
		color: #ffc6c6;
	}

	.v-button:hover:not(:disabled) {
		transform: translateY(-1px);
		filter: brightness(1.06);
		box-shadow:
			0 16px 34px rgb(0 0 0 / 32%),
			0 0 24px rgb(124 92 255 / 18%);
	}

	.v-button:disabled {
		opacity: 0.45;
		cursor: not-allowed;
	}

	.v-button__slot {
		display: inline-flex;
		width: 1rem;
		height: 1rem;
	}

	.v-button__slot :global(svg) {
		width: 100%;
		height: 100%;
	}
</style>
