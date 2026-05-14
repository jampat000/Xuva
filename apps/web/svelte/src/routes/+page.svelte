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
	import Hero from '\$lib/Xuva/Hero.svelte';
	import LandscapeCard from '\$lib/Xuva/LandscapeCard.svelte';
	import XuvaButton from '\$lib/Xuva/XuvaButton.svelte';
	import XuvaEmptyState from '\$lib/Xuva/XuvaEmptyState.svelte';
	import XuvaPanel from '\$lib/Xuva/XuvaPanel.svelte';
	import XuvaShell from '\$lib/Xuva/XuvaShell.svelte';
	import PosterCard from '\$lib/Xuva/PosterCard.svelte';
	import Row from '\$lib/Xuva/Row.svelte';

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
			profile: 'xuva',
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
	<meta name="description" content="Xuva: your personal streaming hub for movies and TV." />
</svelte:head>

<XuvaShell>
	{#if isLoading}
		<XuvaPanel title="Loading Home" subtitle="Fetching your media library from the local server." />
	{:else if loadNotice}
		<section class="px-4 pt-9 sm:px-6 lg:px-8">
			<XuvaEmptyState
				eyebrow="Connection"
				title="Media library unavailable"
				description="Xuva could not reach the media library service. Check that the server is running, then try again."
			>
				{#snippet primaryAction()}
					<XuvaButton variant="primary" onclick={loadHome}>Retry</XuvaButton>
				{/snippet}
				{#snippet secondaryAction()}
					<XuvaButton variant="secondary" href="/settings">Settings</XuvaButton>
				{/snippet}
			</XuvaEmptyState>
		</section>
	{:else if model.trueEmpty}
		<section class="px-4 pt-6 sm:px-6 lg:px-8">
			<XuvaEmptyState
				eyebrow="First run"
				title="Build your Xuva library"
				description="Add your media folders, scan your library, and Xuva will fill this home screen with what you're watching and what's new."
			>
				<nav class="next-step-list" aria-label="Library setup steps">
					<a class="next-step-row" href="/setup">
						<em class="next-step-row__step">01</em>
						<div>
							<strong>Add a library</strong>
							<span>Choose your Movies or TV folder so Xuva knows where to look.</span>
						</div>
					</a>
					<a class="next-step-row" href="/movies">
						<em class="next-step-row__step">02</em>
						<div>
							<strong>Review Movies</strong>
							<span>Use Scan Movies once a movie folder has been added.</span>
						</div>
					</a>
					<a class="next-step-row" href="/tv">
						<em class="next-step-row__step">03</em>
						<div>
							<strong>Review TV</strong>
							<span>Use Scan TV once a TV folder has been added.</span>
						</div>
					</a>
					<a class="next-step-row" href="/settings">
						<em class="next-step-row__step">04</em>
						<div>
							<strong>Check settings</strong>
							<span>Review your library setup and scan status.</span>
						</div>
					</a>
				</nav>
				{#snippet primaryAction()}
					<XuvaButton variant="primary" href="/setup">Add a library</XuvaButton>
				{/snippet}
				{#snippet secondaryAction()}
					<XuvaButton variant="secondary" href="/settings">Settings</XuvaButton>
				{/snippet}
			</XuvaEmptyState>
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
					<XuvaEmptyState
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
					<XuvaEmptyState compact title="No movies have been added yet." description="Add a movie library or run Scan Movies." />
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
					<XuvaEmptyState compact title="No TV shows have been added yet." description="Add a TV library or run Scan TV." />
				</div>
			{/if}
		</Row>
		<div class="h-16"></div>
	{/if}
</XuvaShell>

<style>
	.next-step-list {
		display: grid;
		gap: 0;
	}

	.next-step-row {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		column-gap: 0.9rem;
		row-gap: 0.25rem;
		align-items: start;
		padding: 0.95rem 0;
		border-top: 1px solid rgb(255 255 255 / 8%);
		text-decoration: none;
		transition:
			border-color 0.2s ease,
			background 0.2s ease;
	}

	.next-step-list > :first-child {
		padding-top: 0;
		border-top: 0;
	}

	.next-step-row:hover {
		background: rgb(124 92 255 / 4%);
	}

	.next-step-row__step {
		display: inline-flex;
		align-items: center;
		justify-content: center;
		width: 2rem;
		height: 2rem;
		margin-top: 0.1rem;
		border: 1px solid rgb(255 255 255 / 12%);
		border-radius: 0.35rem;
		background: rgb(255 255 255 / 2%);
		color: rgb(255 255 255 / 48%);
		font-size: 0.72rem;
		font-style: normal;
		font-weight: 800;
		letter-spacing: 0.08em;
	}

	.next-step-row strong {
		display: block;
		color: white;
		font-size: 0.98rem;
		line-height: 1.2;
	}

	.next-step-row span {
		display: block;
		margin-top: 0.25rem;
		color: rgb(255 255 255 / 58%);
		font-size: 0.88rem;
		line-height: 1.45;
	}
</style>
