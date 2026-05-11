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
	hero?: ClientHomeItem;
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
	versions?: MovieDetailVersion[];
}

export interface SeriesDetailSeason {
	episodes?: SeriesDetailEpisode[];
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

export function getClientHome(client: ApiClient = apiClient, limit = 24): Promise<ClientHomeResponse> {
	return client.request<ClientHomeResponse>(`/api/client/home?limit=${encodeURIComponent(String(limit))}`);
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
