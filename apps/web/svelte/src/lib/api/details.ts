import { apiClient, type ApiClient } from './client';
import type { MetadataRecord } from './browse';

export interface MovieVersion {
	mediaSourceId?: string;
	path?: string;
	relPath?: string;
	edition?: string;
	qualityLabel?: string;
	sizeBytes?: number;
	modifiedAt?: string;
}

export interface MovieDetailResponse {
	id?: string;
	title?: string;
	year?: number;
	sortTitle?: string;
	needsReview?: boolean;
	versionCount?: number;
	metadata?: MetadataRecord;
	versions?: MovieVersion[];
}

export interface EpisodeBrief {
	id?: string;
	seasonNumber?: number;
	episodeNumber?: number;
	episodeEnd?: number;
	title?: string;
	needsReview?: boolean;
	versionCount?: number;
	versions?: MovieVersion[];
}

export interface SeasonDetail {
	id?: string;
	seasonNumber?: number;
	episodes?: EpisodeBrief[];
}

export interface SeriesDetailResponse {
	id?: string;
	title?: string;
	sortTitle?: string;
	seasonCount?: number;
	episodeCount?: number;
	metadata?: MetadataRecord;
	seasons?: SeasonDetail[];
}

export interface MediaSourceItem {
	id?: string;
	libraryId?: string;
	kind?: string;
	path?: string;
	relPath?: string;
	name?: string;
	extension?: string;
	sizeBytes?: number;
	modifiedAt?: string;
	probed?: boolean;
	container?: string;
	durationSeconds?: number;
	bitrate?: number;
	videoCodec?: string;
	width?: number;
	height?: number;
	audioStreams?: number;
	subtitleStreams?: number;
}

export interface MediaSourcesResponse {
	mediaSources?: MediaSourceItem[];
}

export interface ProbeTrack {
	index?: number;
	codec?: string;
	language?: string;
	title?: string;
	channels?: number;
	forced?: boolean;
	default?: boolean;
}

export interface MediaSourceTracksResponse {
	audioTracks?: ProbeTrack[];
	subtitleTracks?: ProbeTrack[];
}

export interface SubtitleSidecar {
	path?: string;
	relPath?: string;
	format?: string;
	language?: string;
	forced?: boolean;
	hearingImpaired?: boolean;
	requiresVideoBurn?: boolean;
}

export interface MediaSourceSubtitlesResponse {
	sidecars?: SubtitleSidecar[];
}

export interface PlaybackStateResponse {
	userId?: string;
	mediaSourceId?: string;
	watched?: boolean;
	progressSeconds?: number;
	durationSeconds?: number;
	percent?: number;
	lastPlayedAt?: string;
	updatedAt?: string;
}

export interface PlaybackDecisionResponse {
	mode?: string;
	reason?: string;
	reasonCode?: string;
	reasonText?: string;
	decisionTraceId?: string;
	mediaSourceId?: string;
	clientProfile?: string;
	containerAction?: string;
	videoAction?: string;
	audioAction?: string;
	subtitleAction?: string;
	subtitleClass?: string;
	estimatedCpuCost?: string;
	estimatedGpuCost?: string;
	estimatedNetworkBitrate?: number;
	selected?: Record<string, string>;
	suggestedFixes?: string[];
}

export interface PlaybackRouteResponse {
	route?: string;
	status?: string;
	url?: string;
	manifestUrl?: string;
	protocol?: string;
	policy?: string;
	decision?: PlaybackDecisionResponse;
	fallbackOptions?: Array<{ id?: string; label?: string; description?: string }>;
}

export interface PlaybackQueryOptions {
	clientProfile?: string;
	routeType?: string;
	maxNetworkBitrate?: number;
	audioTrackIndex?: number;
	audioCodec?: string;
	audioChannels?: number;
	subtitleTrackIndex?: number;
	subtitleCodec?: string;
	subtitleMode?: string;
	subtitleTrackActive?: boolean;
	supportsAdaptive?: boolean;
	// Capability-derived fields sent as query params (issue #64)
	supportsHdr?: boolean;
	maxBitDepth?: number;
	videoCodecs?: string[];
	audioCodecs?: string[];
}

