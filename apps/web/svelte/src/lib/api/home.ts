import { apiClient, type ApiClient } from './client';

export interface ClientHomeItem {
	id?: string;
	kind?: string;
	title?: string;
	subtitle?: string;
	description?: string;
	overview?: string;
	mediaSourceId?: string;
	progressPercent?: number;
	percent?: number;
	posterUrl?: string;
	backdropUrl?: string;
	route?: string;
	[key: string]: unknown;
}

export interface ClientHomeRow {
	id?: string;
	title?: string;
	items?: ClientHomeItem[];
}

export interface ClientHomeResponse {
	profile?: string;
	heroes?: ClientHomeItem[];
	rows?: ClientHomeRow[];
	actions?: Record<string, string>;
}

export interface PlaybackRecentItem {
	userId?: string;
	mediaSourceId?: string;
	watched?: boolean;
	progressSeconds?: number;
	durationSeconds?: number;
	percent?: number;
	lastPlayedAt?: string;
	updatedAt?: string;
	name?: string;
	kind?: string;
	relPath?: string;
}

export interface PlaybackRecentResponse {
	recent?: PlaybackRecentItem[];
}

export interface LibraryRecord {
	id?: string;
	name?: string;
	path?: string;
	kind?: string;
	storageType?: string;
	metadataSources?: string[];
}

export interface LibrariesResponse {
	libraries?: LibraryRecord[];
	metadataSources?: unknown[];
}

export interface MovieDetailVersion {
	mediaSourceId?: string;
	qualityLabel?: string;
	edition?: string;
	sizeBytes?: number;
	relPath?: string;
}

export interface MovieDetailResponse {
	id?: string;
	title?: string;
	metadata?: {
		overview?: string;
		[key: string]: unknown;
	};
	versions?: MovieDetailVersion[];
	[key: string]: unknown;
}

export interface SeriesDetailEpisode {
	id?: string;
	title?: string;
	seasonNumber?: number;
	episodeNumber?: number;
	versionCount?: number;
	versions?: MovieDetailVersion[];
	thumbnailUrl?: string;
	overview?: string;
	airDate?: string;
	runtimeMinutes?: number;
	voteAverage?: number;
	[key: string]: unknown;
}

export interface SeriesDetailSeason {
	id?: string;
	seasonNumber?: number;
	episodes?: SeriesDetailEpisode[];
	posterUrl?: string;
	backdropUrl?: string;
	name?: string;
	overview?: string;
	airDate?: string;
	[key: string]: unknown;
}

export interface SeriesDetailResponse {
	id?: string;
	title?: string;
	metadata?: {
		overview?: string;
		[key: string]: unknown;
	};
	seasons?: SeriesDetailSeason[];
	[key: string]: unknown;
}

// ── Home response cache ───────────────────────────────────────────────────────
// The home page is visited frequently. Cache the response for 2 minutes so
// navigating Home → Movies → Home feels instant without stale-data risk.
let _homeCache: { data: ClientHomeResponse; exp: number } | null = null;
const HOME_TTL_MS = 2 * 60_000; // 2 minutes

/** Invalidate the home cache (call after playback stop so continue-watching refreshes). */
export function invalidateHomeCache(): void {
	_homeCache = null;
}

export function getClientHome(client: ApiClient = apiClient, limit = 24): Promise<ClientHomeResponse> {
	if (_homeCache && Date.now() < _homeCache.exp) return Promise.resolve(_homeCache.data);
	const path = `/api/client/home?limit=${encodeURIComponent(String(limit))}`;
	return client.request<ClientHomeResponse>(path).then(data => {
		_homeCache = { data, exp: Date.now() + HOME_TTL_MS };
		return data;
	});
}

export function getPlaybackRecent(
	client: ApiClient = apiClient,
	limit = 12
): Promise<PlaybackRecentResponse> {
	return client.request<PlaybackRecentResponse>(
		`/api/playback/recent?limit=${encodeURIComponent(String(limit))}`
	);
}

export function getLibraries(client: ApiClient = apiClient): Promise<LibrariesResponse> {
	return client.request<LibrariesResponse>('/api/libraries');
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

// ── Collections ─────────────────────────────────────────────────────────────

export interface CollectionMovie {
	id?: string;
	kind?: string;
	title?: string;
	year?: number;
	posterUrl?: string;
	backdropUrl?: string;
	logoUrl?: string;
	voteAverage?: number;
	genres?: string[];
	overview?: string;
	director?: string;
	runtimeMinutes?: number;
}

export interface CollectionDetailResponse {
	collection?: {
		id?: string;
		name?: string;
		posterUrl?: string;
		backdropUrl?: string;
		logoUrl?: string;
	};
	movies?: CollectionMovie[];
}

export function getCollectionDetail(
	id: string,
	client: ApiClient = apiClient
): Promise<CollectionDetailResponse> {
	return client.request<CollectionDetailResponse>(
		`/api/client/collections/${encodeURIComponent(id)}`
	);
}

export interface CollectionListItem {
	id: string;
	name: string;
	posterUrl?: string;
	backdropUrl?: string;
	movieCount: number;
}

export interface CollectionsListResponse {
	collections?: CollectionListItem[];
}

export function getCollections(
	client: ApiClient = apiClient
): Promise<CollectionsListResponse> {
	return client.request<CollectionsListResponse>('/api/client/collections');
}

// ── People ───────────────────────────────────────────────────────────────────

export interface PersonCreditItem {
	id?: string;
	kind?: string;
	title?: string;
	year?: number;
	character?: string;
	role?: string;
	posterUrl?: string;
	backdropUrl?: string;
	voteAverage?: number;
	genres?: string[];
	overview?: string;
}

export interface PersonDetailResponse {
	person?: {
		name?: string;
		profileUrl?: string;
		department?: string;
	};
	credits?: PersonCreditItem[];
}

export function getPersonDetail(
	name: string,
	client: ApiClient = apiClient
): Promise<PersonDetailResponse> {
	return client.request<PersonDetailResponse>(
		`/api/client/people/${encodeURIComponent(name)}`
	);
}

// ── Watchlist ─────────────────────────────────────────────────────────────────

export interface ServerWatchlistItem {
	userId: string;
	mediaId: string;
	kind: 'movie' | 'series';
	title: string;
	year?: number;
	posterUrl?: string;
	backdropUrl?: string;
	genres?: string[];
	addedAt: string;
}

export interface WatchlistListResponse {
	items: ServerWatchlistItem[];
}

export interface WatchlistAddRequest {
	mediaId: string;
	kind: 'movie' | 'series';
	title: string;
	year?: number;
	posterUrl?: string;
	backdropUrl?: string;
	genres?: string[];
}

export function getServerWatchlist(client: ApiClient = apiClient): Promise<WatchlistListResponse> {
	return client.request<WatchlistListResponse>('/api/client/watchlist');
}

export function addToServerWatchlist(
	req: WatchlistAddRequest,
	client: ApiClient = apiClient
): Promise<ServerWatchlistItem> {
	return client.request<ServerWatchlistItem, WatchlistAddRequest>('/api/client/watchlist', {
		method: 'POST',
		body: req
	});
}

export function removeFromServerWatchlist(
	mediaId: string,
	kind: 'movie' | 'series',
	client: ApiClient = apiClient
): Promise<void> {
	return client.request<void>(
		`/api/client/watchlist/${encodeURIComponent(mediaId)}?kind=${kind}`,
		{ method: 'DELETE' }
	);
}
