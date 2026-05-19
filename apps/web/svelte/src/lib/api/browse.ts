import { apiClient, type ApiClient } from './client';

export interface MetadataRecord {
	kind?: string;
	itemId?: string;
	provider?: string;
	title?: string;
	year?: number;
	overview?: string;
	posterUrl?: string;
	backdropUrl?: string;
	[key: string]: unknown;
}

export interface MovieListItem {
	id?: string;
	title?: string;
	year?: number;
	sortTitle?: string;
	needsReview?: boolean;
	versionCount?: number;
	metadata?: MetadataRecord;
}

export interface SeriesListItem {
	id?: string;
	title?: string;
	sortTitle?: string;
	seasonCount?: number;
	episodeCount?: number;
	metadata?: MetadataRecord;
}

export interface MoviesResponse {
	movies?: MovieListItem[];
}

export interface SeriesResponse {
	series?: SeriesListItem[];
}

export interface MetadataRefreshResponse {
	accepted?: number;
	skipped?: number;
	warnings?: string[];
	error?: string;
	[key: string]: unknown;
}

export interface ReviewItem {
	kind?: string;
	id?: string;
	title?: string;
	reviewReason?: string;
}

export interface ReviewItemsResponse {
	items?: ReviewItem[];
}

export interface VersionGroup {
	kind?: string;
	id?: string;
	title?: string;
	versionCount?: number;
}

export interface VersionGroupsResponse {
	versions?: VersionGroup[];
}

export interface MetadataProviderRecord {
	id?: string;
	name?: string;
	status?: string;
	local?: boolean;
}

export interface MetadataRecordsResponse {
	best?: MetadataRecord | null;
	records?: MetadataRecord[];
	providers?: MetadataProviderRecord[];
}

export interface MetadataRefreshRequest {
	kind: string;
	id: string;
	title?: string;
	year?: number;
}

export interface MetadataMatchRequest {
	kind: string;
	id: string;
	title: string;
	year?: number;
	overview?: string;
	provider?: string;
	externalId?: string;
	posterUrl?: string;
	backdropUrl?: string;
	review: boolean;
}

export function getMovies(client: ApiClient = apiClient, limit = 500): Promise<MoviesResponse> {
	return client.request<MoviesResponse>(`/api/movies?limit=${encodeURIComponent(String(limit))}`);
}

export function getSeries(client: ApiClient = apiClient, limit = 500): Promise<SeriesResponse> {
	return client.request<SeriesResponse>(`/api/series?limit=${encodeURIComponent(String(limit))}`);
}

export function scanMovies(
	client: ApiClient = apiClient,
	sampleLimit = 50
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, { sampleLimit: number }>(
		'/api/libraries/movies/scan',
		{ sampleLimit },
		'POST'
	);
}

export function scanTV(
	client: ApiClient = apiClient,
	sampleLimit = 50
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, { sampleLimit: number }>(
		'/api/libraries/tv/scan',
		{ sampleLimit },
		'POST'
	);
}

export function refreshMetadataBatch(
	kind: 'movie' | 'series',
	client: ApiClient = apiClient,
	limit = 25
): Promise<MetadataRefreshResponse> {
	return client.send<MetadataRefreshResponse, { kind: 'movie' | 'series'; limit: number }>(
		'/api/metadata/refresh-batch',
		{ kind, limit },
		'POST'
	);
}

// ── Backfill: fill in metadata for items missing a specific provider ──────
//
// Used by Settings → Metadata to populate TMDB/Fanart for items that
// currently have only filename/Wikipedia data. The server auto-starts a
// TMDB backfill on boot when items are missing TMDB rows; the UI exposes
// the same controls explicitly.

export interface BackfillStatus {
	running: boolean;
	provider?: string;
	kind?: string;
	startedAt?: string;
	finishedAt?: string;
	total: number;
	refreshed: number;
	failed: number;
	remaining: number;
	lastTitle?: string;
	lastError?: string;
}

export interface BackfillResponse {
	status: BackfillStatus;
	missingMovies: number;
	missingSeries: number;
	missingTotal: number;
}

export interface BackfillStartResponse {
	status: BackfillStatus;
}

export function getBackfillStatus(client: ApiClient = apiClient): Promise<BackfillResponse> {
	return client.request<BackfillResponse>('/api/metadata/backfill');
}

export function startBackfill(
	provider: string = 'tmdb',
	client: ApiClient = apiClient
): Promise<BackfillStartResponse> {
	return client.send<BackfillStartResponse, { provider: string }>(
		'/api/metadata/backfill',
		{ provider },
		'POST'
	);
}

export function stopBackfill(
	client: ApiClient = apiClient
): Promise<BackfillStartResponse> {
	return client.send<BackfillStartResponse, Record<string, never>>(
		'/api/metadata/backfill',
		{},
		'DELETE'
	);
}

export function getReviewItems(
	client: ApiClient = apiClient,
	limit = 100
): Promise<ReviewItemsResponse> {
	return client.request<ReviewItemsResponse>(`/api/review?limit=${encodeURIComponent(String(limit))}`);
}

export function getVersionGroups(
	client: ApiClient = apiClient,
	limit = 100
): Promise<VersionGroupsResponse> {
	return client.request<VersionGroupsResponse>(
		`/api/versions?limit=${encodeURIComponent(String(limit))}`
	);
}

export function getMetadataRecords(
	kind: string,
	id: string,
	client: ApiClient = apiClient
): Promise<MetadataRecordsResponse> {
	return client.request<MetadataRecordsResponse>(
		`/api/metadata/${encodeURIComponent(kind)}/${encodeURIComponent(id)}`
	);
}

export function refreshMetadataItem(
	payload: MetadataRefreshRequest,
	client: ApiClient = apiClient
): Promise<MetadataRefreshResponse> {
	return client.send<MetadataRefreshResponse, MetadataRefreshRequest>(
		'/api/metadata/refresh',
		payload,
		'POST'
	);
}

export function applyMetadataMatch(
	payload: MetadataMatchRequest,
	client: ApiClient = apiClient
): Promise<{ match?: MetadataMatchRequest; records?: MetadataRecord[] }> {
	return client.send<{ match?: MetadataMatchRequest; records?: MetadataRecord[] }, MetadataMatchRequest>(
		'/api/metadata/match',
		payload,
		'PUT'
	);
}
