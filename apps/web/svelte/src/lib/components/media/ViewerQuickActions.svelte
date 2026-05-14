<script lang="ts">
	export interface ViewerQuickAction {
		id: string;
		label: string;
		href?: string;
	}

	let { items = [] } = $props<{ items?: ViewerQuickAction[] }>();
</script>

<ul class="viewer-actions">
	{#each items as item (item.id)}
		<li>
			<a href={item.href || '#'}>{item.label}</a>
		</li>
	{/each}
</ul>

<style>
	.viewer-actions {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: var(--xuva-space-2);
	}

	li a {
		display: flex;
		align-items: center;
		justify-content: space-between;
		min-height: var(--xuva-control-height-md);
		padding: 0 var(--xuva-space-3);
		border: 1px solid var(--xuva-color-border-soft);
		border-radius: var(--xuva-radius-md);
		background: rgb(255 246 229 / 3%);
		font-size: 0.88rem;
		font-weight: 630;
		color: var(--xuva-color-text);
		text-decoration: none;
	}

	li a::after {
		content: '>';
		color: var(--xuva-color-text-soft);
	}

	li a:hover {
		background: rgb(255 246 229 / 7%);
	}
</style>
