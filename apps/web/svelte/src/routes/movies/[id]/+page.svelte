<script lang="ts">
	import { onMount } from 'svelte';
	import { Play } from 'lucide-svelte';
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
	import { resolvePreviewMode } from '$lib/home/model';
	import { previewBackdrop, previewPoster } from '$lib/preview/artwork';
	import DetailHero from '$lib/lorivo/DetailHero.svelte';
	import DetailSection from '$lib/lorivo/DetailSection.svelte';
	import LorivoButton from '$lib/lorivo/LorivoButton.svelte';
	import LorivoPanel from '$lib/lorivo/LorivoPanel.svelte';
	import LorivoShell from '$lib/lorivo/LorivoShell.svelte';
	import {
		cleanDescription,
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
	let movie = $state<MovieDetailResponse | null>(null);
	let sourceModels = $state<MovieSourceModel[]>([]);
	let selectedSourceId = $state('');
	let routeResolution = $state<Record<string, RouteResolutionState>>({});
	let previewDetailMode = $state(false);

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
			movie = preview.movie;
			sourceModels = preview.sources;
			selectedSourceId = preview.sources[0]?.mediaSourceId || '';
			isLoading = false;
			return;
		}

		try {
			const [moviePayload, mediaSourcesPayload] = await Promise.all([
				getMovieDetail(asText(params.id), apiClient),
				listMediaSources(apiClient, 1500).catch(() => ({ mediaSources: [] }))
			]);

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
</script>

<svelte:head>
	<title>{movieTitle} - Lorivo Media</title>
</svelte:head>

<LorivoShell>
	{#if isLoading}
		<LorivoPanel title="Loading Movie Details" subtitle="Fetching movie metadata, versions, and playback state." />
	{:else if loadError}
		<LorivoPanel title="Movie details could not load" subtitle={loadError}>
			<div class="flex flex-wrap gap-3">
				<LorivoButton variant="secondary" onclick={loadMovieDetails}>Retry</LorivoButton>
				<LorivoButton variant="ghost" href="/movies">Back to Movies</LorivoButton>
			</div>
		</LorivoPanel>
	{:else if !movie}
		<LorivoPanel title="Movie not found" subtitle="This movie is no longer available in your library.">
			<LorivoButton variant="secondary" href="/movies">Back to Movies</LorivoButton>
		</LorivoPanel>
	{:else}
		<DetailHero
			title={movieTitle}
			meta={heroMeta}
			overview={movieOverview}
			backHref="/movies"
			backLabel="Back to Movies"
			backdropUrl={movieBackdropUrl}
			posterUrl={moviePosterUrl}
			progress={primaryProgress}
			progressLabel={primaryProgress > 0 ? `${primaryProgress}% watched` : ''}
		>
			{#snippet actions()}
				{#if primaryPlayHref}
					<LorivoButton variant="primary" href={primaryPlayHref}>
						<Play size={18} class="fill-white text-white" />
						{primaryPlayLabel}
					</LorivoButton>
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
				<p class="text-sm leading-relaxed text-white/60">
					This movie has no registered media sources yet. Run a library scan to refresh sources.
				</p>
			{:else}
				<div class="grid gap-4 lg:grid-cols-2 xl:grid-cols-3">
					{#each sourceModels as source (source.mediaSourceId)}
						<article class={`rounded-2xl border p-5 shadow-lg shadow-black/20 transition ${isSourceSelected(source.mediaSourceId) ? 'border-[#7C5CFF]/45 bg-[#7C5CFF]/10' : 'border-white/10 bg-white/[0.04]'}`}>
							<button
								class="flex w-full items-start justify-between gap-4 text-left text-white"
								type="button"
								aria-label={`Select ${sourceQualityLabel(source.qualityLabel, formatResolution(Number(source.source?.width || 0), Number(source.source?.height || 0)))}`}
								onclick={() => selectSource(source.mediaSourceId)}
							>
								<span>
									<strong class="block text-lg font-semibold text-white">
										{sourceQualityLabel(
											source.qualityLabel,
											formatResolution(Number(source.source?.width || 0), Number(source.source?.height || 0))
										)}
									</strong>
									<span class="mt-1 block text-sm leading-relaxed text-white/50">{movieSourceMeta(source)}</span>
								</span>
								<em class="shrink-0 rounded-full border border-white/10 bg-[#111827] px-3 py-1 text-xs font-semibold not-italic text-white/70">
									{watchedLabel(source.state)}
								</em>
							</button>
							{#if !previewDetailMode}
								<p class="mt-4 text-sm font-semibold text-white">{playbackModeLabel(source.decision)}</p>
								<p class="mt-1 text-sm leading-relaxed text-white/55">{playbackReasonLabel(source.decision)}</p>
							{/if}
							<div class="mt-5 flex flex-wrap gap-3">
								<LorivoButton variant="primary" size="sm" href={`/play/${encodeURIComponent(source.mediaSourceId)}`}>
									{isResumeState(source.state) ? 'Resume' : 'Play'}
								</LorivoButton>
								<LorivoButton
									variant="secondary"
									size="sm"
									href={`/play/${encodeURIComponent(source.mediaSourceId)}?start=0`}
								>
									Start Over
								</LorivoButton>
								{#if !previewDetailMode}
									<LorivoButton
										variant="ghost"
										size="sm"
										onclick={() => resolvePlaybackRoute(source.mediaSourceId)}
										disabled={routeStateFor(source.mediaSourceId).status === 'loading'}
									>
										{routeStateFor(source.mediaSourceId).status === 'loading' ? 'Checking Route...' : 'Check Route'}
									</LorivoButton>
								{/if}
							</div>
							{#if !previewDetailMode && routeStateFor(source.mediaSourceId).status === 'loaded' && routeStateFor(source.mediaSourceId).payload}
								<p class="mt-4 text-sm text-white/55">
									Route: {asText(routeStateFor(source.mediaSourceId).payload?.route) || 'pending'} -
									{asText(routeStateFor(source.mediaSourceId).payload?.status) || 'pending'}
								</p>
							{:else if !previewDetailMode && routeStateFor(source.mediaSourceId).status === 'error'}
								<p class="mt-4 text-sm text-red-200/80">{routeStateFor(source.mediaSourceId).error}</p>
							{/if}
							{#if !previewDetailMode}
								<dl class="mt-5 grid gap-2 text-sm">
									<div class="flex items-baseline justify-between gap-3">
										<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Duration</dt>
										<dd class="text-right text-white/70">{formatDuration(Number(source.source?.durationSeconds || 0))}</dd>
									</div>
									<div class="flex items-baseline justify-between gap-3">
										<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Bitrate</dt>
										<dd class="text-right text-white/70">{formatBitrate(Number(source.source?.bitrate || 0))}</dd>
									</div>
									<div class="flex items-baseline justify-between gap-3">
										<dt class="text-xs uppercase tracking-[0.08em] text-white/40">File Size</dt>
										<dd class="text-right text-white/70">{formatBytes(source.sizeBytes)}</dd>
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
			{#if !selectedSource}
				<p class="text-sm leading-relaxed text-white/60">Select a source version to view track and subtitle details.</p>
			{:else}
				{#if previewDetailMode}
					<details class="rounded-2xl border border-white/10 bg-white/[0.04] p-4">
						<summary class="cursor-pointer text-sm font-semibold text-white/70">Technical playback details</summary>
						<div class="mt-4 grid gap-4 lg:grid-cols-2">
							<article class="rounded-2xl border border-white/10 bg-[#111827]/70 p-4">
								<h3 class="text-base font-semibold text-white">Audio Tracks</h3>
								{#if (selectedSource.tracks.audioTracks || []).length === 0}
									<p class="mt-3 text-sm text-white/55">No audio track data is available yet.</p>
								{:else}
									<ul class="mt-3 grid gap-2">
										{#each selectedSource.tracks.audioTracks || [] as track, index (index)}
											<li class="flex flex-wrap items-center justify-between gap-2 text-sm text-white/70">
												<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language, Number(track.channels || 0))}</strong>
												{#if track.default}<span class="rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
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
											<li class="flex flex-wrap items-center justify-between gap-2 text-sm text-white/70">
												<strong class="font-medium text-white">{formatTrackSummary(track.codec, track.language)}</strong>
												<span class="flex gap-2">
													{#if track.forced}<span class="rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Forced</span>{/if}
													{#if track.default}<span class="rounded-full bg-white/10 px-2 py-1 text-xs text-white/60">Default</span>{/if}
												</span>
											</li>
										{/each}
									</ul>
								{/if}
								<p class="mt-3 text-sm text-white/50">Sidecar subtitles: {(selectedSource.subtitles.sidecars || []).length}</p>
							</article>
						</div>
					</details>
				{:else}
					<div class="grid gap-4 lg:grid-cols-3">
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
						<article class="rounded-2xl border border-white/10 bg-[#111827]/70 p-4">
							<h3 class="text-base font-semibold text-white">Source Details</h3>
							<dl class="mt-3 grid gap-3 text-sm">
								<div>
									<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Source Path</dt>
									<dd class="mt-1 break-words text-white/70">{selectedSource.relPath || 'Local source'}</dd>
								</div>
								<div>
									<dt class="text-xs uppercase tracking-[0.08em] text-white/40">Last Modified</dt>
									<dd class="mt-1 text-white/70">{selectedSource.modifiedAt ? formatDate(selectedSource.modifiedAt) : 'Unknown'}</dd>
								</div>
							</dl>
						</article>
					</div>
				{/if}
			{/if}
		</DetailSection>
	{/if}
</LorivoShell>
