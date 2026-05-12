import { apiClient, type ApiClient } from './client';

export interface CatalogSummaryResponse {
	libraries?: number;
	mediaSources?: number;
	movies?: number;
	series?: number;
	episodes?: number;
	scanRuns?: number;
	unprobed?: number;
}

export interface CatalogHealthResponse {
	summary?: CatalogSummaryResponse;
	needsReview?: number;
	unprobed?: number;
	unsupported?: number;
	highBitrate?: number;
	withSubtitles?: number;
}

export interface SystemStatusResponse {
	collectedAt?: string;
	cpu?: {
		percent?: number;
		cores?: number;
	};
	memory?: {
		totalBytes?: number;
		availableBytes?: number;
		usedBytes?: number;
		usedPercent?: number;
	};
	process?: {
		goAllocBytes?: number;
		goSysBytes?: number;
		goroutines?: number;
	};
	network?: {
		receiveBps?: number;
		transmitBps?: number;
	};
	disks?: Array<{
		name?: string;
		path?: string;
		usedPercent?: number;
		freeBytes?: number;
		totalBytes?: number;
		writable?: boolean;
	}>;
}

export interface WorkQueueItem {
	id?: string;
	status?: string;
	mediaSourceId?: string;
	mode?: string;
	createdAt?: string;
	updatedAt?: string;
	progress?: number;
}

export interface ScanJobItem {
	id?: string;
	status?: string;
	libraryId?: string;
	kind?: string;
	createdAt?: string;
	updatedAt?: string;
	progress?: number;
}

export interface ProbeJobItem {
	id?: string;
	status?: string;
	createdAt?: string;
	updatedAt?: string;
	progress?: number;
}

export interface DownloadJobItem {
	id?: string;
	status?: string;
	mediaSourceId?: string;
	targetProfile?: string;
	createdAt?: string;
	updatedAt?: string;
	progress?: number;
}

export interface SessionItem {
	id?: string;
	title?: string;
	sourceName?: string;
	deviceId?: string;
	mode?: string;
	route?: string;
	state?: string;
	updatedAt?: string;
}

export interface ScansResponse {
	scans?: ScanJobItem[];
}

export interface ProbesResponse {
	probes?: ProbeJobItem[];
}

export interface WorkResponse {
	work?: WorkQueueItem[];
}

export interface DownloadsResponse {
	downloads?: DownloadJobItem[];
}

export interface SessionsResponse {
	sessions?: SessionItem[];
}

export interface PerformanceSettingsResponse {
	profile?: string;
	playbackPolicy?: {
		id?: string;
		label?: string;
		description?: string;
	};
	limits?: {
		scanWorkers?: number;
		probeWorkers?: number;
		transcodeWorkers?: number;
		gpuWorkers?: number;
	};
	queues?: Array<{
		name?: string;
		class?: string;
		workers?: number;
		active?: number;
		queued?: number;
		workerUtilization?: number;
	}>;
	hardwareAcceleration?: {
		status?: string;
		unlockState?: string;
		gpuWorkers?: number;
		available?: boolean;
	};
}

export interface SettingsResponse {
	restartRequired?: boolean;
	config?: {
		serverName?: string;
		librarySyncMode?: string;
		syncIntervalMins?: number;
		watchDebounceSecs?: number;
		probeBatchLimit?: number;
		playbackPolicy?: string;
		scanWorkers?: number;
		probeWorkers?: number;
		transcodeWorkers?: number;
		gpuWorkers?: number;
		hardwareUnlocked?: boolean;
		metadataProviders?: {
			automatic?: Array<{ id?: string; name?: string; note?: string }>;
			managedOverrides?: Array<{ id?: string; name?: string; configured?: boolean }>;
		};
	};
	libraries?: Array<{
		id?: string;
		name?: string;
		kind?: string;
		path?: string;
		storageType?: string;
	}>;
}

export interface UpdateSettingsRequest {
	serverName?: string;
	librarySyncMode?: string;
	syncIntervalMins?: number;
	watchDebounceSecs?: number;
	probeBatchLimit?: number;
	playbackPolicy?: string;
}

export function getCatalogSummary(client: ApiClient = apiClient): Promise<CatalogSummaryResponse> {
	return client.request<CatalogSummaryResponse>('/api/catalog/summary');
}

export function getCatalogHealth(client: ApiClient = apiClient): Promise<CatalogHealthResponse> {
	return client.request<CatalogHealthResponse>('/api/catalog/health');
}

export function getSystemStatus(client: ApiClient = apiClient): Promise<SystemStatusResponse> {
	return client.request<SystemStatusResponse>('/api/system/status');
}

export function getSettings(client: ApiClient = apiClient): Promise<SettingsResponse> {
	return client.request<SettingsResponse>('/api/settings');
}

export function updateSettings(
	payload: UpdateSettingsRequest,
	client: ApiClient = apiClient
): Promise<SettingsResponse> {
	return client.send<SettingsResponse, UpdateSettingsRequest>('/api/settings', payload, 'PUT');
}

export function getPerformanceSettings(
	client: ApiClient = apiClient
): Promise<PerformanceSettingsResponse> {
	return client.request<PerformanceSettingsResponse>('/api/settings/performance');
}

export function getScans(client: ApiClient = apiClient): Promise<ScansResponse> {
	return client.request<ScansResponse>('/api/scans');
}

export function getProbes(client: ApiClient = apiClient): Promise<ProbesResponse> {
	return client.request<ProbesResponse>('/api/probes');
}

export function getWork(client: ApiClient = apiClient): Promise<WorkResponse> {
	return client.request<WorkResponse>('/api/work');
}

export function getDownloads(client: ApiClient = apiClient): Promise<DownloadsResponse> {
	return client.request<DownloadsResponse>('/api/downloads');
}

export function getSessions(client: ApiClient = apiClient): Promise<SessionsResponse> {
	return client.request<SessionsResponse>('/api/sessions');
}
