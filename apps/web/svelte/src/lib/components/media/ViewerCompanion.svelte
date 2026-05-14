<script lang="ts">
	import ArtworkFallback from './ArtworkFallback.svelte';
	import ViewerQuickActions, { type ViewerQuickAction } from './ViewerQuickActions.svelte';
	import XuvaStat from '../ui/XuvaStat.svelte';
	import XuvaSurface from '../ui/XuvaSurface.svelte';

	interface SummaryModel {
		libraryCount: number;
		movieCount: number;
		tvCount: number;
		inProgressCount: number;
	}

	interface WatchlistItem {
		id: string;
		title: string;
		subtitle?: string;
		meta?: string;
		posterUrl?: string;
		backdropUrl?: string;
		playMediaSourceId?: string;
	}

	let {
		summary,
		watchlistItems = [],
		quickActions = [],
		trueEmpty = false
	} = $props<{
		summary: SummaryModel;
		watchlistItems?: WatchlistItem[];
		quickActions?: ViewerQuickAction[];
		trueEmpty?: boolean;
	}>();

	function formatCount(value: number): string {
		if (!Number.isFinite(value)) return '0';
		return new Intl.NumberFormat().format(Math.max(0, Math.round(value)));
	}
</script>

<div class="home-companion">
	<XuvaSurface tone="elevated" padded={false}>
		<section class="companion-section">
			<header class="companion-heading">
				<h2>Your Library</h2>
			</header>
			<div class="companion-stats">
				<XuvaStat label="Libraries" value={formatCount(summary.libraryCount)} />
				<XuvaStat label="Movies" value={formatCount(summary.movieCount)} />
				<XuvaStat label="TV" value={formatCount(summary.tvCount)} />
				<XuvaStat label="In Progress" value={formatCount(summary.inProgressCount)} />
			</div>
			{#if trueEmpty}
				<p class="companion-note">
					Add your first library to start building Home rows and watch progress.
				</p>
			{/if}
		</section>

		<section class="companion-section" id="homeWatchlist">
			<header class="companion-heading companion-heading--compact">
				<h2>Watchlist</h2>
				{#if watchlistItems.length > 0}
					<a href="#homeWatchlist">View all</a>
				{/if}
			</header>

			{#if watchlistItems.length > 0}
				<ul class="watchlist-items">
					{#each watchlistItems.slice(0, 4) as item (item.id)}
						<li>
							<a
								href={item.playMediaSourceId
									? `/play/${encodeURIComponent(item.playMediaSourceId)}`
									: '#homeHero'}
							>
								<span class="watchlist-thumb">
									<ArtworkFallback
										variant="poster"
										title={item.title}
										meta={item.subtitle || item.meta}
										showCopy={false}
									/>
								</span>
								<span class="watchlist-copy">
									<h3>{item.title}</h3>
									<p>{item.subtitle || item.meta}</p>
								</span>
							</a>
						</li>
					{/each}
				</ul>
			{:else}
				<p class="watchlist-empty">Save titles for later and they will appear here.</p>
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
	.home-companion {
		align-self: start;
	}

	.home-companion :global(.v-surface) {
		border-color: rgb(255 246 229 / 9%);
		border-radius: 17px;
		background:
			linear-gradient(180deg, rgb(255 246 229 / 2%), transparent 18%),
			rgb(20 21 19 / 84%);
		box-shadow: 0 24px 62px rgb(0 0 0 / 24%);
	}

	.companion-section {
		padding: 13px 14px 13px;
	}

	.companion-section + .companion-section {
		border-top: 1px solid rgb(255 246 229 / 8%);
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

	.companion-heading a {
		color: color-mix(in srgb, var(--xuva-color-text-muted) 78%, transparent);
		font-size: 0.82rem;
		text-decoration: none;
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
		border: 1px solid rgb(255 246 229 / 9%);
		border-radius: 13px;
		background: rgb(255 246 229 / 4%);
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

	.watchlist-items {
		list-style: none;
		margin: 0;
		padding: 0;
		display: grid;
		gap: 7px;
	}

	.watchlist-items li a {
		display: grid;
		grid-template-columns: 58px minmax(0, 1fr);
		gap: 10px;
		align-items: center;
		padding: 0;
		text-decoration: none;
	}

	.watchlist-thumb {
		display: inline-flex;
		width: 58px;
		height: 58px;
		border-radius: 10px;
		overflow: hidden;
		border: 1px solid rgb(255 246 229 / 10%);
		background: rgb(22 23 20 / 85%);
	}

	.watchlist-thumb :global(.artwork-fallback) {
		padding: 8px;
	}

	.watchlist-copy {
		display: grid;
		gap: 4px;
		min-width: 0;
	}

	.watchlist-copy h3 {
		margin: 0;
		color: color-mix(in srgb, var(--xuva-color-text) 96%, transparent);
		font-size: 0.9rem;
		font-weight: 620;
	}

	.watchlist-copy h3,
	.watchlist-copy p {
		white-space: nowrap;
		overflow: hidden;
		text-overflow: ellipsis;
	}

	.watchlist-copy p {
		margin: 0;
		color: color-mix(in srgb, var(--xuva-color-text-soft) 82%, transparent);
		font-size: 0.72rem;
		font-weight: 600;
	}

	.watchlist-empty {
		margin: 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 74%, transparent);
		font-size: 0.84rem;
		line-height: 1.45;
	}
</style>
