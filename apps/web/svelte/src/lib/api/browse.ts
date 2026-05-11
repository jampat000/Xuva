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
