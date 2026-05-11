<script lang="ts">
	import { onMount } from 'svelte';
	import { getAuthSession, type AuthSessionUser } from '$lib/api/auth';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import {
		getMediaSourceDetail,
		getMediaSourceSubtitles,
		getMediaSourceTracks,
		getMovieDetail,
		getPlaybackDecision,
		getPlaybackRoute,
		getPlaybackState,
		listMediaSources,
		type MediaSourceItem,
		type MediaSourceSubtitlesResponse,
		type MediaSourceTracksResponse,
		type MovieDetailResponse,
		type MovieVersion,
		type PlaybackDecisionResponse,
		type PlaybackRouteResponse,
		type PlaybackStateResponse
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
		LorivoButton,
		LorivoEmptyState,
		LorivoPanel
	} from '$lib/components';
	import {
		cleanDescription,
		cleanDisplayTitle,
		extractYear,
		formatBitrate,
		formatBytes,
		formatDuration,
		formatResolution,
		formatRuntime,
		formatTrackSummary,
		isResumeState,
		listItemTitle,
		playbackModeLabel,
		playbackPercent,
		playbackReasonLabel,
		resolveArtworkUrl,
		sourceQualityLabel,
		watchedLabel
	} from '$lib/details/model';

	interface MovieSourceModel {
		mediaSourceId: string;
		qualityLabel: string;
		edition: string;
		relPath: string;
		sizeBytes: number;
		modifiedAt: string;
		source: MediaSourceItem | null;
		tracks: MediaSourceTracksResponse;
		subtitles: MediaSourceSubtitlesResponse;
		state: PlaybackStateResponse | null;
		decision: PlaybackDecisionResponse | null;
	}

	interface RouteResolutionState {
		status: 'idle' | 'loading' | 'loaded' | 'error';
		payload: PlaybackRouteResponse | null;
		error: string;
	}

	let { params } = $props<{ params: { id: string } }>();

	let isLoading = $state(true);
	let loadError = $state('');
	let searchValue = $state('');
	let user = $state<AuthSessionUser | null>(null);
	let libraries = $state<LibraryRecord[]>([]);
	let movie = $state<MovieDetailResponse | null>(null);
	let sourceModels = $state<MovieSourceModel[]>([]);
	let selectedSourceId = $state('');
	let routeResolution = $state<Record<string, RouteResolutionState>>({});
	let previewDetailMode = $state(false);

	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
	const movieID = $derived.by(() => asText(movie?.id) || asText(params.id));
	const movieTitle = $derived.by(() =>
		listItemTitle(movie?.metadata?.title, movie?.title || asText(params.id) || 'Movie')
	);
	const movieYear = $derived.by(() => {
		const explicit = Number(movie?.year || movie?.metadata?.year || 0);
		if (Number.isFinite(explicit) && explicit > 1800) return explicit;
		return extractYear(`${movie?.metadata?.title || ''} ${movie?.title || ''}`);
	});
	const moviePosterUrl = $derived.by(() =>
		resolveArtworkUrl('movie', movieID, movie?.metadata?.posterUrl, 'poster')
	);
	const movieBackdropUrl = $derived.by(() =>
		resolveArtworkUrl('movie', movieID, movie?.metadata?.backdropUrl, 'backdrop')
	);
	const movieOverview = $derived.by(() =>
		cleanDescription(movie?.metadata?.overview || '', 360) ||
		'This title is available in your library and can be played with your current playback decisions.'
	);
	const selectedSource = $derived.by(
		() => sourceModels.find((item) => item.mediaSourceId === selectedSourceId) || sourceModels[0] || null
	);
	const primaryMediaSourceId = $derived.by(() => asText(selectedSource?.mediaSourceId));
	const primaryPlayHref = $derived.by(() =>
		primaryMediaSourceId ? `/play/${encodeURIComponent(primaryMediaSourceId)}` : ''
	);
	const primaryStartHref = $derived.by(() =>
		primaryMediaSourceId ? `/play/${encodeURIComponent(primaryMediaSourceId)}?start=0` : ''
	);
	const primaryPlayLabel = $derived.by(() => (isResumeState(selectedSource?.state || null) ? 'Resume' : 'Play'));
	const primaryProgress = $derived.by(() => playbackPercent(selectedSource?.state || null));
	const runtimeLabel = $derived.by(() => formatRuntime(Number(selectedSource?.source?.durationSeconds || 0)));
	const heroMeta = $derived.by(() => {
		const parts = [];
		if (movieYear > 0) parts.push(String(movieYear));
		if (runtimeLabel) parts.push(runtimeLabel);
		parts.push(`${sourceModels.length} version${sourceModels.length === 1 ? '' : 's'}`);
		return parts.join(' - ');
	});

	onMount(() => {
		void loadMovieDetails();
	});

	async function loadMovieDetails(): Promise<void> {
		isLoading = true;
		loadError = '';
		routeResolution = {};
		const previewMode = resolvePreviewMode(new URL(window.location.href).searchParams);
		const routeMovieId = asText(params.id);
		previewDetailMode = previewMode && isMoviePreviewId(routeMovieId);
		if (previewDetailMode) {
			const preview = buildMovieDetailPreview(routeMovieId);
			user = null;
			libraries = [];
			movie = preview.movie;
			sourceModels = preview.sources;
			selectedSourceId = preview.sources[0]?.mediaSourceId || '';
			isLoading = false;
			return;
		}

		try {
			const [sessionPayload, librariesPayload, moviePayload, mediaSourcesPayload] = await Promise.all([
				getAuthSession(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) return { user: null };
					throw error;
				}),
				getLibraries(apiClient),
				getMovieDetail(asText(params.id), apiClient),
				listMediaSources(apiClient, 1500).catch(() => ({ mediaSources: [] }))
			]);

			user = sessionPayload?.user || null;
			libraries = librariesPayload.libraries || [];
			movie = moviePayload;

			const sourceLookup = new Map<string, MediaSourceItem>();
			for (const item of mediaSourcesPayload.mediaSources || []) {
				const id = asText(item.id);
				if (!id) continue;
				sourceLookup.set(id, item);
			}

			const versions = Array.isArray(moviePayload.versions) ? moviePayload.versions : [];
			const loaded = await Promise.all(
				versions.map((version) => loadMovieSource(version, sourceLookup.get(asText(version.mediaSourceId)) || null))
			);
			sourceModels = loaded.filter((item): item is MovieSourceModel => Boolean(item));
			selectedSourceId = sourceModels[0]?.mediaSourceId || '';
		} catch (error) {
			loadError = formatLoadError(error);
		} finally {
			isLoading = false;
		}
	}

	async function loadMovieSource(
		version: MovieVersion,
		fallbackSource: MediaSourceItem | null
	): Promise<MovieSourceModel | null> {
		const mediaSourceId = asText(version.mediaSourceId);
		if (!mediaSourceId) return null;

		const [source, tracks, subtitles, state, decision] = await Promise.all([
			getMediaSourceDetail(mediaSourceId, apiClient).catch(() => fallbackSource),
			getMediaSourceTracks(mediaSourceId, apiClient).catch(() => ({
				audioTracks: [],
				subtitleTracks: []
			})),
			getMediaSourceSubtitles(mediaSourceId, apiClient).catch(() => ({ sidecars: [] })),
			getPlaybackState(mediaSourceId, apiClient).catch(() => null),
			getPlaybackDecision(
				mediaSourceId,
				{ clientProfile: 'web', routeType: 'remote', supportsAdaptive: true },
				apiClient
			).catch(() => null)
		]);

		return {
			mediaSourceId,
			qualityLabel: asText(version.qualityLabel),
			edition: asText(version.edition),
			relPath: asText(version.relPath || source?.relPath || ''),
			sizeBytes: Number(version.sizeBytes || source?.sizeBytes || 0),
			modifiedAt: asText(version.modifiedAt || source?.modifiedAt || ''),
			source,
			tracks,
			subtitles,
			state,
			decision
		};
	}

	async function resolvePlaybackRoute(mediaSourceId: string): Promise<void> {
		const id = asText(mediaSourceId);
		if (!id) return;
		routeResolution = {
			...routeResolution,
			[id]: { status: 'loading', payload: null, error: '' }
		};
		try {
			const payload = await getPlaybackRoute(
				id,
				{ clientProfile: 'web', routeType: 'remote', supportsAdaptive: true },
				apiClient
			);
			routeResolution = {
				...routeResolution,
				[id]: { status: 'loaded', payload, error: '' }
			};
		} catch (error) {
			routeResolution = {
				...routeResolution,
				[id]: { status: 'error', payload: null, error: formatLoadError(error) }
			};
		}
	}

	function routeStateFor(mediaSourceId: string): RouteResolutionState {
		return routeResolution[mediaSourceId] || { status: 'idle', payload: null, error: '' };
	}

	function selectSource(mediaSourceId: string): void {
		selectedSourceId = asText(mediaSourceId);
	}

	function isSourceSelected(mediaSourceId: string): boolean {
		return asText(mediaSourceId) === asText(selectedSourceId || sourceModels[0]?.mediaSourceId || '');
	}

	function initialsForName(name: string): string {
		const words = asText(name).split(/\s+/).filter(Boolean);
		if (words.length === 0) return 'V';
		if (words.length === 1) return words[0].slice(0, 1).toUpperCase();
		return `${words[0][0] || ''}${words[1][0] || ''}`.toUpperCase();
	}

	function movieSourceMeta(source: MovieSourceModel): string {
		const resolution = formatResolution(
			Number(source.source?.width || 0),
			Number(source.source?.height || 0)
		);
		return [
			source.qualityLabel,
			resolution !== 'Unknown resolution' ? resolution : '',
			source.edition,
			source.modifiedAt ? formatDate(source.modifiedAt) : ''
		]
			.filter(Boolean)
			.join(' - ');
	}

	function formatDate(value: string): string {
		const stamp = Date.parse(value);
		if (!Number.isFinite(stamp)) return '';
		return new Intl.DateTimeFormat(undefined, { dateStyle: 'medium' }).format(new Date(stamp));
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
		return 'Movie details could not load.';
	}

	function isMoviePreviewId(id: string): boolean {
		return id === 'preview' || id.startsWith('preview-movie-');
	}

	function buildMovieDetailPreview(id: string): { movie: MovieDetailResponse; sources: MovieSourceModel[] } {
		const movieId = id === 'preview' ? 'movie-ember-harbor' : id;
		const movieTitle = id === 'preview' ? 'Ember Harbor' : cleanIdTitle(id);
		const movie: MovieDetailResponse = {
			id: movieId,
			title: movieTitle,
			year: 2025,
			needsReview: false,
			versionCount: 2,
			metadata: {
				title: movieTitle,
				year: 2025,
				posterUrl: previewPoster(movieTitle),
				backdropUrl: previewBackdrop(movieTitle),
				overview:
					'When long-range navigation lights flicker across the bay, one captain is forced to choose between silence and a broadcast that could reshape the coast.'
			},
			versions: [
				{
					mediaSourceId: `${movieId}-src-1080`,
					qualityLabel: '1080p',
					edition: 'Theatrical',
					relPath: 'Library/Movies/Feature Film Preview',
					sizeBytes: 7_600_000_000,
					modifiedAt: '2026-04-20T13:00:00Z'
				},
				{
					mediaSourceId: `${movieId}-src-4k`,
					qualityLabel: '2160p',
					edition: 'Cinema',
					relPath: 'Library/Movies/Feature Film Preview',
					sizeBytes: 14_200_000_000,
					modifiedAt: '2026-04-28T08:00:00Z'
				}
			]
		};

		const sources: MovieSourceModel[] = [
			buildPreviewSource(`${movieId}-src-1080`, '1080p', 'Theatrical', 6_840, 8_500_000, 1920, 1080),
			buildPreviewSource(`${movieId}-src-4k`, '2160p', 'Cinema', 6_840, 16_500_000, 3840, 2160)
		];
		return { movie, sources };
	}

	function buildPreviewSource(
		mediaSourceId: string,
		qualityLabel: string,
		edition: string,
		durationSeconds: number,
		bitrate: number,
		width: number,
		height: number
	): MovieSourceModel {
		return {
			mediaSourceId,
			qualityLabel,
			edition,
			relPath: 'Preview source',
			sizeBytes: Math.round((durationSeconds * bitrate) / 8),
			modifiedAt: '2026-05-01T09:00:00Z',
			source: {
				id: mediaSourceId,
				kind: 'movie',
				durationSeconds,
				bitrate,
				width,
				height,
				sizeBytes: Math.round((durationSeconds * bitrate) / 8),
				container: 'mkv'
			},
			tracks: {
				audioTracks: [
					{ codec: 'eac3', language: 'en', channels: 6, default: true },
					{ codec: 'aac', language: 'es', channels: 2 }
				],
				subtitleTracks: [
					{ codec: 'srt', language: 'en', default: true },
					{ codec: 'srt', language: 'fr' }
				]
			},
			subtitles: {
				sidecars: [{ relPath: 'preview/en.srt', format: 'srt', language: 'en' }]
			},
			state: { mediaSourceId, watched: false, progressSeconds: 1_420, durationSeconds, percent: 20.8 },
			decision: {
				mode: 'directplay',
				reasonText: 'Browser profile supports this source.',
				estimatedNetworkBitrate: bitrate,
				containerAction: 'copy',
				videoAction: 'copy',
				audioAction: 'copy',
				subtitleAction: 'copy'
			}
		};
	}

	function cleanIdTitle(value: string): string {
		const raw = asText(value).replace(/^preview-movie-/, '').replace(/-/g, ' ');
		if (!raw) return 'Movie Preview';
		return raw.replace(/\b\w/g, (letter) => letter.toUpperCase());
	}

	// preview artwork comes from shared local preview artwork system
