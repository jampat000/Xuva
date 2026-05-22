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
	mediaSourceId?: string;
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

export interface MigrationFormat {
	id?: string;
	label?: string;
	description?: string;
	sources?: string[];
	schema?: string;
	validationRules?: string[];
}

export interface MigrationFormatsResponse {
	formats?: MigrationFormat[];
}

export interface MigrationSummary {
	total?: number;
	importable?: number;
	imported?: number;
	skipped?: number;
	conflicted?: number;
	verified?: number;
	failed?: number;
	rolledBack?: number;
}

export interface MigrationVerification {
	checked?: number;
	passed?: number;
	failed?: number;
}

export interface MigrationItemReport {
	importKey?: string;
	kind?: string;
	title?: string;
	outcome?: string;
	reasonCode?: string;
	reasonText?: string;
	changes?: string[];
	target?: {
		kind?: string;
		itemId?: string;
		mediaSourceId?: string;
		title?: string;
	};
	externalIds?: Record<string, string>;
}

export interface MigrationRunReport {
	runId?: string;
	schema?: string;
	source?: string;
	scopes?: string[];
	status?: string;
	createdAt?: string;
	completedAt?: string;
	rolledBackAt?: string;
	summary?: MigrationSummary;
	verification?: MigrationVerification;
	items?: MigrationItemReport[];
	warnings?: string[];
	error?: string;
}

export interface MigrationRunsResponse {
	runs?: MigrationRunReport[];
}

export interface MigrationRequest {
	payload: string;
	scopes?: string[];
	userId?: string;
	selectedImportKeys?: string[];
}

export interface DiscoveryStatusResponse {
	enabled?: boolean;
	running?: boolean;
	serviceName?: string;
	serviceType?: string;
	hostName?: string;
	webUrl?: string;
	port?: number;
	txtRecords?: string[];
	lastError?: string;
	note?: string;
}

export interface HardwareTestResponse {
	status?: string;
	error?: string;
	working?: number;
	tested?: number;
	testedAt?: string;
	tests?: Array<{
		id?: string;
		label?: string;
		vendor?: string;
		codec?: string;
		ok?: boolean;
		error?: string;
		durationMs?: number;
	}>;
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
		encoders?: Array<{ id?: string; label?: string; vendor?: string; codec?: string }>;
		lastTest?: {
			status?: string;
			working?: number;
			tested?: number;
			testedAt?: string;
			tests?: Array<{
				id?: string;
				label?: string;
				vendor?: string;
				codec?: string;
				ok?: boolean;
				error?: string;
				durationMs?: number;
			}>;
		};
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
			supportsMetadata?: boolean;
			supportsArtwork?: boolean;
			providerHealth?: {
				id?: string;
				managed?: boolean;
				configured?: boolean;
				healthy?: boolean;
				status?: string;
				error?: string;
			};
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
			supportsMetadata?: boolean;
			supportsArtwork?: boolean;
			providerHealth?: {
				id?: string;
				managed?: boolean;
				configured?: boolean;
				healthy?: boolean;
				status?: string;
				error?: string;
			};
		}>;
	};
	metadataSourcePreferences?: {
		movie?: string[];
		series?: string[];
		movieArtwork?: string[];
		seriesArtwork?: string[];
	};
	config?: {
		serverName?: string;
		canonicalWebOrigin?: string;
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
		country?: string;
		timezone?: string;
		metadataLanguage?: string;
		preferTextSubtitles?: boolean;
		originalQualityOnly?: boolean;
		defaultSubtitlesMovies?: boolean;
		defaultSubtitlesTV?: boolean;
		disableTrailers?: boolean;
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
	canonicalWebOrigin?: string;
	dataDir?: string;
	transcodeDir?: string;
	downloadsDir?: string;
	metadataDir?: string;
	cacheDir?: string;
	tempDir?: string;
	hardwareUnlocked?: boolean;
	librarySyncMode?: string;
	syncIntervalMins?: number;
	watchDebounceSecs?: number;
	country?: string;
	timezone?: string;
	metadataLanguage?: string;
	preferTextSubtitles?: boolean;
	originalQualityOnly?: boolean;
	defaultSubtitlesMovies?: boolean;
	defaultSubtitlesTV?: boolean;
	probeBatchLimit?: number;
	playbackPolicy?: string;
	disableTrailers?: boolean;
}

