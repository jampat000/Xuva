<script lang="ts">
	import { onMount } from 'svelte';
	import { Search } from 'lucide-svelte';
	import {
		getMovies,
		refreshMetadataBatch,
		scanMovies,
		type MovieListItem
	} from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import LorivoButton from '$lib/lorivo/LorivoButton.svelte';
	import LorivoPanel from '$lib/lorivo/LorivoPanel.svelte';
	import LorivoPosterLink from '$lib/lorivo/LorivoPosterLink.svelte';
	import LorivoShell from '$lib/lorivo/LorivoShell.svelte';
	import MediaGrid from '$lib/lorivo/MediaGrid.svelte';
	import {
		buildMovieCards,
		filterAndSortMovieCards,
		type MovieSort
	} from '$lib/browse/model';

	let isLoading = $state(true);
	let isScanning = $state(false);
	let isRefreshing = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let searchValue = $state('');
	let movieSort = $state<MovieSort>('title');
	let movieRows = $state<MovieListItem[]>([]);

	const movieCards = $derived.by(() => buildMovieCards(movieRows));
	const visibleCards = $derived.by(() =>
		filterAndSortMovieCards(movieCards, searchValue, 'all', movieSort)
	);
	const renderedCards = $derived.by(() => visibleCards);

	onMount(() => {
		try {
			const params = new URL(window.location.href).searchParams;
			const q = params.get('q');
			if (q) searchValue = q;
		} catch {
			// Keep the page usable if URL parsing fails.
		}
		void loadMovies();
	});

	async function loadMovies(): Promise<void> {
		isLoading = true;
		loadError = '';

		try {
			const moviesPayload = await getMovies(apiClient, 500);
			movieRows = moviesPayload.movies || [];
		} catch (error) {
			loadError = formatLoadError(error);
		} finally {
			isLoading = false;
		}
	}

	async function startMovieScan(): Promise<void> {
		isScanning = true;
		actionMessage = '';
		try {
			await scanMovies(apiClient, 50);
			actionMessage = 'Movie scan started. Refreshing the wall...';
			await loadMovies();
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
			const result = await refreshMetadataBatch('movie', apiClient, 25);
			if (Array.isArray(result.warnings) && result.warnings.length > 0) {
				actionMessage = result.warnings.slice(0, 3).join(' | ');
			} else {
				actionMessage = 'Metadata refresh accepted.';
			}
			await loadMovies();
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

	function formatTitleCount(value: number): string {
		const label = value === 1 ? 'title' : 'titles';
		return `${formatCount(value)} visible ${label}`;
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
		return 'Movies could not load.';
	}

	function chipClass(active: boolean): string {
		const base =
			'inline-flex min-h-9 items-center rounded-full border px-3.5 py-1.5 text-sm font-semibold transition duration-200 active:scale-[0.98] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#7C5CFF]/70 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120]';
		if (active) return `${base} border-[#7C5CFF]/70 bg-[#7C5CFF] text-white shadow-lg shadow-[#7C5CFF]/25`;
		return `${base} border-white/10 bg-white/[0.04] text-white/65 hover:border-white/25 hover:bg-white/[0.08] hover:text-white`;
	}

</script>

<svelte:head>
	<title>Movies - Lorivo Media</title>
</svelte:head>

<LorivoShell>
	<section class="relative mx-4 mt-4 overflow-hidden rounded-2xl bg-[#111827] px-6 py-7 sm:mx-6 sm:px-10 sm:py-8 lg:mx-8 lg:px-12 xl:px-16">
		<div class="absolute inset-0 bg-gradient-to-r from-[#0B1120] via-[#0B1120]/70 to-[#0B1120]/30"></div>
		<div class="relative flex flex-col gap-6 lg:flex-row lg:items-end lg:justify-between">
			<div class="max-w-[600px]">
				<h1 class="text-4xl font-bold leading-tight text-white [text-shadow:0_4px_28px_rgba(0,0,0,0.72)] sm:text-5xl xl:text-6xl">Movies</h1>
				<p class="mt-3 text-base text-white/60">Browse your movie library.</p>
				<p class="mt-4 text-base leading-relaxed text-white/70">
					{formatTitleCount(renderedCards.length)}
				</p>
			</div>
			<div class="flex flex-wrap gap-3">
				<LorivoButton variant="primary" onclick={startMovieScan} disabled={isScanning || isRefreshing}>
					{isScanning ? 'Scanning...' : 'Scan Movies'}
				</LorivoButton>
				<LorivoButton variant="secondary" onclick={runMetadataRefresh} disabled={isScanning || isRefreshing}>
					{isRefreshing ? 'Refreshing...' : 'Refresh Metadata'}
				</LorivoButton>
			</div>
		</div>
	</section>

	{#if isLoading}
		<LorivoPanel title="Loading Movies" subtitle="Fetching your movie library from the media APIs." />
	{:else if loadError}
		<LorivoPanel title="Movies could not load" subtitle={loadError}>
			<div class="flex flex-wrap gap-3">
				<LorivoButton variant="secondary" onclick={loadMovies}>Retry</LorivoButton>
				<LorivoButton variant="ghost" href="/">Back to Home</LorivoButton>
			</div>
		</LorivoPanel>
	{:else}
		<section class="relative px-4 pt-7 sm:px-6 lg:px-8">
			<div class="flex flex-col gap-3 rounded-2xl border border-white/10 bg-white/[0.04] p-3 shadow-lg shadow-black/20 backdrop-blur sm:p-4 lg:flex-row lg:items-center lg:justify-between">
				<div class="relative w-full lg:max-w-[400px]">
					<Search size={16} class="pointer-events-none absolute left-4 top-1/2 -translate-y-1/2 text-white/40" />
					<input
						type="text"
						placeholder="Search movies"
						bind:value={searchValue}
						class="h-10 w-full rounded-full border border-white/10 bg-[#111827]/80 pl-11 pr-4 text-sm text-white shadow-inner shadow-black/20 placeholder:text-white/40 transition focus:border-[#7C5CFF]/50 focus:outline-none focus:ring-2 focus:ring-[#7C5CFF]/25"
					/>
				</div>
				<div class="flex flex-col gap-3 sm:flex-row sm:items-center sm:justify-between lg:justify-end">
					<span class="inline-flex min-h-9 items-center rounded-full border border-white/10 bg-[#111827]/70 px-3.5 py-1.5 text-sm font-medium text-white/65">
						{formatTitleCount(renderedCards.length)}
					</span>
					<div class="flex flex-wrap gap-2">
						<button type="button" class={chipClass(movieSort === 'title')} onclick={() => (movieSort = 'title')}>Title</button>
						<button type="button" class={chipClass(movieSort === 'year')} onclick={() => (movieSort = 'year')}>Year</button>
					</div>
				</div>
			</div>
			{#if actionMessage}
				<p class="mt-3 text-sm text-white/60">{actionMessage}</p>
			{/if}
		</section>

		{#if movieCards.length === 0}
			<LorivoPanel title="No movies found" subtitle="Try adding a movie library or running a scan." />
		{:else if renderedCards.length === 0}
			<LorivoPanel title="No matching movies" subtitle="Try another search term." />
		{:else}
			<MediaGrid title="Movies" subtitle={formatTitleCount(renderedCards.length)}>
				{#each renderedCards as item (item.id)}
					<LorivoPosterLink
						title={item.title}
						meta={item.meta}
						img={item.posterUrl}
						href={`/movies/${encodeURIComponent(item.id)}`}
					/>
				{/each}
			</MediaGrid>
		{/if}
	{/if}
</LorivoShell>