</script>

<MediaShell active="movies" bind:searchValue {userInitials}>
	<DetailPage>
		{#if isLoading}
			<LorivoPanel title="Loading Movie Details" subtitle="Fetching movie metadata, versions, and playback state." />
		{:else if loadError}
			<LorivoPanel title="Movie details could not load" subtitle={loadError}>
				<div class="status-actions">
					<LorivoButton variant="secondary" onclick={loadMovieDetails}>Retry</LorivoButton>
					<LorivoButton variant="ghost" href="/movies">Back to Movies</LorivoButton>
				</div>
			</LorivoPanel>
		{:else if !movie}
			<LorivoEmptyState title="Movie not found" message="This movie is no longer available in your library.">
				{#snippet action()}
					<LorivoButton variant="secondary" href="/movies">Back to Movies</LorivoButton>
				{/snippet}
			</LorivoEmptyState>
		{:else}
			<DetailHero
				title={movieTitle}
				meta={heroMeta}
				overview={movieOverview}
				backHref="/movies"
				backLabel="Back to Movies"
				backdropUrl={movieBackdropUrl}
				posterUrl={moviePosterUrl}
				progressLabel={primaryProgress > 0 ? `${primaryProgress}% watched` : ''}
			>
				{#snippet actions()}
					{#if primaryPlayHref}
						<LorivoButton variant="primary" href={primaryPlayHref}>{primaryPlayLabel}</LorivoButton>
						<LorivoButton variant="secondary" href={primaryStartHref}>Play From Start</LorivoButton>
					{:else}
						<LorivoButton variant="primary" disabled>Play</LorivoButton>
					{/if}
				{/snippet}
			</DetailHero>

			<DetailSection
				title={previewDetailMode ? 'Playback Options' : 'Versions'}
				subtitle={previewDetailMode ? 'Pick how you want to watch this title.' : 'Choose a source and start playback.'}
			>
				{#if sourceModels.length === 0}
					<LorivoEmptyState
						title="No playable versions"
						message="This movie has no registered media sources yet. Run a library scan to refresh sources."
					/>
				{:else}
					<div class="version-grid">
						{#each sourceModels as source (source.mediaSourceId)}
							<article class="version-card" class:version-card--selected={isSourceSelected(source.mediaSourceId)}>
								<button
									class="version-card__selector"
									type="button"
									aria-label={`Select ${sourceQualityLabel(source.qualityLabel, formatResolution(Number(source.source?.width || 0), Number(source.source?.height || 0)))}`}
									onclick={() => selectSource(source.mediaSourceId)}
								>
									<div>
										<strong>
											{sourceQualityLabel(
												source.qualityLabel,
												formatResolution(Number(source.source?.width || 0), Number(source.source?.height || 0))
											)}
										</strong>
										<span>{movieSourceMeta(source)}</span>
									</div>
									<em>{watchedLabel(source.state)}</em>
								</button>
								{#if !previewDetailMode}
									<p class="version-card__decision">{playbackModeLabel(source.decision)}</p>
									<p class="version-card__reason">{playbackReasonLabel(source.decision)}</p>
								{/if}
								<div class="version-card__actions">
									<LorivoButton variant="primary" href={`/play/${encodeURIComponent(source.mediaSourceId)}`}>
										{isResumeState(source.state) ? 'Resume' : 'Play'}
									</LorivoButton>
									<LorivoButton
										variant="secondary"
										href={`/play/${encodeURIComponent(source.mediaSourceId)}?start=0`}
									>
										Start Over
									</LorivoButton>
									{#if !previewDetailMode}
										<LorivoButton
											variant="ghost"
											onclick={() => resolvePlaybackRoute(source.mediaSourceId)}
											disabled={routeStateFor(source.mediaSourceId).status === 'loading'}
										>
											{routeStateFor(source.mediaSourceId).status === 'loading'
												? 'Checking route...'
												: 'Check Route'}
										</LorivoButton>
									{/if}
								</div>
								{#if !previewDetailMode && routeStateFor(source.mediaSourceId).status === 'loaded' && routeStateFor(source.mediaSourceId).payload}
									<p class="route-note">
										Route: {asText(routeStateFor(source.mediaSourceId).payload?.route) || 'pending'} -
										{asText(routeStateFor(source.mediaSourceId).payload?.status) || 'pending'}
									</p>
								{:else if !previewDetailMode && routeStateFor(source.mediaSourceId).status === 'error'}
									<p class="route-note route-note--error">{routeStateFor(source.mediaSourceId).error}</p>
								{/if}
								{#if !previewDetailMode}
									<dl class="version-card__facts">
										<div>
											<dt>Duration</dt>
											<dd>{formatDuration(Number(source.source?.durationSeconds || 0))}</dd>
										</div>
										<div>
											<dt>Bitrate</dt>
											<dd>{formatBitrate(Number(source.source?.bitrate || 0))}</dd>
										</div>
										<div>
											<dt>File Size</dt>
											<dd>{formatBytes(source.sizeBytes)}</dd>
										</div>
									</dl>
								{/if}
							</article>
						{/each}
					</div>
				{/if}
			</DetailSection>

			<DetailSection
				title={previewDetailMode ? 'Source Details' : 'Audio and Subtitles'}
				subtitle={previewDetailMode
					? 'Technical source details are available when needed.'
					: 'Technical stream details remain secondary and tied to the selected source.'}
			>
				<div class="technical-wrap" class:technical-wrap--preview={previewDetailMode}>
				{#if !selectedSource}
					<LorivoEmptyState
						title="No source selected"
						message="Select a source version to view track and subtitle details."
					/>
				{:else if previewDetailMode}
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
						<DetailTechnicalPanel title="Source Details">
							<ul>
								<li>
									<strong>Source Path</strong>
									<span>{selectedSource.relPath || 'Local source'}</span>
								</li>
								<li>
									<strong>Last Modified</strong>
									<span>{selectedSource.modifiedAt ? formatDate(selectedSource.modifiedAt) : 'Unknown'}</span>
								</li>
							</ul>
							<p class="track-sidecar">Technical source fields are intentionally secondary.</p>
						</DetailTechnicalPanel>
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
		gap: var(--lorivo-space-2);
	}

	.version-grid {
		display: grid;
		grid-template-columns: repeat(auto-fill, minmax(290px, 1fr));
		gap: 12px;
	}

	.version-card {
		display: grid;
		gap: 10px;
		padding: 12px;
		border: 1px solid rgb(255 246 229 / 10%);
		border-radius: 12px;
		background: linear-gradient(180deg, rgb(255 246 229 / 4%), rgb(255 246 229 / 2%));
	}

	.version-card--selected {
		border-color: color-mix(in srgb, var(--lorivo-color-accent-teal) 42%, transparent);
		background:
			linear-gradient(180deg, rgb(255 246 229 / 6%), rgb(255 246 229 / 2%)),
			color-mix(in srgb, var(--lorivo-color-accent-teal) 10%, transparent);
	}

	.version-card__selector {
		display: flex;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
		padding: 0;
		border: 0;
		background: none;
		text-align: left;
		color: inherit;
		cursor: pointer;
	}

	.version-card__selector strong {
		display: block;
		font-size: 0.96rem;
		font-weight: 690;
		letter-spacing: 0.005em;
	}

	.version-card__selector span {
		display: block;
		margin-top: 3px;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 84%, transparent);
		font-size: 0.8rem;
		line-height: 1.33;
	}

	.version-card__selector em {
		flex-shrink: 0;
		display: inline-flex;
		align-items: center;
		min-height: 26px;
		padding: 0 9px;
		border-radius: 999px;
		border: 1px solid rgb(255 246 229 / 14%);
		background: rgb(255 246 229 / 5%);
		color: color-mix(in srgb, var(--lorivo-color-text) 88%, transparent);
		font-size: 0.74rem;
		font-style: normal;
		font-weight: 630;
	}

	.version-card__decision {
		margin: 0;
		color: color-mix(in srgb, var(--lorivo-color-text) 94%, transparent);
		font-size: 0.85rem;
		font-weight: 660;
	}

	.version-card__reason {
		margin: 0;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 82%, transparent);
		font-size: 0.8rem;
		line-height: 1.38;
	}

	.version-card__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
	}

	.route-note {
		margin: 0;
		font-size: 0.78rem;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 88%, transparent);
	}

	.route-note--error {
		color: color-mix(in srgb, var(--lorivo-color-danger) 84%, white 16%);
	}

	.version-card__facts {
		display: grid;
		gap: 8px;
		margin: 0;
	}

	.version-card__facts div {
		display: flex;
		align-items: baseline;
		justify-content: space-between;
		gap: 10px;
	}

	.version-card__facts dt {
		color: color-mix(in srgb, var(--lorivo-color-text-soft) 90%, transparent);
		font-size: 0.73rem;
		letter-spacing: 0.05em;
		text-transform: uppercase;
	}

	.version-card__facts dd {
		margin: 0;
		color: color-mix(in srgb, var(--lorivo-color-text) 90%, transparent);
		font-size: 0.78rem;
		font-family: var(--lorivo-font-mono);
		text-align: right;
		word-break: break-word;
	}

	.track-grid {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(220px, 1fr));
		gap: 11px;
	}

	.technical-wrap--preview {
		opacity: 0.88;
	}

	.technical-wrap--preview :global(.detail-tech) {
		border-color: rgb(255 246 229 / 8%);
		background: linear-gradient(180deg, rgb(255 246 229 / 3%), rgb(255 246 229 / 1%));
	}

	.technical-collapse {
		border: 1px solid rgb(255 246 229 / 10%);
		border-radius: 11px;
		background: rgb(255 246 229 / 2%);
		padding: 10px 10px 12px;
	}

	.technical-collapse summary {
		cursor: pointer;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 88%, transparent);
		font-size: 0.82rem;
		font-weight: 620;
		margin-bottom: 10px;
	}

	.track-empty {
		margin: 0;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 84%, transparent);
		font-size: 0.82rem;
	}

	.track-sidecar {
		margin: 10px 0 0;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 82%, transparent);
		font-size: 0.78rem;
	}

	@media (max-width: 920px) {
		.track-grid {
			grid-template-columns: minmax(0, 1fr);
		}
	}

	@media (max-width: 680px) {
		.version-grid {
			grid-template-columns: minmax(0, 1fr);
		}

		.version-card__facts div {
			flex-direction: column;
			align-items: flex-start;
		}

		.version-card__facts dd {
			text-align: left;
		}
	}
</style>
