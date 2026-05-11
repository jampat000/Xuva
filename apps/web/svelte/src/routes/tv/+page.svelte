<script lang="ts">
	import { onMount } from 'svelte';
	import { Search } from 'lucide-svelte';
	import { getSeries, refreshMetadataBatch, scanTV, type SeriesListItem } from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { resolvePreviewMode } from '$lib/home/model';
	import { previewPoster } from '$lib/preview/artwork';
	import LorivoButton from '$lib/lorivo/LorivoButton.svelte';
	import LorivoPanel from '$lib/lorivo/LorivoPanel.svelte';
	import LorivoPosterLink from '$lib/lorivo/LorivoPosterLink.svelte';
	import LorivoShell from '$lib/lorivo/LorivoShell.svelte';
	import MediaGrid from '$lib/lorivo/MediaGrid.svelte';
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
	let seriesRows = $state<SeriesListItem[]>([]);

	const seriesCards = $derived.by(() => buildSeriesCards(seriesRows));
	const visibleCards = $derived.by(() =>
		filterAndSortSeriesCards(seriesCards, searchValue, seriesFilter, seriesSort)
	);
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
			// Keep the page usable if URL parsing fails.
		}
		void loadSeries();
	});

	async function loadSeries(): Promise<void> {
		isLoading = true;
		loadError = '';
		const activePreviewMode = resolvePreviewMode(new URL(window.location.href).searchParams);
		if (activePreviewMode) {
			previewMode = true;
			seriesRows = previewSeriesRows();
			isLoading = false;
			return;
		}

		try {
			const seriesPayload = await getSeries(apiClient, 500);
			previewMode = false;
			seriesRows = seriesPayload.series || [];
		} catch (error) {
			loadError = formatLoadError(error);
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

	function filterClass(active: boolean): string {
		const base =
			'rounded-full border px-4 py-2 text-sm font-medium transition duration-200 focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#7C5CFF]/70 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120]';
		if (active) return `${base} border-[#7C5CFF]/60 bg-[#7C5CFF] text-white shadow-lg shadow-[#7C5CFF]/20`;
		return `${base} border-white/10 bg-[#111827] text-white/60 hover:border-white/25 hover:bg-white/10 hover:text-white`;
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

<svelte:head>
	<title>TV - Lorivo Media</title>
</svelte:head>

<LorivoShell>
	<section class="relative mx-4 mt-4 overflow-hidden rounded-2xl bg-[#111827] px-6 py-10 sm:mx-6 sm:px-10 lg:mx-8 lg:px-12 xl:px-16">
		<div class="absolute inset-0 bg-gradient-to-r from-[#0B1120] via-[#0B1120]/70 to-[#0B1120]/30"></div>
		<div class="relative flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-[600px]">
				<h1 class="text-5xl font-bold leading-tight text-white [text-shadow:0_4px_28px_rgba(0,0,0,0.72)] sm:text-6xl xl:text-7xl">TV Shows</h1>
				<p class="mt-4 text-base text-white/60">Browse your TV library.</p>
				<p class="mt-5 text-base leading-relaxed text-white/70">
					{formatCount(visibleCards.length)} visible shows - {formatCount(totalSeasons)} seasons - {formatCount(totalEpisodes)} episodes
				</p>
			</div>
			{#if !previewMode}
				<div class="flex flex-wrap gap-3">
					<LorivoButton variant="primary" onclick={startTVScan} disabled={isScanning || isRefreshing}>
						{isScanning ? 'Scanning...' : 'Scan TV'}
					</LorivoButton>
					<LorivoButton variant="secondary" onclick={runMetadataRefresh} disabled={isScanning || isRefreshing}>
						{isRefreshing ? 'Refreshing...' : 'Refresh Metadata'}
					</LorivoButton>
				</div>
			{/if}
		</div>
	</section>

	{#if isLoading}
		<LorivoPanel title="Loading TV Shows" subtitle="Fetching your TV library from the media APIs." />
	{:else if loadError}
		<LorivoPanel title="TV Shows could not load" subtitle={loadError}>
			<div class="flex flex-wrap gap-3">
				<LorivoButton variant="secondary" onclick={loadSeries}>Retry</LorivoButton>
				<LorivoButton variant="ghost" href="/">Back to Home</LorivoButton>
			</div>
		</LorivoPanel>
	{:else}
		<section class="relative px-4 pt-9 sm:px-6 sm:pt-10 lg:px-8 lg:pt-11">
			<div class="flex flex-col gap-4 rounded-2xl border border-white/10 bg-white/[0.04] p-4 shadow-lg shadow-black/20 backdrop-blur lg:flex-row lg:items-center lg:justify-between">
				<div class="relative w-full lg:max-w-[420px]">
					<Search size={16} class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-white/40" />
					<input
						type="text"
						placeholder="Search TV"
						bind:value={searchValue}
						class="h-10 w-full rounded-full border border-white/5 bg-[#111827] pl-11 pr-4 text-sm text-white placeholder:text-white/40 focus:outline-none focus:ring-1 focus:ring-[#7C5CFF]/50"
					/>
				</div>
				<div class="flex flex-wrap gap-2">
					{#if !previewMode}
						<button type="button" class={filterClass(seriesFilter === 'all')} onclick={() => (seriesFilter = 'all')}>All</button>
						<button type="button" class={filterClass(seriesFilter === 'multi-season')} onclick={() => (seriesFilter = 'multi-season')}>Multi-Season</button>
						<button type="button" class={filterClass(seriesFilter === 'with-episodes')} onclick={() => (seriesFilter = 'with-episodes')}>With Episodes</button>
						<button type="button" class={filterClass(seriesFilter === 'unknown-year')} onclick={() => (seriesFilter = 'unknown-year')}>Unknown Year</button>
					{/if}
					<button type="button" class={filterClass(seriesSort === 'title')} onclick={() => (seriesSort = 'title')}>Title</button>
					<button type="button" class={filterClass(seriesSort === 'year')} onclick={() => (seriesSort = 'year')}>Year</button>
					<button type="button" class={filterClass(seriesSort === 'seasons')} onclick={() => (seriesSort = 'seasons')}>Seasons</button>
					<button type="button" class={filterClass(seriesSort === 'episodes')} onclick={() => (seriesSort = 'episodes')}>Episodes</button>
				</div>
			</div>
			{#if actionMessage}
				<p class="mt-3 text-sm text-white/60">{actionMessage}</p>
			{/if}
		</section>

		{#if seriesCards.length === 0}
			<LorivoPanel title="No TV shows found" subtitle="Try adding a TV library or running a scan." />
		{:else if visibleCards.length === 0}
			<LorivoPanel title="No TV shows found" subtitle="Try changing search terms." />
		{:else}
			<MediaGrid title="TV Shows" subtitle={`${formatCount(visibleCards.length)} shows`}>
				{#each visibleCards as item (item.id)}
					<LorivoPosterLink
						title={item.title}
						meta={item.meta}
						img={item.posterUrl}
						href={`/tv/${encodeURIComponent(item.id)}${previewMode ? '?preview=1' : ''}`}
					/>
				{/each}
			</MediaGrid>
		{/if}
	{/if}
</LorivoShell>
