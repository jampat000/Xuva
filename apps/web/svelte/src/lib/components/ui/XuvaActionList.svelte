<script lang="ts">
	export interface XuvaActionItem {
		id: string;
		label: string;
		description?: string;
		status?: string;
		actionLabel?: string;
	}

	let {
		items = [],
		emptyLabel = 'No actions available'
	} = $props<{
		items?: XuvaActionItem[];
		emptyLabel?: string;
	}>();
</script>

{#if items.length === 0}
	<p class="v-action-list__empty">{emptyLabel}</p>
{:else}
	<ul class="v-action-list">
		{#each items as item (item.id)}
			<li>
				<div class="v-action-list__copy">
					<strong>{item.label}</strong>
					{#if item.description}
						<p>{item.description}</p>
					{/if}
				</div>
				{#if item.status}
					<span class="v-action-list__status">{item.status}</span>
				{/if}
				{#if item.actionLabel}
					<button type="button" aria-label={item.actionLabel}>{item.actionLabel}</button>
				{/if}
			</li>
		{/each}
	</ul>
{/if}

<style>
	.v-action-list {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: 0;
	}

	li {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto auto;
		align-items: center;
		gap: var(--xuva-space-3);
		padding: 0.75rem 0;
		border-top: 1px solid var(--xuva-color-border-soft);
		background: transparent;
	}

	.v-action-list > :first-child {
		padding-top: 0;
		border-top: 0;
	}

	.v-action-list__copy {
		min-width: 0;
	}

	strong {
		display: block;
		font-size: 0.92rem;
		letter-spacing: -0.008em;
	}

	p {
		margin: var(--xuva-space-1) 0 0;
		font-size: 0.81rem;
		color: var(--xuva-color-text-muted);
	}

	.v-action-list__status {
		font-size: 0.73rem;
		color: var(--xuva-color-text-soft);
		white-space: nowrap;
	}

	button {
		min-height: var(--xuva-control-height-sm);
		padding: 0 var(--xuva-space-3);
		border-radius: var(--xuva-radius-sm);
		background: rgb(255 246 229 / 8%);
		color: var(--xuva-color-text);
		font-size: 0.82rem;
		font-weight: 620;
	}

	.v-action-list__empty {
		margin: 0;
		color: var(--xuva-color-text-muted);
		font-size: 0.88rem;
	}

	@media (max-width: 640px) {
		li {
			grid-template-columns: minmax(0, 1fr);
		}
	}
</style>
