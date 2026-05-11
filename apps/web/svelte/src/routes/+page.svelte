<script lang="ts">
	import { onMount } from 'svelte';
	import { getMovies, getSeries, type MovieListItem, type SeriesListItem } from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import {
		getClientHome,
		getLibraries,
		getPlaybackRecent,
		type ClientHomeItem,
		type ClientHomeResponse,
		type LibrariesResponse,
		type PlaybackRecentResponse
	} from '$lib/api/home';
	import {
		buildHomeViewModel,
		createEmptyHomeViewModel,
		type HomeDisplayItem,
		type HomeViewModel
	} from '$lib/home/model';
	import Hero from '$lib/lorivo/Hero.svelte';
	import LandscapeCard from '$lib/lorivo/LandscapeCard.svelte';
	import LorivoButton from '$lib/lorivo/LorivoButton.svelte';
	import LorivoEmptyState from '$lib/lorivo/LorivoEmptyState.svelte';
	import LorivoPanel from '$lib/lorivo/LorivoPanel.svelte';
	import LorivoShell from '$lib/lorivo/LorivoShell.svelte';
	import PosterCard from '$lib/lorivo/PosterCard.svelte';
	import Row from '$lib/lorivo/Row.svelte';

	let isLoading = $state(true);
	let loadNotice = $state('');
	let model = $state<HomeViewModel>(createEmptyHomeViewModel());

	const hero = $derived(model.hero);
	const heroPlayHref = $derived(playHref(hero));
	const heroDetailHref = $derived(detailHref(hero));
	const continueWatching = $derived.by(() => model.continueItems.map(toContinueCard));
	const recentMovies = $derived.by(() => model.movieItems.map(toPosterCard));
	const recentTV = $derived.by(() => model.tvItems.map(toTVPosterCard));

	onMount(() => {
		void loadHome();
	});

	async function loadHome(): Promise<void> {
		isLoading = true;
		loadNotice = '';
		try {
			const { homePayload, playbackRecentPayload, librariesPayload } = await loadHomeData();
			model = buildHomeViewModel({
				homePayload,
				playbackRecentPayload,
				librariesPayload,
				sessionPayload: null,
				forceEmpty: false
			});
		} catch (error) {
			model = createEmptyHomeViewModel();
			loadNotice = formatLoadError(error);
		} finally {
			isLoading = false;
		}
	}

	async function loadHomeData(): Promise<{
		homePayload: ClientHomeResponse;
		playbackRecentPayload: PlaybackRecentResponse;
		librariesPayload: LibrariesResponse;
	}> {
		const [homeResult, playbackRecentResult, librariesResult] = await Promise.allSettled([
			getClientHome(apiClient, 24),
			getPlaybackRecent(apiClient, 12),
			getLibraries(apiClient)
		]);

		const playbackRecentPayload =
			playbackRecentResult.status === 'fulfilled' ? playbackRecentResult.value : { recent: [] };
		const librariesPayload =
			librariesResult.status === 'fulfilled' ? librariesResult.value : { libraries: [] };

		if (homeResult.status === 'fulfilled') {
			return {
				homePayload: homeResult.value,
				playbackRecentPayload,
				librariesPayload
			};
		}

		const fallbackHomePayload = await loadCatalogHomeFallback();
		if (fallbackHomePayload) {
			return {
				homePayload: fallbackHomePayload,
				playbackRecentPayload,
				librariesPayload
			};
		}

		throw homeResult.reason;
	}

	async function loadCatalogHomeFallback(): Promise<ClientHomeResponse | null> {
		const [moviesResult, seriesResult] = await Promise.allSettled([
			getMovies(apiClient, 24),
			getSeries(apiClient, 24)
		]);

		const movies = moviesResult.status === 'fulfilled' ? moviesResult.value.movies || [] : [];
		const series = seriesResult.status === 'fulfilled' ? seriesResult.value.series || [] : [];
		if (movies.length === 0 && series.length === 0) {
			if (moviesResult.status === 'rejected' && seriesResult.status === 'rejected') return null;
		}

		const movieItems = movies.map(movieToHomeItem).filter(hasHomeIdentity);
		const seriesItems = series.map(seriesToHomeItem).filter(hasHomeIdentity);
		const recentlyAddedItems = [...movieItems, ...seriesItems].slice(0, 24);

		return {
			profile: 'lorivo',
			hero: movieItems[0] || seriesItems[0],
			rows: [
				{ id: 'continue', title: 'Continue Watching', items: [] },
				{ id: 'movies', title: 'Movies', items: movieItems },
				{ id: 'tv', title: 'TV Shows', items: seriesItems },
				{ id: 'recently-added', title: 'Recently Added', items: recentlyAddedItems }
			]
		};
	}

	function movieToHomeItem(item: MovieListItem): ClientHomeItem {
		const id = asText(item.id);
		const title = asText(item.metadata?.title) || asText(item.title) || 'Untitled';
		const year = Number(item.metadata?.year || item.year || 0);
		return {
			id,
			kind: 'movie',
			title,
			subtitle: year > 0 ? String(year) : '',
			description: asText(item.metadata?.overview),
			posterUrl: asText(item.metadata?.posterUrl),
			backdropUrl: asText(item.metadata?.backdropUrl)
		};
	}

	function seriesToHomeItem(item: SeriesListItem): ClientHomeItem {
		const id = asText(item.id);
		const title = asText(item.metadata?.title) || asText(item.title) || 'Untitled';
		const seasonCount = Number(item.seasonCount || 0);
		const episodeCount = Number(item.episodeCount || 0);
		const subtitle =
			seasonCount > 0 || episodeCount > 0
				? `${seasonCount} season${seasonCount === 1 ? '' : 's'} - ${episodeCount} episode${episodeCount === 1 ? '' : 's'}`
				: '';
		return {
			id,
			kind: 'series',
			title,
			subtitle,
			description: asText(item.metadata?.overview),
			posterUrl: asText(item.metadata?.posterUrl),
			backdropUrl: asText(item.metadata?.backdropUrl)
		};
	}

	function hasHomeIdentity(item: ClientHomeItem): boolean {
		return Boolean(asText(item.id) || asText(item.title));
	}

	function toContinueCard(item: HomeDisplayItem): {
		title: string;
		sub: string;
		progress: number;
		img?: string;
	} {
		return {
			title: item.title,
			sub: item.meta || item.subtitle || 'Resume playback',
			progress: item.progressPercent,
			img: item.backdropUrl || item.posterUrl
		};
	}

	function toPosterCard(item: HomeDisplayItem): { title: string; img?: string } {
		return {
			title: item.title,
			img: item.posterUrl || item.backdropUrl
		};
	}

	function toTVPosterCard(item: HomeDisplayItem): { title: string; img?: string; ep?: string } {
		return {
			title: item.title,
			img: item.posterUrl || item.backdropUrl,
			ep: item.subtitle || item.meta
		};
	}

	function detailHref(item: HomeDisplayItem): string {
		if (!item.id) return '';
		if (item.kind === 'movie') return `/movies/${encodeURIComponent(item.id)}`;
		if (item.kind === 'series') return `/tv/${encodeURIComponent(item.id)}`;
		return '';
	}

	function playHref(item: HomeDisplayItem): string {
		const mediaSourceId = item.playMediaSourceId || item.mediaSourceId;
		return mediaSourceId ? `/play/${encodeURIComponent(mediaSourceId)}` : '';
	}

	function formatLoadError(error: unknown): string {
		if (error instanceof ApiClientError) return error.userMessage || error.message;
		if (error instanceof Error) return error.message;
		return 'Home could not load.';
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}
</script>