export interface DeviceProfile {
	id?: string;
	name?: string;
	containers?: string[];
	videoCodecs?: string[];
	audioCodecs?: string[];
	subtitleCodecs?: string[];
	supportsHdr?: boolean;
	supportsToneMapping?: boolean;
	supportsHls?: boolean;
}

export function getMovieDetail(
	id: string,
	client: ApiClient = apiClient
): Promise<MovieDetailResponse> {
	return client.request<MovieDetailResponse>(`/api/movies/${encodeURIComponent(id)}`);
}

export function getSeriesDetail(
	id: string,
	client: ApiClient = apiClient
): Promise<SeriesDetailResponse> {
	return client.request<SeriesDetailResponse>(`/api/series/${encodeURIComponent(id)}`);
}

export function listMediaSources(
	client: ApiClient = apiClient,
	limit = 500
): Promise<MediaSourcesResponse> {
	return client.request<MediaSourcesResponse>(`/api/media-sources?limit=${encodeURIComponent(String(limit))}`);
}

export function getMediaSourceDetail(
	id: string,
	client: ApiClient = apiClient
): Promise<MediaSourceItem> {
	return client.request<MediaSourceItem>(`/api/media-sources/${encodeURIComponent(id)}`);
}

export function getMediaSourceTracks(
	id: string,
	client: ApiClient = apiClient
): Promise<MediaSourceTracksResponse> {
	return client.request<MediaSourceTracksResponse>(`/api/media-sources/${encodeURIComponent(id)}/tracks`);
}

export function getMediaSourceSubtitles(
	id: string,
	client: ApiClient = apiClient
): Promise<MediaSourceSubtitlesResponse> {
	return client.request<MediaSourceSubtitlesResponse>(
		`/api/media-sources/${encodeURIComponent(id)}/subtitles`
	);
}

export function getPlaybackState(
	mediaSourceId: string,
	client: ApiClient = apiClient
): Promise<PlaybackStateResponse> {
	return client.request<PlaybackStateResponse>(
		`/api/playback/state/${encodeURIComponent(mediaSourceId)}`
	);
}

export function getPlaybackDecision(
	mediaSourceId: string,
	options: PlaybackQueryOptions = {},
	client: ApiClient = apiClient
): Promise<PlaybackDecisionResponse> {
	const query = playbackQueryString(mediaSourceId, options);
	return client.request<PlaybackDecisionResponse>(`/api/playback/decision?${query}`);
}

export function getPlaybackRoute(
	mediaSourceId: string,
	options: PlaybackQueryOptions = {},
	client: ApiClient = apiClient
): Promise<PlaybackRouteResponse> {
	const query = playbackQueryString(mediaSourceId, options);
	return client.request<PlaybackRouteResponse>(`/api/playback/route?${query}`);
}

export function getDeviceProfiles(
	client: ApiClient = apiClient
): Promise<{ profiles?: DeviceProfile[] }> {
	return client.request<{ profiles?: DeviceProfile[] }>('/api/devices/profiles');
}

export function startMediaProbe(
	mediaSourceId: string,
	client: ApiClient = apiClient
): Promise<{ id?: string; status?: string }> {
	return client.send<{ id?: string; status?: string }, Record<string, never>>(
		`/api/media-sources/${encodeURIComponent(mediaSourceId)}/probe`,
		{},
		'POST'
	);
}

// ─── Client playback session lifecycle ──────────────────────────────────────

export interface ClientPlaybackSession {
	id?: string;
	mediaSourceId?: string;
	status?: string;
	startedAt?: string;
	defaultSubtitlesEnabled?: boolean;
}

export interface ClientCapabilities {
	containers: string[];
	videoCodecs: string[];
	audioCodecs: string[];
	subtitleCodecs: string[];
	maxVideoBitDepth: number;
	maxVideoFrameRate: number;
	supportsHdr: boolean;
	supportsDolbyVision: boolean;
	supportsHls: boolean;
}

export interface ClientPlaybackStartRequest {
	mediaSourceId: string;
	positionSeconds?: number;
	clientProfile?: string;
	deviceId?: string;
	clientCapabilities?: ClientCapabilities;
}

