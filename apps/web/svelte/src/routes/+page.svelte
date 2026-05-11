<script lang="ts">
	import { page } from '$app/state';
	import { onMount } from 'svelte';
	import { getAuthSession, type AuthSessionResponse } from '$lib/api/auth';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import {
		getClientHome,
		getLibraries,
		getPlaybackRecent,
		type ClientHomeResponse,
		type LibrariesResponse,
		type PlaybackRecentResponse
	} from '$lib/api/home';
	import LorivoMediaHome from '$lib/components/home/LorivoMediaHome.svelte';
	import {
		buildHomeViewModel,
		resolveForceEmptyMode,
		resolvePreviewMode,
		type HomeDisplayItem,
		type HomeViewModel
	} from '$lib/home/model';

	interface LorivoHeroProps {
		title: string;
		meta: string;
		description: string;
		progressLabel: string;
		progressPercent: number;
		runtime: string;
		backdropUrl: string;
		posterUrl: string;
		resumeHref: string;
		detailsHref: string;
	}

	interface LorivoWideItemProps {
		title: string;
		context: string;
		progressLabel: string;
		progressPercent: number;
		imageUrl: string;
		playHref: string;
	}

	interface LorivoPosterItemProps {
		title: string;
		meta: string;
		imageUrl: string;
		href: string;
	}

	const previewMode = resolvePreviewMode(page.url.searchParams);
	const forceEmpty = resolveForceEmptyMode(page.url.searchParams);
	const emptyPayload = {};
	const fallbackBackdrop = svgDataUri(
		'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 1200 675"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop stop-color="#050b18"/><stop offset="0.55" stop-color="#111827"/><stop offset="1" stop-color="#33218c"/></linearGradient></defs><rect width="1200" height="675" fill="url(#g)"/><circle cx="885" cy="155" r="260" fill="#7c5cff" opacity="0.18"/><text x="84" y="392" fill="#f6f7ff" font-family="Inter,Arial,sans-serif" font-size="86" font-weight="750">Lorivo Media</text></svg>'
	);
	const fallbackPoster = svgDataUri(
		'<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 600 900"><defs><linearGradient id="g" x1="0" y1="0" x2="1" y2="1"><stop stop-color="#070d1b"/><stop offset="0.62" stop-color="#121a30"/><stop offset="1" stop-color="#4a33c5"/></linearGradient></defs><rect width="600" height="900" fill="url(#g)"/><circle cx="470" cy="150" r="190" fill="#7c5cff" opacity="0.18"/><text x="58" y="486" fill="#f6f7ff" font-family="Inter,Arial,sans-serif" font-size="62" font-weight="750">Lorivo</text></svg>'
	);

	let viewModel = $state<HomeViewModel>(
		buildHomeViewModel({
			homePayload: emptyPayload,
			playbackRecentPayload: emptyPayload,
			librariesPayload: emptyPayload,
			sessionPayload: null,
			previewMode,
			forceEmpty
		})
	);

	const hero = $derived(toHero(viewModel.hero));
	const continueItems = $derived(viewModel.continueItems.map(toWideItem));
	const movieItems = $derived(viewModel.movieItems.map(toPosterItem));
	const tvItems = $derived(viewModel.tvItems.map(toPosterItem));

	onMount(() => {
		void loadHome();
	});

	async function loadHome(): Promise<void> {
		try {
			const [homePayload, playbackRecentPayload, librariesPayload, sessionPayload] =
				await Promise.all([
					getClientHome(apiClient, 24).catch((error: unknown) => fallbackHome(error)),
					getPlaybackRecent(apiClient, 12).catch((error: unknown) => fallbackRecent(error)),
					getLibraries(apiClient).catch((error: unknown) => fallbackLibraries(error)),
					getAuthSession(apiClient).catch((error: unknown) => fallbackSession(error))
				]);

			viewModel = buildHomeViewModel({
				homePayload,
				playbackRecentPayload,
				librariesPayload,
				sessionPayload,
				previewMode,
				forceEmpty
			});
		} catch {
			viewModel = buildHomeViewModel({
				homePayload: emptyPayload,
				playbackRecentPayload: emptyPayload,
				librariesPayload: emptyPayload,
				sessionPayload: null,
				previewMode,
				forceEmpty: previewMode ? forceEmpty : true
			});
		}
	}

	function fallbackHome(error: unknown): ClientHomeResponse {
		if (previewMode || isApiStatus(error, 401)) return {};
		throw error;
	}

	function fallbackRecent(error: unknown): PlaybackRecentResponse {
		if (previewMode || isApiStatus(error, 401)) return {};
		throw error;
	}

	function fallbackLibraries(error: unknown): LibrariesResponse {
		if (previewMode || isApiStatus(error, 401)) return { libraries: [] };
		throw error;
	}

	function fallbackSession(error: unknown): AuthSessionResponse | null {
		if (previewMode || isApiStatus(error, 401)) return null;
		throw error;
	}

	function toHero(item: HomeDisplayItem): LorivoHeroProps {
		const detailsHref = detailsHrefFor(item);
		return {
			title: item.title || 'Lorivo Media',
			meta: item.meta || item.subtitle || 'Personal media',
			description:
				item.description ||
				'Your movies and TV shows will appear here once Lorivo has scanned your media folders.',
			progressLabel: progressLabelFor(item) || 'Ready to watch',
			progressPercent: item.progressPercent,
			runtime: item.subtitle || item.meta || 'Local library',
			backdropUrl: item.backdropUrl || item.posterUrl || fallbackBackdrop,
			posterUrl: item.posterUrl || item.backdropUrl || fallbackPoster,
			resumeHref: playbackHrefFor(item, detailsHref),
			detailsHref
		};
	}

	function toWideItem(item: HomeDisplayItem): LorivoWideItemProps {
		const detailsHref = detailsHrefFor(item);
		return {
			title: item.title,
			context: item.subtitle || item.meta || 'Resume',
			progressLabel: progressLabelFor(item),
			progressPercent: item.progressPercent,
			imageUrl: item.backdropUrl || item.posterUrl || fallbackBackdrop,
			playHref: playbackHrefFor(item, detailsHref)
		};
	}

	function toPosterItem(item: HomeDisplayItem): LorivoPosterItemProps {
		return {
			title: item.title,
			meta: item.meta || item.subtitle || mediaLabelFor(item),
			imageUrl: item.posterUrl || item.backdropUrl || fallbackPoster,
			href: detailsHrefFor(item)
		};
	}

	function playbackHrefFor(item: HomeDisplayItem, fallbackHref: string): string {
		const mediaSourceId = item.playMediaSourceId || item.mediaSourceId;
		if (!mediaSourceId || item.kind === 'empty') return fallbackHref;
		return `/play/${encodeURIComponent(mediaSourceId)}`;
	}

	function detailsHrefFor(item: HomeDisplayItem): string {
		if (!item.id || item.kind === 'empty') return '/settings';
		if (item.kind === 'series' || item.kind === 'episode') return `/tv/${encodeURIComponent(item.id)}`;
		return `/movies/${encodeURIComponent(item.id)}`;
	}

	function progressLabelFor(item: HomeDisplayItem): string {
		const value = Math.max(0, Math.min(100, Number(item.progressPercent) || 0));
		return value > 0 ? `Resume from ${Math.round(value)}%` : item.meta;
	}

	function mediaLabelFor(item: HomeDisplayItem): string {
		if (item.kind === 'series' || item.kind === 'episode') return 'TV';
		if (item.kind === 'movie') return 'Movie';
		return 'Media';
	}

	function isApiStatus(error: unknown, status: number): boolean {
		return error instanceof ApiClientError && error.status === status;
	}

	function svgDataUri(svg: string): string {
		return `data:image/svg+xml,${encodeURIComponent(svg)}`;
	}
</script>

<svelte:head>
	<title>Lorivo Media</title>
	<meta
		name="description"
		content="Lorivo Media is a local-first home for streaming your personal movies and TV library."
	/>
</svelte:head>

<LorivoMediaHome
	{hero}
	{continueItems}
	{movieItems}
	{tvItems}
	previewMode={previewMode}
/>
