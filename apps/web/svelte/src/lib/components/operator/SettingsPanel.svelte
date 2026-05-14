<script lang="ts">
	import type { Snippet } from 'svelte';
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

<section class="settings-panel">
	<header>
		<div>
			<h2>{title}</h2>
			{#if description}
				<p>{description}</p>
			{/if}
		</div>
		<div class="settings-panel__header-right">
			<LiveStatusBadge {status} />
			{#if actions}
				<div class="settings-panel__actions">
					{@render actions()}
				</div>
			{/if}
		</div>
	</header>
	<div>
		{@render children?.()}
	</div>
</section>

<style>
	.settings-panel {
		display: grid;
		gap: var(--xuva-space-3);
		padding: 0.95rem 0 0.45rem;
		border-top: 1px solid var(--xuva-color-border-soft);
	}

	header {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: var(--xuva-space-3);
		padding-bottom: 0.8rem;
		border-bottom: 1px solid color-mix(in srgb, var(--xuva-color-border-soft) 88%, transparent);
	}

	h2 {
		margin: 0;
		font-size: 1.02rem;
		font-weight: 720;
		letter-spacing: -0.01em;
	}

	p {
		margin: var(--xuva-space-1) 0 0;
		color: var(--xuva-color-text-muted);
		font-size: 0.84rem;
	}

	.settings-panel > div:last-child {
		padding-top: 0.2rem;
	}

	.settings-panel__header-right {
		display: flex;
		flex-wrap: wrap;
		gap: var(--xuva-space-2);
		align-items: center;
		justify-content: flex-end;
	}

	.settings-panel__actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--xuva-space-2);
		justify-content: flex-end;
	}

	@media (max-width: 720px) {
		header {
			flex-direction: column;
		}

		.settings-panel__header-right,
		.settings-panel__actions {
			justify-content: flex-start;
		}
	}
</style>
