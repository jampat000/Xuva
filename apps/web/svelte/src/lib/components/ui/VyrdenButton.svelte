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
		gap: var(--vyrden-space-2);
		padding: 0 var(--vyrden-space-4);
		border: 1px solid transparent;
		border-radius: var(--vyrden-radius-md);
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
		min-height: var(--vyrden-control-height-sm);
		font-size: 0.85rem;
	}

	.v-button[data-size='md'] {
		min-height: var(--vyrden-control-height-md);
		font-size: 0.95rem;
	}

	.v-button[data-size='lg'] {
		min-height: var(--vyrden-control-height-lg);
		font-size: 1rem;
	}

	.v-button[data-variant='primary'] {
		background: linear-gradient(180deg, #6fd3ba, #49b59b);
		color: #f7f3ea;
		border-color: rgb(111 211 186 / 42%);
		box-shadow: 0 10px 24px rgb(0 0 0 / 26%);
	}

	.v-button[data-variant='secondary'] {
		background: rgb(255 246 229 / 4%);
		border-color: var(--vyrden-color-border-soft);
		color: var(--vyrden-color-text);
	}

	.v-button[data-variant='ghost'] {
		color: var(--vyrden-color-text-muted);
		background: rgb(255 246 229 / 2%);
	}

	.v-button[data-variant='danger'] {
		background: rgb(220 139 131 / 15%);
		border-color: rgb(220 139 131 / 28%);
		color: #ffc6c6;
	}

	.v-button:hover:not(:disabled) {
		transform: translateY(-1px);
		box-shadow: 0 12px 28px rgb(0 0 0 / 26%);
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
