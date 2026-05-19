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
