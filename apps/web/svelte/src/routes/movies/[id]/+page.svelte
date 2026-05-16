<script lang="ts">
	import { onMount } from 'svelte';
	import { Play } from 'lucide-svelte';
	import { scanMovies } from '$lib/api/browse';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getSessions, type SessionItem } from '$lib/api/operator';
	import {
		getDeviceProfiles,
		getMediaSourceDetail,
		getMediaSourceSubtitles,
		getMediaSourceTracks,
		getMovieDetail,
		getPlaybackDecision,
		getPlaybackRoute,
		getPlaybackState,
		startMediaProbe,
		listMediaSources,
		type DeviceProfile,
		type MediaSourceItem,
		type MediaSourceSubtitlesResponse,
		type MediaSourceTracksResponse,
		type MovieDetailResponse,
		type MovieVersion,
		type PlaybackDecisionResponse,
		type PlaybackRouteResponse,
		type PlaybackStateResponse
	} from '$lib/api/details';
	import DetailHero from '$lib/Xuva/DetailHero.svelte';
	import DetailSection from '$lib/Xuva/DetailSection.svelte';
	import XuvaButton from '$lib/Xuva/XuvaButton.svelte';
	import XuvaEmptyState from '$lib/Xuva/XuvaEmptyState.svelte';
	import XuvaPanel from '$lib/Xuva/XuvaPanel.svelte';
	import XuvaShell from '$lib/Xuva/XuvaShell.svelte';
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
	let isScanning = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let movie = $state<MovieDetailResponse | null>(null);
	let sourceModels = $state<MovieSourceModel[]>([]);
	let selectedSourceId = $state('');
	let routeResolution = $state<Record<string, RouteResolutionState>>({});
	let probingSourceId = $state('');
	let probeStatusSourceId = $state('');
	let probeStatusTone = $state<'info' | 'success' | 'error'>('info');
	let probeStatusMessage = $state('');
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
	let livePlaybackSourceId = $state('');
	let livePlaybackSession = $state<SessionItem | null>(null);
	let livePlaybackState = $state<PlaybackStateResponse | null>(null);
	let livePlaybackError = $state('');

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
		primaryMediaSourceId ? playHrefWithOptions(primaryMediaSourceId) : ''
	);
	const primaryStartHref = $derived.by(() =>
		primaryMediaSourceId ? playHrefWithOptions(primaryMediaSourceId, true) : ''
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
		const ticker = window.setInterval(() => {
			void refreshLivePlayback();
		}, 2500);
		return () => window.clearInterval(ticker);
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

	async function loadMovieDetails(): Promise<void> {
		isLoading = true;
		loadError = '';
		routeResolution = {};

		try {
			const [moviePayload, mediaSourcesPayload, profilesPayload] = await Promise.all([
				getMovieDetail(asText(params.id), apiClient),
				listMediaSources(apiClient, 1500).catch(() => ({ mediaSources: [] })),
				getDeviceProfiles(apiClient).catch(() => ({ profiles: [] }))
			]);
			deviceProfiles = profilesPayload.profiles || [];

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
			if (isApiStatus(error, 404)) {
				movie = null;
				sourceModels = [];
				selectedSourceId = '';
			} else {
				loadError = formatLoadError(error);
			}
		} finally {
			isLoading = false;
		}
	}

	async function startMovieScan(): Promise<void> {
		isScanning = true;
		actionMessage = '';
		try {
			await scanMovies(apiClient, 50);
			actionMessage = 'Movie scan started.';
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isScanning = false;
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
			probeStatusMessage = 'Media check started. Refreshing file details.';
			actionMessage = 'Media check started. Refreshing file details...';
			await loadMovieDetails();
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

	function playHrefWithOptions(sourceId: string, startFromZero = false): string {
		const id = asText(sourceId);
		if (!id) return '';
		const params = new URLSearchParams();
		const options = currentPlaybackOptions();
		params.set('clientProfile', options.clientProfile);
		params.set('routeType', options.routeType);
		if (options.maxNetworkBitrate > 0) params.set('maxNetworkBitrate', String(options.maxNetworkBitrate));
		params.set('audioTrackIndex', String(options.audioTrackIndex));
		params.set('subtitleTrackIndex', String(options.subtitleTrackIndex));
		params.set('subtitleTrackActive', options.subtitleTrackActive ? 'true' : 'false');
		params.set('supportsAdaptive', options.supportsAdaptive ? 'true' : 'false');
		params.set('autoplayIntent', '1');
		params.set('strictAutoplay', '1');
		params.set('forcePlayable', 'true');
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
		const sourceId = asText(livePlaybackSourceId || selectedSource?.mediaSourceId);
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

</script>

<XuvaShell>
	{#if isLoading}
		<XuvaPanel title="Loading Movie Details" subtitle="Fetching movie metadata, versions, and playback state." />
	{:else if loadError}
		<section class="px-4 pt-9 sm:px-6 lg:px-8">
			<XuvaEmptyState
				eyebrow="Connection"
				title="Media library unavailable"
				description="Xuva could not reach the media library service. Check that the server is running, then try again."
			>
				{#snippet primaryAction()}
					<XuvaButton variant="primary" onclick={loadMovieDetails}>Retry</XuvaButton>
				{/snippet}
				{#snippet secondaryAction()}
					<XuvaButton variant="secondary" href="/settings">Settings</XuvaButton>
					<XuvaButton variant="ghost" href="/">Back Home</XuvaButton>
				{/snippet}
			</XuvaEmptyState>
		</section>
	{:else if !movie}
		<section class="px-4 pt-9 sm:px-6 lg:px-8">
			<XuvaEmptyState
				eyebrow="Movies"
				title="Movie not found"
				description="Xuva could not find that movie. It may have been removed, renamed, or not scanned yet."
			>
				{#snippet primaryAction()}
					<XuvaButton variant="primary" href="/movies">Back to Movies</XuvaButton>
				{/snippet}
				{#snippet secondaryAction()}
					<XuvaButton variant="secondary" onclick={startMovieScan} disabled={isScanning}>
						{isScanning ? 'Scanning...' : 'Scan Movies'}
					</XuvaButton>
				{/snippet}
			</XuvaEmptyState>
			{#if actionMessage}
				<p class="mt-3 text-sm text-white/60">{actionMessage}</p>
			{/if}
		</section>
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
					<XuvaButton
						variant="primary"
						href={primaryPlayHref}
						onclick={() => markPlayingLive(primaryMediaSourceId)}
					>
						<Play size={18} class="fill-white text-white" />
						{primaryPlayLabel}
					</XuvaButton>
					<XuvaButton
						variant="secondary"
						href={primaryStartHref}
						onclick={() => markPlayingLive(primaryMediaSourceId)}
					>
						Play From Start
					</XuvaButton>
				{:else}
					<XuvaButton variant="primary" disabled>Play</XuvaButton>
				{/if}
			{/snippet}
		</DetailHero>

		{#if asText(livePlaybackSourceId || primaryMediaSourceId)}
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

		<DetailSection
			title="Versions"
			subtitle="Choose a source and start playback."
		>
			{#if sourceModels.length === 0}
				<p class="text-sm leading-relaxed text-white/60">
					This movie has no registered media sources yet. Run a library scan to refresh sources.
				</p>
			{:else}
				<div class="grid gap-4 lg:grid-cols-2 xl:grid-cols-3">
					{#each sourceModels as source (source.mediaSourceId)}
						<article class={`rounded-md border p-4 transition ${isSourceSelected(source.mediaSourceId) ? 'border-[#7C5CFF]/45 bg-[#7C5CFF]/8' : 'border-white/10 bg-white/[0.02]'}`}>
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
								<em class="shrink-0 rounded-sm border border-white/10 bg-[#111827] px-3 py-1 text-xs font-semibold not-italic text-white/70">
									{watchedLabel(source.state)}
								</em>
							</button>
							<p class="mt-4 text-sm font-semibold text-white">{playbackModeLabel(source.decision)}</p>
							<p class="mt-1 text-sm leading-relaxed text-white/55">{playbackReasonLabel(source.decision)}</p>
							<div class="mt-5 flex flex-wrap gap-3">
								<XuvaButton
									variant="primary"
									size="sm"
									href={playHrefWithOptions(source.mediaSourceId)}
									onclick={() => markPlayingLive(source.mediaSourceId)}
								>
									{isResumeState(source.state) ? 'Resume' : 'Play'}
								</XuvaButton>
								<XuvaButton
									variant="secondary"
									size="sm"
									href={playHrefWithOptions(source.mediaSourceId, true)}
									onclick={() => markPlayingLive(source.mediaSourceId)}
								>
									Start Over
								</XuvaButton>
								<XuvaButton
									variant="ghost"
									size="sm"
									onclick={() => resolvePlaybackRoute(source.mediaSourceId)}
									disabled={routeStateFor(source.mediaSourceId).status === 'loading'}
								>
									{routeStateFor(source.mediaSourceId).status === 'loading' ? 'Checking Route...' : 'Check Route'}
								</XuvaButton>
								{#if !Boolean(source.source?.probed)}
									<XuvaButton
										variant="secondary"
										size="sm"
										onclick={() => runMediaCheck(source.mediaSourceId)}
										disabled={probingSourceId === source.mediaSourceId}
									>
										{probingSourceId === source.mediaSourceId ? 'Checking File...' : 'Run Media Check'}
									</XuvaButton>
									{#if probeStatusMessage && probeStatusSourceId === source.mediaSourceId}
										<p class={`text-xs ${probeStatusTone === 'error' ? 'text-red-200/80' : probeStatusTone === 'success' ? 'text-emerald-200/80' : 'text-white/60'}`}>
											{probeStatusMessage}
										</p>
									{/if}
								{/if}
							</div>
							{#if routeStateFor(source.mediaSourceId).status === 'loaded' && routeStateFor(source.mediaSourceId).payload}
								<p class="mt-4 text-sm text-white/55">
									Route: {asText(routeStateFor(source.mediaSourceId).payload?.route) || 'pending'} -
									{asText(routeStateFor(source.mediaSourceId).payload?.status) || 'pending'}
								</p>
								{#if asText(routeStateFor(source.mediaSourceId).payload?.route) === 'blocked' && (routeStateFor(source.mediaSourceId).payload?.fallbackOptions || []).length > 0}
									<p class="mt-1 text-xs text-white/50">
										Fallbacks:
										{#each routeStateFor(source.mediaSourceId).payload?.fallbackOptions || [] as option, idx}
											{idx > 0 ? ', ' : ''}{asText(option.label) || asText(option.id) || 'Option'}
										{/each}
									</p>
								{/if}
							{:else if routeStateFor(source.mediaSourceId).status === 'error'}
								<p class="mt-4 text-sm text-red-200/80">{routeStateFor(source.mediaSourceId).error}</p>
							{/if}
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
						</article>
					{/each}
				</div>
			{/if}
		</DetailSection>

		<DetailSection
			title="Playback Analyzer"
			subtitle="Test this file against device profiles and choose playback options before you press Play."
		>
			{#if !selectedSource}
				<p class="text-sm leading-relaxed text-white/60">Select a source version to analyze playback behavior.</p>
			{:else}
				<div class="grid gap-4 lg:grid-cols-2">
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
						<select
							bind:value={analyzerSubtitleEnabled}
							onchange={() => {
								void refreshAnalyzer();
								void refreshDeviceMatrix();
							}}
						>
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
					{#if selectedSource}
						<XuvaButton
							variant="ghost"
							href={playHrefWithOptions(selectedSource.mediaSourceId)}
							onclick={() => markPlayingLive(selectedSource.mediaSourceId)}
						>
							Play With These Options
						</XuvaButton>
					{/if}
				</div>
				{#if analyzerError}
					<p class="mt-3 text-sm text-red-200/80">{analyzerError}</p>
				{/if}
				{#if analyzerDecision}
					<div class="mt-4 rounded-md border border-white/10 bg-[#111827]/55 p-4">
						<p class="text-sm font-semibold text-white">{playbackModeLabel(analyzerDecision)}</p>
						<p class="mt-1 text-sm text-white/60">{playbackReasonLabel(analyzerDecision)}</p>
						{#if analyzerRoute}
							<p class="mt-2 text-sm text-white/55">
								Route: {asText(analyzerRoute.route) || 'pending'} - {asText(analyzerRoute.status) || 'pending'}
							</p>
						{/if}
					</div>
				{/if}
				<div class="mt-4 rounded-md border border-white/10 bg-[#111827]/55 p-4">
					<h3 class="text-sm font-semibold text-white">Device Compatibility Matrix</h3>
					{#if analyzerMatrixError}
						<p class="mt-2 text-sm text-red-200/80">{analyzerMatrixError}</p>
					{:else if analyzerMatrixBusy}
						<p class="mt-2 text-sm text-white/55">Analyzing device profiles...</p>
					{:else if analyzerMatrix.length === 0}
						<p class="mt-2 text-sm text-white/55">No device profile results yet.</p>
					{:else}
						<div class="mt-3 grid gap-2">
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
			{/if}
		</DetailSection>

		<DetailSection
			title="Audio and Subtitles"
			subtitle="Technical stream details remain secondary and tied to the selected source."
		>
			{#if !selectedSource}
				<p class="text-sm leading-relaxed text-white/60">Select a source version to view track and subtitle details.</p>
			{:else}
				<div class="grid gap-4 lg:grid-cols-3">
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
						<article class="rounded-md border border-white/10 bg-[#111827]/55 p-4">
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
		</DetailSection>
	{/if}
</XuvaShell>
