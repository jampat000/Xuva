<script lang="ts">
	import type { ViewerQuickAction } from './ViewerQuickActions.svelte';
	import ViewerQuickActions from './ViewerQuickActions.svelte';
	import XuvaStat from '../ui/XuvaStat.svelte';
	import XuvaSurface from '../ui/XuvaSurface.svelte';

	interface CompanionStat {
		label: string;
		value: string;
	}

	let {
		title = 'Library Snapshot',
		stats = [],
		note = '',
		quickActions = []
	} = $props<{
		title?: string;
		stats?: CompanionStat[];
		note?: string;
		quickActions?: ViewerQuickAction[];
	}>();
</script>

<div class="browse-companion">
	<XuvaSurface tone="elevated" padded={false}>
		<section class="companion-section">
			<header class="companion-heading">
				<h2>{title}</h2>
			</header>
			<div class="companion-stats">
				{#each stats as stat (stat.label)}
					<XuvaStat label={stat.label} value={stat.value} />
				{/each}
			</div>
			{#if note}
				<p class="companion-note">{note}</p>
			{/if}
		</section>

		<section class="companion-section">
			<header class="companion-heading companion-heading--compact">
				<h2>Quick Actions</h2>
			</header>
			<ViewerQuickActions items={quickActions} />
		</section>
	</XuvaSurface>
</div>

<style>
	.browse-companion {
		align-self: start;
	}

	.browse-companion :global(.v-surface) {
		border-color: rgb(255 255 255 / 7.5%);
		border-radius: 17px;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 1.4%), transparent 18%),
			rgb(16 19 24 / 82%);
		box-shadow: 0 24px 62px rgb(0 0 0 / 28%);
	}

	.companion-section {
		padding: 13px 14px 13px;
	}

	.companion-section + .companion-section {
		border-top: 1px solid rgb(255 255 255 / 7.5%);
	}

	.companion-heading {
		display: flex;
		align-items: center;
		justify-content: space-between;
		gap: 14px;
		margin-bottom: 9px;
	}

	.companion-heading--compact h2 {
		font-size: 0.98rem;
		font-weight: 670;
		letter-spacing: -0.02em;
	}

	.companion-heading h2 {
		margin: 0;
		font-family: var(--xuva-font-display);
		font-size: 1rem;
		font-weight: 670;
		letter-spacing: -0.02em;
	}

	.companion-stats {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 8px;
	}

	.companion-stats :global(.v-stat) {
		gap: 6px;
		min-height: 72px;
		padding: 13px 12px;
		border: 1px solid rgb(255 255 255 / 7%);
		border-radius: 13px;
		background: rgb(255 255 255 / 3%);
	}

	.companion-stats :global(.v-stat span) {
		color: color-mix(in srgb, var(--xuva-color-text-muted) 80%, transparent);
		font-size: 0.79rem;
		font-weight: 700;
		letter-spacing: 0.01em;
	}

	.companion-stats :global(.v-stat strong) {
		color: color-mix(in srgb, var(--xuva-color-text) 96%, transparent);
		font-size: 1.22rem;
		font-weight: 800;
		letter-spacing: -0.03em;
	}

	.companion-note {
		margin: 10px 0 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 74%, transparent);
		font-size: 0.84rem;
		line-height: 1.45;
	}
</style>
