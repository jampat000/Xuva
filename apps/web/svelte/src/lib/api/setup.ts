import { apiClient, type ApiClient } from './client';
import type { LibraryRecord } from './home';

export interface FolderBrowseEntry {
	name?: string;
	path?: string;
}

export interface FolderBrowseResponse {
	path?: string;
	parent?: string;
	entries?: FolderBrowseEntry[];
	writable?: boolean;
	message?: string;
	error?: string;
}

export interface SaveLibraryRequest {
	name: string;
	path: string;
	kind: 'movies' | 'tv';
}

export interface ScanJobResponse {
	id?: string;
	status?: string;
	kind?: string;
	libraryId?: string;
	[key: string]: unknown;
}

export function browseFolder(path = '', client: ApiClient = apiClient): Promise<FolderBrowseResponse> {
	const query = path ? `?path=${encodeURIComponent(path)}` : '';
	return client.request<FolderBrowseResponse>(`/api/settings/folders/browse${query}`);
}

export function saveLibrary(
	payload: SaveLibraryRequest,
	client: ApiClient = apiClient
): Promise<LibraryRecord> {
	return client.send<LibraryRecord, SaveLibraryRequest>('/api/libraries', payload, 'POST');
}

export function startLibraryScan(
	libraryID: string,
	client: ApiClient = apiClient
): Promise<ScanJobResponse> {
	return client.send<ScanJobResponse, Record<string, never>>(
		`/api/libraries/${encodeURIComponent(libraryID)}/scan`,
		{},
		'POST'
	);
}
