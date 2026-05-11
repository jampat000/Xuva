<script lang="ts">
	import type { Snippet } from 'svelte';
	import VyrdenSurface from '../ui/VyrdenSurface.svelte';
	import LiveStatusBadge from './LiveStatusBadge.svelte';

	type Status = 'healthy' | 'warning' | 'critical' | 'idle';

	let {
		title,
		description = '',
		status = 'idle',
		actions,
		children
	} = $props<{
		title: string;
		description?: string;
		status?: Status;
		actions?: Snippet;
		children?: Snippet;
	}>();
</script>

<VyrdenSurface tone="elevated">
	<section class="admin-panel">
		<header>
			<div>
				<h2>{title}</h2>
				{#if description}
					<p>{description}</p>
				{/if}
			</div>
			<div class="admin-panel__header-right">
				<LiveStatusBadge {status} />
				{@render actions?.()}
			</div>
		</header>
		<div>
			{@render children?.()}
		</div>
	</section>
</VyrdenSurface>

<style>
	.admin-panel {
		display: grid;
		gap: var(--vyrden-space-4);
	}

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--vyrden-space-3);
	}

	h2 {
		margin: 0;
		font-size: 1.04rem;
		font-weight: 680;
		letter-spacing: -0.01em;
	}

	p {
		margin: var(--vyrden-space-1) 0 0;
		color: var(--vyrden-color-text-muted);
		font-size: 0.84rem;
	}

	.admin-panel__header-right {
		display: flex;
		gap: var(--vyrden-space-2);
		align-items: center;
	}
</style>
