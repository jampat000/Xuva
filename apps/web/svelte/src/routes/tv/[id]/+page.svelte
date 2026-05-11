<script lang="ts">
	import { onMount } from 'svelte';
	import { Play } from 'lucide-svelte';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import {
		getMediaSourceDetail,
		getMediaSourceSubtitles,
		getMediaSourceTracks,
		getPlaybackDecision,
		getPlaybackState,
		getSeriesDetail,
		listMediaSources,
		type EpisodeBrief,
		type MediaSourceItem,
		type MediaSourceSubtitlesResponse,
		type MediaSourceTracksResponse,
		type PlaybackDecisionResponse,
		type PlaybackStateResponse,
		type SeasonDetail,
		type SeriesDetailResponse
	} from '$lib/api/details';
	import { resolvePreviewMode } from '$lib/home/model';
	import { previewBackdrop, previewPoster } from '$lib/preview/artwork';
	import DetailHero from '$lib/lorivo/DetailHero.svelte';
	import DetailSection from '$lib/lorivo/DetailSection.svelte';
	import LorivoButton from '$lib/lorivo/LorivoButton.svelte';
	import LorivoPanel from '$lib/lorivo/LorivoPanel.svelte';
	import LorivoShell from '$lib/lorivo/LorivoShell.svelte';
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
	let loadError = $state('');
	let series = $state<SeriesDetailResponse | null>(null);
	let seasons = $state<SeasonModel[]>([]);
	let selectedEpisodeId = $state('');
	let selectedSource = $state<EpisodeSourceModel | null>(null);
	let selectedSourceLoading = $state(false);
	let selectedSourceError = $state('');
	let previewDetailMode = $state(false);

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
		return mediaSourceId ? `/play/${encodeURIComponent(mediaSourceId)}` : '';
	});
	const selectedStartHref = $derived.by(() => {
		const mediaSourceId = asText(selectedEpisode?.mediaSourceId);
		return mediaSourceId ? `/play/${encodeURIComponent(mediaSourceId)}?start=0` : '';
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
	});

	$effect(() => {
		const mediaSourceId = asText(selectedEpisode?.mediaSourceId);
		if (!mediaSourceId) {
			selectedSource = null;
			selectedSourceError = '';
			selectedSourceLoading = false;
			return;
		}
		void loadSelectedEpisodeSource(mediaSourceId, selectedEpisode?.state || null);
	});

	async function loadSeriesDetails(): Promise<void> {
		isLoading = true;
		loadError = '';
		selectedSource = null;
		selectedSourceError = '';
		const previewMode = resolvePreviewMode(new URL(window.location.href).searchParams);
		const routeSeriesId = asText(params.id);
		previewDetailMode = previewMode && isSeriesPreviewId(routeSeriesId);
		if (previewDetailMode) {
			const preview = buildSeriesDetailPreview(routeSeriesId);
			series = preview.series;
			seasons = preview.seasons;
			selectedEpisodeId = preview.firstEpisodeId;
			isLoading = false;
			return;
		}

		try {
			const [seriesPayload, mediaSourcePayload] = await Promise.all([
				getSeriesDetail(asText(params.id), apiClient),
				listMediaSources(apiClient, 1500).catch(() => ({ mediaSources: [] }))
			]);

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
			loadError = formatLoadError(error);
		} finally {
			isLoading = false;
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
		if (mediaSourceId.startsWith('preview-tv-') || mediaSourceId.startsWith('series-')) {
			selectedSource = {
				mediaSourceId,
				source: {
					id: mediaSourceId,
					kind: 'episode',
					durationSeconds: 2_700,
					bitrate: 7_200_000,
					width: 1920,
					height: 1080,
					sizeBytes: 2_430_000_000,
					container: 'mkv'
				},
				tracks: {
					audioTracks: [
						{ codec: 'eac3', language: 'en', channels: 6, default: true },
						{ codec: 'aac', language: 'es', channels: 2 }
					],
					subtitleTracks: [{ codec: 'srt', language: 'en', default: true }]
				},
				subtitles: { sidecars: [{ relPath: 'preview/en.srt', format: 'srt', language: 'en' }] },
				decision: {
					mode: 'directplay',
					reasonText: 'Browser profile supports this source.',
					estimatedNetworkBitrate: 7_200_000,
					containerAction: 'copy',
					videoAction: 'copy',
					audioAction: 'copy',
					subtitleAction: 'copy'
				},
				state
			};
			selectedSourceLoading = false;
			return;
		}
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

	function isSeriesPreviewId(id: string): boolean {
		return id === 'preview' || id.startsWith('preview-tv-');
	}

	function buildSeriesDetailPreview(id: string): {
		series: SeriesDetailResponse;
		seasons: SeasonModel[];
		firstEpisodeId: string;
	} {
		const seriesId = id === 'preview' ? 'series-violet-signal' : id;
		const title = id === 'preview' ? 'Violet Signal' : cleanIdTitle(id, 'preview-tv-');
		const episodesSeason1: EpisodeItemModel[] = [
			previewEpisode(seriesId, 1, 1, 'Pilot', true),
			previewEpisode(seriesId, 1, 2, 'Second Signal', false),
			previewEpisode(seriesId, 1, 3, 'Crossfade', false)
		];
		const episodesSeason2: EpisodeItemModel[] = [
			previewEpisode(seriesId, 2, 1, 'Return Vector', false),
			previewEpisode(seriesId, 2, 2, 'Night Archive', false)
		];
		const seasons: SeasonModel[] = [
			{ seasonId: `${seriesId}-season-1`, seasonNumber: 1, episodes: episodesSeason1 },
			{ seasonId: `${seriesId}-season-2`, seasonNumber: 2, episodes: episodesSeason2 }
		];
		const series: SeriesDetailResponse = {
			id: seriesId,
			title,
			seasonCount: 2,
			episodeCount: 5,
			metadata: {
				title,
				year: 2025,
				posterUrl: previewPoster(title),
				backdropUrl: previewBackdrop(title),
				overview:
					'A late-night observatory feed reveals repeating anomalies, and a fragmented crew races to decode what returns with every tide.'
			}
		};
		return { series, seasons, firstEpisodeId: episodesSeason1[0]?.episodeId || '' };
	}

	function previewEpisode(
		seriesId: string,
		seasonNumber: number,
		episodeNumber: number,
		title: string,
		resume: boolean
	): EpisodeItemModel {
		const mediaSourceId = `${seriesId}-s${seasonNumber}e${episodeNumber}`;
		return {
			episodeId: `${mediaSourceId}-episode`,
			seasonId: `${seriesId}-season-${seasonNumber}`,
			seasonNumber,
			episodeNumber,
			episodeEnd: 0,
			title,
			label: `S${seasonNumber} E${episodeNumber}`,
			versionCount: 1,
			needsReview: false,
			mediaSourceId,
			state: resume
				? { mediaSourceId, watched: false, progressSeconds: 680, durationSeconds: 2700, percent: 25.1 }
				: { mediaSourceId, watched: false, progressSeconds: 0, durationSeconds: 2700, percent: 0 }
		};
	}

	function cleanIdTitle(value: string, prefix: string): string {
		const raw = asText(value).replace(prefix, '').replace(/-/g, ' ');
		if (!raw) return 'TV Preview';
		return raw.replace(/\b\w/g, (letter) => letter.toUpperCase());
	}
</script>

<svelte:head>
	<title>{seriesTitle} - Lorivo Media</title>
</svelte:head>

<LorivoShell>
	{#if isLoading}
		<LorivoPanel title="Loading TV Details" subtitle="Fetching series, seasons, episodes, and playback state." />
	{:else if loadError}
		<LorivoPanel title="TV details could not load" subtitle={loadError}>
			<div class="flex flex-wrap gap-3">
				<LorivoButton variant="secondary" onclick={loadSeriesDetails}>Retry</LorivoButton>
				<LorivoButton variant="ghost" href="/tv">Back to TV</LorivoButton>
			</div>
		</LorivoPanel>
	{:else if !series}
		<LorivoPanel title="TV show not found" subtitle="This show is no longer available in your library.">
			<LorivoButton variant="secondary" href="/tv">Back to TV</LorivoButton>
		</LorivoPanel>
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
					<LorivoButton variant="primary" href={selectedPlayHref}>
						<Play size={18} class="fill-white text-white" />
						{selectedPlayLabel}
					</LorivoButton>
					<LorivoButton variant="secondary" href={selectedStartHref}>Start Over</LorivoButton>
				{:else}
					<LorivoButton variant="primary" disabled>Play</LorivoButton>
				{/if}
			{/snippet}
		</DetailHero>

		<DetailSection title="Seasons and Episodes" subtitle="Browse episodes and start playback.">
			{#if seasons.length === 0}
				<p class="text-sm leading-relaxed text-white/60">
					Series metadata does not include season or episode rows yet.
				</p>
			{:else}
				<div class="grid gap-4">
					{#each seasons as season (season.seasonId)}
						<article class="rounded-2xl border border-white/10 bg-white/[0.04] p-5 shadow-lg shadow-black/20">
							<header class="mb-4 flex flex-wrap items-baseline justify-between gap-3">
								<h3 class="text-lg font-semibold text-white">Season {season.seasonNumber || 'Unknown'}</h3>
								<span class="text-sm text-white/50">{season.episodes.length} episodes</span>
							</header>
							{#if season.episodes.length === 0}
								<p class="text-sm text-white/55">No episodes were returned for this season.</p>
							{:else}
								<div class="grid gap-3">
									{#each season.episodes as episode (episode.episodeId)}
										<div class={`grid gap-3 rounded-2xl border p-4 transition lg:grid-cols-[minmax(0,1fr)_auto] lg:items-center ${isEpisodeSelected(episode.episodeId) ? 'border-[#7C5CFF]/45 bg-[#7C5CFF]/10' : 'border-white/10 bg-[#111827]/60'}`}>
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
												<em class="shrink-0 rounded-full border border-white/10 bg-black/30 px-3 py-1 text-xs font-semibold not-italic text-white/65">
													{episodeSummary(episode)}
												</em>
											</button>
											<div class="flex flex-wrap gap-3 lg:justify-end">
												{#if episode.mediaSourceId}
													<LorivoButton
														variant="primary"
														size="sm"
														href={`/play/${encodeURIComponent(episode.mediaSourceId)}`}
													>
														{isResumeState(episode.state) ? 'Resume' : 'Play'}
													</LorivoButton>
													<LorivoButton
														variant="secondary"
														size="sm"
														href={`/play/${encodeURIComponent(episode.mediaSourceId)}?start=0`}
													>
														Start Over
													</LorivoButton>
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
			subtitle={previewDetailMode
				? 'Technical source details stay secondary to browsing and playback.'
				: 'Technical details stay secondary to episode browsing and playback.'}
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
				<div class="rounded-2xl border border-white/10 bg-[#111827]/70 p-5">
					<div class="flex flex-wrap items-baseline justify-between gap-3">
						<div>
							<h3 class="text-lg font-semibold text-white">{selectedEpisode.title}</h3>
							<p class="mt-1 text-sm text-white/50">{sourceEpisodeMeta()}</p>
						</div>
						<span class="rounded-full border border-white/10 bg-white/10 px-3 py-1 text-xs font-semibold text-white/70">
							{watchedLabel(selectedSource.state)}
						</span>
					</div>
					{#if !previewDetailMode}
						<p class="mt-5 text-sm font-semibold text-white">{playbackModeLabel(selectedSource.decision)}</p>
						<p class="mt-1 text-sm leading-relaxed text-white/55">{playbackReasonLabel(selectedSource.decision)}</p>
						<dl class="mt-5 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
							<div class="rounded-xl border border-white/10 bg-white/[0.04] p-3">
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
							<div class="rounded-xl border border-white/10 bg-white/[0.04] p-3">
								<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Duration</dt>
								<dd class="mt-1 text-sm text-white/75">{formatDuration(Number(selectedSource.source?.durationSeconds || 0))}</dd>
							</div>
							<div class="rounded-xl border border-white/10 bg-white/[0.04] p-3">
								<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Bitrate</dt>
								<dd class="mt-1 text-sm text-white/75">{formatBitrate(Number(selectedSource.source?.bitrate || 0))}</dd>
							</div>
							<div class="rounded-xl border border-white/10 bg-white/[0.04] p-3">
								<dt class="text-xs uppercase tracking-[0.08em] text-white/40">File Size</dt>
								<dd class="mt-1 text-sm text-white/75">{formatBytes(Number(selectedSource.source?.sizeBytes || 0))}</dd>
							</div>
						</dl>
					{/if}
				</div>

				{#if previewDetailMode}
					<details class="mt-4 rounded-2xl border border-white/10 bg-white/[0.04] p-4">
						<summary class="cursor-pointer text-sm font-semibold text-white/70">Technical playback details</summary>
						<div class="mt-4 grid gap-4 lg:grid-cols-2">
							<article class="rounded-2xl border border-white/10 bg-[#111827]/70 p-4">
								<h3 class="text-base font-semibold text-white">Audio Tracks</h3>
								{#if (selectedSource.tracks.audioTracks || []).length === 0}
									<p class="mt-3 text-sm text-white/55">No audio track data is available yet.</p>
								{:else}
									<ul class="mt-3 grid gap-2">
										{#each selectedSource.tracks.audioTracks || [] as track, index (index)}
											<li class="text-sm text-white/70">
												<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language, Number(track.channels || 0))}</strong>
												{#if track.default}<span class="ml-2 rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
											</li>
										{/each}
									</ul>
								{/if}
							</article>
							<article class="rounded-2xl border border-white/10 bg-[#111827]/70 p-4">
								<h3 class="text-base font-semibold text-white">Subtitle Tracks</h3>
								{#if (selectedSource.tracks.subtitleTracks || []).length === 0}
									<p class="mt-3 text-sm text-white/55">No embedded subtitle tracks were detected.</p>
								{:else}
									<ul class="mt-3 grid gap-2">
										{#each selectedSource.tracks.subtitleTracks || [] as track, index (index)}
											<li class="text-sm text-white/70">
												<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language)}</strong>
												{#if track.forced}<span class="ml-2 rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Forced</span>{/if}
												{#if track.default}<span class="ml-2 rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
											</li>
										{/each}
									</ul>
								{/if}
								<p class="mt-3 text-sm text-white/50">Sidecar subtitles: {(selectedSource.subtitles.sidecars || []).length}</p>
							</article>
						</div>
					</details>
				{:else}
					<div class="mt-4 grid gap-4 lg:grid-cols-2">
						<article class="rounded-2xl border border-white/10 bg-[#111827]/70 p-4">
							<h3 class="text-base font-semibold text-white">Audio Tracks</h3>
							{#if (selectedSource.tracks.audioTracks || []).length === 0}
								<p class="mt-3 text-sm text-white/55">No audio track data is available yet.</p>
							{:else}
								<ul class="mt-3 grid gap-2">
									{#each selectedSource.tracks.audioTracks || [] as track, index (index)}
										<li class="text-sm text-white/70">
											<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language, Number(track.channels || 0))}</strong>
											{#if track.default}<span class="ml-2 rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
										</li>
									{/each}
								</ul>
							{/if}
						</article>
						<article class="rounded-2xl border border-white/10 bg-[#111827]/70 p-4">
							<h3 class="text-base font-semibold text-white">Subtitle Tracks</h3>
							{#if (selectedSource.tracks.subtitleTracks || []).length === 0}
								<p class="mt-3 text-sm text-white/55">No embedded subtitle tracks were detected.</p>
							{:else}
								<ul class="mt-3 grid gap-2">
									{#each selectedSource.tracks.subtitleTracks || [] as track, index (index)}
										<li class="text-sm text-white/70">
											<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language)}</strong>
											{#if track.forced}<span class="ml-2 rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Forced</span>{/if}
											{#if track.default}<span class="ml-2 rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
										</li>
									{/each}
								</ul>
							{/if}
							<p class="mt-3 text-sm text-white/50">Sidecar subtitles: {(selectedSource.subtitles.sidecars || []).length}</p>
						</article>
					</div>
				{/if}
			{/if}
		</DetailSection>
	{/if}
</LorivoShell>
