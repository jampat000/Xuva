<script lang="ts">
	import { onMount } from 'svelte';
	import { Play } from 'lucide-svelte';
	import { scanTV } from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getSessions, type SessionItem } from '$lib/api/operator';
	import {
		getDeviceProfiles,
		getMediaSourceDetail,
		getMediaSourceSubtitles,
		getMediaSourceTracks,
		getPlaybackDecision,
		getPlaybackRoute,
		getPlaybackState,
		getSeriesDetail,
		startMediaProbe,
		listMediaSources,
		type DeviceProfile,
		type EpisodeBrief,
		type MediaSourceItem,
		type MediaSourceSubtitlesResponse,
		type MediaSourceTracksResponse,
		type PlaybackDecisionResponse,
		type PlaybackRouteResponse,
		type PlaybackStateResponse,
		type SeasonDetail,
		type SeriesDetailResponse
	} from '$lib/api/details';
	import DetailHero from '$lib/Xuva/DetailHero.svelte';
	import DetailSection from '$lib/Xuva/DetailSection.svelte';
	import XuvaButton from '$lib/Xuva/XuvaButton.svelte';
	import XuvaEmptyState from '$lib/Xuva/XuvaEmptyState.svelte';
	import XuvaPanel from '$lib/Xuva/XuvaPanel.svelte';
	import XuvaShell from '$lib/Xuva/XuvaShell.svelte';
	import {
		buildSeriesCardMeta,
		cleanDescription,
		episodeLabel,
		formatBitrate,
		formatBytes,
		formatDuration,
		formatResolution,
		formatRuntime,
		formatTrackSummary,
		isResumeState,
		listItemTitle,
		playbackModeLabel,
		playbackReasonLabel,
		resolveArtworkUrl,
		sourceQualityLabel,
		watchedLabel
	} from '$lib/details/model';

	interface EpisodeItemModel {
		episodeId: string;
		seasonId: string;
		seasonNumber: number;
		episodeNumber: number;
		episodeEnd: number;
		title: string;
		label: string;
		versionCount: number;
		needsReview: boolean;
		mediaSourceId: string;
		state: PlaybackStateResponse | null;
	}

	interface SeasonModel {
		seasonId: string;
		seasonNumber: number;
		episodes: EpisodeItemModel[];
	}

	interface EpisodeSourceModel {
		mediaSourceId: string;
		source: MediaSourceItem | null;
		tracks: MediaSourceTracksResponse;
		subtitles: MediaSourceSubtitlesResponse;
		decision: PlaybackDecisionResponse | null;
		state: PlaybackStateResponse | null;
	}

	let { params } = $props<{ params: { id: string } }>();

	let isLoading = $state(true);
	let isScanning = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let series = $state<SeriesDetailResponse | null>(null);
	let seasons = $state<SeasonModel[]>([]);
	let selectedEpisodeId = $state('');
	let selectedSource = $state<EpisodeSourceModel | null>(null);
	let selectedSourceLoading = $state(false);
	let selectedSourceError = $state('');
	let deviceProfiles = $state<DeviceProfile[]>([]);
	let analyzerProfileId = $state('web');
	let analyzerRouteType = $state<'lan' | 'remote'>('lan');
	let analyzerMaxBitrate = $state('');
	let analyzerAudioTrackIndex = $state(0);
	let analyzerSubtitleTrackIndex = $state(0);
	let analyzerSubtitleEnabled = $state(false);
	let analyzerBusy = $state(false);
	let analyzerError = $state('');
	let analyzerDecision = $state<PlaybackDecisionResponse | null>(null);
	let analyzerRoute = $state<PlaybackRouteResponse | null>(null);
	let analyzerMatrixBusy = $state(false);
	let analyzerMatrixError = $state('');
	let analyzerMatrix = $state<Array<{ profileId: string; profileName: string; mode: string; reason: string }>>([]);
	let probingSourceId = $state('');
	let probeStatusSourceId = $state('');
	let probeStatusTone = $state<'info' | 'success' | 'error'>('info');
	let probeStatusMessage = $state('');
	let livePlaybackSourceId = $state('');
	let livePlaybackSession = $state<SessionItem | null>(null);
	let livePlaybackState = $state<PlaybackStateResponse | null>(null);
	let livePlaybackError = $state('');

	const seriesID = $derived.by(() => asText(series?.id) || asText(params.id));
	const seriesTitle = $derived.by(() =>
		listItemTitle(series?.metadata?.title, series?.title || asText(params.id) || 'TV Show')
	);
	const seriesPosterUrl = $derived.by(() =>
		resolveArtworkUrl('series', seriesID, series?.metadata?.posterUrl, 'poster')
	);
	const seriesBackdropUrl = $derived.by(() =>
		resolveArtworkUrl('series', seriesID, series?.metadata?.backdropUrl, 'backdrop')
	);
	const seriesSeasonCount = $derived.by(() => Number(series?.seasonCount || 0));
	const seriesEpisodeCount = $derived.by(() => Number(series?.episodeCount || 0));
	const seriesMeta = $derived.by(() => buildSeriesCardMeta(seriesSeasonCount, seriesEpisodeCount));
	const seriesOverview = $derived.by(() =>
		cleanDescription(series?.metadata?.overview || '', 360) ||
		'Browse seasons and episodes, then start playback.'
	);
	const selectedEpisode = $derived.by(() => findEpisodeById(seasons, selectedEpisodeId));
	const selectedPlayHref = $derived.by(() => {
		const mediaSourceId = asText(selectedEpisode?.mediaSourceId);
		return mediaSourceId ? playHrefWithOptions(mediaSourceId) : '';
	});
	const selectedStartHref = $derived.by(() => {
		const mediaSourceId = asText(selectedEpisode?.mediaSourceId);
		return mediaSourceId ? playHrefWithOptions(mediaSourceId, true) : '';
	});
	const selectedPlayLabel = $derived.by(() =>
		isResumeState(selectedEpisode?.state || null) ? 'Resume' : 'Play'
	);
	const selectedRuntime = $derived.by(() =>
		formatRuntime(Number(selectedSource?.source?.durationSeconds || 0))
	);
	const selectedProgress = $derived.by(() => {
		const percent = Number(selectedEpisode?.state?.percent ?? 0);
		if (Number.isFinite(percent) && percent > 1) return Math.round(Math.max(0, Math.min(100, percent)));
		if (Number.isFinite(percent) && percent > 0) return Math.round(percent * 100);
		return 0;
	});

	onMount(() => {
		void loadSeriesDetails();
		const ticker = window.setInterval(() => {
			void refreshLivePlayback();
		}, 2500);
		return () => window.clearInterval(ticker);
	});

	$effect(() => {
		const mediaSourceId = asText(selectedEpisode?.mediaSourceId);
		if (!mediaSourceId) {
			selectedSource = null;
			selectedSourceError = '';
			selectedSourceLoading = false;
			analyzerDecision = null;
			analyzerRoute = null;
			analyzerMatrix = [];
			return;
		}
		void loadSelectedEpisodeSource(mediaSourceId, selectedEpisode?.state || null);
	});

	$effect(() => {
		const source = selectedSource;
		if (!source) return;
		analyzerAudioTrackIndex = preferredTrackIndex(source.tracks.audioTracks);
		analyzerSubtitleTrackIndex = preferredTrackIndex(source.tracks.subtitleTracks);
		analyzerSubtitleEnabled = false;
		void refreshAnalyzer();
		void refreshDeviceMatrix();
	});

	async function loadSeriesDetails(): Promise<void> {
		isLoading = true;
		loadError = '';
		selectedSource = null;
		selectedSourceError = '';
		try {
			const [seriesPayload, mediaSourcePayload, profilesPayload] = await Promise.all([
				getSeriesDetail(asText(params.id), apiClient),
				listMediaSources(apiClient, 1500).catch(() => ({ mediaSources: [] })),
				getDeviceProfiles(apiClient).catch(() => ({ profiles: [] }))
			]);
			deviceProfiles = profilesPayload.profiles || [];

			series = seriesPayload;

			const knownSourceIDs = new Set<string>();
			for (const item of mediaSourcePayload.mediaSources || []) {
				const id = asText(item.id);
				if (id) knownSourceIDs.add(id);
			}

			const builtSeasons = await buildSeasonModels(seriesPayload.seasons || [], knownSourceIDs);
			seasons = builtSeasons;

			const firstPlayableEpisode = builtSeasons
				.flatMap((item) => item.episodes)
				.find((episode) => Boolean(episode.mediaSourceId));
			selectedEpisodeId = asText(firstPlayableEpisode?.episodeId);
		} catch (error) {
			if (isApiStatus(error, 404)) {
				series = null;
				seasons = [];
				selectedEpisodeId = '';
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
			actionMessage = 'TV scan started.';
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isScanning = false;
		}
	}

	async function buildSeasonModels(
		items: SeasonDetail[],
		knownSourceIDs: Set<string>
	): Promise<SeasonModel[]> {
		const output: SeasonModel[] = [];
		for (const season of items || []) {
			const episodeModels = await Promise.all(
				(season.episodes || []).map((episode) => buildEpisodeModel(season, episode, knownSourceIDs))
			);
			output.push({
				seasonId: asText(season.id) || `season-${Number(season.seasonNumber || 0)}`,
				seasonNumber: Number(season.seasonNumber || 0),
				episodes: episodeModels.filter((item): item is EpisodeItemModel => Boolean(item))
			});
		}
		return output.sort((a, b) => a.seasonNumber - b.seasonNumber);
	}

	async function buildEpisodeModel(
		season: SeasonDetail,
		episode: EpisodeBrief,
		knownSourceIDs: Set<string>
	): Promise<EpisodeItemModel | null> {
		const episodeID = asText(episode.id);
		if (!episodeID) return null;
		const mediaSourceId = asText(episode.versions?.[0]?.mediaSourceId);
		const canFetchState = mediaSourceId && knownSourceIDs.has(mediaSourceId);
		const state = canFetchState ? await getPlaybackState(mediaSourceId, apiClient).catch(() => null) : null;

		return {
			episodeId: episodeID,
			seasonId: asText(season.id) || '',
			seasonNumber: Number(episode.seasonNumber || season.seasonNumber || 0),
			episodeNumber: Number(episode.episodeNumber || 0),
			episodeEnd: Number(episode.episodeEnd || 0),
			title: listItemTitle(episode.title, episodeLabel(episode)),
			label: episodeLabel(episode),
			versionCount: Math.max(0, Number(episode.versionCount || 0)),
			needsReview: Boolean(episode.needsReview),
			mediaSourceId,
			state
		};
	}

	async function loadSelectedEpisodeSource(
		mediaSourceId: string,
		state: PlaybackStateResponse | null
	): Promise<void> {
		selectedSourceLoading = true;
		selectedSourceError = '';
		try {
			const [source, tracks, subtitles, decision] = await Promise.all([
				getMediaSourceDetail(mediaSourceId, apiClient).catch(() => null),
				getMediaSourceTracks(mediaSourceId, apiClient).catch(() => ({ audioTracks: [], subtitleTracks: [] })),
				getMediaSourceSubtitles(mediaSourceId, apiClient).catch(() => ({ sidecars: [] })),
				getPlaybackDecision(
					mediaSourceId,
					{ clientProfile: 'web', routeType: 'remote', supportsAdaptive: true },
					apiClient
				).catch(() => null)
			]);
			selectedSource = {
				mediaSourceId,
				source,
				tracks,
				subtitles,
				decision,
				state
			};
		} catch (error) {
			selectedSourceError = formatLoadError(error);
			selectedSource = null;
		} finally {
			selectedSourceLoading = false;
		}
	}

	function selectEpisode(episodeId: string): void {
		selectedEpisodeId = asText(episodeId);
	}

	function isEpisodeSelected(episodeId: string): boolean {
		return asText(episodeId) === asText(selectedEpisodeId);
	}

	function findEpisodeById(items: SeasonModel[], episodeId: string): EpisodeItemModel | null {
		const target = asText(episodeId);
		if (!target) return null;
		for (const season of items) {
			for (const episode of season.episodes) {
				if (episode.episodeId === target) return episode;
			}
		}
		return null;
	}

	function episodeSummary(episode: EpisodeItemModel): string {
		if (episode.needsReview) return 'Needs review';
		if (episode.state?.watched) return 'Watched';
		if (isResumeState(episode.state)) return 'Resume available';
		if (episode.versionCount > 0) return `${episode.versionCount} version${episode.versionCount === 1 ? '' : 's'}`;
		return 'No source';
	}

	function sourceEpisodeMeta(): string {
		if (!selectedEpisode) return '';
		const parts = [selectedEpisode.label];
		if (selectedRuntime) parts.push(selectedRuntime);
		return parts.join(' - ');
	}

	function currentPlaybackOptions() {
		const source = selectedSource;
		const profileId = asText(analyzerProfileId) || 'web';
		const profile = deviceProfiles.find((item) => asText(item.id) === profileId);
		const maxBitrate = Number.parseInt(asText(analyzerMaxBitrate), 10);
		const normalizedMaxBitrate = Number.isFinite(maxBitrate) && maxBitrate > 0 ? maxBitrate : 0;
		const audioTrack = Number.isFinite(analyzerAudioTrackIndex) ? analyzerAudioTrackIndex : 0;
		const subtitleTrack = Number.isFinite(analyzerSubtitleTrackIndex) ? analyzerSubtitleTrackIndex : 0;
		const audio = source?.tracks.audioTracks?.find((item) => Number(item.index || 0) === audioTrack);
		const subtitle = source?.tracks.subtitleTracks?.find((item) => Number(item.index || 0) === subtitleTrack);
		return {
			clientProfile: profileId,
			routeType: analyzerRouteType,
			maxNetworkBitrate: normalizedMaxBitrate,
			audioTrackIndex: audioTrack,
			audioCodec: asText(audio?.codec),
			audioChannels: Number(audio?.channels || 0),
			subtitleTrackIndex: subtitleTrack,
			subtitleCodec: asText(subtitle?.codec),
			subtitleTrackActive: analyzerSubtitleEnabled,
			supportsAdaptive: Boolean(profile?.supportsHls)
		};
	}

	function preferredTrackIndex(
		tracks: Array<{ index?: number; default?: boolean }> | null | undefined
	): number {
		if (!Array.isArray(tracks) || tracks.length === 0) return 0;
		const preferred = tracks.find((item) => Boolean(item?.default));
		if (preferred && Number.isFinite(Number(preferred.index))) return Number(preferred.index);
		const first = tracks[0];
		return Number.isFinite(Number(first?.index)) ? Number(first?.index) : 0;
	}

	async function refreshAnalyzer(): Promise<void> {
		const source = selectedSource;
		if (!source) return;
		analyzerBusy = true;
		analyzerError = '';
		try {
			const options = currentPlaybackOptions();
			const [decision, route] = await Promise.all([
				getPlaybackDecision(source.mediaSourceId, options, apiClient),
				getPlaybackRoute(source.mediaSourceId, options, apiClient)
			]);
			analyzerDecision = decision;
			analyzerRoute = route;
		} catch (error) {
			analyzerError = formatLoadError(error);
			analyzerDecision = null;
			analyzerRoute = null;
		} finally {
			analyzerBusy = false;
		}
	}

	async function refreshDeviceMatrix(): Promise<void> {
		const source = selectedSource;
		if (!source || deviceProfiles.length === 0) {
			analyzerMatrix = [];
			return;
		}
		analyzerMatrixBusy = true;
		analyzerMatrixError = '';
		try {
			const rows = await Promise.all(
				deviceProfiles.map(async (profile) => {
					const profileId = asText(profile.id) || 'web';
					const decision = await getPlaybackDecision(
						source.mediaSourceId,
						{
							clientProfile: profileId,
							routeType: analyzerRouteType,
							maxNetworkBitrate: 0,
							audioTrackIndex: analyzerAudioTrackIndex,
							subtitleTrackIndex: analyzerSubtitleTrackIndex,
							subtitleTrackActive: analyzerSubtitleEnabled,
							supportsAdaptive: Boolean(profile.supportsHls)
						},
						apiClient
					);
					return {
						profileId,
						profileName: asText(profile.name) || profileId,
						mode: asText(decision.mode) || 'Pending',
						reason: asText(decision.reasonText || decision.reason) || 'No reason available.'
					};
				})
			);
			analyzerMatrix = rows;
		} catch (error) {
			analyzerMatrixError = formatLoadError(error);
			analyzerMatrix = [];
		} finally {
			analyzerMatrixBusy = false;
		}
	}

	async function runMediaCheck(mediaSourceId: string): Promise<void> {
		const id = asText(mediaSourceId);
		if (!id || probingSourceId) return;
		probingSourceId = id;
		probeStatusSourceId = id;
		probeStatusTone = 'info';
		probeStatusMessage = 'Queued media check. This can take a moment on large files.';
		actionMessage = '';
		try {
			await startMediaProbe(id, apiClient);
			probeStatusTone = 'success';
			probeStatusMessage = 'Media check started. Refreshing source details.';
			actionMessage = 'Media check started. Refreshing source details...';
			await loadSelectedEpisodeSource(id, selectedEpisode?.state || null);
			await refreshAnalyzer();
			await refreshDeviceMatrix();
		} catch (error) {
			probeStatusTone = 'error';
			probeStatusMessage = formatLoadError(error);
			actionMessage = formatLoadError(error);
		} finally {
			probingSourceId = '';
		}
	}

	function playHrefWithOptions(sourceId: string, startFromZero = false): string {
		const id = asText(sourceId);
		if (!id) return '';
		const params = new URLSearchParams();
		const options = currentPlaybackOptions();
		params.set('clientProfile', options.clientProfile);
		params.set('routeType', options.routeType);
		params.set('supportsAdaptive', options.supportsAdaptive ? 'true' : 'false');
		params.set('audioTrackIndex', String(options.audioTrackIndex));
		params.set('subtitleTrackIndex', String(options.subtitleTrackIndex));
		params.set('subtitleTrackActive', options.subtitleTrackActive ? 'true' : 'false');
		params.set('autoplayIntent', '1');
		params.set('strictAutoplay', '1');
		params.set('forcePlayable', 'true');
		if (options.maxNetworkBitrate > 0) params.set('maxNetworkBitrate', String(options.maxNetworkBitrate));
		if (startFromZero) params.set('start', '0');
		return `/play/${encodeURIComponent(id)}?${params.toString()}`;
	}

	function playbackRouteLabel(mode: string): string {
		const value = asText(mode).toLowerCase();
		const normalized = value.replace(/[^a-z]+/g, ' ').trim();
		if (!value) return '';
		if (
			normalized === 'updating' ||
			normalized === 'starting' ||
			normalized === 'playing' ||
			normalized === 'paused' ||
			normalized === 'stopped' ||
			normalized === 'idle' ||
			normalized === 'pending' ||
			normalized.includes('updat') ||
			normalized.includes('wait') ||
			normalized.includes('check') ||
			normalized.includes('load')
		) {
			return '';
		}
		if (value.includes('audio-transcode')) return 'Transcoding audio';
		if (value.includes('transcode')) return 'Transcoding video/audio';
		if (value.includes('remux')) return 'Repackaging stream';
		if (value.includes('adaptive')) return 'Adaptive stream';
		if (value.includes('direct')) return 'Direct play';
		if (value.includes('decision deferred')) return 'Decision pending';
		return '';
	}

	function markPlayingLive(sourceId: string): void {
		livePlaybackSourceId = asText(sourceId);
		void refreshLivePlayback();
	}

	async function refreshLivePlayback(): Promise<void> {
		const sourceId = asText(livePlaybackSourceId || selectedSource?.mediaSourceId || selectedEpisode?.mediaSourceId);
		if (!sourceId) return;
		livePlaybackError = '';
		try {
			const [sessionsPayload, statePayload] = await Promise.all([
				getSessions(apiClient),
				getPlaybackState(sourceId, apiClient).catch(() => null)
			]);
			const session = (sessionsPayload.sessions || []).find((item) => asText(item.mediaSourceId) === sourceId) || null;
			livePlaybackSession = session;
			livePlaybackState = statePayload;
		} catch (error) {
			livePlaybackError = formatLoadError(error);
		}
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
		return 'TV details could not load.';
	}

</script>

<XuvaShell>
	{#if isLoading}
		<XuvaPanel title="Loading TV Details" subtitle="Fetching series, seasons, episodes, and playback state." />
	{:else if loadError}
		<section class="px-4 pt-9 sm:px-6 lg:px-8">
			<XuvaEmptyState
				eyebrow="Connection"
				title="Media library unavailable"
				description="Xuva could not reach the media library service. Check that the server is running, then try again."
			>
				{#snippet primaryAction()}
					<XuvaButton variant="primary" onclick={loadSeriesDetails}>Retry</XuvaButton>
				{/snippet}
				{#snippet secondaryAction()}
					<XuvaButton variant="secondary" href="/settings">Settings</XuvaButton>
					<XuvaButton variant="ghost" href="/">Back Home</XuvaButton>
				{/snippet}
			</XuvaEmptyState>
		</section>
	{:else if !series}
		<section class="px-4 pt-9 sm:px-6 lg:px-8">
			<XuvaEmptyState
				eyebrow="TV"
				title="TV show not found"
				description="Xuva could not find that show. It may have been removed, renamed, or not scanned yet."
			>
				{#snippet primaryAction()}
					<XuvaButton variant="primary" href="/tv">Back to TV</XuvaButton>
				{/snippet}
				{#snippet secondaryAction()}
					<XuvaButton variant="secondary" onclick={startTVScan} disabled={isScanning}>
						{isScanning ? 'Scanning...' : 'Scan TV'}
					</XuvaButton>
				{/snippet}
			</XuvaEmptyState>
			{#if actionMessage}
				<p class="mt-3 text-sm text-white/60">{actionMessage}</p>
			{/if}
		</section>
	{:else}
		<DetailHero
			title={seriesTitle}
			meta={seriesMeta}
			overview={seriesOverview}
			backHref="/tv"
			backLabel="Back to TV"
			backdropUrl={seriesBackdropUrl}
			posterUrl={seriesPosterUrl}
			progress={selectedProgress}
			progressLabel={selectedProgress > 0 ? `${selectedProgress}% watched` : ''}
		>
			{#snippet actions()}
				{#if selectedPlayHref}
					<XuvaButton
						variant="primary"
						href={selectedPlayHref}
						onclick={() => selectedEpisode && markPlayingLive(selectedEpisode.mediaSourceId)}
					>
						<Play size={18} class="fill-white text-white" />
						{selectedPlayLabel}
					</XuvaButton>
					<XuvaButton
						variant="secondary"
						href={selectedStartHref}
						onclick={() => selectedEpisode && markPlayingLive(selectedEpisode.mediaSourceId)}
					>
						Start Over
					</XuvaButton>
				{:else}
					<XuvaButton variant="primary" disabled>Play</XuvaButton>
				{/if}
			{/snippet}
		</DetailHero>

		{#if asText(livePlaybackSourceId || selectedEpisode?.mediaSourceId || selectedSource?.mediaSourceId)}
			<section class="px-4 pt-4 sm:px-6 lg:px-8">
				<div class="rounded-md border border-[#34D3E6]/25 bg-[#081826]/70 p-4">
					<div class="flex flex-wrap items-center justify-between gap-3">
						<p class="text-sm font-semibold text-[#8DE6EC]">Playing Live</p>
						{#if playbackRouteLabel(asText(livePlaybackSession?.mode || livePlaybackSession?.route))}
							<p class="text-xs text-white/60">{playbackRouteLabel(asText(livePlaybackSession?.mode || livePlaybackSession?.route))}</p>
						{/if}
					</div>
					<div class="mt-3 h-2 overflow-hidden rounded-full bg-white/10">
						<div class="h-full rounded-full bg-[#34D3E6]" style={`width: ${Math.max(0, Math.min(100, Number(livePlaybackState?.percent || 0)))}%`}></div>
					</div>
					<div class="mt-2 flex flex-wrap gap-x-4 gap-y-1 text-xs text-white/65">
						<span>{Math.round(Math.max(0, Math.min(100, Number(livePlaybackState?.percent || 0)))) > 0 ? `${Math.round(Math.max(0, Math.min(100, Number(livePlaybackState?.percent || 0))))}% watched` : 'Unplayed'}</span>
					</div>
					{#if livePlaybackError}
						<p class="mt-2 text-xs text-red-200/80">{livePlaybackError}</p>
					{/if}
				</div>
			</section>
		{/if}

		<DetailSection title="Seasons and Episodes" subtitle="Browse episodes and start playback.">
			{#if seasons.length === 0}
				<p class="text-sm leading-relaxed text-white/60">
					Series metadata does not include season or episode rows yet.
				</p>
			{:else}
				<div class="grid gap-4">
					{#each seasons as season (season.seasonId)}
						<article class="rounded-md border border-white/10 bg-white/[0.02] p-4">
							<header class="mb-4 flex flex-wrap items-baseline justify-between gap-3">
								<h3 class="text-lg font-semibold text-white">Season {season.seasonNumber || 'Unknown'}</h3>
								<span class="text-sm text-white/50">{season.episodes.length} episodes</span>
							</header>
							{#if season.episodes.length === 0}
								<p class="text-sm text-white/55">No episodes were returned for this season.</p>
							{:else}
								<div class="grid gap-3">
									{#each season.episodes as episode (episode.episodeId)}
										<div class={`grid gap-3 rounded-md border p-4 transition lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center ${isEpisodeSelected(episode.episodeId) ? 'border-[#7C5CFF]/45 bg-[#7C5CFF]/8' : 'border-white/10 bg-[#111827]/55'}`}>
											<button
												class="flex w-full items-start justify-between gap-4 text-left text-white"
												type="button"
												aria-label={`Open ${episode.title}`}
												onclick={() => selectEpisode(episode.episodeId)}
											>
												<span>
													<strong class="block text-sm font-semibold text-white">{episode.label}</strong>
													<span class="mt-1 block text-base font-medium text-white">{episode.title}</span>
												</span>
												<em class="shrink-0 rounded-sm border border-white/10 bg-black/30 px-3 py-1 text-xs font-semibold not-italic text-white/65">
													{episodeSummary(episode)}
												</em>
											</button>
											<div class="flex flex-wrap gap-3 lg:justify-end">
												{#if episode.mediaSourceId}
													<XuvaButton
														variant="primary"
														size="sm"
														href={playHrefWithOptions(episode.mediaSourceId)}
														onclick={() => markPlayingLive(episode.mediaSourceId)}
													>
														{isResumeState(episode.state) ? 'Resume' : 'Play'}
													</XuvaButton>
													<XuvaButton
														variant="secondary"
														size="sm"
														href={playHrefWithOptions(episode.mediaSourceId, true)}
														onclick={() => markPlayingLive(episode.mediaSourceId)}
													>
														Start Over
													</XuvaButton>
												{:else}
													<span class="text-sm text-white/50">No source</span>
												{/if}
											</div>
										</div>
									{/each}
								</div>
							{/if}
						</article>
					{/each}
				</div>
			{/if}
		</DetailSection>

		<DetailSection
			title="Selected Episode Source"
			subtitle="Technical details stay secondary to episode browsing and playback."
		>
			{#if !selectedEpisode}
				<p class="text-sm leading-relaxed text-white/60">Choose a playable episode to inspect source, tracks, and subtitles.</p>
			{:else if selectedSourceLoading}
				<p class="text-sm leading-relaxed text-white/60">Loading source metadata, track details, and playback decision.</p>
			{:else if selectedSourceError}
				<p class="text-sm leading-relaxed text-red-200/80">{selectedSourceError}</p>
			{:else if !selectedSource}
				<p class="text-sm leading-relaxed text-white/60">No media source metadata is available for this episode.</p>
			{:else}
				<div class="rounded-md border border-white/10 bg-[#111827]/55 p-4">
					<div class="flex flex-wrap items-baseline justify-between gap-3">
						<div>
							<h3 class="text-lg font-semibold text-white">{selectedEpisode.title}</h3>
							<p class="mt-1 text-sm text-white/50">{sourceEpisodeMeta()}</p>
						</div>
						<span class="rounded-sm border border-white/10 bg-white/10 px-3 py-1 text-xs font-semibold text-white/70">
							{watchedLabel(selectedSource.state)}
						</span>
					</div>
					<p class="mt-5 text-sm font-semibold text-white">{playbackModeLabel(selectedSource.decision)}</p>
					<p class="mt-1 text-sm leading-relaxed text-white/55">{playbackReasonLabel(selectedSource.decision)}</p>
					{#if analyzerRoute}
						<p class="mt-2 text-sm text-white/55">
							Route: {asText(analyzerRoute.route) || 'pending'} - {asText(analyzerRoute.status) || 'pending'}
						</p>
					{/if}
					<div class="mt-3 flex flex-wrap gap-3">
						{#if !Boolean(selectedSource.source?.probed)}
							<XuvaButton
								variant="secondary"
								size="sm"
								onclick={() => selectedSource && runMediaCheck(selectedSource.mediaSourceId)}
								disabled={probingSourceId === (selectedSource?.mediaSourceId ?? '')}
							>
								{probingSourceId === (selectedSource?.mediaSourceId ?? '') ? 'Checking File...' : 'Run Media Check'}
							</XuvaButton>
							{#if probeStatusMessage && probeStatusSourceId === (selectedSource?.mediaSourceId ?? '')}
								<p class={`text-xs ${probeStatusTone === 'error' ? 'text-red-200/80' : probeStatusTone === 'success' ? 'text-emerald-200/80' : 'text-white/60'}`}>
									{probeStatusMessage}
								</p>
							{/if}
						{/if}
						<XuvaButton variant="ghost" size="sm" onclick={refreshAnalyzer} disabled={analyzerBusy}>
							{analyzerBusy ? 'Refreshing...' : 'Refresh Decision'}
						</XuvaButton>
					</div>
					<dl class="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
							<div class="rounded-sm border border-white/10 bg-white/[0.02] p-3">
								<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Quality</dt>
								<dd class="mt-1 text-sm text-white/75">
									{sourceQualityLabel(
										'',
										formatResolution(
											Number(selectedSource.source?.width || 0),
											Number(selectedSource.source?.height || 0)
										)
									)}
								</dd>
							</div>
							<div class="rounded-sm border border-white/10 bg-white/[0.02] p-3">
								<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Duration</dt>
								<dd class="mt-1 text-sm text-white/75">{formatDuration(Number(selectedSource.source?.durationSeconds || 0))}</dd>
							</div>
							<div class="rounded-sm border border-white/10 bg-white/[0.02] p-3">
								<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Bitrate</dt>
								<dd class="mt-1 text-sm text-white/75">{formatBitrate(Number(selectedSource.source?.bitrate || 0))}</dd>
							</div>
							<div class="rounded-sm border border-white/10 bg-white/[0.02] p-3">
								<dt class="text-xs uppercase tracking-[0.08em] text-white/40">File Size</dt>
								<dd class="mt-1 text-sm text-white/75">{formatBytes(Number(selectedSource.source?.sizeBytes || 0))}</dd>
							</div>
					</dl>
				</div>

					<div class="mt-4 grid gap-4 lg:grid-cols-2">
						<article class="rounded-md border border-white/10 bg-[#111827]/55 p-4">
							<h3 class="text-base font-semibold text-white">Audio Tracks</h3>
							{#if (selectedSource.tracks.audioTracks || []).length === 0}
								<p class="mt-3 text-sm text-white/55">No audio track data is available yet.</p>
							{:else}
								<ul class="mt-3 grid gap-2">
									{#each selectedSource.tracks.audioTracks || [] as track, index (index)}
										<li class="text-sm text-white/70">
											<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language, Number(track.channels || 0))}</strong>
											{#if track.default}<span class="ml-2 rounded-sm bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
										</li>
									{/each}
								</ul>
							{/if}
						</article>
						<article class="rounded-md border border-white/10 bg-[#111827]/55 p-4">
							<h3 class="text-base font-semibold text-white">Subtitle Tracks</h3>
							{#if (selectedSource.tracks.subtitleTracks || []).length === 0}
								<p class="mt-3 text-sm text-white/55">No embedded subtitle tracks were detected.</p>
							{:else}
								<ul class="mt-3 grid gap-2">
									{#each selectedSource.tracks.subtitleTracks || [] as track, index (index)}
										<li class="text-sm text-white/70">
											<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language)}</strong>
											{#if track.forced}<span class="ml-2 rounded-sm bg-white/10 px-2 py-1 text-xs text-white/60">Forced</span>{/if}
											{#if track.default}<span class="ml-2 rounded-sm bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
										</li>
									{/each}
								</ul>
							{/if}
							<p class="mt-3 text-sm text-white/50">Sidecar subtitles: {(selectedSource.subtitles.sidecars || []).length}</p>
						</article>
					</div>
					<div class="mt-4 rounded-md border border-white/10 bg-[#111827]/55 p-4">
						<h3 class="text-sm font-semibold text-white">Playback Analyzer</h3>
						<div class="mt-3 grid gap-4 lg:grid-cols-2">
							<label class="settings-field">
								<span>Device profile</span>
								<select bind:value={analyzerProfileId} onchange={() => void refreshAnalyzer()}>
									{#each deviceProfiles as profile (profile.id)}
										<option value={asText(profile.id)}>{asText(profile.name) || asText(profile.id)}</option>
									{/each}
								</select>
							</label>
							<label class="settings-field">
								<span>Route type</span>
								<select bind:value={analyzerRouteType} onchange={() => { void refreshAnalyzer(); void refreshDeviceMatrix(); }}>
									<option value="lan">LAN</option>
									<option value="remote">Remote</option>
								</select>
							</label>
							<label class="settings-field">
								<span>Max network bitrate (bps, optional)</span>
								<input bind:value={analyzerMaxBitrate} inputmode="numeric" placeholder="8000000" onblur={() => void refreshAnalyzer()} />
							</label>
							<label class="settings-field">
								<span>Audio track</span>
								<select bind:value={analyzerAudioTrackIndex} onchange={() => { void refreshAnalyzer(); void refreshDeviceMatrix(); }}>
									{#each selectedSource.tracks.audioTracks || [] as track, index (index)}
										<option value={Number(track.index || 0)}>
											{formatTrackSummary(track.codec, track.language, Number(track.channels || 0))}
										</option>
									{/each}
								</select>
							</label>
							<label class="settings-field">
								<span>Subtitle mode</span>
								<select bind:value={analyzerSubtitleEnabled} onchange={() => { void refreshAnalyzer(); void refreshDeviceMatrix(); }}>
									<option value={false}>Off</option>
									<option value={true}>On</option>
								</select>
							</label>
							<label class="settings-field">
								<span>Subtitle track</span>
								<select bind:value={analyzerSubtitleTrackIndex} onchange={() => { void refreshAnalyzer(); void refreshDeviceMatrix(); }}>
									{#each selectedSource.tracks.subtitleTracks || [] as track, index (index)}
										<option value={Number(track.index || 0)}>
											{formatTrackSummary(track.codec, track.language)}
										</option>
									{/each}
								</select>
							</label>
						</div>
						<div class="mt-4 flex flex-wrap gap-3">
							<XuvaButton variant="primary" onclick={refreshAnalyzer} disabled={analyzerBusy}>
								{analyzerBusy ? 'Analyzing...' : 'Analyze Playback'}
							</XuvaButton>
							<XuvaButton variant="secondary" onclick={refreshDeviceMatrix} disabled={analyzerMatrixBusy}>
								{analyzerMatrixBusy ? 'Checking Devices...' : 'Refresh Device Matrix'}
							</XuvaButton>
							<XuvaButton
								variant="ghost"
								href={playHrefWithOptions(selectedSource?.mediaSourceId ?? '')}
								onclick={() => selectedSource && markPlayingLive(selectedSource.mediaSourceId)}
							>
								Play With These Options
							</XuvaButton>
						</div>
						{#if analyzerError}
							<p class="mt-3 text-sm text-red-200/80">{analyzerError}</p>
						{/if}
						{#if analyzerDecision}
							<div class="mt-4 rounded-sm border border-white/10 bg-white/[0.02] p-3">
								<p class="text-sm font-semibold text-white">{playbackModeLabel(analyzerDecision)}</p>
								<p class="mt-1 text-sm text-white/60">{playbackReasonLabel(analyzerDecision)}</p>
								{#if analyzerRoute}
									<p class="mt-2 text-sm text-white/55">
										Route: {asText(analyzerRoute.route) || 'pending'} - {asText(analyzerRoute.status) || 'pending'}
									</p>
								{/if}
							</div>
						{/if}
						<div class="mt-4">
							<h4 class="text-xs font-semibold uppercase tracking-[0.08em] text-white/50">Device Compatibility Matrix</h4>
							{#if analyzerMatrixError}
								<p class="mt-2 text-sm text-red-200/80">{analyzerMatrixError}</p>
							{:else if analyzerMatrixBusy}
								<p class="mt-2 text-sm text-white/55">Analyzing device profiles...</p>
							{:else if analyzerMatrix.length === 0}
								<p class="mt-2 text-sm text-white/55">No device profile results yet.</p>
							{:else}
								<div class="mt-2 grid gap-2">
									{#each analyzerMatrix as row (row.profileId)}
										<div class="rounded-sm border border-white/10 bg-white/[0.02] p-3">
											<p class="text-sm font-semibold text-white">{row.profileName}</p>
											<p class="text-sm text-white/70">{row.mode}</p>
											<p class="text-xs text-white/50">{row.reason}</p>
										</div>
									{/each}
								</div>
							{/if}
						</div>
					</div>
			{/if}
		</DetailSection>
	{/if}
</XuvaShell>
