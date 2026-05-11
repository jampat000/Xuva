<script lang="ts">
	import { onMount } from 'svelte';
	import { getAuthSession, type AuthSessionUser } from '$lib/api/auth';
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
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
import { resolvePreviewMode } from '$lib/home/model';
import { previewBackdrop, previewPoster } from '$lib/preview/artwork';
	import {
		DetailHero,
		DetailPage,
		DetailSection,
		DetailTechnicalPanel,
		MediaShell,
		VyrdenButton,
		VyrdenEmptyState,
		VyrdenPanel
	} from '$lib/components';
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
	let searchValue = $state('');
	let user = $state<AuthSessionUser | null>(null);
	let libraries = $state<LibraryRecord[]>([]);
	let series = $state<SeriesDetailResponse | null>(null);
	let seasons = $state<SeasonModel[]>([]);
	let selectedEpisodeId = $state('');
	let selectedSource = $state<EpisodeSourceModel | null>(null);
	let selectedSourceLoading = $state(false);
	let selectedSourceError = $state('');
	let previewDetailMode = $state(false);

	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
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
			user = null;
			libraries = [];
			series = preview.series;
			seasons = preview.seasons;
			selectedEpisodeId = preview.firstEpisodeId;
			isLoading = false;
			return;
		}

		try {
			const [sessionPayload, librariesPayload, seriesPayload, mediaSourcePayload] = await Promise.all([
				getAuthSession(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) return { user: null };
					throw error;
				}),
				getLibraries(apiClient),
				getSeriesDetail(asText(params.id), apiClient),
				listMediaSources(apiClient, 1500).catch(() => ({ mediaSources: [] }))
			]);

			user = sessionPayload?.user || null;
			libraries = librariesPayload.libraries || [];
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
		if (mediaSourceId.startsWith('preview-tv-')) {
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

	// preview artwork comes from shared local preview artwork system
</script>

<MediaShell active="tv" bind:searchValue {userInitials}>
	<DetailPage>
		{#if isLoading}
			<VyrdenPanel title="Loading TV Details" subtitle="Fetching series, seasons, episodes, and playback state." />
		{:else if loadError}
			<VyrdenPanel title="TV details could not load" subtitle={loadError}>
				<div class="status-actions">
					<VyrdenButton variant="secondary" onclick={loadSeriesDetails}>Retry</VyrdenButton>
					<VyrdenButton variant="ghost" href="/tv">Back to TV</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else if !series}
			<VyrdenEmptyState title="TV show not found" message="This show is no longer available in your library.">
				{#snippet action()}
					<VyrdenButton variant="secondary" href="/tv">Back to TV</VyrdenButton>
				{/snippet}
			</VyrdenEmptyState>
		{:else}
			<DetailHero
				title={seriesTitle}
				meta={seriesMeta}
				overview={seriesOverview}
				backHref="/tv"
				backLabel="Back to TV"
				backdropUrl={seriesBackdropUrl}
				posterUrl={seriesPosterUrl}
			>
				{#snippet actions()}
					{#if selectedPlayHref}
						<VyrdenButton variant="primary" href={selectedPlayHref}>{selectedPlayLabel}</VyrdenButton>
						<VyrdenButton variant="secondary" href={selectedStartHref}>Start Over</VyrdenButton>
					{:else}
						<VyrdenButton variant="primary" disabled>Play</VyrdenButton>
					{/if}
				{/snippet}
			</DetailHero>

			<DetailSection
				title="Seasons and Episodes"
				subtitle="Browse episodes and start playback."
			>
				{#if seasons.length === 0}
					<VyrdenEmptyState
						title="No seasons found"
						message="Series metadata does not include season or episode rows yet."
					/>
				{:else}
					<div class="season-stack">
						{#each seasons as season (season.seasonId)}
							<article class="season-card">
								<header>
									<h3>Season {season.seasonNumber || 'Unknown'}</h3>
									<span>{season.episodes.length} episodes</span>
								</header>
								{#if season.episodes.length === 0}
									<p class="season-empty">No episodes were returned for this season.</p>
								{:else}
									<div class="episode-list">
										{#each season.episodes as episode (episode.episodeId)}
											<div class="episode-row" class:episode-row--selected={isEpisodeSelected(episode.episodeId)}>
												<button
													class="episode-row__select"
													type="button"
													aria-label={`Open ${episode.title}`}
													onclick={() => selectEpisode(episode.episodeId)}
												>
													<div>
														<strong>{episode.label}</strong>
														<span>{episode.title}</span>
													</div>
													<em>{episodeSummary(episode)}</em>
												</button>
												<div class="episode-row__actions">
													{#if episode.mediaSourceId}
														<VyrdenButton
															variant="primary"
															size="sm"
															href={`/play/${encodeURIComponent(episode.mediaSourceId)}`}
														>
															{isResumeState(episode.state) ? 'Resume' : 'Play'}
														</VyrdenButton>
														<VyrdenButton
															variant="secondary"
															size="sm"
															href={`/play/${encodeURIComponent(episode.mediaSourceId)}?start=0`}
														>
															Start Over
														</VyrdenButton>
													{:else}
														<span class="episode-row__missing">No source</span>
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
				<div class="technical-wrap" class:technical-wrap--preview={previewDetailMode}>
				{#if !selectedEpisode}
					<VyrdenEmptyState
						title="Select an episode"
						message="Choose a playable episode to inspect source, tracks, and subtitles."
					/>
				{:else if selectedSourceLoading}
					<VyrdenPanel
						title={selectedEpisode.title}
						subtitle="Loading source metadata, track details, and playback decision."
					/>
				{:else if selectedSourceError}
					<VyrdenPanel title="Episode source could not load" subtitle={selectedSourceError} />
				{:else if !selectedSource}
					<VyrdenEmptyState
						title="No source details"
						message="No media source metadata is available for this episode."
					/>
				{:else}
					<div class="source-panel">
						<div class="source-panel__header">
							<div>
								<h3>{selectedEpisode.title}</h3>
								<p>{sourceEpisodeMeta()}</p>
							</div>
							<span>{watchedLabel(selectedSource.state)}</span>
						</div>
						{#if !previewDetailMode}
							<p class="source-panel__decision">{playbackModeLabel(selectedSource.decision)}</p>
							<p class="source-panel__reason">{playbackReasonLabel(selectedSource.decision)}</p>
							<dl class="source-panel__facts">
								<div>
									<dt>Quality</dt>
									<dd>
										{sourceQualityLabel(
											'',
											formatResolution(
												Number(selectedSource.source?.width || 0),
												Number(selectedSource.source?.height || 0)
											)
										)}
									</dd>
								</div>
								<div>
									<dt>Duration</dt>
									<dd>{formatDuration(Number(selectedSource.source?.durationSeconds || 0))}</dd>
								</div>
								<div>
									<dt>Bitrate</dt>
									<dd>{formatBitrate(Number(selectedSource.source?.bitrate || 0))}</dd>
								</div>
								<div>
									<dt>File Size</dt>
									<dd>{formatBytes(Number(selectedSource.source?.sizeBytes || 0))}</dd>
								</div>
							</dl>
						{/if}
						{#if previewDetailMode}
							<details class="technical-collapse">
								<summary>Technical playback details</summary>
								<div class="track-grid">
									<DetailTechnicalPanel title="Audio Tracks">
										{#if (selectedSource.tracks.audioTracks || []).length === 0}
											<p class="track-empty">No audio track data is available yet.</p>
										{:else}
											<ul>
												{#each selectedSource.tracks.audioTracks || [] as track, index (index)}
													<li>
														<strong>{formatTrackSummary(track.codec, track.language, Number(track.channels || 0))}</strong>
														{#if track.default}<span>Default</span>{/if}
													</li>
												{/each}
											</ul>
										{/if}
									</DetailTechnicalPanel>
									<DetailTechnicalPanel title="Subtitle Tracks">
										{#if (selectedSource.tracks.subtitleTracks || []).length === 0}
											<p class="track-empty">No embedded subtitle tracks were detected.</p>
										{:else}
											<ul>
												{#each selectedSource.tracks.subtitleTracks || [] as track, index (index)}
													<li>
														<strong>{formatTrackSummary(track.codec, track.language)}</strong>
														{#if track.forced}<span>Forced</span>{/if}
														{#if track.default}<span>Default</span>{/if}
													</li>
												{/each}
											</ul>
										{/if}
										<p class="track-sidecar">
											Sidecar subtitles: {(selectedSource.subtitles.sidecars || []).length}
										</p>
									</DetailTechnicalPanel>
								</div>
							</details>
						{:else}
						<div class="track-grid">
							<DetailTechnicalPanel title="Audio Tracks">
								{#if (selectedSource.tracks.audioTracks || []).length === 0}
									<p class="track-empty">No audio track data is available yet.</p>
								{:else}
									<ul>
										{#each selectedSource.tracks.audioTracks || [] as track, index (index)}
											<li>
												<strong>{formatTrackSummary(track.codec, track.language, Number(track.channels || 0))}</strong>
												{#if track.default}<span>Default</span>{/if}
											</li>
										{/each}
									</ul>
								{/if}
							</DetailTechnicalPanel>
							<DetailTechnicalPanel title="Subtitle Tracks">
								{#if (selectedSource.tracks.subtitleTracks || []).length === 0}
									<p class="track-empty">No embedded subtitle tracks were detected.</p>
								{:else}
									<ul>
										{#each selectedSource.tracks.subtitleTracks || [] as track, index (index)}
											<li>
												<strong>{formatTrackSummary(track.codec, track.language)}</strong>
												{#if track.forced}<span>Forced</span>{/if}
												{#if track.default}<span>Default</span>{/if}
											</li>
										{/each}
									</ul>
								{/if}
								<p class="track-sidecar">
									Sidecar subtitles: {(selectedSource.subtitles.sidecars || []).length}
								</p>
							</DetailTechnicalPanel>
						</div>
						{/if}
					</div>
				{/if}
				</div>
			</DetailSection>
		{/if}
	</DetailPage>
</MediaShell>

<style>
	.status-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--vyrden-space-2);
	}

	.season-stack {
		display: grid;
		gap: 16px;
	}

	.season-card {
		padding: 18px;
		border: 1px solid rgb(255 255 255 / 12%);
		border-radius: 16px;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 6%), rgb(255 255 255 / 2%)),
			rgb(17 24 39 / 58%);
		box-shadow: 0 16px 34px rgb(0 0 0 / 20%);
	}

	.season-card header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 10px;
		margin-bottom: 10px;
	}

	.season-card header h3 {
		margin: 0;
		font-size: 1.12rem;
		font-weight: 680;
	}

	.season-card header span {
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 86%, transparent);
		font-size: 0.78rem;
	}

	.season-empty {
		margin: 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 84%, transparent);
		font-size: 0.82rem;
	}

	.episode-list {
		display: grid;
		gap: 10px;
	}

	.episode-row {
		display: grid;
		grid-template-columns: minmax(0, 1fr) auto;
		gap: 12px;
		padding: 12px;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 13px;
		background: rgb(255 255 255 / 4%);
	}

	.episode-row--selected {
		border-color: color-mix(in srgb, var(--vyrden-color-accent-teal) 40%, transparent);
		background:
			linear-gradient(180deg, rgb(255 255 255 / 6%), rgb(255 255 255 / 2%)),
			color-mix(in srgb, var(--vyrden-color-accent-teal) 8%, transparent);
	}

	.episode-row__select {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 8px;
		padding: 0;
		border: 0;
		background: none;
		color: inherit;
		text-align: left;
		cursor: pointer;
	}

	.episode-row__select strong {
		display: block;
		font-size: 0.94rem;
		font-weight: 700;
	}

	.episode-row__select span {
		display: block;
		margin-top: 3px;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 84%, transparent);
		font-size: 0.82rem;
		line-height: 1.35;
	}

	.episode-row__select em {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		min-height: 28px;
		padding: 0 10px;
		border-radius: 999px;
		border: 1px solid rgb(255 255 255 / 14%);
		background: rgb(255 255 255 / 7%);
		color: color-mix(in srgb, var(--vyrden-color-text) 88%, transparent);
		font-size: 0.72rem;
		font-style: normal;
		font-weight: 620;
	}

	.episode-row__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 6px;
		align-items: center;
		justify-content: flex-end;
	}

	.episode-row__missing {
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 80%, transparent);
		font-size: 0.78rem;
	}

	.source-panel {
		display: grid;
		gap: 14px;
		padding: 16px;
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 15px;
		background: rgb(255 255 255 / 4%);
	}

	.source-panel__header {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 10px;
	}

	.source-panel__header h3 {
		margin: 0;
		font-size: 1rem;
		font-weight: 690;
	}

	.source-panel__header p {
		margin: 4px 0 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 85%, transparent);
		font-size: 0.81rem;
	}

	.source-panel__header span {
		display: inline-flex;
		align-items: center;
		min-height: 24px;
		padding: 0 8px;
		border-radius: 999px;
		border: 1px solid rgb(255 255 255 / 14%);
		background: rgb(255 255 255 / 5%);
		color: color-mix(in srgb, var(--vyrden-color-text) 88%, transparent);
		font-size: 0.72rem;
		font-weight: 620;
	}

	.source-panel__decision {
		margin: 0;
		color: color-mix(in srgb, var(--vyrden-color-text) 94%, transparent);
		font-size: 0.85rem;
		font-weight: 660;
	}

	.source-panel__reason {
		margin: 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 82%, transparent);
		font-size: 0.8rem;
		line-height: 1.38;
	}

	.source-panel__facts {
		display: grid;
		gap: 8px;
		margin: 0;
	}

	.source-panel__facts div {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 10px;
	}

	.source-panel__facts dt {
		color: color-mix(in srgb, var(--vyrden-color-text-soft) 90%, transparent);
		font-size: 0.73rem;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	.source-panel__facts dd {
		margin: 0;
		color: color-mix(in srgb, var(--vyrden-color-text) 90%, transparent);
		font-size: 0.78rem;
		font-family: var(--vyrden-font-mono);
		text-align: right;
	}

	.track-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 11px;
	}

	.technical-wrap--preview {
		opacity: 0.88;
	}

	.technical-wrap--preview :global(.detail-tech),
	.technical-wrap--preview .source-panel {
		border-color: rgb(255 255 255 / 8%);
		background: linear-gradient(180deg, rgb(255 255 255 / 3%), rgb(255 255 255 / 1%));
	}

	.technical-collapse {
		border: 1px solid rgb(255 255 255 / 10%);
		border-radius: 14px;
		background: rgb(255 255 255 / 4%);
		padding: 13px 14px 15px;
	}

	.technical-collapse summary {
		cursor: pointer;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 88%, transparent);
		font-size: 0.82rem;
		font-weight: 620;
		margin-bottom: 10px;
	}

	.track-empty {
		margin: 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 84%, transparent);
		font-size: 0.82rem;
	}

	.track-sidecar {
		margin: 10px 0 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 82%, transparent);
		font-size: 0.78rem;
	}

	@media (max-width: 980px) {
		.track-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 760px) {
		.episode-row {
			grid-template-columns: minmax(0, 1fr);
		}

		.episode-row__actions {
			justify-content: flex-start;
		}
	}

	@media (max-width: 620px) {
		.source-panel__facts div {
			flex-direction: column;
			align-items: flex-start;
		}

		.source-panel__facts dd {
			text-align: left;
		}
	}

</style>
