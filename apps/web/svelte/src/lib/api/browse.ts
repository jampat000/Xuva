import { apiClient, type ApiClient } from './client';
import { invalidateSwrPrefix, subscribeSwr, swrFetch } from './cache/swr-cache';

export interface MetadataRecord {
	kind?: string;
	itemId?: string;
	provider?: string;
	title?: string;
	year?: number;
	overview?: string;
	tagline?: string;
	posterUrl?: string;
	backdropUrl?: string;
	thumbnailUrl?: string;
	logoUrl?: string;
	bannerUrl?: string;
	videoKey?: string;        // YouTube trailer key
	trailerPath?: string;     // Local MP4 path once downloaded
	contentRating?: string;   // G / PG / PG-13 / R / TV-MA …
	releaseDate?: string;
	firstAirDate?: string;
	runtimeMinutes?: number;
	genres?: string[];
	directors?: string[];
	writers?: string[];
	studios?: string[];
	productionCompanies?: string[];
	networks?: string[];
	statusText?: string;      // "Running" | "Ended" for series
	cast?: MetadataCredit[];
	voteAverage?: number;
	runtime?: string;
	[key: string]: unknown;
}

export interface MetadataCredit {
	name?: string;
	character?: string;
	role?: string;
	profileUrl?: string;
	sortOrder?: number;
}

export interface MovieListItem {
	id?: string;
	title?: string;
	year?: number;
	sortTitle?: string;
	needsReview?: boolean;
	probed?: boolean;
	versionCount?: number;
	addedAt?: string;
	watched?: boolean;
	metadata?: MetadataRecord;
}

export interface SeriesListItem {
	id?: string;
	title?: string;
	sortTitle?: string;
	seasonCount?: number;
	episodeCount?: number;
	addedAt?: string;
	watched?: boolean;
	needsReview?: boolean;
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
	tmdbOverrideId?: number;
}

export interface TMDBCandidate {
	id: number;
	title: string;
	year?: number;
	overview?: string;
	posterUrl?: string;
	backdropUrl?: string;
	voteAverage?: number;
	voteCount?: number;
}

