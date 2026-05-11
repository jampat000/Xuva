<script lang="ts">
	import { onMount } from 'svelte';
	import { Search } from 'lucide-svelte';
	import { getSeries, refreshMetadataBatch, scanTV, type SeriesListItem } from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import LorivoButton from '$lib/lorivo/LorivoButton.svelte';
	import LorivoEmptyState from '$lib/lorivo/LorivoEmptyState.svelte';
	import LorivoPanel from '$lib/lorivo/LorivoPanel.svelte';
	import LorivoPosterLink from '$lib/lorivo/LorivoPosterLink.svelte';
	import LorivoShell from '$lib/lorivo/LorivoShell.svelte';
	import MediaGrid from '$lib/lorivo/MediaGrid.svelte';
	import {
		buildSeriesCards,
		filterAndSortSeriesCards,
		type SeriesSort
	} from '$lib/browse/model';

	let isLoading = $state(true);
	let isScanning = $state(false);
	let isRefreshing = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let searchValue = $state('');
	let seriesSort = $state<SeriesSort>('title');
	let seriesRows = $state<SeriesListItem[]>([]);

	const seriesCards = $derived.by(() => buildSeriesCards(seriesRows));
	const visibleCards = $derived.by(() =>
		filterAndSortSeriesCards(seriesCards, searchValue, 'all', seriesSort)
	);
	const visibleEpisodes = $derived.by(() =>
		visibleCards.reduce((total, item) => total + item.episodeCount, 0)
	);
	const visibleSeasons = $derived.by(() =>
		visibleCards.reduce((total, item) => total + item.seasonCount, 0)
	);

	onMount(() => {
		try {
			const params = new URL(window.location.href).searchParams;
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

		try {
			const seriesPayload = await getSeries(apiClient, 500);
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

	function formatVisibleCount(shows: number, seasons: number, episodes: number): string {
		const showLabel = shows === 1 ? 'show' : 'shows';
		const seasonLabel = seasons === 1 ? 'season' : 'seasons';
		const episodeLabel = episodes === 1 ? 'episode' : 'episodes';
		return `${formatCount(shows)} visible ${showLabel} - ${formatCount(seasons)} ${seasonLabel} - ${formatCount(episodes)} ${episodeLabel}`;
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

	function chipClass(active: boolean): string {
		const base =
			'inline-flex min-h-9 items-center rounded-full border px-3.5 py-1.5 text-sm font-semibold transition duration-200 active:scale-[0.98] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#7C5CFF]/70 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120]';
		if (active) return `${base} border-[#7C5CFF]/70 bg-[#7C5CFF] text-white shadow-lg shadow-[#7C5CFF]/25`;
		return `${base} border-white/10 bg-white/[0.04] text-white/65 hover:border-white/25 hover:bg-white/[0.08] hover:text-white`;
	}

</script>

<svelte:head>
	<title>TV - Lorivo Media</title>
</svelte:head>

<LorivoShell>
	<section class="relative mx-4 mt-4 overflow-hidden rounded-2xl bg-[#111827] px-6 py-7 sm:mx-6 sm:px-10 sm:py-8 lg:mx-8 lg:px-12 xl:px-16">
		<div class="absolute inset-0 bg-gradient-to-r from-[#0B1120] via-[#0B1120]/70 to-[#0B1120]/30"></div>
		<div class="relative flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-[600px]">
				<h1 class="text-4xl font-bold leading-tight text-white [text-shadow:0_4px_28px_rgba(0,0,0,0.72)] sm:text-5xl xl:text-6xl">TV Shows</h1>
				<p class="mt-3 text-base text-white/60">Browse your TV library.</p>
				<p class="mt-4 text-base leading-relaxed text-white/70">
					{formatVisibleCount(visibleCards.length, visibleSeasons, visibleEpisodes)}
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<LorivoButton variant="primary" onclick={startTVScan} disabled={isScanning || isRefreshing}>
					{isScanning ? 'Scanning...' : 'Scan TV'}
				</LorivoButton>
				<LorivoButton variant="secondary" onclick={runMetadataRefresh} disabled={isScanning || isRefreshing}>
					{isRefreshing ? 'Refreshing...' : 'Refresh Metadata'}
				</LorivoButton>
			</div>
		</div>
	</section>

	{#if isLoading}
		<LorivoPanel title="Loading TV Shows" subtitle="Fetching your TV library from the media APIs." />
	{:else if loadError}
		<section class="px-4 pt-9 sm:px-6 lg:px-8">
			<LorivoEmptyState
				eyebrow="Connection"
				title="Media library unavailable"
				description="Lorivo could not reach the media library service. Check that the server is running, then try again."
			>
				{#snippet primaryAction()}
					<LorivoButton variant="primary" onclick={loadSeries}>Retry</LorivoButton>
				{/snippet}
				{#snippet secondaryAction()}
					<LorivoButton variant="secondary" href="/settings">Settings</LorivoButton>
					<LorivoButton variant="ghost" href="/">Back Home</LorivoButton>
				{/snippet}
			</LorivoEmptyState>
		</section>
	{:else}
		<section class="relative px-4 pt-7 sm:px-6 lg:px-8">
			<div class="flex flex-col gap-3 rounded-2xl border border-white/10 bg-white/[0.04] p-3 shadow-lg shadow-black/20 backdrop-blur sm:p-4 lg:flex-row lg:items-center lg:justify-between">
				<div class="relative w-full lg:max-w-[400px]">
					<Search size={16} class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-white/40" />
					<input
						type="text"
						placeholder="Search TV"
						bind:value={searchValue}
						class="h-10 w-full rounded-full border border-white/10 bg-[#111827]/80 pl-11 pr-4 text-sm text-white shadow-inner shadow-black/20 placeholder:text-white/40 transition focus:border-[#7C5CFF]/50 focus:outline-none focus:ring-2 focus:ring-[#7C5CFF]/25"
					/>
				</div>
				<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between lg:justify-end">
					<span class="inline-flex min-h-9 items-center rounded-full border border-white/10 bg-[#111827]/70 px-3.5 py-1.5 text-sm font-medium text-white/65">
						{formatVisibleCount(visibleCards.length, visibleSeasons, visibleEpisodes)}
					</span>
					<div class="flex flex-wrap gap-2">
						<button type="button" class={chipClass(seriesSort === 'title')} onclick={() => (seriesSort = 'title')}>Title</button>
						<button type="button" class={chipClass(seriesSort === 'year')} onclick={() => (seriesSort = 'year')}>Year</button>
						<button type="button" class={chipClass(seriesSort === 'seasons')} onclick={() => (seriesSort = 'seasons')}>Seasons</button>
						<button type="button" class={chipClass(seriesSort === 'episodes')} onclick={() => (seriesSort = 'episodes')}>Episodes</button>
					</div>
				</div>
			</div>
			{#if actionMessage}
				<p class="mt-3 text-sm text-white/60">{actionMessage}</p>
			{/if}
		</section>

		{#if seriesCards.length === 0}
			<section class="px-4 pt-7 sm:px-6 lg:px-8">
				<LorivoEmptyState
					eyebrow="TV"
					title="No TV shows found yet"
					description="Add a TV library or run a scan, and your shows will appear here."
				>
					{#snippet primaryAction()}
						<LorivoButton variant="primary" onclick={startTVScan} disabled={isScanning || isRefreshing}>
							{isScanning ? 'Scanning...' : 'Scan TV'}
						</LorivoButton>
					{/snippet}
					{#snippet secondaryAction()}
						<LorivoButton variant="secondary" href="/setup">Add Library</LorivoButton>
					{/snippet}
				</LorivoEmptyState>
			</section>
		{:else if visibleCards.length === 0}
			<section class="px-4 pt-7 sm:px-6 lg:px-8">
				<LorivoEmptyState
					compact
					title="No shows match that search"
					description="Try a different search or reset the sort."
				>
					{#snippet primaryAction()}
						<LorivoButton variant="secondary" size="sm" onclick={() => (searchValue = '')}>Clear search</LorivoButton>
					{/snippet}
					{#snippet secondaryAction()}
						<LorivoButton variant="ghost" size="sm" onclick={() => (seriesSort = 'title')}>Reset sort</LorivoButton>
					{/snippet}
				</LorivoEmptyState>
			</section>
		{:else}
			<MediaGrid title="TV Shows" subtitle={formatVisibleCount(visibleCards.length, visibleSeasons, visibleEpisodes)}>
				{#each visibleCards as item (item.id)}
					<LorivoPosterLink
						title={item.title}
						meta={item.meta}
						img={item.posterUrl}
						href={`/tv/${encodeURIComponent(item.id)}`}
					/>
				{/each}
			</MediaGrid>
		{/if}
	{/if}
</LorivoShell>