export interface UpdateMetadataSourcePreferencesRequest {
	movie: string[];
	series: string[];
	movieArtwork: string[];
	seriesArtwork: string[];
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

export function getMigrationFormats(
	client: ApiClient = apiClient
): Promise<MigrationFormatsResponse> {
	return client.request<MigrationFormatsResponse>('/api/migrations/formats');
}

export function getMigrationRuns(client: ApiClient = apiClient): Promise<MigrationRunsResponse> {
	return client.request<MigrationRunsResponse>('/api/migrations/runs');
}

export function getMigrationRun(
	id: string,
	client: ApiClient = apiClient
): Promise<MigrationRunReport> {
	return client.request<MigrationRunReport>(`/api/migrations/runs/${encodeURIComponent(id)}`);
}

export function runMigrationDryRun(
	payload: MigrationRequest,
	client: ApiClient = apiClient
): Promise<MigrationRunReport> {
	return client.send<MigrationRunReport, MigrationRequest>(
		'/api/migrations/dry-run',
		payload,
		'POST'
	);
}

export function runMigrationImport(
	payload: MigrationRequest,
	client: ApiClient = apiClient
): Promise<MigrationRunReport> {
	return client.send<MigrationRunReport, MigrationRequest>(
		'/api/migrations/import',
		payload,
		'POST'
	);
}

export function rollbackMigrationRun(
	id: string,
	client: ApiClient = apiClient
): Promise<MigrationRunReport> {
	return client.send<MigrationRunReport, Record<string, never>>(
		`/api/migrations/runs/${encodeURIComponent(id)}/rollback`,
		{},
		'POST'
	);
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
): Promise<void> {
	return client.send<void, Record<string, never>>(
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

export function runHardwareTest(
	client: ApiClient = apiClient
): Promise<HardwareTestResponse> {
	return client.send<HardwareTestResponse, Record<string, never>>(
		'/api/settings/hardware/test',
		{},
		'POST'
	);
}

// ─── Library management ────────────────────────────────────────────────────

export interface LibraryItem {
	id?: string;
	name?: string;
	kind?: string;
	path?: string;
	storageType?: string;
}

export interface LibrariesResponse {
	libraries?: LibraryItem[];
}

export interface LibrarySaveRequest {
	name: string;
	kind: string;
	path: string;
	storageType: string;
}

export interface FolderEntry {
	name?: string;
	path?: string;
	isDir?: boolean;
}

export interface FolderBrowseResponse {
	currentPath?: string;
	parentPath?: string;
	entries?: FolderEntry[];
}

export function getLibraries(client: ApiClient = apiClient): Promise<LibrariesResponse> {
	return client.request<LibrariesResponse>('/api/libraries');
}

export function saveLibrary(
	payload: LibrarySaveRequest,
	client: ApiClient = apiClient
): Promise<LibraryItem> {
	return client.send<LibraryItem, LibrarySaveRequest>('/api/libraries', payload, 'POST');
}

export function deleteLibrary(
	id: string,
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.request<Record<string, unknown>>(
		`/api/libraries/${encodeURIComponent(id)}`,
		{ method: 'DELETE' }
	);
}

export function scanLibrary(
	id: string,
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, Record<string, never>>(
		`/api/libraries/${encodeURIComponent(id)}/scan`,
		{},
		'POST'
	);
}

export function browseFolders(
	path?: string,
	client: ApiClient = apiClient
): Promise<FolderBrowseResponse> {
	const url = path
		? `/api/settings/folders/browse?path=${encodeURIComponent(path)}`
		: '/api/settings/folders/browse';
	return client.request<FolderBrowseResponse>(url);
}

// ─── Users ────────────────────────────────────────────────────────────────

export interface UserItem {
	id?: string;
	username?: string;
	displayName?: string;
	avatarUrl?: string;
	avatarPreset?: string;
	avatarColor?: string;
	role?: string;
	isRestricted?: boolean;
	maxRating?: string;
	hasPin?: boolean;
	createdAt?: string;
	updatedAt?: string;
}

export interface UsersResponse {
	users?: UserItem[];
}

export interface CreateUserRequest {
	username: string;
	displayName?: string;
	password: string;
	role?: string;
}

export interface UpdateUserRequest {
	displayName?: string;
	role?: string;
}

export interface UpdateUserPasswordRequest {
	password: string;
	currentPassword?: string;
}

export function getUsers(client: ApiClient = apiClient): Promise<UsersResponse> {
	return client.request<UsersResponse>('/api/users');
}

export function createUser(
	payload: CreateUserRequest,
	client: ApiClient = apiClient
): Promise<UserItem> {
	return client.send<UserItem, CreateUserRequest>('/api/users', payload, 'POST');
}

export function updateUser(
	id: string,
	payload: UpdateUserRequest,
	client: ApiClient = apiClient
): Promise<UserItem> {
	return client.send<UserItem, UpdateUserRequest>(
		`/api/users/${encodeURIComponent(id)}`,
		payload,
		'PATCH'
	);
}

export function deleteUser(
	id: string,
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.request<Record<string, unknown>>(
		`/api/users/${encodeURIComponent(id)}`,
		{ method: 'DELETE' }
	);
}

export function updateUserPassword(
	id: string,
	payload: UpdateUserPasswordRequest,
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, UpdateUserPasswordRequest>(
		`/api/users/${encodeURIComponent(id)}/password`,
		payload,
		'POST'
	);
}

// ─── Device profiles ──────────────────────────────────────────────────────

export interface DeviceProfile {
	id?: string;
	name?: string;
	description?: string;
	maxBitrate?: number;
	supportsHevc?: boolean;
	supportsAv1?: boolean;
	supportsDolbyVision?: boolean;
	supportsHdr?: boolean;
	preferDirectPlay?: boolean;
}

export interface DeviceProfilesResponse {
	profiles?: DeviceProfile[];
}

export function getDeviceProfiles(
	client: ApiClient = apiClient
): Promise<DeviceProfilesResponse> {
	return client.request<DeviceProfilesResponse>('/api/devices/profiles');
}

// ─── Scan all libraries ───────────────────────────────────────────────────

export function scanAllLibraries(
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, Record<string, never>>(
		'/api/libraries/scan',
		{},
		'POST'
	);
}

// ─── Backup / export ──────────────────────────────────────────────────────────

export interface BackupManifest {
	version?: number;
	createdAt?: string;
	dataDir?: string;
	mediaPaths?: { movies?: string; tv?: string };
}

export interface BackupImportResponse {
	status?: string;
	requiresRestart?: boolean;
	manifest?: BackupManifest;
}

/** Triggers a streamed archive download of the current database and settings. */
export async function exportBackup(): Promise<void> {
	const resp = await fetch('/api/backup/export', { credentials: 'include' });
	if (!resp.ok) throw new Error(`Export failed: ${resp.status}`);
	const blob = await resp.blob();
	const url = URL.createObjectURL(blob);
	const a = document.createElement('a');
	const cd = resp.headers.get('Content-Disposition') ?? '';
	const m = cd.match(/filename="([^"]+)"/);
	a.href = url;
	a.download = m ? m[1] : `xuva-backup.tar.gz`;
	document.body.appendChild(a);
	a.click();
	document.body.removeChild(a);
	URL.revokeObjectURL(url);
}

/** Uploads an archive file and stages a restore (applied on next server restart). */
export async function importBackup(file: File): Promise<BackupImportResponse> {
	const csrfToken = (() => {
		if (typeof document === 'undefined') return '';
		const m = document.cookie.match(/(?:^|; )xuva_csrf=([^;]*)/);
		return m ? decodeURIComponent(m[1]) : '';
	})();
	const form = new FormData();
	form.append('archive', file);
	const resp = await fetch('/api/backup/import', {
		method: 'POST',
		credentials: 'include',
		headers: { 'X-CSRF-Token': csrfToken },
		body: form,
	});
	if (!resp.ok) {
		const text = await resp.text();
		throw new Error(text || `Import failed: ${resp.status}`);
	}
	return resp.json();
}

// ─── Notifications ────────────────────────────────────────────────────────────

export interface NotificationItem {
	id?: string;
	kind?: string;
	title?: string;
	message?: string;
	link?: string;
	dismissed?: boolean;
	createdAt?: string;
}

export interface NotificationsResponse {
	notifications?: NotificationItem[];
}

export function getNotifications(
	client: ApiClient = apiClient
): Promise<NotificationsResponse> {
	return client.request<NotificationsResponse>('/api/notifications');
}

export function dismissNotification(
	id: string,
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, Record<string, never>>(
		`/api/notifications/${encodeURIComponent(id)}/dismiss`,
		{} as Record<string, never>,
		'POST'
	);
}

export function dismissAllNotifications(
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, Record<string, never>>(
		'/api/notifications/dismiss-all',
		{} as Record<string, never>,
		'POST'
	);
}

export interface ChapterSegment {
	start: number;
	end: number;
}

export interface ChaptersResponse {
	mediaSourceId: string;
	intro?: ChapterSegment;
	credits?: ChapterSegment;
	analyzedAt?: string;
}

export interface UserPreferences {
	autoSkipIntros?: boolean;
	posterSize?: 'S' | 'M' | 'L';
}

export function getChapters(
	mediaSourceId: string,
	client: ApiClient = apiClient
): Promise<ChaptersResponse> {
	return client.request<ChaptersResponse>(
		`/api/media-sources/${encodeURIComponent(mediaSourceId)}/chapters`
	);
}

export function analyzeChapters(
	mediaSourceId: string,
	client: ApiClient = apiClient
): Promise<Record<string, unknown>> {
	return client.send<Record<string, unknown>, Record<string, never>>(
		`/api/media-sources/${encodeURIComponent(mediaSourceId)}/chapters/analyze`,
		{} as Record<string, never>,
		'POST'
	);
}

export function updateUserPreferences(
	prefs: UserPreferences,
	client: ApiClient = apiClient
): Promise<UserPreferences> {
	return client.send<UserPreferences, UserPreferences>(
		'/api/users/me/preferences',
		prefs,
		'PATCH'
	);
}

// ── QR pair token ─────────────────────────────────────────────────────────────

export interface QRTokenResponse {
	token: string;
	claimUrl: string;
	imageUrl: string;
	expiresAt: string;
}

export function generateQRPairToken(client: ApiClient = apiClient): Promise<QRTokenResponse> {
	return client.send<QRTokenResponse, Record<string, never>>(
		'/api/pairing/qr',
		{} as Record<string, never>,
		'POST'
	);
}