export interface TMDBCandidatesResponse {
	kind: string;
	title: string;
	year: number;
	candidates: TMDBCandidate[];
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

// ── Library search ────────────────────────────────────────────────────────
//
// Aggregated search across movies, series, people (cast/crew), and
// collections (TMDB franchises). Server returns up to `limit` per type
// (default 8, max 40). Each item carries a `kind` discriminator.

export interface SearchMovieHit {
	kind: 'movie';
	id: string;
	title: string;
	year?: number;
	posterUrl?: string;
	backdropUrl?: string;
	logoUrl?: string;
	overview?: string;
	genres?: string[];
	voteAverage?: number;
}

export interface SearchSeriesHit {
	kind: 'series';
	id: string;
	title: string;
	year?: number;
	seasonCount?: number;
	episodeCount?: number;
	posterUrl?: string;
	backdropUrl?: string;
	logoUrl?: string;
	overview?: string;
	genres?: string[];
	voteAverage?: number;
}

export interface SearchPersonHit {
	kind: 'person';
	name: string;
	profileUrl?: string;
	department?: string;
	creditCount: number;
}

export interface SearchCollectionHit {
	kind: 'collection';
	id: string;
	name: string;
	posterUrl?: string;
	backdropUrl?: string;
	logoUrl?: string;
	movieCount: number;
}

export type SearchHit =
	| SearchMovieHit
	| SearchSeriesHit
	| SearchPersonHit
	| SearchCollectionHit;

export interface SearchResponse {
	query: string;
	movies: SearchMovieHit[];
	series: SearchSeriesHit[];
	people: SearchPersonHit[];
	collections: SearchCollectionHit[];
}

export function searchLibrary(
	q: string,
	limit = 8,
	client: ApiClient = apiClient,
): Promise<SearchResponse> {
	const params = new URLSearchParams({ q, limit: String(limit) });
	return client.request<SearchResponse>(`/api/client/search?${params}`);
}

// ── Response cache ────────────────────────────────────────────────────────────
// Stale-while-revalidate cache for the two heavy list endpoints, backed by
// IndexedDB so cold starts after a tab close / hard refresh are also instant.
//
// Pages get the cached value synchronously on second-and-later access, even
// when it is stale. A background refresh fires whenever a stale value is
// served; subscribers (via subscribeMovies / subscribeSeries) receive the
// fresh data and update the grid in place. Mutation endpoints invalidate via
// invalidateListCache so the next browse fetches new data.
//
// freshMs was previously 5 minutes — too short for typical browsing
// sessions. On a 4000-item library, every visit beyond the 5 min mark
// triggered a ~2.4 s background fetch even though the data was unchanged.
// Bumped to 24 h here; library mutations still flow in via the SSE
// invalidation handler in lib/api/events.ts (library.updated, scan.completed,
// metadata.updated → invalidateListCache), so the only cost of the longer
// TTL is missing TMDB-side metadata changes for up to a day.

const MOVIES_FRESH_MS  = 24 * 60 * 60_000;       // 24 h — invalidated by SSE on mutations
const MOVIES_MAX_AGE   = 7 * 24 * 60 * 60_000;   // 7 d before discarding cache entirely
const SERIES_FRESH_MS  = MOVIES_FRESH_MS;
const SERIES_MAX_AGE   = MOVIES_MAX_AGE;

const MOVIES_KEY_PREFIX = '/api/movies';
const SERIES_KEY_PREFIX = '/api/series';

/** Invalidate the list cache (call after a scan or library change). */
export function invalidateListCache(): Promise<void> {
	return Promise.all([
		invalidateSwrPrefix(MOVIES_KEY_PREFIX),
		invalidateSwrPrefix(SERIES_KEY_PREFIX),
	]).then(() => undefined);
}

export function getMovies(client: ApiClient = apiClient, limit = 0): Promise<MoviesResponse> {
	const path = `/api/movies?limit=${encodeURIComponent(String(limit))}`;
	return swrFetch<MoviesResponse>(
		path,
		() => client.request<MoviesResponse>(path),
		{ freshMs: MOVIES_FRESH_MS, maxAgeMs: MOVIES_MAX_AGE },
	);
}

export function getSeries(client: ApiClient = apiClient, limit = 0): Promise<SeriesResponse> {
	const path = `/api/series?limit=${encodeURIComponent(String(limit))}`;
	return swrFetch<SeriesResponse>(
		path,
		() => client.request<SeriesResponse>(path),
		{ freshMs: SERIES_FRESH_MS, maxAgeMs: SERIES_MAX_AGE },
	);
}

/**
 * Subscribe to fresh /api/movies data for a given limit. Fires when a
 * background SWR refresh succeeds. Return value is the unsubscribe handle —
 * call it from onMount cleanup.
 */
export function subscribeMovies(
	limit: number,
	callback: (data: MoviesResponse) => void,
): () => void {
	return subscribeSwr<MoviesResponse>(`/api/movies?limit=${encodeURIComponent(String(limit))}`, callback);
}

export function subscribeSeries(
	limit: number,
	callback: (data: SeriesResponse) => void,
): () => void {
	return subscribeSwr<SeriesResponse>(`/api/series?limit=${encodeURIComponent(String(limit))}`, callback);
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

export function getMetadataCandidates(
	kind: 'movie' | 'series',
	title: string,
	year?: number,
	limit = 8,
	client: ApiClient = apiClient
): Promise<TMDBCandidatesResponse> {
	const params = new URLSearchParams({ kind, title, limit: String(limit) });
	if (year) params.set('year', String(year));
	return client.request<TMDBCandidatesResponse>(`/api/metadata/candidates?${params}`);
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