<svelte:head>
	<title>Lorivo - Stream Movies & TV</title>
	<meta name="description" content="Lorivo: your personal streaming hub for movies and TV." />
</svelte:head>

<LorivoShell>
	{#if isLoading}
		<LorivoPanel title="Loading Home" subtitle="Fetching your media library from the local server." />
	{:else if loadNotice}
		<section class="px-4 pt-9 sm:px-6 lg:px-8">
			<LorivoEmptyState
				eyebrow="Connection"
				title="Media library unavailable"
				description="Lorivo could not reach the media library service. Check that the server is running, then try again."
			>
				{#snippet primaryAction()}
					<LorivoButton variant="primary" onclick={loadHome}>Retry</LorivoButton>
				{/snippet}
				{#snippet secondaryAction()}
					<LorivoButton variant="secondary" href="/settings">Settings</LorivoButton>
				{/snippet}
			</LorivoEmptyState>
		</section>
	{:else if model.trueEmpty}
		<section class="px-4 pt-6 sm:px-6 lg:px-8">
			<LorivoEmptyState
				eyebrow="First run"
				title="Build your Lorivo library"
				description="Add your media folders, scan your library, and Lorivo will fill this home screen with what you're watching and what's new."
			>
				<div class="grid gap-3 sm:grid-cols-2 xl:grid-cols-4">
					<a class="next-step-card" href="/setup">
						<strong>Add a library</strong>
						<span>Choose your Movies or TV folder so Lorivo knows where to look.</span>
					</a>
					<a class="next-step-card" href="/movies">
						<strong>Review Movies</strong>
						<span>Use Scan Movies once a movie folder has been added.</span>
					</a>
					<a class="next-step-card" href="/tv">
						<strong>Review TV</strong>
						<span>Use Scan TV once a TV folder has been added.</span>
					</a>
					<a class="next-step-card" href="/settings">
						<strong>Check settings</strong>
						<span>Review configured libraries, providers, and server status.</span>
					</a>
				</div>
				{#snippet primaryAction()}
					<LorivoButton variant="primary" href="/setup">Add a library</LorivoButton>
				{/snippet}
				{#snippet secondaryAction()}
					<LorivoButton variant="secondary" href="/settings">Settings</LorivoButton>
				{/snippet}
			</LorivoEmptyState>
		</section>
	{:else}
		<Hero
			heroPoster={hero.posterUrl}
			heroBackdrop={hero.backdropUrl || hero.posterUrl}
			title={hero.title}
			meta={hero.meta}
			description={hero.description}
			progress={hero.progressPercent}
			progressLabel={hero.progressPercent > 0 ? `${hero.progressPercent}% watched` : ''}
			playHref={heroPlayHref}
			detailHref={heroDetailHref}
		/>
		<Row title="Continue Watching">
			{#if continueWatching.length > 0}
				{#each continueWatching as m (m.title)}
					<LandscapeCard item={m} />
				{/each}
			{:else}
				<div class="min-w-[280px] flex-1">
					<LorivoEmptyState
						compact
						title="Nothing in progress yet."
						description="Start a movie or episode and it will appear here."
					/>
				</div>
			{/if}
		</Row>
		<Row title="Recently Added Movies">
			{#if recentMovies.length > 0}
				{#each recentMovies as m (m.title)}
					<PosterCard img={m.img} title={m.title} />
				{/each}
			{:else}
				<div class="min-w-[280px] flex-1">
					<LorivoEmptyState compact title="No movies have been added yet." description="Add a movie library or run Scan Movies." />
				</div>
			{/if}
		</Row>
		<Row title="Recently Added TV">
			{#if recentTV.length > 0}
				{#each recentTV as m (m.title)}
					<PosterCard img={m.img} title={m.title} ep={m.ep} />
				{/each}
			{:else}
				<div class="min-w-[280px] flex-1">
					<LorivoEmptyState compact title="No TV shows have been added yet." description="Add a TV library or run Scan TV." />
				</div>
			{/if}
		</Row>
		<div class="h-16"></div>
	{/if}
</LorivoShell>

<style>
	.next-step-card {
		display: grid;
		gap: 0.45rem;
		min-height: 9rem;
		align-content: start;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 1rem;
		background: rgb(11 17 32 / 58%);
		padding: 1rem;
		text-decoration: none;
		transition:
			transform 0.2s ease,
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.next-step-card:hover {
		transform: translateY(-2px);
		border-color: rgb(124 92 255 / 45%);
		background: rgb(124 92 255 / 10%);
	}

	.next-step-card strong {
		color: white;
		font-size: 1rem;
	}

	.next-step-card span {
		color: rgb(255 255 255 / 58%);
		font-size: 0.9rem;
		line-height: 1.45;
	}
</style>
