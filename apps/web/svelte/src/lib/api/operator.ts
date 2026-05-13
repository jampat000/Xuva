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
		usedBytes?: number;
		usedPercent?: number;
		freeBytes?: number;
		totalBytes?: number;
		writable?: boolean;
		error?: string;
		sharedWithData?: boolean;
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

export interface PairingRequestItem {
	id?: string;
	code?: string;
	deviceName?: string;
	clientProfile?: string;
	deviceId?: string;
	status?: string;
	approvedBy?: string;
	expiresAt?: string;
	createdAt?: string;
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

export interface PairingRequestsResponse {
	requests?: PairingRequestItem[];
}

export interface ApprovedDeviceItem {
	id?: string;
	deviceName?: string;
	clientProfile?: string;
	displayName?: string;
	status?: string;
	approvedAt?: string;
	approvedBy?: string;
	createdAt?: string;
	updatedAt?: string;
}

export interface ApprovedDevicesResponse {
	devices?: ApprovedDeviceItem[];
}

export interface DiscoveryStatusResponse {
	enabled?: boolean;
	running?: boolean;
	serviceName?: string;
	serviceType?: string;
	port?: number;
	txtRecords?: string[];
	lastError?: string;
	note?: string;
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
	metadataSources?: {
		movie?: Array<{
			id?: string;
			name?: string;
			description?: string;
			coverage?: string;
			note?: string;
			local?: boolean;
			managed?: boolean;
			requiresConfig?: boolean;
			available?: boolean;
			runtimeReady?: boolean;
			status?: string;
		}>;
		series?: Array<{
			id?: string;
			name?: string;
			description?: string;
			coverage?: string;
			note?: string;
			local?: boolean;
			managed?: boolean;
			requiresConfig?: boolean;
			available?: boolean;
			runtimeReady?: boolean;
			status?: string;
		}>;
	};
	metadataSourcePreferences?: {
		movie?: string[];
		series?: string[];
	};
	config?: {
		serverName?: string;
		dataDir?: string;
		transcodeDir?: string;
		downloadsDir?: string;
		metadataDir?: string;
		cacheDir?: string;
		tempDir?: string;
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
	dataDir?: string;
	transcodeDir?: string;
	downloadsDir?: string;
	metadataDir?: string;
	cacheDir?: string;
	tempDir?: string;
	librarySyncMode?: string;
	syncIntervalMins?: number;
	watchDebounceSecs?: number;
	probeBatchLimit?: number;
	playbackPolicy?: string;
}

export interface UpdateMetadataSourcePreferencesRequest {
	movie: string[];
	series: string[];
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

export function updateMetadataSourcePreferences(
	payload: UpdateMetadataSourcePreferencesRequest,
	client: ApiClient = apiClient
): Promise<SettingsResponse> {
	return client.send<SettingsResponse, UpdateMetadataSourcePreferencesRequest>(
		'/api/settings/metadata-sources',
		payload,
		'PUT'
	);
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

export function getPairingRequests(client: ApiClient = apiClient): Promise<PairingRequestsResponse> {
	return client.request<PairingRequestsResponse>('/api/pairing/requests');
}

export function getApprovedDevices(client: ApiClient = apiClient): Promise<ApprovedDevicesResponse> {
	return client.request<ApprovedDevicesResponse>('/api/devices');
}

export function getDiscoveryStatus(
	client: ApiClient = apiClient
): Promise<DiscoveryStatusResponse> {
	return client.request<DiscoveryStatusResponse>('/api/discovery/status');
}

export function approvePairingRequest(
	id: string,
	client: ApiClient = apiClient
): Promise<PairingRequestItem> {
	return client.send<PairingRequestItem, Record<string, never>>(
		`/api/pairing/requests/${encodeURIComponent(id)}/approve`,
		{},
		'POST'
	);
}

export function denyPairingRequest(
	id: string,
	client: ApiClient = apiClient
): Promise<PairingRequestItem> {
	return client.send<PairingRequestItem, Record<string, never>>(
		`/api/pairing/requests/${encodeURIComponent(id)}/deny`,
		{},
		'POST'
	);
}

export function revokeApprovedDevice(
	id: string,
	client: ApiClient = apiClient
): Promise<ApprovedDeviceItem> {
	return client.send<ApprovedDeviceItem, Record<string, never>>(
		`/api/devices/${encodeURIComponent(id)}/revoke`,
		{},
		'POST'
	);
}
