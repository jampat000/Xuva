<script lang="ts">
	import type { Snippet } from 'svelte';

	type ButtonVariant = 'primary' | 'secondary' | 'ghost';
	type ButtonSize = 'md' | 'sm';

	let {
		variant = 'primary',
		size = 'md',
		href = '',
		disabled = false,
		onclick,
		children
	}: {
		variant?: ButtonVariant;
		size?: ButtonSize;
		href?: string;
		disabled?: boolean;
		onclick?: (event: MouseEvent) => void;
		children: Snippet;
	} = $props();

	const baseClass =
		'inline-flex items-center justify-center gap-3 rounded-xl font-semibold text-white transition duration-200 active:translate-y-0 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120]';
	const sizeClass = $derived(
		size === 'sm' ? 'min-h-[42px] px-4 py-2 text-sm' : 'min-h-[60px] px-8 py-4 text-base sm:px-10'
	);
	const variantClass = $derived.by(() => {
		if (variant === 'secondary') {
			return '!border !border-white/25 !bg-white/10 shadow-lg shadow-black/20 backdrop-blur hover:-translate-y-0.5 hover:!border-white/40 hover:!bg-white/15 focus-visible:ring-white/35';
		}
		if (variant === 'ghost') {
			return '!border !border-white/10 !bg-white/[0.03] text-white/85 hover:-translate-y-0.5 hover:!border-white/25 hover:!bg-white/10 hover:text-white focus-visible:ring-white/30';
		}
		return '!bg-[#7C5CFF] shadow-xl shadow-[#7C5CFF]/40 hover:-translate-y-0.5 hover:!bg-[#6a4af0] hover:brightness-110 focus-visible:ring-[#7C5CFF]/70';
	});
	const disabledClass = $derived(disabled ? 'pointer-events-none opacity-50' : '');
	const className = $derived(`${baseClass} ${sizeClass} ${variantClass} ${disabledClass}`);
</script>

{#if href && !disabled}
	<a class={className} {href}>
		{@render children()}
	</a>
{:else if href && disabled}
	<span class={className} aria-disabled="true">
		{@render children()}
	</span>
{:else}
	<button type="button" class={className} {disabled} {onclick}>
		{@render children()}
	</button>
{/if}
