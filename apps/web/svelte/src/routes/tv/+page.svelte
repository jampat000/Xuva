<script lang="ts">
	import { onMount } from 'svelte';
	import { getAuthSession, type AuthSessionUser } from '$lib/api/auth';
	import { getSeries, refreshMetadataBatch, scanTV, type SeriesListItem } from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { resolvePreviewMode } from '$lib/home/model';
	import { previewPoster } from '$lib/preview/artwork';
	import {
		BrowseFilterGroup,
		BrowseGrid,
		BrowseHeader,
		MediaShell,
		BrowsePage,
		BrowseStatChip,
		BrowseToolbar,
		PosterCard,
		VyrdenButton,
		VyrdenEmptyState,
		VyrdenPanel
	} from '$lib/components';
	import {
		buildSeriesCards,
		filterAndSortSeriesCards,
		type SeriesFilter,
		type SeriesSort
	} from '$lib/browse/model';

	let isLoading = $state(true);
	let isScanning = $state(false);
	let isRefreshing = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let searchValue = $state('');
	let previewMode = $state(false);
	let seriesFilter = $state<SeriesFilter>('all');
	let seriesSort = $state<SeriesSort>('title');
	let user = $state<AuthSessionUser | null>(null);
	let libraries = $state<LibraryRecord[]>([]);
	let seriesRows = $state<SeriesListItem[]>([]);

	const seriesCards = $derived.by(() => buildSeriesCards(seriesRows));
	const visibleCards = $derived.by(() =>
		filterAndSortSeriesCards(seriesCards, searchValue, seriesFilter, seriesSort)
	);
	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
	const totalEpisodes = $derived.by(() =>
		seriesCards.reduce((total, item) => total + item.episodeCount, 0)
	);
	const totalSeasons = $derived.by(() =>
		seriesCards.reduce((total, item) => total + item.seasonCount, 0)
	);

	onMount(() => {
		try {
			const params = new URL(window.location.href).searchParams;
			previewMode = resolvePreviewMode(params);
			const q = params.get('q');
			if (q) searchValue = q;
		} catch {
			// Ignore URL parsing errors and use default search state.
		}
		void loadSeries();
	});

	async function loadSeries(): Promise<void> {
		isLoading = true;
		loadError = '';
		const activePreviewMode = resolvePreviewMode(new URL(window.location.href).searchParams);
		if (activePreviewMode) {
			previewMode = true;
			user = null;
			libraries = [];
			seriesRows = previewSeriesRows();
			isLoading = false;
			return;
		}
		try {
			const [sessionPayload, librariesPayload, seriesPayload] = await Promise.all([
				getAuthSession(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) return { user: null };
					throw error;
				}),
				getLibraries(apiClient),
				getSeries(apiClient, 500)
			]);
			user = sessionPayload?.user || null;
			libraries = librariesPayload.libraries || [];
			seriesRows = seriesPayload.series || [];
			if (activePreviewMode && seriesRows.length === 0) {
				seriesRows = previewSeriesRows();
			}
		} catch (error) {
			if (activePreviewMode && isApiStatus(error, 401)) {
				user = null;
				seriesRows = previewSeriesRows();
				loadError = '';
			} else {
				loadError = formatLoadError(error);
			}
		} finally {
			isLoading = false;
		}
	}

	async function startTVScan(): Promise<void> {
		isScanning = true;
		actionMessage = '';
		try {
			await scanTV(apiClient, 50);
			actionMessage = 'TV scan started. Refreshing the wall...';
			await loadSeries();
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isScanning = false;
		}
	}

	async function runMetadataRefresh(): Promise<void> {
		isRefreshing = true;
		actionMessage = '';
		try {
			const result = await refreshMetadataBatch('series', apiClient, 25);
			if (Array.isArray(result.warnings) && result.warnings.length > 0) {
				actionMessage = result.warnings.slice(0, 3).join(' | ');
			} else {
				actionMessage = 'Metadata refresh accepted.';
			}
			await loadSeries();
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isRefreshing = false;
		}
	}

	function formatCount(value: number): string {
		if (!Number.isFinite(value)) return '0';
		return new Intl.NumberFormat().format(Math.max(0, Math.round(value)));
	}

	function initialsForName(name: string): string {
		const words = asText(name).split(/\s+/).filter(Boolean);
		if (words.length === 0) return 'V';
		if (words.length === 1) return words[0].slice(0, 1).toUpperCase();
		return `${words[0][0] || ''}${words[1][0] || ''}`.toUpperCase();
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function isApiStatus(error: unknown, expectedStatus: number): boolean {
		if (error instanceof ApiClientError) return error.status === expectedStatus;
		if (typeof error !== 'object' || !error) return false;
		const candidate = (error as { status?: unknown }).status;
		return Number(candidate) === expectedStatus;
	}

	function formatLoadError(error: unknown): string {
		if (error instanceof ApiClientError) return error.userMessage || error.message;
		if (isApiStatus(error, 401)) return 'Your session is no longer active. Sign in again to continue.';
		if (error instanceof Error) return error.message;
		return 'TV could not load.';
	}

	function previewSeriesRows(): SeriesListItem[] {
		return [
			{
				id: 'preview-tv-coastline',
				title: 'Coastline',
				seasonCount: 3,
				episodeCount: 24,
				metadata: { title: 'Coastline', year: 2024, overview: 'Preview series item.', posterUrl: previewArtwork('Coastline') }
			},
			{
				id: 'preview-tv-violet-signal',
				title: 'Violet Signal',
				seasonCount: 2,
				episodeCount: 18,
				metadata: { title: 'Violet Signal', year: 2023, overview: 'Preview series item.', posterUrl: previewArtwork('Violet Signal') }
			},
			{
				id: 'preview-tv-return-vector',
				title: 'Return Vector',
				seasonCount: 1,
				episodeCount: 8,
				metadata: { title: 'Return Vector', year: 2022, overview: 'Preview series item.', posterUrl: previewArtwork('Return Vector') }
			},
			{
				id: 'preview-tv-low-country',
				title: 'Low Country',
				seasonCount: 4,
				episodeCount: 36,
				metadata: { title: 'Low Country', year: 2021, overview: 'Preview series item.', posterUrl: previewArtwork('Low Country') }
			},
			{
				id: 'preview-tv-night-archive',
				title: 'Night Archive',
				seasonCount: 2,
				episodeCount: 20,
				metadata: { title: 'Night Archive', year: 2022, overview: 'Preview series item.', posterUrl: previewArtwork('Night Archive') }
			},
			{
				id: 'preview-tv-orchard-line',
				title: 'Orchard Line',
				seasonCount: 3,
				episodeCount: 27,
				metadata: { title: 'Orchard Line', year: 2020, overview: 'Preview series item.', posterUrl: previewArtwork('Orchard Line') }
			},
			{
				id: 'preview-tv-atlas-watch',
				title: 'Atlas Watch',
				seasonCount: 1,
				episodeCount: 10,
				metadata: { title: 'Atlas Watch', year: 2024, overview: 'Preview series item.', posterUrl: previewArtwork('Atlas Watch') }
			},
			{
				id: 'preview-tv-ember-shore',
				title: 'Ember Shore',
				seasonCount: 2,
				episodeCount: 16,
				metadata: { title: 'Ember Shore', year: 2021, overview: 'Preview series item.', posterUrl: previewArtwork('Ember Shore') }
			},
			{
				id: 'preview-tv-sunward',
				title: 'Sunward',
				seasonCount: 4,
				episodeCount: 40,
				metadata: { title: 'Sunward', year: 2023, overview: 'Preview series item.', posterUrl: previewArtwork('Sunward') }
			},
			{
				id: 'preview-tv-littoral',
				title: 'Littoral',
				seasonCount: 1,
				episodeCount: 12,
				metadata: { title: 'Littoral', year: 2024, overview: 'Preview series item.', posterUrl: previewArtwork('Littoral') }
			}
		];
	}

	function previewArtwork(title: string): string {
		return previewPoster(title);
	}
</script>

<MediaShell active="tv" bind:searchValue {userInitials}>
	<BrowsePage>
		{#if isLoading}
			<VyrdenPanel title="Loading TV Shows" subtitle="Fetching your TV library from the media APIs." />
		{:else if loadError}
			<VyrdenPanel title="TV Shows could not load" subtitle={loadError}>
				<div class="status-actions">
					<VyrdenButton variant="secondary" onclick={loadSeries}>Retry</VyrdenButton>
					<VyrdenButton variant="ghost" href="/">Back to Home</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else}
			<BrowseHeader title="TV Shows" subtitle="Browse your TV library.">
				{#snippet chips()}
					<BrowseStatChip label={`${formatCount(visibleCards.length)} visible`} />
					<BrowseStatChip label={`${formatCount(totalSeasons)} seasons`} />
					<BrowseStatChip label={`${formatCount(totalEpisodes)} episodes`} />
				{/snippet}
			</BrowseHeader>

			<BrowseToolbar message={actionMessage}>
				{#snippet controls()}
					{#if !previewMode}
						<BrowseFilterGroup ariaLabel="TV filters">
							<button type="button" class:selected={seriesFilter === 'all'} onclick={() => (seriesFilter = 'all')}>
								All
							</button>
							<button
								type="button"
								class:selected={seriesFilter === 'multi-season'}
								onclick={() => (seriesFilter = 'multi-season')}
							>
								Multi-Season
							</button>
							<button
								type="button"
								class:selected={seriesFilter === 'with-episodes'}
								onclick={() => (seriesFilter = 'with-episodes')}
							>
								With Episodes
							</button>
							<button
								type="button"
								class:selected={seriesFilter === 'unknown-year'}
								onclick={() => (seriesFilter = 'unknown-year')}
							>
								Unknown Year
							</button>
						</BrowseFilterGroup>
					{/if}
					<BrowseFilterGroup segmented ariaLabel="TV sorting">
						<button type="button" class:selected={seriesSort === 'title'} onclick={() => (seriesSort = 'title')}>
							Title
						</button>
						<button type="button" class:selected={seriesSort === 'year'} onclick={() => (seriesSort = 'year')}>
							Year
						</button>
						<button
							type="button"
							class:selected={seriesSort === 'seasons'}
							onclick={() => (seriesSort = 'seasons')}
						>
							Seasons
						</button>
						<button
							type="button"
							class:selected={seriesSort === 'episodes'}
							onclick={() => (seriesSort = 'episodes')}
						>
							Episodes
						</button>
					</BrowseFilterGroup>
				{/snippet}
				{#snippet actions()}
					{#if !previewMode}
						<VyrdenButton variant="primary" onclick={startTVScan} disabled={isScanning || isRefreshing}>
							{isScanning ? 'Scanning...' : 'Scan TV'}
						</VyrdenButton>
						<VyrdenButton variant="secondary" onclick={runMetadataRefresh} disabled={isScanning || isRefreshing}>
							{isRefreshing ? 'Refreshing...' : 'Refresh Metadata'}
						</VyrdenButton>
					{/if}
				{/snippet}
			</BrowseToolbar>

			{#if seriesCards.length === 0}
				<VyrdenEmptyState
					title="No TV shows found"
					message="Try adding a TV library or running a scan."
				/>
			{:else if visibleCards.length === 0}
				<VyrdenEmptyState title="No TV shows found" message="Try changing search terms." />
			{:else}
				<BrowseGrid>
					{#each visibleCards as item (item.id)}
						<PosterCard
							title={item.title}
							meta={item.meta}
							imageUrl={item.posterUrl}
							href={`/tv/${encodeURIComponent(item.id)}${previewMode ? '?preview=1' : ''}`}
						/>
					{/each}
				</BrowseGrid>
			{/if}
		{/if}
	</BrowsePage>
</MediaShell>

<style>
	.status-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--vyrden-space-2);
	}
</style>