export function startClientPlayback(
	req: ClientPlaybackStartRequest,
	client: ApiClient = apiClient
): Promise<ClientPlaybackSession> {
	return client.send<ClientPlaybackSession, ClientPlaybackStartRequest>(
		'/api/client/playback/start',
		req,
		'POST'
	);
}

export interface ClientPlaybackHeartbeatRequest {
	positionSeconds: number;
	isPaused?: boolean;
}

export function heartbeatClientPlayback(
	sessionId: string,
	req: ClientPlaybackHeartbeatRequest,
	client: ApiClient = apiClient
): Promise<void> {
	return client.send<void, ClientPlaybackHeartbeatRequest>(
		`/api/client/playback/${encodeURIComponent(sessionId)}`,
		req,
		'PATCH'
	);
}

export interface ClientPlaybackStopRequest {
	positionSeconds: number;
	completed?: boolean;
}

export function stopClientPlayback(
	sessionId: string,
	req: ClientPlaybackStopRequest,
	client: ApiClient = apiClient
): Promise<void> {
	return client.send<void, ClientPlaybackStopRequest>(
		`/api/client/playback/${encodeURIComponent(sessionId)}/stop`,
		req,
		'POST'
	);
}

export interface SetPlaybackStateRequest {
	progressSeconds: number;
	durationSeconds?: number;
	watched?: boolean;
}

export function setPlaybackState(
	mediaSourceId: string,
	req: SetPlaybackStateRequest,
	client: ApiClient = apiClient
): Promise<void> {
	return client.send<void, SetPlaybackStateRequest>(
		`/api/playback/state/${encodeURIComponent(mediaSourceId)}`,
		req,
		'PUT'
	);
}

function playbackQueryString(mediaSourceId: string, options: PlaybackQueryOptions): string {
	const params = new URLSearchParams();
	params.set('mediaSourceId', mediaSourceId);
	const clientProfile = asText(options.clientProfile) || 'web';
	params.set('clientProfile', clientProfile);
	if (asText(options.routeType)) params.set('routeType', asText(options.routeType));
	if (Number.isFinite(options.maxNetworkBitrate)) {
		params.set('maxNetworkBitrate', String(Math.max(0, Math.round(Number(options.maxNetworkBitrate)))));
	}
	if (Number.isFinite(options.audioTrackIndex)) {
		params.set('audioTrackIndex', String(Math.max(0, Math.round(Number(options.audioTrackIndex)))));
	}
	if (asText(options.audioCodec)) params.set('audioCodec', asText(options.audioCodec));
	if (Number.isFinite(options.audioChannels)) {
		params.set('audioChannels', String(Math.max(0, Math.round(Number(options.audioChannels)))));
	}
	if (Number.isFinite(options.subtitleTrackIndex)) {
		params.set(
			'subtitleTrackIndex',
			String(Math.max(0, Math.round(Number(options.subtitleTrackIndex))))
		);
	}
	if (asText(options.subtitleCodec)) params.set('subtitleCodec', asText(options.subtitleCodec));
	if (asText(options.subtitleMode)) params.set('subtitleMode', asText(options.subtitleMode));
	if (typeof options.subtitleTrackActive === 'boolean') {
		params.set('subtitleTrackActive', options.subtitleTrackActive ? 'true' : 'false');
	}
	if (typeof options.supportsAdaptive === 'boolean') {
		params.set('supportsAdaptive', options.supportsAdaptive ? 'true' : 'false');
	}
	// Capability flags derived from browser detection (issue #64)
	if (typeof options.supportsHdr === 'boolean') {
		params.set('supportsHdr', options.supportsHdr ? 'true' : 'false');
	}
	if (typeof options.maxBitDepth === 'number' && options.maxBitDepth > 0) {
		params.set('maxBitDepth', String(options.maxBitDepth));
	}
	if (Array.isArray(options.videoCodecs) && options.videoCodecs.length > 0) {
		params.set('videoCodecs', options.videoCodecs.join(','));
	}
	if (Array.isArray(options.audioCodecs) && options.audioCodecs.length > 0) {
		params.set('audioCodecs', options.audioCodecs.join(','));
	}
	return params.toString();
}

function asText(value: unknown): string {
	return String(value ?? '').trim();
}
