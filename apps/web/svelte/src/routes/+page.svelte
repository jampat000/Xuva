<script lang="ts">
	import { onMount } from 'svelte';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getClientHome, getLibraries, getPlaybackRecent } from '$lib/api/home';
	import {
		buildHomeViewModel,
		createEmptyHomeViewModel,
		type HomeDisplayItem,
		type HomeViewModel
	} from '$lib/home/model';
	import Hero from '$lib/lorivo/Hero.svelte';
	import LandscapeCard from '$lib/lorivo/LandscapeCard.svelte';
	import LorivoPanel from '$lib/lorivo/LorivoPanel.svelte';
	import PosterCard from '$lib/lorivo/PosterCard.svelte';
	import Row from '$lib/lorivo/Row.svelte';
	import TopBar from '$lib/lorivo/TopBar.svelte';

	let isLoading = $state(true);
	let loadError = $state('');
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
		loadError = '';
		try {
			const [homePayload, playbackRecentPayload, librariesPayload] = await Promise.all([
				getClientHome(apiClient, 24),
				getPlaybackRecent(apiClient, 12),
				getLibraries(apiClient)
			]);
			model = buildHomeViewModel({
				homePayload,
				playbackRecentPayload,
				librariesPayload,
				sessionPayload: null,
				forceEmpty: false
			});
		} catch (error) {
			model = createEmptyHomeViewModel();
			loadError = formatLoadError(error);
		} finally {
			isLoading = false;
		}
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
</script>

<svelte:head>
	<title>Lorivo - Stream Movies & TV</title>
	<meta name="description" content="Lorivo: your personal streaming hub for movies and TV." />
</svelte:head>

<div class="min-h-screen bg-[#0B1120] font-sans text-white antialiased">
	<TopBar />
	{#if isLoading}
		<LorivoPanel title="Loading Home" subtitle="Fetching your media library from the local server." />
	{:else if loadError}
		<LorivoPanel title="Home could not load" subtitle={loadError}>
			<div class="flex flex-wrap gap-3">
				<button
					type="button"
					class="inline-flex min-h-11 items-center rounded-xl !bg-[#7C5CFF] px-5 py-3 text-sm font-semibold text-white shadow-lg shadow-[#7C5CFF]/30 transition hover:!bg-[#6a4af0] focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-[#7C5CFF]/70 focus-visible:ring-offset-2 focus-visible:ring-offset-[#0B1120]"
					onclick={loadHome}
				>
					Retry
				</button>
			</div>
		</LorivoPanel>
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
		{#if continueWatching.length > 0}
			<Row title="Continue Watching">
				{#each continueWatching as m (m.title)}
					<LandscapeCard item={m} />
				{/each}
			</Row>
		{/if}
		{#if recentMovies.length > 0}
			<Row title="Recently Added Movies">
				{#each recentMovies as m (m.title)}
					<PosterCard img={m.img} title={m.title} />
				{/each}
			</Row>
		{/if}
		{#if recentTV.length > 0}
			<Row title="Recently Added TV">
				{#each recentTV as m (m.title)}
					<PosterCard img={m.img} title={m.title} ep={m.ep} />
				{/each}
			</Row>
		{/if}
		{#if model.trueEmpty}
			<LorivoPanel
				title="No media library yet"
				subtitle="Add a Movies or TV folder, then scan your library to populate Lorivo."
			/>
		{/if}
		<div class="h-16"></div>
	{/if}
</div>
