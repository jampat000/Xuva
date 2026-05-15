	<script lang="ts">
	import { onMount } from 'svelte';
	import {
		applyMetadataMatch,
		getMetadataRecords,
		getReviewItems,
		getVersionGroups,
		refreshMetadataBatch,
		refreshMetadataItem,
		scanMovies,
		scanTV,
		type MetadataMatchRequest,
		type MetadataRecord,
		type MetadataRecordsResponse,
		type ReviewItem,
		type VersionGroup
	} from '$lib/api/browse';
	import {
		createUser,
		deleteUser,
		getAuthSessionIfAvailable,
		getClientBootstrap,
		getUsers,
		logout,
		updateUserPassword,
		type AuthSessionResponse,
		type AuthSessionUser,
		type ClientBootstrapResponse,
		type UserAccount
	} from '$lib/api/auth';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import {
		browseFolder,
		deleteLibrary,
		saveLibrary,
		startLibraryScan,
		type FolderBrowseResponse
	} from '$lib/api/setup';
	import {
		getApprovedDevices,
		approvePairingRequest,
		denyPairingRequest,
		getCatalogHealth,
		getCatalogSummary,
		getDiscoveryStatus,
		getDownloads,
		getPairingRequests,
		getPerformanceSettings,
		getProbes,
		runHardwareTest,
		getScans,
		getSessions,
		getSettings,
		getSystemStatus,
		getWork,
		revokeApprovedDevice,
		updateSettings,
		updateMetadataSourcePreferences,
		type ApprovedDeviceItem,
		type CatalogHealthResponse,
		type CatalogSummaryResponse,
		type DiscoveryStatusResponse,
		type DownloadJobItem,
		type PerformanceSettingsResponse,
		type PairingRequestItem,
		type ProbeJobItem,
		type ScanJobItem,
		type SessionItem,
		type SettingsResponse,
		type HardwareTestResponse,
		type SystemStatusResponse,
		type UpdateSettingsRequest,
		type WorkQueueItem
	} from '$lib/api/operator';
	import { createEventStream } from '$lib/events/stream';
	import { xuvaTitle, normalizeServerName } from '$lib/server-name';
	import {
		ActivityListShell,
		ServerShell,
		XuvaActionList,
		XuvaButton,
		XuvaPanel,
		XuvaStat
	} from '$lib/components';
	import SettingsPanel from '$lib/components/operator/SettingsPanel.svelte';
	import FolderBrowserPanel from '$lib/components/operator/FolderBrowserPanel.svelte';

	type SettingsSection =
		| 'dashboard'
		| 'general'
		| 'libraries'
		| 'scanning'
		| 'metadata'
		| 'playback'
		| 'transcoding'
		| 'storage'
		| 'network'
		| 'pairing'
		| 'approved-devices'
		| 'discovery'
		| 'admin-access'
		| 'planned-tools'
		| 'about';

	type LibraryActionKind = 'scan' | 'remove' | '';
	type PairingActionKind = 'approve' | 'deny' | '';

	type StorageFieldKey =
		| 'dataDir'
		| 'transcodeDir'
		| 'downloadsDir'
		| 'metadataDir'
		| 'cacheDir'
		| 'tempDir';

	type EditableStorageFieldKey = Exclude<StorageFieldKey, 'dataDir'>;
	type SystemDisk = NonNullable<SystemStatusResponse['disks']>[number];

	interface LibrarySyncModeOption {
		id: 'manual' | 'daily' | 'watch';
		label: string;
		description: string;
	}

	interface PlaybackPolicyOption {
		id: 'original_only' | 'light' | 'full' | 'cinema';
		label: string;
		description: string;
	}

	interface BuildInfo {
		buildID?: string;
		publishedAt?: string;
		gitCommit?: string | null;
		sourceApp?: string;
	}

	interface StorageFieldDefinition {
		key: StorageFieldKey;
		diskName: 'data' | 'transcode' | 'downloads' | 'metadata' | 'cache' | 'temp';
		label: string;
		helper: string;
		group: 'processing' | 'library' | 'app';
		editable: boolean;
	}

	interface MetadataSourceDefinition {
		id?: string;
		name?: string;
		description?: string;
		coverage?: string;
		note?: string;
		local?: boolean;
		managed?: boolean;
		available?: boolean;
		runtimeReady?: boolean;
		status?: string;
	}

	type MetadataKind = 'movie' | 'series';

	interface ReviewDraft {
		title: string;
		year: string;
	}

	interface SettingsShellIdentity {
		serverName: string;
		userDisplayName: string;
		userRole: string;
	}

	let isLoading = $state(true);
	let isScanningMovies = $state(false);
	let isScanningTV = $state(false);
	let isRefreshingMovies = $state(false);
	let isRefreshingTV = $state(false);
	let isSavingServerName = $state(false);
	let isSavingScanningAutomation = $state(false);
	let isSavingPlaybackPolicy = $state(false);
	let isSavingStorage = $state(false);
	let isSavingTranscoding = $state(false);
	let isSavingMetadataSources = $state(false);
	let isTestingHardware = $state(false);
	let isSigningOut = $state(false);
	let isSavingPassword = $state(false);
	let isCreatingUser = $state(false);
	let isDeletingUser = $state(false);
	let isSavingLibrary = $state(false);
	let isBrowsingStorage = $state(false);
	let isBrowsingLibrary = $state(false);
	let isLoadingMetadataReview = $state(false);
	let isLoadingPairingRequests = $state(false);
	let isLoadingApprovedDevices = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let scanningSettingsError = $state('');
	let serverNameError = $state('');
	let storageSettingsError = $state('');
	let transcodingSettingsError = $state('');
	let metadataSourceError = $state('');
	let metadataReviewError = $state('');
	let pairingRequestsError = $state('');
	let approvedDevicesError = $state('');
	let usersError = $state('');
	let usersActionMessage = $state('');
	let libraryLoadError = $state('');
	let librarySettingsError = $state('');
	let librarySettingsMessage = $state('');
	let liveStatusError = $state('');
	let discoveryStatusError = $state('');
	let pairingActionMessage = $state('');
	let approvedDevicesActionMessage = $state('');
	let lastUpdatedLabel = $state('');
	let sessionsUnavailable = $state(false);
	let authDisabled = $state(false);
	let activeSection = $state<SettingsSection>('dashboard');
	let activeLibraryActionID = $state('');
	let activeLibraryActionKind = $state<LibraryActionKind>('');
	let activePairingRequestID = $state('');
	let activePairingActionKind = $state<PairingActionKind>('');
	let activeApprovedDeviceID = $state('');
	let activeUserID = $state('');

	let user = $state<AuthSessionUser | null>(null);
	let clientBootstrap = $state<ClientBootstrapResponse>({});
	let libraries = $state<LibraryRecord[]>([]);
	let summary = $state<CatalogSummaryResponse>({});
	let health = $state<CatalogHealthResponse>({});
	let system = $state<SystemStatusResponse>({});
	let settings = $state<SettingsResponse>({});
	let performance = $state<PerformanceSettingsResponse>({});
	let scans = $state<ScanJobItem[]>([]);
	let probes = $state<ProbeJobItem[]>([]);
	let work = $state<WorkQueueItem[]>([]);
	let downloads = $state<DownloadJobItem[]>([]);
	let sessions = $state<SessionItem[]>([]);
	let pairingRequests = $state<PairingRequestItem[]>([]);
	let approvedDevices = $state<ApprovedDeviceItem[]>([]);
	let users = $state<UserAccount[]>([]);
	let reviewItems = $state<ReviewItem[]>([]);
	let versionGroups = $state<VersionGroup[]>([]);
	let metadataRecordsByItem = $state<Record<string, MetadataRecordsResponse>>({});
	let metadataRecordsLoading = $state<Record<string, boolean>>({});
	let metadataRecordsError = $state<Record<string, string>>({});
	let reviewDrafts = $state<Record<string, ReviewDraft>>({});
	let reviewExpanded = $state<Record<string, boolean>>({});
	let reviewRefreshState = $state<Record<string, boolean>>({});
	let reviewApplyState = $state<Record<string, boolean>>({});
	let reviewMessages = $state<Record<string, string>>({});
	let reviewErrors = $state<Record<string, string>>({});
	let buildInfo = $state<BuildInfo | null>(null);
	let discoveryStatus = $state<DiscoveryStatusResponse>({});
	let storageFolderBrowse = $state<FolderBrowseResponse | null>(null);
	let libraryFolderBrowse = $state<FolderBrowseResponse | null>(null);
	let activeStorageBrowseField = $state<EditableStorageFieldKey | ''>('');
	let hardwareModeDraft = $state<'software' | 'hardware'>('software');
	let hardwareTestResult = $state<HardwareTestResponse | null>(null);
	let serverNameDraft = $state('Xuva');
	let librarySyncModeDraft = $state<LibrarySyncModeOption['id']>('daily');
	let syncIntervalDraft = $state('1440');
	let watchDebounceDraft = $state('30');
	let probeBatchLimitDraft = $state('50');
	let playbackPolicyDraft = $state<'original_only' | 'light' | 'full' | 'cinema'>('original_only');
	let metadataSourceOrderDraft = $state<Record<MetadataKind, string[]>>({
		movie: [],
		series: []
	});
	let metadataSourceEnabledDraft = $state<Record<MetadataKind, Record<string, boolean>>>({
		movie: {},
		series: {}
	});
	let storageDraft = $state<Record<StorageFieldKey, string>>({
		dataDir: '',
		transcodeDir: '',
		downloadsDir: '',
		metadataDir: '',
		cacheDir: '',
		tempDir: ''
	});
	let sessionExpiresAt = $state('');
	let passwordDraft = $state('');
	let confirmPasswordDraft = $state('');
	let newUserUsernameDraft = $state('');
	let newUserDisplayNameDraft = $state('');
	let newUserPasswordDraft = $state('');
	let newUserRoleDraft = $state<'admin' | 'standard'>('standard');
	let newLibraryKindDraft = $state<'movies' | 'tv'>('movies');
	let newLibraryNameDraft = $state('');
	let newLibraryPathDraft = $state('');
	let shellIdentity = $state<SettingsShellIdentity>({
		serverName: 'Xuva',
		userDisplayName: '',
		userRole: ''
	});

	let refreshTimer: ReturnType<typeof setTimeout> | null = null;
	const SETTINGS_SHELL_IDENTITY_KEY = 'xuva.settings.shell.identity.v1';

	const serverIdentityHelpText =
		'Xuva uses this name in the browser title and advertises it to local clients when local discovery is running.';

	const storageFieldDefinitions: StorageFieldDefinition[] = [
		{
			key: 'transcodeDir',
			diskName: 'transcode',
			label: 'Transcoding folder',
			helper: 'Where Xuva stores temporary files while preparing playback.',
			group: 'processing',
			editable: true
		},
		{
			key: 'downloadsDir',
			diskName: 'downloads',
			label: 'Optimized versions folder',
			helper: 'Where Xuva stores optimized versions created for playback or travel.',
			group: 'processing',
			editable: true
		},
		{
			key: 'metadataDir',
			diskName: 'metadata',
			label: 'Metadata folder',
			helper: 'Where Xuva stores artwork and metadata it downloads for your library.',
			group: 'library',
			editable: true
		},
		{
			key: 'cacheDir',
			diskName: 'cache',
			label: 'Cache folder',
			helper: 'Where Xuva stores temporary cached data.',
			group: 'library',
			editable: true
		},
		{
			key: 'tempDir',
			diskName: 'temp',
			label: 'Scratch/temp folder',
			helper: 'Where Xuva stores short-lived working files.',
			group: 'library',
			editable: true
		},
		{
			key: 'dataDir',
			diskName: 'data',
			label: 'Data folder',
			helper: 'Where Xuva stores its main settings and local data.',
			group: 'app',
			editable: false
		}
	];

	const librarySyncModeOptions: LibrarySyncModeOption[] = [
		{
			id: 'manual',
			label: 'Manual only',
			description: 'Xuva scans your libraries only when you start a scan.'
		},
		{
			id: 'daily',
			label: 'Scheduled',
			description: 'Xuva checks your libraries on a repeating schedule.'
		},
		{
			id: 'watch',
			label: 'Watch folders',
			description: 'Xuva waits for folder changes, then starts a scan after a short delay.'
		}
	];

	const playbackPolicyOptions: PlaybackPolicyOption[] = [
		{
			id: 'original_only',
			label: 'Original files only',
			description: 'Keep playback as close to the original file as possible. If a device needs help, Xuva offers fallback choices instead of converting automatically.'
		},
		{
			id: 'light',
			label: 'Direct play with audio fixes',
			description: 'Prefer the original video. Xuva can repackage playback or convert audio when that is enough.'
		},
		{
			id: 'full',
			label: 'Compatibility preferred',
			description: 'Allow temporary video conversion while playing when a device needs more help.'
		},
		{
			id: 'cinema',
			label: 'Broadest device support',
			description: 'Allow heavier live compatibility work for the widest range of devices.'
		}
	];
	const metadataKinds: MetadataKind[] = ['movie', 'series'];

	const userDisplayName = $derived.by(() => user?.displayName || user?.username || '');
	const userRoleLabel = $derived.by(() => accountTypeLabel(user?.role));
	const shellServerName = $derived.by(() => {
		const configured = asText(settings.config?.serverName);
		if (configured) return displayServerName(configured);
		const cachedName = asText(shellIdentity.serverName);
		if (cachedName) return displayServerName(cachedName);
		const draftName = asText(serverNameDraft);
		if (draftName) return displayServerName(draftName);
		return 'Xuva';
	});
	const shellUserDisplayName = $derived.by(() => userDisplayName || asText(shellIdentity.userDisplayName));
	const shellUserRole = $derived.by(() => {
		if (user) return userRoleLabel;
		return asText(shellIdentity.userRole);
	});
	const shellDisplayNameForShell = $derived.by(() => {
		if (isLoading) return asText(shellIdentity.userDisplayName) || userDisplayName;
		return userDisplayName;
	});
	const shellRoleForShell = $derived.by(() => {
		if (isLoading) return asText(shellIdentity.userRole) || userRoleLabel;
		return userRoleLabel;
	});
	const userInitials = $derived.by(() => initialsForName(shellUserDisplayName || 'User'));
	const serverDisplayName = $derived.by(() => displayServerName(settings.config?.serverName));
	const canManageSettings = $derived.by(() => authDisabled || asText(user?.role).toLowerCase() === 'admin');
	const hasLibraryLoadError = $derived.by(() => Boolean(libraryLoadError));
	const hasLiveStatusError = $derived.by(() => Boolean(liveStatusError));
	const canShowSignIn = $derived.by(() => !authDisabled && !user);
	const ownerSetupPending = $derived.by(() => !authDisabled && !user && Boolean(clientBootstrap.auth?.bootstrapAllowed));
	const requiresOwnerSignIn = $derived.by(() => !authDisabled && !canManageSettings);
	const ownerAccessMessage = $derived.by(() => 'Sign in as an admin to manage Xuva settings.');
	const ownerActionLabel = $derived.by(() => (ownerSetupPending ? 'Create Admin Account' : 'Sign In'));
	const ownerActionHref = $derived.by(() => (ownerSetupPending ? '/setup-wizard' : '/signin'));
	const ownerActionDetail = $derived.by(() => {
		const defaultUsername = asText(clientBootstrap.auth?.defaultUsername);
		if (ownerSetupPending) {
			return defaultUsername
				? `This server still needs its first admin account. Open Sign In to create ${defaultUsername}.`
				: 'This server still needs its first admin account. Open Sign In to create it.';
		}
		return 'Open Sign In to continue with an admin account.';
	});
	const accessCardLabel = $derived.by(() => {
		if (user) return 'Signed in';
		if (authDisabled) return 'Local access';
		if (ownerSetupPending) return 'Create admin account';
		return 'Sign in';
	});
	const accessCardDetail = $derived.by(() => {
		if (user) return userDisplayName;
		if (authDisabled) return 'sign-in is not required on this server';
		if (ownerSetupPending) return 'first admin account needed for changes';
		return 'admin account needed for changes';
	});
	const canManageUsers = $derived.by(() => !authDisabled && Boolean(user) && canManageSettings);
	const usersCountLabel = $derived.by(() => (users.length === 1 ? '1 account' : `${users.length} accounts`));
	const currentUserRoleLabel = $derived.by(() => accountTypeLabel(user?.role));
	const canSubmitPasswordChange = $derived.by(
		() =>
			canManageUsers &&
			Boolean(asText(passwordDraft)) &&
			asText(passwordDraft) === asText(confirmPasswordDraft) &&
			asText(passwordDraft).length >= 10
	);
	const canSubmitNewUser = $derived.by(
		() =>
			canManageUsers &&
			Boolean(asText(newUserUsernameDraft)) &&
			Boolean(asText(newUserPasswordDraft)) &&
			asText(newUserPasswordDraft).length >= 10
	);
	const accessAccountValue = $derived.by(() => {
		if (user) return userDisplayName;
		if (authDisabled) return 'Local access';
		return 'Not signed in';
	});
	const accessAccountMeta = $derived.by(() => {
		if (user) return userRoleLabel;
		if (authDisabled) return 'Sign-in is not required on this server.';
		return 'Sign in to change protected settings.';
	});
	const metadataSourceCatalog = $derived.by(
		() =>
			(settings.metadataSources || {
				movie: [],
				series: []
			}) as Record<MetadataKind, MetadataSourceDefinition[]>
	);
	const metadataSourcePreferences = $derived.by(
		() =>
			(settings.metadataSourcePreferences || {
				movie: [],
				series: []
			}) as Record<MetadataKind, string[]>
	);
	const metadataSourcesChanged = $derived.by(
		() =>
			joinMetadataSourceIDs(enabledMetadataSourceIDsForKind('movie')) !==
				joinMetadataSourceIDs(metadataSourcePreferences.movie || []) ||
			joinMetadataSourceIDs(enabledMetadataSourceIDsForKind('series')) !==
				joinMetadataSourceIDs(metadataSourcePreferences.series || [])
	);
	const accessSessionValue = $derived.by(() => {
		if (user) return 'Active';
		if (authDisabled) return 'Local';
		return 'Signed out';
	});
	const accessSessionMeta = $derived.by(() => {
		if (user && sessionExpiresAt) return `Current session expires ${formatDateTime(sessionExpiresAt)}.`;
		if (user) return 'This browser has an active Xuva session.';
		if (authDisabled) return 'This server is running without account sign-in.';
		return 'Open Sign In to continue.';
	});
	const activeQueueCount = $derived.by(
		() => [...scans, ...probes, ...work, ...downloads].filter((item) => isActiveStatus(item.status)).length
	);
	const activeSessionCount = $derived.by(() => sessions.filter((item) => Boolean(asText(item.id))).length);
	const appStatus = $derived.by(() => {
		const cpu = Number(system.cpu?.percent || 0);
		const memory = Number(system.memory?.usedPercent || 0);
		if (cpu >= 90 || memory >= 92) return 'critical';
		if (cpu >= 75 || memory >= 85 || activeQueueCount > 0) return 'warning';
		if (cpu > 0 || memory > 0) return 'healthy';
		return 'idle';
	});
	const discoveryRunning = $derived.by(() => Boolean(discoveryStatus.running));
	const discoveryServiceName = $derived.by(
		() => asText(discoveryStatus.serviceName) || serverDisplayName || 'Xuva'
	);
	const discoveryStatusLabel = $derived.by(() => {
		if (discoveryStatusError) return 'Unavailable';
		if (discoveryRunning) return 'Running';
		if (asText(discoveryStatus.lastError)) return 'Needs attention';
		return 'Not running';
	});
	const discoveryStatusTone = $derived.by(() => {
		if (discoveryStatusError) return 'warn';
		if (discoveryRunning) return 'good';
		if (asText(discoveryStatus.lastError)) return 'warn';
		return 'neutral';
	});
	const discoveryStatusMessage = $derived.by(() => {
		if (discoveryStatusError) {
			return 'Local discovery status is unavailable right now.';
		}
		if (discoveryRunning) {
			return `Devices on your home network can find this server as ${discoveryServiceName}.`;
		}
		if (asText(discoveryStatus.lastError)) {
			return 'Xuva could not start local discovery. The server is still available at this web address.';
		}
		return 'Local discovery is not running.';
	});
	const discoveryStatusDetail = $derived.by(() => {
		if (discoveryStatusError) return discoveryStatusError;
		const note = asText(discoveryStatus.note);
		if (note) return note;
		if (discoveryRunning) {
			const portLabel = Number(discoveryStatus.port || 0) > 0 ? String(discoveryStatus.port) : 'current server port';
			return `${asText(discoveryStatus.serviceType) || '_xuva._tcp.local.'} on port ${portLabel}.`;
		}
		return 'Xuva keeps working normally through its web address even when discovery is off.';
	});
	const hasScanningAutomationChanges = $derived.by(
		() =>
			librarySyncModeDraft !== normalizeLibrarySyncMode(settings.config?.librarySyncMode) ||
			syncIntervalDraft !== stringDraft(settings.config?.syncIntervalMins, 1440) ||
			watchDebounceDraft !== stringDraft(settings.config?.watchDebounceSecs, 30) ||
			probeBatchLimitDraft !== stringDraft(settings.config?.probeBatchLimit, 50)
	);
	const hasStorageChanges = $derived.by(
		() =>
			asText(storageDraft.metadataDir) !== asText(settings.config?.metadataDir) ||
			asText(storageDraft.cacheDir) !== asText(settings.config?.cacheDir) ||
			asText(storageDraft.tempDir) !== asText(settings.config?.tempDir)
	);
	const hasTranscodingChanges = $derived.by(
		() =>
			asText(storageDraft.transcodeDir) !== asText(settings.config?.transcodeDir) ||
			asText(storageDraft.downloadsDir) !== asText(settings.config?.downloadsDir) ||
			(hardwareModeDraft === 'hardware') !== Boolean(settings.config?.hardwareUnlocked)
	);
	const configuredLibraries = $derived.by(() => settings.libraries || libraries || []);
	const libraryCards = $derived.by(() =>
		configuredLibraries.map((item) => ({
			id: asText(item.id),
			name: asText(item.name) || libraryTypeLabel(item.kind),
			typeLabel: libraryTypeLabel(item.kind),
			path: asText(item.path),
			storageLabel: libraryStorageLabel(item.storageType),
			statusLabel: libraryStatusLabel(item),
			lastScanLabel: libraryLastScanLabel(item)
		}))
	);
	const scanItems = $derived.by(() => scans.map((item) => scanListItem(item)));
	const sessionItems = $derived.by(() =>
		sessions.map((item, index) => ({
			id: asText(item.id) || `session-${index}`,
			label: asText(item.title) || asText(item.sourceName) || 'Active playback',
			description: [asText(item.deviceId), asText(item.mode) || asText(item.route)].filter(Boolean).join(' - '),
			status: humanStatus(item.state || item.mode || 'active')
		}))
	);
	const warningItems = $derived.by(() => {
		const output: Array<{ id: string; label: string; description: string; status: string }> = [];
		if (!hasLibraryLoadError && libraryCards.length === 0) {
			output.push({
				id: 'warn-library',
				label: 'Library needed',
				description: 'Add a Movies or TV folder so Xuva can build your media library.',
				status: 'Action needed'
			});
		}
		if (Number(health.needsReview || 0) > 0) {
			output.push({
				id: 'warn-review',
				label: 'Metadata review pending',
				description: `${asCount(health.needsReview)} items may need metadata attention.`,
				status: 'Warning'
			});
		}
		if (Number(health.unprobed || 0) > 0) {
			output.push({
				id: 'warn-unprobed',
				label: 'Media check pending',
				description: `${asCount(health.unprobed)} files still need Xuva's media check.`,
				status: 'Warning'
			});
		}
		if (Number(health.unsupported || 0) > 0) {
			output.push({
				id: 'warn-unsupported',
				label: 'Playback review needed',
				description: `${asCount(health.unsupported)} items may need review before playback.`,
				status: 'Critical'
			});
		}
		return output;
	});
	const primaryWarningItem = $derived.by(() => warningItems[0] || null);
	const libraryPanelStatus = $derived.by(() => {
		if (hasLibraryLoadError) return 'warning';
		if (libraryCards.length > 0) return 'healthy';
		return 'idle';
	});
	const scanningPanelStatus = $derived.by(() => {
		if (hasLiveStatusError) return 'warning';
		if (activeQueueCount > 0) return 'warning';
		return 'healthy';
	});
	const metadataPanelStatus = $derived.by(() => {
		if (hasLiveStatusError || metadataReviewError) return 'warning';
		if (Number(health.needsReview || 0) > 0) return 'warning';
		return 'healthy';
	});
	const playbackPanelStatus = $derived.by(() => {
		if (hasLiveStatusError) return 'warning';
		if (activeSessionCount > 0) return 'warning';
		return 'healthy';
	});
	const storageFields = $derived.by(() =>
		storageFieldDefinitions.map((field) => {
			const disk = storageDiskFor(field.diskName);
			const value = asText(storageDraft[field.key]);
			const diskForValue = asText(disk?.path) === value ? disk : undefined;
			const readiness = storageReadinessForDisk(diskForValue, value);
			return {
				...field,
				value,
				readinessLabel: readiness.label,
				readinessTone: readiness.tone,
				readinessDetail: readiness.detail,
				capacityLabel: storageCapacityLabel(diskForValue),
				sharedWithData: Boolean(diskForValue?.sharedWithData),
				error: asText(diskForValue?.error),
				browseAvailable: field.editable && canManageSettings
			};
		})
	);
	const storageConfiguredCount = $derived.by(() => storageFields.filter((field) => Boolean(asText(field.value))).length);
	const storageNeedsAttentionCount = $derived.by(
		() => storageFields.filter((field) => field.readinessLabel === 'Needs attention').length
	);
	const storageDashboardLabel = $derived.by(() => {
		if (hasLiveStatusError) return 'Status unavailable';
		if (storageNeedsAttentionCount > 0) return 'Needs attention';
		if (storageConfiguredCount > 0) return `${asCount(storageConfiguredCount)} folders configured`;
		return 'Folders not set yet';
	});
	const storageDashboardDetail = $derived.by(() => {
		if (hasLiveStatusError) return liveStatusError;
		if (storageNeedsAttentionCount > 0) {
			return `${asCount(storageNeedsAttentionCount)} folder${storageNeedsAttentionCount === 1 ? '' : 's'} need attention before restart.`;
		}
		return 'Review where Xuva keeps its media processing, artwork, cache, and local data.';
	});
	const storagePanelStatus = $derived.by(() => {
		if (hasLiveStatusError) return 'warning';
		if (storageNeedsAttentionCount > 0) return 'warning';
		if (storageConfiguredCount > 0) return 'healthy';
		return 'idle';
	});
	const activeStorageBrowseFieldLabel = $derived.by(() => storageFieldLabel(activeStorageBrowseField));
	const scanningSaveMessage = $derived.by(() =>
		actionMessage === 'Scanning settings saved.' ||
		actionMessage === 'Saved. Restart Xuva for this change to fully take effect.'
			? actionMessage
			: ''
	);
	const storageSaveMessage = $derived.by(() =>
		actionMessage === 'Storage settings saved.' ||
		actionMessage === 'Saved. Restart Xuva for these folder changes to fully take effect.'
			? actionMessage
			: ''
	);
	const transcodingSaveMessage = $derived.by(() =>
		actionMessage === 'Transcoding settings saved.' ||
		actionMessage === 'Saved. Restart Xuva for these transcoding changes to fully take effect.'
			? actionMessage
			: ''
	);
	const playbackSaveMessage = $derived.by(() =>
		actionMessage === 'Playback setting saved.' ||
		actionMessage === 'Playback setting saved. Restart Xuva to apply it.'
			? actionMessage
			: ''
	);
	const serverNameSaveMessage = $derived.by(() => (actionMessage === 'Server name saved.' ? actionMessage : ''));
	const selectedTitle = $derived.by(() => {
		return sectionTitle(activeSection);
	});
	const selectedDescription = $derived.by(() => {
		return sectionDescription(activeSection);
	});

	onMount(() => {
		shellIdentity = readSettingsShellIdentity();
		syncSectionFromHash();
		void loadSettingsSurface();
		void loadBuildInfo();
		window.addEventListener('hashchange', syncSectionFromHash);
		const stream = createEventStream();
		stream.connect();
		const unsubscribe = stream.subscribeAny(({ type }) => {
			if (!shouldRefreshForEvent(type)) return;
			queueSilentRefresh();
		});
		return () => {
			window.removeEventListener('hashchange', syncSectionFromHash);
			unsubscribe();
			stream.disconnect();
			if (refreshTimer) {
				clearTimeout(refreshTimer);
				refreshTimer = null;
			}
		};
	});

	async function loadSettingsSurface(silent = false): Promise<void> {
	if (!silent) {
		isLoading = true;
		loadError = '';
	}
	librarySettingsMessage = '';
	metadataReviewError = '';
	pairingRequestsError = '';
		approvedDevicesError = '';
		usersError = '';
		libraryLoadError = '';
		liveStatusError = '';
		discoveryStatusError = '';
		isLoadingMetadataReview = true;
		isLoadingPairingRequests = true;
		isLoadingApprovedDevices = true;
		sessionsUnavailable = false;
		try {
			const [bootstrapPayload, sessionPayload, settingsPayload] = await Promise.all([
				getClientBootstrap(apiClient).catch(() => ({} as ClientBootstrapResponse)),
				getAuthSessionIfAvailable(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) return {} as AuthSessionResponse;
					throw error;
				}),
				getSettings(apiClient),
			]);

			const [
				librariesPayload,
				summaryPayload,
				healthPayload,
				systemPayload,
				performancePayload,
				scansPayload,
				probesPayload,
				workPayload,
				downloadsPayload,
				sessionsPayload,
				discoveryPayload,
				reviewPayload,
				versionGroupPayload
			] = await Promise.all([
				getLibraries(apiClient).catch((error: unknown) => {
					libraryLoadError = formatLoadError(error);
					return { libraries: [] };
				}),
				getCatalogSummary(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return {} as CatalogSummaryResponse;
				}),
				getCatalogHealth(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return {} as CatalogHealthResponse;
				}),
				getSystemStatus(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return {} as SystemStatusResponse;
				}),
				getPerformanceSettings(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return {} as PerformanceSettingsResponse;
				}),
				getScans(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return { scans: [] };
				}),
				getProbes(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return { probes: [] };
				}),
				getWork(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return { work: [] };
				}),
				getDownloads(apiClient).catch((error: unknown) => {
					liveStatusError = liveStatusError || formatLoadError(error);
					return { downloads: [] };
				}),
				getSessions(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) {
						sessionsUnavailable = true;
						return { sessions: [] };
					}
					liveStatusError = liveStatusError || formatLoadError(error);
					return { sessions: [] };
				}),
				getDiscoveryStatus(apiClient).catch((error: unknown) => {
					discoveryStatusError = formatLoadError(error);
					return {} as DiscoveryStatusResponse;
				}),
				getReviewItems(apiClient).catch((error: unknown) => {
					metadataReviewError = formatLoadError(error);
					return { items: [] };
				}),
				getVersionGroups(apiClient).catch((error: unknown) => {
					metadataReviewError = metadataReviewError || formatLoadError(error);
					return { versions: [] };
				})
			]);

			clientBootstrap = bootstrapPayload || {};
			authDisabled = Boolean(sessionPayload?.authDisabled);
			user = authDisabled ? null : sessionPayload?.user || null;
			sessionExpiresAt = authDisabled ? '' : asText(sessionPayload?.session?.expiresAt);
			libraries = librariesPayload.libraries || [];
			summary = summaryPayload || {};
			health = healthPayload || {};
			system = systemPayload || {};
			settings = settingsPayload || {};
			serverNameDraft = displayServerName(settingsPayload.config?.serverName);
			syncStorageDraft(settingsPayload.config);
			syncMetadataSourceDraft(settingsPayload);
			librarySyncModeDraft = normalizeLibrarySyncMode(settingsPayload.config?.librarySyncMode);
			syncIntervalDraft = stringDraft(settingsPayload.config?.syncIntervalMins, 1440);
			watchDebounceDraft = stringDraft(settingsPayload.config?.watchDebounceSecs, 30);
			probeBatchLimitDraft = stringDraft(settingsPayload.config?.probeBatchLimit, 50);
			scanningSettingsError = '';
			storageSettingsError = '';
			transcodingSettingsError = '';
			hardwareTestResult = null;
			playbackPolicyDraft = normalizePlaybackPolicy(settingsPayload.config?.playbackPolicy);
			performance = performancePayload || {};
			scans = scansPayload.scans || [];
			probes = probesPayload.probes || [];
			work = workPayload.work || [];
			downloads = downloadsPayload.downloads || [];
			sessions = sessionsPayload.sessions || [];
			discoveryStatus = discoveryPayload || {};
			reviewItems = reviewPayload.items || [];
			versionGroups = versionGroupPayload.versions || [];
			const nextSessionUser = sessionPayload?.user;
			const nextUserDisplayName = asText(nextSessionUser?.displayName || nextSessionUser?.username);
			const nextUserRole = nextSessionUser ? accountTypeLabel(nextSessionUser.role) : '';
			shellIdentity = {
				serverName: displayServerName(settingsPayload.config?.serverName),
				userDisplayName: nextSessionUser ? nextUserDisplayName : '',
				userRole: nextSessionUser ? nextUserRole : ''
			};
			writeSettingsShellIdentity(shellIdentity);
			syncReviewDrafts(reviewPayload.items || []);
			const canLoadPairingRequests = Boolean(sessionPayload?.authDisabled || sessionPayload?.user);
			if (canLoadPairingRequests) {
				const pairingPayload = await getPairingRequests(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401) || isApiStatus(error, 403)) {
						pairingRequestsError = '';
						return { requests: [] };
					}
					pairingRequestsError = formatLoadError(error);
					return { requests: [] };
				});
				pairingRequests = pairingPayload.requests || [];
				const approvedDevicesPayload = await getApprovedDevices(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401) || isApiStatus(error, 403)) {
						approvedDevicesError = '';
						return { devices: [] };
					}
					approvedDevicesError = formatLoadError(error);
					return { devices: [] };
				});
				approvedDevices = approvedDevicesPayload.devices || [];
			} else {
				pairingRequests = [];
				approvedDevices = [];
			}
			if (sessionPayload?.user && !sessionPayload?.authDisabled) {
				const usersPayload = await getUsers(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401) || isApiStatus(error, 403)) {
						usersError = '';
						return { users: [] as UserAccount[] };
					}
					usersError = formatLoadError(error);
					return { users: [] as UserAccount[] };
				});
				users = usersPayload.users || [];
			} else {
				users = [];
			}
			lastUpdatedLabel = new Date().toLocaleTimeString();
		} catch (error) {
			loadError = formatLoadError(error);
		} finally {
			isLoadingMetadataReview = false;
			isLoadingPairingRequests = false;
			isLoadingApprovedDevices = false;
			isLoading = false;
		}
	}

	async function loadBuildInfo(): Promise<void> {
		try {
			const response = await fetch('/build-info.json', { cache: 'no-store' });
			if (!response.ok) return;
			buildInfo = (await response.json()) as BuildInfo;
		} catch {
			buildInfo = null;
		}
	}

async function startMovieScan(): Promise<void> {
	isScanningMovies = true;
	actionMessage = '';
	librarySettingsMessage = '';
	try {
		await scanMovies(apiClient, 50);
		actionMessage = 'Movie scan started.';
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isScanningMovies = false;
		}
	}

async function startTVScan(): Promise<void> {
	isScanningTV = true;
	actionMessage = '';
	librarySettingsMessage = '';
	try {
		await scanTV(apiClient, 50);
		actionMessage = 'TV scan started.';
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isScanningTV = false;
		}
	}

async function refreshMovieMetadata(): Promise<void> {
	isRefreshingMovies = true;
	actionMessage = '';
	librarySettingsMessage = '';
	try {
		const result = await refreshMetadataBatch('movie', apiClient, 25);
			actionMessage = metadataRefreshMessage(result.warnings, 'Movie metadata refresh accepted.');
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isRefreshingMovies = false;
		}
	}

async function refreshTVMetadata(): Promise<void> {
	isRefreshingTV = true;
	actionMessage = '';
	librarySettingsMessage = '';
	try {
		const result = await refreshMetadataBatch('series', apiClient, 25);
			actionMessage = metadataRefreshMessage(result.warnings, 'TV metadata refresh accepted.');
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isRefreshingTV = false;
		}
	}

	async function toggleReviewDetails(item: ReviewItem): Promise<void> {
		const key = reviewItemKey(item);
		if (!key) return;
		const next = !Boolean(reviewExpanded[key]);
		reviewExpanded[key] = next;
		if (next && !metadataRecordsByItem[key] && !metadataRecordsLoading[key]) {
			await loadMetadataRecords(item);
		}
	}

	async function loadMetadataRecords(item: ReviewItem, force = false): Promise<void> {
		const key = reviewItemKey(item);
		const kind = asText(item.kind);
		const id = asText(item.id);
		if (!key || !kind || !id) return;
		if (!force && metadataRecordsByItem[key]) return;
		metadataRecordsLoading[key] = true;
		metadataRecordsError[key] = '';
		try {
			metadataRecordsByItem[key] = await getMetadataRecords(kind, id, apiClient);
		} catch (error) {
			metadataRecordsError[key] = formatLoadError(error);
		} finally {
			metadataRecordsLoading[key] = false;
		}
	}

	async function refreshReviewItem(item: ReviewItem): Promise<void> {
		if (!canManageSettings) return;
		const key = reviewItemKey(item);
		const kind = asText(item.kind);
		const id = asText(item.id);
		const title = asText(reviewDrafts[key]?.title || item.title);
		if (!key || !kind || !id) return;
		reviewErrors[key] = '';
		reviewMessages[key] = '';
		if (!title) {
			reviewErrors[key] = 'Enter a title before refreshing metadata.';
			return;
		}
		const year = parseReviewYear(reviewDrafts[key]?.year);
		if (reviewDraftSupportsYear(kind) && reviewDrafts[key]?.year && year === null) {
			reviewErrors[key] = 'Enter a valid year or leave it blank.';
			return;
		}
		reviewRefreshState[key] = true;
		actionMessage = '';
		try {
			const result = await refreshMetadataItem(
				{
					kind,
					id,
					title,
					...(year !== null ? { year } : {})
				},
				apiClient
			);
			reviewMessages[key] = metadataRefreshMessage(result.warnings, 'Metadata refresh finished.');
			actionMessage = reviewMessages[key];
			await loadMetadataRecords(item, true);
			await loadSettingsSurface(true);
		} catch (error) {
			reviewErrors[key] = formatLoadError(error);
		} finally {
			reviewRefreshState[key] = false;
		}
	}

	async function applyManualCorrection(item: ReviewItem): Promise<void> {
		if (!canManageSettings) return;
		const key = reviewItemKey(item);
		const kind = asText(item.kind);
		const id = asText(item.id);
		const title = asText(reviewDrafts[key]?.title || item.title);
		if (!key || !kind || !id) return;
		reviewErrors[key] = '';
		reviewMessages[key] = '';
		if (!title) {
			reviewErrors[key] = 'Enter a title before applying a manual correction.';
			return;
		}
		const year = parseReviewYear(reviewDrafts[key]?.year);
		if (reviewDraftSupportsYear(kind) && reviewDrafts[key]?.year && year === null) {
			reviewErrors[key] = 'Enter a valid year or leave it blank.';
			return;
		}
		reviewApplyState[key] = true;
		actionMessage = '';
		try {
			await applyMetadataMatch(
				{
					kind,
					id,
					title,
					...(year !== null ? { year } : {}),
					provider: 'manual',
					review: false
				},
				apiClient
			);
			reviewMessages[key] = 'Manual correction applied.';
			actionMessage = 'Manual correction applied.';
			await loadMetadataRecords(item, true);
			await loadSettingsSurface(true);
		} catch (error) {
			reviewErrors[key] = formatLoadError(error);
		} finally {
			reviewApplyState[key] = false;
		}
	}

	async function applyRecordMatch(item: ReviewItem, record: MetadataRecord): Promise<void> {
		if (!canManageSettings) return;
		const key = reviewItemKey(item);
		const kind = asText(item.kind);
		const id = asText(item.id);
		const title = asText(record.title);
		if (!key || !kind || !id || !title) return;
		reviewApplyState[key] = true;
		actionMessage = '';
		reviewErrors[key] = '';
		reviewMessages[key] = '';
		const payload: MetadataMatchRequest = {
			kind,
			id,
			title,
			year: Number(record.year || 0) || undefined,
			overview: asText(record.overview),
			provider: asText(record.provider),
			externalId: asText(record.externalId),
			posterUrl: asText(record.posterUrl),
			backdropUrl: asText(record.backdropUrl),
			review: false
		};
		try {
			await applyMetadataMatch(payload, apiClient);
			reviewDrafts[key] = {
				title,
				year: Number(record.year || 0) > 0 ? String(record.year) : ''
			};
			reviewMessages[key] = 'Match applied.';
			actionMessage = 'Match applied.';
			await loadMetadataRecords(item, true);
			await loadSettingsSurface(true);
		} catch (error) {
			reviewErrors[key] = formatLoadError(error);
		} finally {
			reviewApplyState[key] = false;
		}
	}

	async function saveMetadataSources(): Promise<void> {
		if (isSavingMetadataSources || !canManageSettings) return;
		const movie = enabledMetadataSourceIDsForKind('movie');
		const series = enabledMetadataSourceIDsForKind('series');
		metadataSourceError = '';
		actionMessage = '';

		if (movie.length === 0) {
			metadataSourceError = 'Enable at least one movie metadata source.';
			return;
		}
		if (series.length === 0) {
			metadataSourceError = 'Enable at least one TV metadata source.';
			return;
		}

		isSavingMetadataSources = true;
		try {
			const updated = await updateMetadataSourcePreferences({ movie, series }, apiClient);
			settings = {
				...settings,
				...updated,
				config: { ...settings.config, ...updated.config },
				metadataSources: updated.metadataSources || settings.metadataSources,
				metadataSourcePreferences:
					updated.metadataSourcePreferences || settings.metadataSourcePreferences
			};
			syncMetadataSourceDraft(settings);
			actionMessage = updated.restartRequired
				? 'Saved. Restart Xuva for metadata source changes to fully take effect.'
				: 'Metadata source settings saved.';
			await loadSettingsSurface(true);
		} catch (error) {
			metadataSourceError = formatLoadError(error);
		} finally {
			isSavingMetadataSources = false;
		}
	}

	async function saveServerName(): Promise<void> {
		if (isSavingServerName || !canManageSettings) return;
		const nextName = asText(serverNameDraft);
		serverNameError = '';
		actionMessage = '';
		if (!nextName) {
			serverNameError = 'Enter a server name.';
			return;
		}
		if ([...nextName].length > 50) {
			serverNameError = 'Server name must be 50 characters or fewer.';
			return;
		}
		isSavingServerName = true;
		try {
			const updated = await updateSettings({ serverName: nextName }, apiClient);
			settings = { ...settings, ...updated, config: { ...settings.config, ...updated.config } };
			serverNameDraft = displayServerName(settings.config?.serverName);
			shellIdentity = {
				...shellIdentity,
				serverName: displayServerName(settings.config?.serverName)
			};
			writeSettingsShellIdentity(shellIdentity);
			announceServerName();
			actionMessage = 'Server name saved.';
		} catch (error) {
			serverNameError = formatLoadError(error);
		} finally {
			isSavingServerName = false;
		}
	}

	async function saveScanningAutomation(): Promise<void> {
		if (isSavingScanningAutomation || !canManageSettings) return;
		scanningSettingsError = '';
		actionMessage = '';
		const librarySyncMode = normalizeLibrarySyncMode(librarySyncModeDraft);
		const syncIntervalMins = parseWholeNumber(syncIntervalDraft);
		const watchDebounceSecs = parseWholeNumber(watchDebounceDraft);
		const probeBatchLimit = parseWholeNumber(probeBatchLimitDraft);

		if (librarySyncMode === 'daily') {
			if (syncIntervalMins === null || syncIntervalMins < 15) {
				scanningSettingsError = 'Enter a scan interval of at least 15 minutes.';
				return;
			}
		}
		if (librarySyncMode === 'watch') {
			if (watchDebounceSecs === null || watchDebounceSecs < 5 || watchDebounceSecs > 300) {
				scanningSettingsError = 'Enter a folder watch delay between 5 seconds and 5 minutes.';
				return;
			}
		}
		if (probeBatchLimit === null || probeBatchLimit <= 0) {
			scanningSettingsError = 'Enter a media check batch size greater than 0.';
			return;
		}

		isSavingScanningAutomation = true;
		try {
			const updated = await updateSettings(
				{
					librarySyncMode,
					syncIntervalMins: syncIntervalMins ?? 1440,
					watchDebounceSecs: watchDebounceSecs ?? 30,
					probeBatchLimit
				},
				apiClient
			);
			settings = { ...settings, ...updated, config: { ...settings.config, ...updated.config } };
			librarySyncModeDraft = normalizeLibrarySyncMode(updated.config?.librarySyncMode);
			syncIntervalDraft = stringDraft(updated.config?.syncIntervalMins, 1440);
			watchDebounceDraft = stringDraft(updated.config?.watchDebounceSecs, 30);
			probeBatchLimitDraft = stringDraft(updated.config?.probeBatchLimit, 50);
			actionMessage = updated.restartRequired
				? 'Saved. Restart Xuva for this change to fully take effect.'
				: 'Scanning settings saved.';
		} catch (error) {
			scanningSettingsError = formatLoadError(error);
		} finally {
			isSavingScanningAutomation = false;
		}
	}

	async function savePlaybackPolicy(): Promise<void> {
		if (isSavingPlaybackPolicy || !canManageSettings) return;
		isSavingPlaybackPolicy = true;
		actionMessage = '';
		try {
			const updated = await updateSettings({ playbackPolicy: playbackPolicyDraft }, apiClient);
			settings = { ...settings, ...updated, config: { ...settings.config, ...updated.config } };
			playbackPolicyDraft = normalizePlaybackPolicy(updated.config?.playbackPolicy);
			actionMessage = updated.restartRequired
				? 'Playback setting saved. Restart Xuva to apply it.'
				: 'Playback setting saved.';
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isSavingPlaybackPolicy = false;
		}
	}

	async function openStorageBrowser(field: EditableStorageFieldKey, path = ''): Promise<void> {
		if (!canManageSettings) return;
		isBrowsingStorage = true;
		storageSettingsError = '';
		try {
			activeStorageBrowseField = field;
			storageFolderBrowse = await browseFolder(path || asText(storageDraft[field]), apiClient);
		} catch (error) {
			storageSettingsError = formatLoadError(error);
		} finally {
			isBrowsingStorage = false;
		}
	}

	function useStorageBrowsePath(path: string): void {
		if (!activeStorageBrowseField) return;
		storageDraft[activeStorageBrowseField] = asText(path);
		storageSettingsError = '';
	}

	function browseStorageField(field: StorageFieldKey): void {
		if (field === 'dataDir') return;
		void openStorageBrowser(field);
	}

	function browseActiveStoragePath(path: string): Promise<void> {
		if (!activeStorageBrowseField) return Promise.resolve();
		return openStorageBrowser(activeStorageBrowseField, path);
	}

	async function saveStorageSettings(): Promise<void> {
		if (isSavingStorage || !canManageSettings) return;
		storageSettingsError = '';
		actionMessage = '';
		const nextMetadataDir = asText(storageDraft.metadataDir);
		const nextCacheDir = asText(storageDraft.cacheDir);
		const nextTempDir = asText(storageDraft.tempDir);
		if (!nextMetadataDir) {
			storageSettingsError = 'Choose a folder for the Metadata folder.';
			return;
		}
		if (!nextCacheDir) {
			storageSettingsError = 'Choose a folder for the Cache folder.';
			return;
		}
		if (!nextTempDir) {
			storageSettingsError = 'Choose a folder for the Scratch/temp folder.';
			return;
		}

		const payload: UpdateSettingsRequest = {};
		if (nextMetadataDir !== asText(settings.config?.metadataDir)) payload.metadataDir = nextMetadataDir;
		if (nextCacheDir !== asText(settings.config?.cacheDir)) payload.cacheDir = nextCacheDir;
		if (nextTempDir !== asText(settings.config?.tempDir)) payload.tempDir = nextTempDir;

		if (Object.keys(payload).length === 0) return;

		isSavingStorage = true;
		try {
			const updated = await updateSettings(payload, apiClient);
			settings = { ...settings, ...updated, config: { ...settings.config, ...updated.config } };
			syncStorageDraft(updated.config);
			actionMessage = updated.restartRequired
				? 'Saved. Restart Xuva for these folder changes to fully take effect.'
				: 'Storage settings saved.';
			await loadSettingsSurface(true);
		} catch (error) {
			storageSettingsError = formatLoadError(error);
		} finally {
			isSavingStorage = false;
		}
	}

	async function saveTranscodingSettings(): Promise<void> {
		if (isSavingTranscoding || !canManageSettings) return;
		transcodingSettingsError = '';
		actionMessage = '';
		const nextTranscodeDir = asText(storageDraft.transcodeDir);
		const nextDownloadsDir = asText(storageDraft.downloadsDir);
		if (!nextTranscodeDir) {
			transcodingSettingsError = 'Choose a folder for the Transcoding folder.';
			return;
		}
		if (!nextDownloadsDir) {
			transcodingSettingsError = 'Choose a folder for the Optimized versions folder.';
			return;
		}
		const payload: UpdateSettingsRequest = {};
		if (nextTranscodeDir !== asText(settings.config?.transcodeDir)) payload.transcodeDir = nextTranscodeDir;
		if (nextDownloadsDir !== asText(settings.config?.downloadsDir)) payload.downloadsDir = nextDownloadsDir;
		const nextHardwareUnlocked = hardwareModeDraft === 'hardware';
		if (nextHardwareUnlocked !== Boolean(settings.config?.hardwareUnlocked)) {
			payload.hardwareUnlocked = nextHardwareUnlocked;
		}
		if (Object.keys(payload).length === 0) return;
		isSavingTranscoding = true;
		try {
			const updated = await updateSettings(payload, apiClient);
			settings = { ...settings, ...updated, config: { ...settings.config, ...updated.config } };
			syncStorageDraft(updated.config);
			actionMessage = updated.restartRequired
				? 'Saved. Restart Xuva for these transcoding changes to fully take effect.'
				: 'Transcoding settings saved.';
			await loadSettingsSurface(true);
		} catch (error) {
			transcodingSettingsError = formatLoadError(error);
		} finally {
			isSavingTranscoding = false;
		}
	}

	async function runTranscodingHardwareTest(): Promise<void> {
		if (isTestingHardware || !canManageSettings) return;
		isTestingHardware = true;
		transcodingSettingsError = '';
		hardwareTestResult = null;
		try {
			hardwareTestResult = await runHardwareTest(apiClient);
		} catch (error) {
			transcodingSettingsError = formatLoadError(error);
		} finally {
			isTestingHardware = false;
		}
	}

async function scanLibraryItem(library: LibraryRecord): Promise<void> {
	const id = asText(library.id);
	if (!id || !canManageSettings) return;
	activeLibraryActionID = id;
	activeLibraryActionKind = 'scan';
	actionMessage = '';
	librarySettingsMessage = '';
	try {
		await startLibraryScan(id, apiClient);
			actionMessage = `${libraryDisplayName(library)} scan started.`;
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			activeLibraryActionID = '';
			activeLibraryActionKind = '';
		}
	}

async function removeLibraryItem(library: LibraryRecord): Promise<void> {
	const id = asText(library.id);
	if (!id || !canManageSettings || typeof window === 'undefined') return;
	const confirmed = window.confirm('Remove this library from Xuva? Media files are not deleted.');
	if (!confirmed) return;
	activeLibraryActionID = id;
	activeLibraryActionKind = 'remove';
	actionMessage = '';
	librarySettingsMessage = '';
	try {
		await deleteLibrary(id, apiClient);
			actionMessage = `${libraryDisplayName(library)} removed.`;
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			activeLibraryActionID = '';
			activeLibraryActionKind = '';
		}
	}

	async function addLibraryItem(): Promise<void> {
		if (isSavingLibrary || !canManageSettings) return;

		const pathValue = asText(newLibraryPathDraft).trim();
		const nameValue = asText(newLibraryNameDraft).trim();
		if (!pathValue) {
			librarySettingsError = 'Enter a folder path.';
			librarySettingsMessage = '';
			return;
		}

		isSavingLibrary = true;
		librarySettingsError = '';
		librarySettingsMessage = '';
		try {
			const addedLibraryMessage = `${newLibraryKindDraft === 'movies' ? 'Movies' : 'TV'} library added.`;
			await saveLibrary(
				{
					kind: newLibraryKindDraft,
					path: pathValue,
					name: nameValue || (newLibraryKindDraft === 'movies' ? 'Movies' : 'TV')
				},
				apiClient
			);
			newLibraryNameDraft = '';
			newLibraryPathDraft = '';
			await loadSettingsSurface(false);
			librarySettingsMessage = addedLibraryMessage;
		} catch (error) {
			librarySettingsError = formatLoadError(error);
		} finally {
			isSavingLibrary = false;
		}
	}

	async function openLibraryBrowser(path = ''): Promise<void> {
		if (!canManageSettings) return;
		isBrowsingLibrary = true;
		librarySettingsError = '';
		try {
			libraryFolderBrowse = await browseFolder(path || asText(newLibraryPathDraft), apiClient);
		} catch (error) {
			librarySettingsError = formatLoadError(error);
		} finally {
			isBrowsingLibrary = false;
		}
	}

	function useLibraryBrowsePath(path: string): void {
		newLibraryPathDraft = asText(path);
		librarySettingsError = '';
	}

	function browseLibraryPath(path: string): Promise<void> {
		return openLibraryBrowser(path);
	}

	async function signOut(): Promise<void> {
		if (isSigningOut || authDisabled || !user) return;
		isSigningOut = true;
		actionMessage = '';
		try {
			await logout(apiClient);
			shellIdentity = {
				...shellIdentity,
				userDisplayName: '',
				userRole: ''
			};
			writeSettingsShellIdentity(shellIdentity);
			actionMessage = 'Signed out.';
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isSigningOut = false;
		}
	}

	async function changeCurrentPassword(): Promise<void> {
		if (!user || !canManageUsers || isSavingPassword) return;
		usersActionMessage = '';
		usersError = '';
		const nextPassword = asText(passwordDraft);
		const confirmPassword = asText(confirmPasswordDraft);
		if (nextPassword.length < 10) {
			usersError = 'Use at least 10 characters for a password.';
			return;
		}
		if (nextPassword !== confirmPassword) {
			usersError = 'Passwords do not match yet.';
			return;
		}
		isSavingPassword = true;
		try {
			await updateUserPassword(asText(user.id), nextPassword, apiClient);
			passwordDraft = '';
			confirmPasswordDraft = '';
			usersActionMessage = 'Password updated.';
			await loadSettingsSurface(true);
		} catch (error) {
			usersError = formatLoadError(error);
		} finally {
			isSavingPassword = false;
		}
	}

	async function createUserAccount(): Promise<void> {
		if (!canManageUsers || isCreatingUser) return;
		usersActionMessage = '';
		usersError = '';
		const username = asText(newUserUsernameDraft).toLowerCase();
		const displayName = asText(newUserDisplayNameDraft);
		const password = asText(newUserPasswordDraft);
		if (!username) {
			usersError = 'Enter a username.';
			return;
		}
		if (password.length < 10) {
			usersError = 'Use at least 10 characters for a password.';
			return;
		}
		isCreatingUser = true;
		try {
			await createUser(
				{
					username,
					password,
					displayName,
					role: newUserRoleDraft
				},
				apiClient
			);
			newUserUsernameDraft = '';
			newUserDisplayNameDraft = '';
			newUserPasswordDraft = '';
			newUserRoleDraft = 'standard';
			usersActionMessage = 'User account created.';
			await loadSettingsSurface(true);
		} catch (error) {
			usersError = formatLoadError(error);
		} finally {
			isCreatingUser = false;
		}
	}

	async function removeUserAccount(account: UserAccount): Promise<void> {
		if (!canManageUsers || isDeletingUser || typeof window === 'undefined') return;
		const targetID = asText(account.id);
		if (!targetID) return;
		const targetName = asText(account.displayName) || asText(account.username) || 'this account';
		if (!window.confirm(`Remove ${targetName}?`)) return;
		usersActionMessage = '';
		usersError = '';
		activeUserID = targetID;
		isDeletingUser = true;
		try {
			await deleteUser(targetID, apiClient);
			usersActionMessage = `${targetName} removed.`;
			await loadSettingsSurface(true);
		} catch (error) {
			usersError = formatLoadError(error);
		} finally {
			activeUserID = '';
			isDeletingUser = false;
		}
	}

	async function updatePairingRequest(
		request: PairingRequestItem,
		action: Exclude<PairingActionKind, ''>
	): Promise<void> {
		const id = asText(request.id);
		if (!id || !canManageSettings) return;
		activePairingRequestID = id;
		activePairingActionKind = action;
		pairingActionMessage = '';
		pairingRequestsError = '';
		try {
			if (action === 'approve') {
				await approvePairingRequest(id, apiClient);
				pairingActionMessage = `${pairingDeviceName(request)} approved.`;
			} else {
				await denyPairingRequest(id, apiClient);
				pairingActionMessage = `${pairingDeviceName(request)} denied.`;
			}
			await loadSettingsSurface(true);
		} catch (error) {
			pairingRequestsError = formatLoadError(error);
		} finally {
			activePairingRequestID = '';
			activePairingActionKind = '';
		}
	}

	async function revokeApprovedDeviceEntry(device: ApprovedDeviceItem): Promise<void> {
		const id = asText(device.id);
		if (!id || !canManageSettings) return;
		activeApprovedDeviceID = id;
		approvedDevicesActionMessage = '';
		approvedDevicesError = '';
		try {
			await revokeApprovedDevice(id, apiClient);
			approvedDevicesActionMessage = `${approvedDeviceName(device)} removed.`;
			await loadSettingsSurface(true);
		} catch (error) {
			approvedDevicesError = formatLoadError(error);
		} finally {
			activeApprovedDeviceID = '';
		}
	}

	function syncSectionFromHash(): void {
		if (typeof window === 'undefined') return;
		const candidate = window.location.hash.replace(/^#/, '');
		if (!candidate) {
			activeSection = 'dashboard';
			return;
		}
		const normalized = resolveSettingsSection(candidate);
		activeSection = normalized || 'dashboard';
	}

	function isSettingsSection(value: string): value is SettingsSection {
		return [
			'dashboard',
			'general',
			'libraries',
			'scanning',
			'metadata',
			'playback',
			'transcoding',
			'storage',
			'network',
			'pairing',
			'approved-devices',
			'discovery',
			'admin-access',
			'planned-tools',
			'about'
		].includes(value);
	}

	function resolveSettingsSection(value: string): SettingsSection | '' {
		const normalized = asText(value).toLowerCase();
		const aliases: Record<string, SettingsSection> = {
			dashboard: 'dashboard',
			general: 'general',
			library: 'libraries',
			libraries: 'libraries',
			scanning: 'scanning',
			metadata: 'metadata',
			playback: 'playback',
			transcoding: 'transcoding',
			storage: 'storage',
			network: 'network',
			access: 'admin-access',
			users: 'admin-access',
			'owner-access': 'admin-access',
			'admin-access': 'admin-access',
			pairing: 'pairing',
			'approved-devices': 'approved-devices',
			discovery: 'discovery',
			'planned-tools': 'planned-tools',
			planned: 'planned-tools',
			backlog: 'planned-tools',
			about: 'about'
		};
		const candidate = aliases[normalized];
		return candidate && isSettingsSection(candidate) ? candidate : '';
	}

	function queueSilentRefresh(): void {
		if (refreshTimer) clearTimeout(refreshTimer);
		refreshTimer = setTimeout(() => {
			refreshTimer = null;
			void loadSettingsSurface(true);
		}, 220);
	}

	function shouldRefreshForEvent(eventType: string): boolean {
		const normalized = asText(eventType);
		if (!normalized || normalized === 'message' || normalized === 'ready') return false;
		return (
			normalized.startsWith('scan.') ||
			normalized.startsWith('probe.') ||
			normalized.startsWith('download.') ||
			normalized.startsWith('transcode.') ||
			normalized.startsWith('session.') ||
			normalized.startsWith('playback.state.') ||
			normalized.startsWith('metadata.') ||
			normalized.startsWith('settings.') ||
			normalized.startsWith('library.')
		);
	}

	function scanListItem(item: ScanJobItem): { id: string; label: string; description: string; status: string } {
		const kind = libraryKindLabel(asText(item.kind));
		return {
			id: `scan-${asText(item.id) || randomId('scan')}`,
			label: `${kind} scan`,
			description: scanProgressLabel(item),
			status: humanStatus(item.status)
		};
	}

	function scanProgressLabel(item: ScanJobItem): string {
		const progress = Number(item.progress || 0);
		if (Number.isFinite(progress) && progress > 0) return `${Math.round(progress * 100)}% complete`;
		const updated = asText(item.updatedAt || item.createdAt);
		if (updated) return `Updated ${updated}`;
		return 'Scan activity from the current library.';
	}

	function metadataRefreshMessage(warnings: string[] | undefined, fallback: string): string {
		if (Array.isArray(warnings) && warnings.length > 0) return warnings.slice(0, 3).join(' | ');
		return fallback;
	}

	function warningItemHref(itemID: string): string {
		if (itemID === 'warn-library') return '#libraries';
		if (itemID === 'warn-unsupported') return '#playback';
		return '#metadata';
	}

	function warningItemActionLabel(itemID: string): string {
		if (itemID === 'warn-library') return 'Open Libraries';
		if (itemID === 'warn-unsupported') return 'Open Playback';
		return 'Open Metadata';
	}

	function isActiveStatus(value: unknown): boolean {
		const normalized = asText(value).toLowerCase();
		return (
			normalized === 'queued' ||
			normalized === 'running' ||
			normalized === 'active' ||
			normalized === 'started' ||
			normalized === 'in_progress'
		);
	}

	function humanStatus(value: unknown): string {
		const normalized = asText(value).toLowerCase();
		if (!normalized) return 'Unknown';
		if (normalized === 'queued') return 'Waiting';
		if (normalized === 'in_progress') return 'In Progress';
		return normalized
			.split('_')
			.map((part) => capitalize(part))
			.join(' ');
	}

	function pairingDeviceName(item: PairingRequestItem): string {
		return asText(item.deviceName) || 'Device';
	}

	function pairingProfileLabel(value: unknown): string {
		const id = asText(value).toLowerCase();
		if (!id) return 'Unknown device';
		const knownProfiles = Array.isArray(clientBootstrap.profiles) ? clientBootstrap.profiles : [];
		const matched = knownProfiles.find((profile) => asText(profile?.id).toLowerCase() === id);
		if (matched) return asText(matched.name) || capitalize(id);
		if (id === 'apple-tv') return 'Apple TV';
		if (id === 'android-tv') return 'Android TV';
		if (id === 'ios') return 'iPhone / iPad';
		if (id === 'chromecast') return 'Chromecast';
		if (id === 'web') return 'Web Player';
		return id
			.split(/[-_]/)
			.map((part) => capitalize(part))
			.join(' ');
	}

	function pairingStatusSummary(item: PairingRequestItem): string {
		const status = asText(item.status).toLowerCase();
		if (status === 'pending') return 'Waiting for your approval.';
		if (status === 'approved') return 'Approved for this server.';
		if (status === 'denied') return 'Request denied.';
		if (status === 'expired') return 'Pairing request expired.';
		return 'Request status is not available.';
	}

	function canUpdatePairingRequest(item: PairingRequestItem): boolean {
		return canManageSettings && asText(item.status).toLowerCase() === 'pending';
	}

	function approvedDeviceName(item: ApprovedDeviceItem): string {
		return asText(item.displayName) || asText(item.deviceName) || pairingProfileLabel(item.clientProfile);
	}

	function approvedDeviceSummary(item: ApprovedDeviceItem): string {
		const approvedAt = asText(item.approvedAt);
		if (approvedAt) return `Approved ${formatDateTime(approvedAt)}.`;
		return 'Approved for this server.';
	}

	function asCount(value: unknown): string {
		const parsed = Number(value || 0);
		if (!Number.isFinite(parsed)) return '0';
		return new Intl.NumberFormat().format(Math.max(0, Math.round(parsed)));
	}

	function libraryKindLabel(kind: string): string {
		const normalized = asText(kind).toLowerCase();
		if (normalized === 'movies' || normalized === 'movie') return 'Movies';
		if (normalized === 'tv' || normalized === 'series') return 'TV';
		return 'Library';
	}

	function libraryTypeLabel(kind: unknown): string {
		const normalized = asText(kind).toLowerCase();
		if (normalized === 'movies' || normalized === 'movie') return 'Movies library';
		if (normalized === 'tv' || normalized === 'series') return 'TV library';
		return 'Library';
	}

	function libraryDisplayName(library: LibraryRecord): string {
		return asText(library.name) || libraryTypeLabel(library.kind);
	}

	function libraryStorageLabel(storageType: unknown): string {
		const normalized = asText(storageType).toLowerCase();
		if (normalized === 'local') return 'Local drive';
		if (normalized === 'network') return 'Network drive';
		if (normalized === 'mounted') return 'Mounted drive';
		if (normalized === 'removable') return 'Removable drive';
		return 'Storage not known yet';
	}

	function latestScanForLibrary(libraryID: unknown): ScanJobItem | null {
		const id = asText(libraryID);
		if (!id) return null;
		const matches = scans.filter((item) => asText(item.libraryId) === id);
		if (matches.length === 0) return null;
		return [...matches].sort((left, right) => latestTimestamp(right) - latestTimestamp(left))[0] || null;
	}

	function latestTimestamp(item: { updatedAt?: unknown; createdAt?: unknown }): number {
		const value = asText(item.updatedAt || item.createdAt);
		if (!value) return 0;
		const parsed = Date.parse(value);
		return Number.isFinite(parsed) ? parsed : 0;
	}

	function libraryStatusLabel(library: LibraryRecord): string {
		const scan = latestScanForLibrary(library.id);
		if (scan) return humanStatus(scan.status);
		return 'Ready';
	}

	function libraryLastScanLabel(library: LibraryRecord): string {
		const scan = latestScanForLibrary(library.id);
		if (!scan) return '';
		const when = asText(scan.updatedAt || scan.createdAt);
		if (!when) return '';
		return formatDateTime(when);
	}

	function formatDateTime(value: unknown): string {
		const text = asText(value);
		if (!text) return '';
		const parsed = new Date(text);
		if (Number.isNaN(parsed.getTime())) return text;
		return parsed.toLocaleString([], {
			month: 'short',
			day: 'numeric',
			hour: 'numeric',
			minute: '2-digit'
		});
	}

	function normalizeLibrarySyncMode(value: unknown): LibrarySyncModeOption['id'] {
		const normalized = asText(value).toLowerCase();
		if (normalized === 'manual' || normalized === 'watch') return normalized;
		return 'daily';
	}

	function scanningModeDetails(value: unknown): LibrarySyncModeOption {
		const mode = normalizeLibrarySyncMode(value);
		return librarySyncModeOptions.find((item) => item.id === mode) || librarySyncModeOptions[1];
	}

	function scanningModeSummary(): string {
		return scanningModeDetails(settings.config?.librarySyncMode).label;
	}

	function scanningModeMeta(): string {
		const mode = normalizeLibrarySyncMode(settings.config?.librarySyncMode);
		if (mode === 'watch') {
			return `Watches folders and waits ${friendlySeconds(settings.config?.watchDebounceSecs, 30)} before scanning.`;
		}
		if (mode === 'manual') {
			return activeQueueCount > 0
				? `${asCount(activeQueueCount)} active scan tasks right now.`
				: 'Scans start only when you ask Xuva to run them.';
		}
		return `Checks libraries every ${friendlyMinutes(settings.config?.syncIntervalMins, 1440)}.`;
	}

	function stringDraft(value: unknown, fallback: number): string {
		const parsed = parseWholeNumber(value);
		if (parsed === null || parsed <= 0) return String(fallback);
		return String(parsed);
	}

	function parseWholeNumber(value: unknown): number | null {
		const text = asText(value);
		if (!text) return null;
		if (!/^\d+$/.test(text)) return null;
		const parsed = Number(text);
		if (!Number.isFinite(parsed)) return null;
		return Math.trunc(parsed);
	}

	function friendlyMinutes(value: unknown, fallback: number): string {
		const minutes = parseWholeNumber(value) ?? fallback;
		if (minutes % 1440 === 0) {
			const days = minutes / 1440;
			return days === 1 ? '24 hours' : `${days} days`;
		}
		if (minutes % 60 === 0) {
			const hours = minutes / 60;
			return hours === 1 ? '1 hour' : `${hours} hours`;
		}
		return minutes === 1 ? '1 minute' : `${minutes} minutes`;
	}

	function friendlySeconds(value: unknown, fallback: number): string {
		const seconds = parseWholeNumber(value) ?? fallback;
		if (seconds % 60 === 0) {
			const minutes = seconds / 60;
			return minutes === 1 ? '1 minute' : `${minutes} minutes`;
		}
		return seconds === 1 ? '1 second' : `${seconds} seconds`;
	}

	function normalizePlaybackPolicy(value: unknown): PlaybackPolicyOption['id'] {
		const normalized = asText(value).toLowerCase();
		if (normalized === 'light' || normalized === 'full' || normalized === 'cinema') return normalized;
		return 'original_only';
	}

	function playbackPolicyDetails(value: unknown): PlaybackPolicyOption {
		const policy = normalizePlaybackPolicy(value);
		return playbackPolicyOptions.find((item) => item.id === policy) || playbackPolicyOptions[0];
	}

	function initialsForName(name: string): string {
		const words = asText(name).split(/\s+/).filter(Boolean);
		if (words.length === 0) return 'L';
		if (words.length === 1) return words[0].slice(0, 1).toUpperCase();
		return `${words[0][0] || ''}${words[1][0] || ''}`.toUpperCase();
	}

	function capitalize(value: string): string {
		if (!value) return value;
		return `${value.slice(0, 1).toUpperCase()}${value.slice(1)}`;
	}

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function readSettingsShellIdentity(): SettingsShellIdentity {
		if (typeof window === 'undefined') {
			return { serverName: 'Xuva', userDisplayName: '', userRole: '' };
		}
		try {
			const raw = window.localStorage.getItem(SETTINGS_SHELL_IDENTITY_KEY);
			if (!raw) return { serverName: 'Xuva', userDisplayName: '', userRole: '' };
			const parsed = JSON.parse(raw) as Partial<SettingsShellIdentity>;
			return {
				serverName: displayServerName(parsed?.serverName),
				userDisplayName: asText(parsed?.userDisplayName),
				userRole: asText(parsed?.userRole)
			};
		} catch {
			return { serverName: 'Xuva', userDisplayName: '', userRole: '' };
		}
	}

	function writeSettingsShellIdentity(value: SettingsShellIdentity): void {
		if (typeof window === 'undefined') return;
		try {
			window.localStorage.setItem(SETTINGS_SHELL_IDENTITY_KEY, JSON.stringify(value));
		} catch {
			// Ignore storage write failures.
		}
	}

	function displayText(value: unknown): string {
		return asText(value).replace(/_/g, ' ');
	}

	function displayServerName(value: unknown): string {
		return normalizeServerName(value);
	}

	function titleForServerName(value: unknown): string {
		return xuvaTitle(value);
	}

	function announceServerName(): void {
		if (typeof window === 'undefined') return;
		window.dispatchEvent(new CustomEvent('xuva:server-name-changed', { detail: { serverName: serverDisplayName } }));
	}

	function randomId(prefix: string): string {
		return `${prefix.toLowerCase()}-${Math.random().toString(36).slice(2, 8)}`;
	}

	function isApiStatus(error: unknown, expectedStatus: number): boolean {
		if (error instanceof ApiClientError) return error.status === expectedStatus;
		if (typeof error !== 'object' || !error) return false;
		const candidate = (error as { status?: unknown }).status;
		return Number(candidate) === expectedStatus;
	}

	function formatLoadError(error: unknown): string {
		if (error instanceof ApiClientError) return error.userMessage || error.message;
		if (isApiStatus(error, 401)) return 'Your session is no longer active. Sign in again to continue.';
		if (error instanceof Error) return error.message;
		return 'Settings could not load.';
	}

	function accountTypeLabel(role: unknown): string {
		const normalized = asText(role).toLowerCase();
		if (!normalized) return 'User';
		if (normalized === 'admin') return 'Admin';
		if (normalized === 'standard') return 'User';
		return capitalize(normalized);
	}

	function syncStorageDraft(configValues: SettingsResponse['config'] | undefined): void {
		storageDraft.dataDir = asText(configValues?.dataDir);
		storageDraft.transcodeDir = asText(configValues?.transcodeDir);
		storageDraft.downloadsDir = asText(configValues?.downloadsDir);
		storageDraft.metadataDir = asText(configValues?.metadataDir);
		storageDraft.cacheDir = asText(configValues?.cacheDir);
		storageDraft.tempDir = asText(configValues?.tempDir);
		hardwareModeDraft = configValues?.hardwareUnlocked ? 'hardware' : 'software';
	}

	function syncMetadataSourceDraft(sourceSettings: SettingsResponse): void {
		const catalog = (sourceSettings.metadataSources || {
			movie: [],
			series: []
		}) as Record<MetadataKind, MetadataSourceDefinition[]>;
		const preferences = (sourceSettings.metadataSourcePreferences || {
			movie: [],
			series: []
		}) as Record<MetadataKind, string[]>;
		for (const kind of ['movie', 'series'] as MetadataKind[]) {
			const ids = (catalog[kind] || []).map((item) => asText(item.id)).filter(Boolean);
			const preferred = (preferences[kind] || []).map((item) => asText(item)).filter(Boolean);
			const seen = new Set<string>();
			const order: string[] = [];
			for (const id of preferred) {
				if (!ids.includes(id) || seen.has(id)) continue;
				seen.add(id);
				order.push(id);
			}
			for (const id of ids) {
				if (seen.has(id)) continue;
				seen.add(id);
				order.push(id);
			}
			metadataSourceOrderDraft[kind] = order;
			const enabled: Record<string, boolean> = {};
			for (const id of ids) enabled[id] = preferred.includes(id);
			metadataSourceEnabledDraft[kind] = enabled;
		}
		metadataSourceError = '';
	}

	function metadataSourceDefinition(kind: MetadataKind, id: string): MetadataSourceDefinition | undefined {
		return (metadataSourceCatalog[kind] || []).find((item) => asText(item.id) === id);
	}

	function metadataSourceRows(kind: MetadataKind): Array<MetadataSourceDefinition & { id: string; enabled: boolean; unavailable: boolean }> {
		return metadataSourceOrderDraft[kind]
			.map((id) => {
				const source = metadataSourceDefinition(kind, id);
				if (!source) return null;
				const enabled = Boolean(metadataSourceEnabledDraft[kind]?.[id]);
				const unavailable =
					Boolean(source.managed) && !Boolean(source.available);
				return {
					...source,
					id,
					enabled,
					unavailable
				};
			})
			.filter(Boolean) as Array<MetadataSourceDefinition & { id: string; enabled: boolean; unavailable: boolean }>;
	}

	function enabledMetadataSourceIDsForKind(kind: MetadataKind): string[] {
		return metadataSourceOrderDraft[kind].filter((id) => Boolean(metadataSourceEnabledDraft[kind]?.[id]));
	}

	function joinMetadataSourceIDs(values: string[]): string {
		return values.map((value) => asText(value)).filter(Boolean).join('|');
	}

	function metadataSourceStatusLabel(source: MetadataSourceDefinition & { enabled: boolean; unavailable: boolean }): string {
		if (source.unavailable) return 'Unavailable in this build';
		return source.enabled ? 'Enabled' : 'Disabled';
	}

	function metadataSourceStatusTone(source: MetadataSourceDefinition & { enabled: boolean; unavailable: boolean }): 'good' | 'warn' | 'neutral' {
		if (source.unavailable) return 'warn';
		return source.enabled ? 'good' : 'neutral';
	}

	function toggleMetadataSource(kind: MetadataKind, id: string): void {
		const source = metadataSourceDefinition(kind, id);
		if (!source) return;
		metadataSourceEnabledDraft[kind] = {
			...metadataSourceEnabledDraft[kind],
			[id]: !Boolean(metadataSourceEnabledDraft[kind]?.[id])
		};
		metadataSourceError = '';
	}

	function moveMetadataSource(kind: MetadataKind, id: string, direction: -1 | 1): void {
		const order = [...metadataSourceOrderDraft[kind]];
		const index = order.indexOf(id);
		if (index < 0) return;
		const target = index + direction;
		if (target < 0 || target >= order.length) return;
		[order[index], order[target]] = [order[target], order[index]];
		metadataSourceOrderDraft[kind] = order;
	}

	function canMoveMetadataSource(kind: MetadataKind, id: string, direction: -1 | 1): boolean {
		const order = metadataSourceOrderDraft[kind];
		const index = order.indexOf(id);
		const target = index + direction;
		return index >= 0 && target >= 0 && target < order.length;
	}

	function syncReviewDrafts(items: ReviewItem[]): void {
		for (const item of items) {
			const key = reviewItemKey(item);
			if (!key || reviewDrafts[key]) continue;
			reviewDrafts[key] = {
				title: asText(item.title),
				year: ''
			};
		}
	}

	function reviewItemKey(item: ReviewItem): string {
		const kind = asText(item.kind);
		const id = asText(item.id);
		if (!kind || !id) return '';
		return `${kind}:${id}`;
	}

	function reviewKindLabel(kind: unknown): string {
		const normalized = asText(kind).toLowerCase();
		if (normalized === 'movie') return 'Movie';
		if (normalized === 'series') return 'TV series';
		if (normalized === 'episode') return 'TV episode';
		return 'Media item';
	}

	function reviewReasonSummary(reason: unknown): string {
		const normalized = asText(reason).toLowerCase();
		if (
			normalized.includes('unable to infer') ||
			normalized.includes('missing') ||
			normalized.includes('no metadata')
		) {
			return 'Missing metadata';
		}
		if (
			normalized.includes('wrong') ||
			normalized.includes('mismatch') ||
			normalized.includes('match')
		) {
			return 'Wrong match';
		}
		return 'Needs review';
	}

	function reviewReasonDetail(reason: unknown): string {
		const text = asText(reason);
		if (!text) return 'Xuva could not confirm this match automatically.';
		return capitalize(text);
	}

	function reviewDraftFor(item: ReviewItem): ReviewDraft {
		const key = reviewItemKey(item);
		if (!reviewDrafts[key]) {
			reviewDrafts[key] = {
				title: asText(item.title),
				year: ''
			};
		}
		return reviewDrafts[key];
	}

	function parseReviewYear(value: unknown): number | null {
		const text = asText(value);
		if (!text) return null;
		if (!/^\d{4}$/.test(text)) return null;
		const parsed = Number(text);
		if (!Number.isFinite(parsed) || parsed < 1888 || parsed > new Date().getFullYear() + 5) {
			return null;
		}
		return Math.trunc(parsed);
	}

	function reviewDraftSupportsYear(kind: unknown): boolean {
		return asText(kind).toLowerCase() === 'movie';
	}

	function providerLabel(id: unknown): string {
		const normalized = asText(id).toLowerCase();
		if (!normalized) return 'Metadata source';
		const allSources = [...(metadataSourceCatalog.movie || []), ...(metadataSourceCatalog.series || [])];
		const found = allSources.find((item) => asText(item.id).toLowerCase() === normalized);
		if (found?.name) return found.name;
		if (normalized === 'manual') return 'Manual correction';
		if (normalized === 'artwork') return 'Local artwork';
		if (normalized === 'nfo') return 'Local NFO';
		if (normalized === 'filename') return 'Filename and folders';
		if (normalized === 'tvdb') return 'TheTVDB';
		if (normalized === 'tmdb') return 'TMDB';
		if (normalized === 'omdb') return 'OMDb';
		if (normalized === 'tvmaze') return 'TVMaze';
		if (normalized === 'wikidata') return 'Wikidata';
		if (normalized === 'wikipedia') return 'Wikipedia';
		return capitalize(normalized);
	}

	function reviewRecordsForItem(item: ReviewItem): MetadataRecord[] {
		return metadataRecordsByItem[reviewItemKey(item)]?.records || [];
	}

	function reviewBestRecordForItem(item: ReviewItem): MetadataRecord | null {
		return (metadataRecordsByItem[reviewItemKey(item)]?.best as MetadataRecord | null | undefined) || null;
	}

	function versionGroupLink(item: VersionGroup): string {
		const kind = asText(item.kind).toLowerCase();
		const id = asText(item.id);
		if (!id) return '';
		if (kind === 'movie') return `/movies/${encodeURIComponent(id)}`;
		return '';
	}

	function storageDiskFor(name: StorageFieldDefinition['diskName']): SystemDisk | undefined {
		return (system.disks || []).find((disk) => asText(disk.name).toLowerCase() === name);
	}

	function storageFieldLabel(field: StorageFieldKey | EditableStorageFieldKey | ''): string {
		return storageFieldDefinitions.find((item) => item.key === field)?.label || 'Folder';
	}

	function storageReadinessForDisk(
		disk: SystemDisk | undefined,
		path: string
	): { label: 'Ready' | 'Needs attention' | 'Not checked'; tone: 'good' | 'warn' | 'neutral'; detail: string } {
		if (!path) {
			return { label: 'Not checked', tone: 'neutral', detail: 'No folder is set yet.' };
		}
		if (!disk) {
			return { label: 'Not checked', tone: 'neutral', detail: 'Xuva has not checked this folder yet.' };
		}
		if (disk.error || disk.writable === false) {
			return {
				label: 'Needs attention',
				tone: 'warn',
				detail: asText(disk.error) || 'Xuva could not confirm that this folder is writable.'
			};
		}
		return { label: 'Ready', tone: 'good', detail: 'Xuva can use this folder.' };
	}

	function storageCapacityLabel(disk: SystemDisk | undefined): string {
		const total = Number(disk?.totalBytes || 0);
		const free = Number(disk?.freeBytes || 0);
		if (!Number.isFinite(total) || total <= 0 || !Number.isFinite(free) || free < 0) return '';
		return `${formatBytes(free)} free of ${formatBytes(total)}`;
	}

	function formatBytes(value: number): string {
		if (!Number.isFinite(value) || value <= 0) return '0 B';
		const units = ['B', 'KB', 'MB', 'GB', 'TB'];
		let size = value;
		let unitIndex = 0;
		while (size >= 1024 && unitIndex < units.length - 1) {
			size /= 1024;
			unitIndex += 1;
		}
		const rounded = size >= 100 || unitIndex === 0 ? Math.round(size) : Math.round(size * 10) / 10;
		return `${rounded} ${units[unitIndex]}`;
	}

	function sectionTitle(section: SettingsSection): string {
		return {
			dashboard: 'Dashboard',
			general: 'General',
			libraries: 'Libraries',
			scanning: 'Scanning',
			metadata: 'Metadata',
			playback: 'Playback',
			transcoding: 'Transcoding',
			storage: 'Storage',
			network: 'Network',
			pairing: 'Pairing',
			'approved-devices': 'Approved Devices',
			discovery: 'Discovery',
			'admin-access': 'Users',
			'planned-tools': 'Planned (Placeholder)',
			about: 'About'
		}[section];
	}

	function sectionDescription(section: SettingsSection): string {
		return {
			dashboard: 'Check whether your Xuva library is ready and jump to the next useful setting.',
			general: 'Server identity and core server details.',
			libraries: 'Media folders, library status, and setup access.',
			scanning: 'Scan controls, automation, and scan activity.',
			metadata: 'Choose metadata sources, review matches, and refresh movie and TV information.',
			playback: 'Choose how Xuva handles playback compatibility and review active sessions.',
			transcoding: 'Transcoding folder and conversion mode controls.',
			storage: 'Folder paths for metadata, cache, scratch, and app data.',
			network: 'Network visibility and local discovery status.',
			pairing: 'Pending device pairing requests that need admin review.',
			'approved-devices': 'Approved devices that can connect to this server.',
			discovery: 'Local mDNS / Bonjour advertisement status and service identity.',
			'admin-access': 'User accounts, passwords, and admin session state.',
			'planned-tools': 'Placeholder-only section for deferred features.',
			about: 'Xuva identity, build, and local-first details.'
		}[section];
	}
</script>

<svelte:head>
	<title>{titleForServerName(settings.config?.serverName)}</title>
</svelte:head>

<ServerShell
	active={activeSection}
	showStorage={true}
	serverGroupLabel={shellServerName}
	userDisplayName={shellDisplayNameForShell}
	userRole={shellRoleForShell}
	{userInitials}
	canSignOut={!isLoading && Boolean(user) && !authDisabled}
>
	<div class="settings-page">
		{#if isLoading}
			<XuvaPanel title="Loading Settings" subtitle="Checking your Xuva library." />
		{:else if loadError}
			<XuvaPanel title="Settings could not load" subtitle={loadError}>
				<div class="status-actions">
					<XuvaButton variant="secondary" onclick={() => loadSettingsSurface(false)}>Retry</XuvaButton>
					<XuvaButton variant="ghost" href="/">Back to Media</XuvaButton>
				</div>
			</XuvaPanel>
		{:else}
			<header class="settings-head">
				<div>
					<p class="settings-head__eyebrow">Settings</p>
					<strong class="settings-head__server-name" data-testid="settings-server-name">{serverDisplayName}</strong>
					<h1>{selectedTitle}</h1>
					<p>{selectedDescription}</p>
				</div>
				<div class="settings-head__meta">
					<span>Updated {lastUpdatedLabel || '--'}</span>
				</div>
			</header>

			{#if actionMessage}
				<XuvaPanel title="Latest action" subtitle={actionMessage} />
			{/if}

			{#if requiresOwnerSignIn}
				<XuvaPanel title="Admin sign-in required" subtitle={ownerAccessMessage}>
					<p class="settings-note settings-note--inline">{ownerActionDetail}</p>
					<div class="status-actions">
						<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
					</div>
				</XuvaPanel>
			{/if}

			{#if activeSection === 'dashboard'}
				<section class="settings-dashboard" aria-label="Settings dashboard" data-testid="settings-dashboard">
					<article class="settings-dashboard-card settings-dashboard-card--identity">
						<span>Server name</span>
						<strong>{serverDisplayName}</strong>
						<small>{serverIdentityHelpText}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="secondary" size="sm" href="#general">Edit in General</XuvaButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Library</span>
						<strong>{libraryCards.length > 0 ? `${asCount(libraryCards.length)} folders ready` : 'Library setup needed'}</strong>
						<small>{libraryCards.length > 0 ? 'Review folders, run scans, or remove a library.' : 'Add a Movies or TV folder so Xuva can start building your library.'}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="secondary" size="sm" href="#libraries">Open Libraries</XuvaButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Media</span>
						<strong>{Number(summary.movies || 0) + Number(summary.series || 0) > 0 ? `${asCount(summary.movies)} movies` : 'No media found yet'}</strong>
						<small>{Number(summary.movies || 0) + Number(summary.series || 0) > 0 ? `${asCount(summary.series)} shows and ${asCount(summary.episodes)} episodes ready to browse.` : 'Add a library and run a scan to populate Xuva.'}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="secondary" size="sm" href="/movies">Open Movies</XuvaButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Scanning</span>
						<strong>{scanningModeSummary()}</strong>
						<small>{scanningModeMeta()}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="secondary" size="sm" href="#scanning">Open Scanning</XuvaButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Metadata</span>
						<strong>{Number(health.needsReview || 0) > 0 ? 'Review' : 'Ready'}</strong>
						<small>{Number(health.needsReview || 0) > 0 ? `${asCount(health.needsReview)} items need attention.` : 'Movie and TV details look up to date.'}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="secondary" size="sm" href="#metadata">Open Metadata</XuvaButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Playback</span>
						<strong>{playbackPolicyDetails(settings.config?.playbackPolicy).label}</strong>
						<small>{activeSessionCount > 0 ? `${asCount(activeSessionCount)} active sessions right now.` : 'Playback is idle right now.'}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="secondary" size="sm" href="#playback">Open Playback</XuvaButton>
						</div>
					</article>
					{#if canManageSettings}
						<article class="settings-dashboard-card">
							<span>Storage</span>
							<strong>{storageDashboardLabel}</strong>
							<small>{storageDashboardDetail}</small>
							<div class="dashboard-card-actions">
								<XuvaButton variant="secondary" size="sm" href="#storage">Open Storage</XuvaButton>
							</div>
						</article>
					{/if}
					<article class="settings-dashboard-card">
						<span>Access</span>
						<strong>{accessCardLabel}</strong>
						<small>{accessCardDetail}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="secondary" size="sm" href="#admin-access">Open Users</XuvaButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Needs attention</span>
						<strong>{warningItems.length > 0 ? `${asCount(warningItems.length)} item${warningItems.length === 1 ? '' : 's'}` : 'Ready'}</strong>
						<small>{warningItems.length > 0 ? 'A few items need a quick review.' : 'Everything looks ready.'}</small>
						{#if warningItems.length > 0}
							<ul class="dashboard-warning-list">
								{#each warningItems.slice(0, 2) as item (item.id)}
									<li>{item.label}</li>
								{/each}
							</ul>
						{/if}
						<div class="dashboard-card-actions">
							<XuvaButton
								variant="ghost"
								size="sm"
								href={primaryWarningItem ? warningItemHref(primaryWarningItem.id) : '#dashboard'}
							>
								{primaryWarningItem ? warningItemActionLabel(primaryWarningItem.id) : 'Review Settings'}
							</XuvaButton>
						</div>
					</article>
					<article class="settings-dashboard-card settings-dashboard-card--quiet">
						<span>About</span>
						<strong>Xuva</strong>
						<small>{asText(buildInfo?.buildID) || 'Local build details'}</small>
						<div class="dashboard-card-actions">
							<XuvaButton variant="ghost" size="sm" href="#about">Open About</XuvaButton>
						</div>
					</article>
				</section>
				{#if libraryLoadError || liveStatusError}
					<XuvaPanel title="Some settings details are unavailable" subtitle={libraryLoadError || liveStatusError} />
				{/if}
			{:else if activeSection === 'libraries'}
			<section id="libraries" class="settings-section" data-testid="settings-section-content" data-section="libraries">
				<SettingsPanel title="Libraries" description="Media folders and library status." status={libraryPanelStatus}>
					<div class="stat-grid stat-grid--compact">
						<XuvaStat label="Libraries" value={asCount(libraryCards.length)} meta={libraryLoadError ? 'Current library details are unavailable right now.' : libraryCards.length > 0 ? 'Configured folders' : 'Add a library to begin'} tone={libraryLoadError ? 'warn' : libraryCards.length > 0 ? 'good' : 'warn'} />
						<XuvaStat label="Movies" value={asCount(summary.movies)} meta="Current movie catalog count." />
						<XuvaStat label="Shows" value={asCount(summary.series)} meta={`${asCount(summary.episodes)} episodes in the current TV catalog.`} />
					</div>
					{#if libraryLoadError}
						<p class="settings-error">{libraryLoadError}</p>
					{/if}
					{#if canManageSettings}
						<form class="settings-stack users-create-form" onsubmit={(event) => { event.preventDefault(); void addLibraryItem(); }}>
							<label class="settings-field">
								<span>Library type</span>
								<select bind:value={newLibraryKindDraft}>
									<option value="movies">Movies</option>
									<option value="tv">TV</option>
								</select>
							</label>
							<label class="settings-field">
								<span>Library name (optional)</span>
								<input bind:value={newLibraryNameDraft} maxlength="80" placeholder={newLibraryKindDraft === 'movies' ? 'Movies' : 'TV'} />
							</label>
							<label class="settings-field" style="grid-column: 1 / -1;">
								<span>Folder path</span>
								<div class="storage-input-row">
									<input bind:value={newLibraryPathDraft} placeholder="D:\\Media\\Movies" required />
									<XuvaButton
										variant="secondary"
										size="sm"
										type="button"
										disabled={isBrowsingLibrary}
										onclick={() => void openLibraryBrowser()}
									>
										{isBrowsingLibrary ? 'Loading folders...' : 'Browse'}
									</XuvaButton>
								</div>
								<small>Type a local path or UNC share (example: \\NAS\Media\Movies), or choose Browse.</small>
							</label>
							<div class="status-actions">
								<XuvaButton variant="primary" disabled={isSavingLibrary} onclick={addLibraryItem}>
									{isSavingLibrary ? 'Saving...' : 'Save Library'}
								</XuvaButton>
							</div>
						</form>
						{#if librarySettingsMessage}
							<p class="settings-feedback">{librarySettingsMessage}</p>
						{/if}
						{#if librarySettingsError}
							<p class="settings-error">{librarySettingsError}</p>
						{/if}
						<FolderBrowserPanel
							browser={libraryFolderBrowse}
							title="Library folder browser"
							subtitle={libraryFolderBrowse?.path || 'Select a library folder path.'}
							onBrowse={browseLibraryPath}
							onUsePath={useLibraryBrowsePath}
						/>
					{/if}
					{#if libraryCards.length > 0}
						<div class="settings-subsection">
							<div class="settings-subsection__head">
								<div>
									<h3>Configured libraries</h3>
									<p>Review folder paths, run a scan, or remove a library that is no longer needed.</p>
								</div>
							</div>
							<div class="library-list" data-testid="library-list">
							{#each libraryCards as library (library.id)}
								<article class="library-card" data-testid="library-item">
									<div class="library-card__body">
										<div class="library-card__heading">
											<div>
												<h3>{library.name}</h3>
												<p>{library.typeLabel}</p>
											</div>
											<span class="library-card__status">{library.statusLabel}</span>
										</div>
										<dl class="library-card__facts">
											<div>
												<dt>Folder</dt>
												<dd>{library.path || 'Not set'}</dd>
											</div>
											<div>
												<dt>Storage</dt>
												<dd>{library.storageLabel}</dd>
											</div>
											{#if library.lastScanLabel}
												<div>
													<dt>Last scan</dt>
													<dd>{library.lastScanLabel}</dd>
												</div>
											{/if}
										</dl>
										<p class="library-card__note">Item counts are available for the full catalog, but not per library yet.</p>
									</div>
									{#if canManageSettings}
										<div class="library-card__actions">
											<XuvaButton
												variant="secondary"
												size="sm"
												disabled={activeLibraryActionID === library.id}
												onclick={() => scanLibraryItem(configuredLibraries.find((item) => asText(item.id) === library.id) || {})}
											>
												{activeLibraryActionID === library.id && activeLibraryActionKind === 'scan' ? 'Scanning...' : 'Scan'}
											</XuvaButton>
											<XuvaButton
												variant="danger"
												size="sm"
												disabled={activeLibraryActionID === library.id}
												onclick={() => removeLibraryItem(configuredLibraries.find((item) => asText(item.id) === library.id) || {})}
											>
												{activeLibraryActionID === library.id && activeLibraryActionKind === 'remove' ? 'Removing...' : 'Remove'}
											</XuvaButton>
										</div>
									{/if}
								</article>
							{/each}
							</div>
						</div>
					{:else}
						<div class="settings-subsection settings-subsection--quiet settings-subsection--empty">
							<div class="settings-subsection__head">
								<div>
									<h3>No library folders yet</h3>
									<p>Add a Movies or TV folder so Xuva can scan it and build your media library.</p>
								</div>
							</div>
						</div>
					{/if}
					{#if !canManageSettings && !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as an admin to manage Xuva settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
							</div>
						</div>
					{/if}
				</SettingsPanel>
			</section>

			{:else if activeSection === 'scanning'}
			<section id="scanning" class="settings-section" data-testid="settings-section-content" data-section="scanning">
				<SettingsPanel title="Scanning" description="Start real library scans, choose how Xuva checks libraries, and review current scan activity." status={scanningPanelStatus}>
					{#snippet actions()}
						{#if canManageSettings}
							<XuvaButton
								variant="primary"
								disabled={isSavingScanningAutomation || !hasScanningAutomationChanges}
								onclick={saveScanningAutomation}
							>
								{isSavingScanningAutomation ? 'Saving...' : 'Save Scanning Settings'}
							</XuvaButton>
						{:else if !authDisabled}
							<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
						{/if}
					{/snippet}
					{#if liveStatusError}
						<p class="settings-error">{liveStatusError}</p>
					{/if}
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Manual scans</h3>
								<p>Run a scan when you want Xuva to check movies or TV right now.</p>
							</div>
						</div>
						<div class="status-actions">
							<XuvaButton variant="primary" onclick={startMovieScan} disabled={isScanningMovies || isScanningTV}>
								{isScanningMovies ? 'Scanning Movies...' : 'Scan Movies'}
							</XuvaButton>
							<XuvaButton variant="secondary" onclick={startTVScan} disabled={isScanningMovies || isScanningTV}>
								{isScanningTV ? 'Scanning TV...' : 'Scan TV'}
							</XuvaButton>
						</div>
					</div>
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Automation</h3>
								<p>Choose how Xuva checks your libraries for new or changed media.</p>
							</div>
						</div>
						<div class="stat-grid stat-grid--compact">
						<XuvaStat
							label="Library Scan Mode"
							value={scanningModeDetails(settings.config?.librarySyncMode).label}
							meta={scanningModeDetails(settings.config?.librarySyncMode).description}
						/>
						{#if normalizeLibrarySyncMode(settings.config?.librarySyncMode) === 'watch'}
							<XuvaStat
								label="Folder Watch Delay"
								value={friendlySeconds(settings.config?.watchDebounceSecs, 30)}
								meta="How long Xuva waits after a folder change before starting a scan."
							/>
						{:else}
							<XuvaStat
								label="Scan Interval"
								value={normalizeLibrarySyncMode(settings.config?.librarySyncMode) === 'manual' ? 'Manual' : friendlyMinutes(settings.config?.syncIntervalMins, 1440)}
								meta={normalizeLibrarySyncMode(settings.config?.librarySyncMode) === 'manual' ? 'Scans start only when you ask Xuva to run them.' : 'Time between scheduled library scans.'}
							/>
						{/if}
						<XuvaStat
							label="Media check batch size"
							value={asCount(settings.config?.probeBatchLimit || 50)}
							meta="Items Xuva checks at a time after a library scan."
						/>
						</div>
					</div>
					{#if canManageSettings}
						<form class="scanning-automation-form" data-testid="scanning-automation-form" onsubmit={(event) => { event.preventDefault(); void saveScanningAutomation(); }}>
							<div class="settings-field">
								<span>Library scan mode</span>
								<small>Choose how Xuva checks your libraries for new or changed media.</small>
							</div>
							<div class="playback-policy-options">
								{#each librarySyncModeOptions as option (option.id)}
									<label class="playback-policy-option">
										<input
											type="radio"
											name="librarySyncMode"
											value={option.id}
											checked={librarySyncModeDraft === option.id}
											onchange={() => {
												librarySyncModeDraft = option.id;
												scanningSettingsError = '';
											}}
										/>
										<div>
											<strong>{option.label}</strong>
											<span>{option.description}</span>
										</div>
									</label>
								{/each}
							</div>
							{#if librarySyncModeDraft === 'daily'}
								<label class="settings-field">
									<span>Scan interval</span>
									<input
										bind:value={syncIntervalDraft}
										type="number"
										min="15"
										step="1"
										inputmode="numeric"
										placeholder="1440"
										oninput={() => (scanningSettingsError = '')}
									/>
									<small>How often Xuva runs a scheduled library scan. Minimum 15 minutes.</small>
								</label>
							{:else if librarySyncModeDraft === 'watch'}
								<label class="settings-field">
									<span>Folder watch delay</span>
									<input
										bind:value={watchDebounceDraft}
										type="number"
										min="5"
										max="300"
										step="1"
										inputmode="numeric"
										placeholder="30"
										oninput={() => (scanningSettingsError = '')}
									/>
									<small>How long Xuva waits after a folder change before starting a scan.</small>
								</label>
							{/if}
							<details class="settings-advanced" data-testid="scanning-advanced">
								<summary>Advanced scanning</summary>
								<div class="settings-advanced__content">
									<label class="settings-field">
										<span>Media check batch size</span>
										<input
											bind:value={probeBatchLimitDraft}
											type="number"
											min="1"
											step="1"
											inputmode="numeric"
											placeholder="50"
											oninput={() => (scanningSettingsError = '')}
										/>
										<small>How many items Xuva checks in one batch during library scanning.</small>
									</label>
								</div>
							</details>
							{#if scanningSaveMessage}
								<p class="settings-feedback">{scanningSaveMessage}</p>
							{/if}
							{#if scanningSettingsError}
								<p class="settings-error">{scanningSettingsError}</p>
							{/if}
						</form>
					{:else if !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as an admin to manage Xuva settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
							</div>
						</div>
					{/if}
					<ActivityListShell title="Recent Scans">
						<XuvaActionList
							items={scanItems}
							emptyLabel="No scan activity right now."
						/>
					</ActivityListShell>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'metadata'}
			<section id="metadata" class="settings-section" data-testid="settings-section-content" data-section="metadata">
				<SettingsPanel title="Metadata" description="Choose metadata sources, review matches, and refresh movie and TV information." status={metadataPanelStatus}>
					{#snippet actions()}
						{#if canManageSettings}
							<XuvaButton variant="primary" disabled={isSavingMetadataSources || !metadataSourcesChanged} onclick={saveMetadataSources}>
								{isSavingMetadataSources ? 'Saving Metadata Sources...' : 'Save Metadata Sources'}
							</XuvaButton>
						{:else if canShowSignIn}
							<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
						{/if}
					{/snippet}
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Metadata Sources</h3>
								<p>Xuva checks enabled sources in order. Move your preferred source higher.</p>
							</div>
						</div>
						{#if canManageSettings}
							<form class="metadata-source-form" data-testid="metadata-source-form" onsubmit={(event) => { event.preventDefault(); void saveMetadataSources(); }}>
								<div class="metadata-source-groups">
									{#each metadataKinds as kind (kind)}
										<div class="metadata-source-group">
											<div class="settings-subsection__head settings-subsection__head--tight">
												<div>
													<h4>{kind === 'movie' ? 'Movie metadata sources' : 'TV metadata sources'}</h4>
													<p>{kind === 'movie' ? 'Choose where Xuva looks first for movie titles, artwork, and ratings.' : 'Choose where Xuva looks first for series titles, artwork, and ratings.'}</p>
												</div>
											</div>
											<div class="metadata-source-list" data-testid={`metadata-source-list-${kind}`}>
												{#each metadataSourceRows(kind) as source (source.id)}
													<div class="metadata-source-card" class:metadata-source-card--disabled={!source.enabled}>
														<div class="metadata-source-card__main">
															<label class="metadata-source-card__toggle">
																<input
																	type="checkbox"
																	checked={source.enabled}
																	onchange={() => toggleMetadataSource(kind, source.id)}
																/>
																<div>
																	<strong>{source.name || 'Metadata source'}</strong>
																	<span>{source.description || source.coverage || source.note || ''}</span>
																</div>
															</label>
															<div class="metadata-source-card__meta">
																<span class={`status-pill status-pill--${metadataSourceStatusTone(source)}`}>{metadataSourceStatusLabel(source)}</span>
																{#if source.note && !source.unavailable}
																	<small>{source.note}</small>
																{:else if source.unavailable}
																	<small>This source is managed by Xuva and is unavailable in this build.</small>
																{/if}
															</div>
														</div>
														<div class="metadata-source-card__actions">
															<XuvaButton variant="ghost" size="sm" onclick={() => moveMetadataSource(kind, source.id, -1)} disabled={!canMoveMetadataSource(kind, source.id, -1)}>
																Move Up
															</XuvaButton>
															<XuvaButton variant="ghost" size="sm" onclick={() => moveMetadataSource(kind, source.id, 1)} disabled={!canMoveMetadataSource(kind, source.id, 1)}>
																Move Down
															</XuvaButton>
														</div>
													</div>
												{/each}
											</div>
										</div>
									{/each}
								</div>
								{#if metadataSourceError}
									<p class="status-copy status-copy--warn">{metadataSourceError}</p>
								{/if}
							</form>
						{:else}
							<div class="settings-subsection settings-subsection--quiet">
								<div class="settings-subsection__head">
									<div>
										<h4>Metadata Sources</h4>
										<p>{ownerAccessMessage}</p>
									</div>
								</div>
								<div class="status-actions">
									{#if canShowSignIn}
										<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
									{/if}
								</div>
							</div>
						{/if}
					</div>

					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Metadata Review</h3>
								<p>Review items that still need metadata attention and fix missing metadata or a wrong match.</p>
							</div>
						</div>
						<div class="stat-grid stat-grid--compact">
							<XuvaStat label="Needs review" value={asCount(health.needsReview)} meta="Missing metadata or wrong matches that still need attention." tone={Number(health.needsReview || 0) > 0 ? 'warn' : 'good'} />
							<XuvaStat label="Missing media checks" value={asCount(health.unprobed)} meta="Files still waiting for Xuva's media check." tone={Number(health.unprobed || 0) > 0 ? 'warn' : 'good'} />
							<XuvaStat label="Playback review" value={asCount(health.unsupported)} meta="Items that may need extra playback attention." tone={Number(health.unsupported || 0) > 0 ? 'warn' : 'good'} />
							<XuvaStat label="High bitrate" value={asCount(health.highBitrate)} meta="Large files that may need stronger device support." />
							<XuvaStat label="Subtitles found" value={asCount(health.withSubtitles)} meta="Available for playback when supported." />
						</div>
						{#if liveStatusError}
							<p class="settings-error">{liveStatusError}</p>
						{/if}
						{#if requiresOwnerSignIn}
							<div class="settings-auth-callout">
								<p class="settings-note">Sign in as an admin to update metadata.</p>
								<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
								<div class="status-actions">
									<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
								</div>
							</div>
						{/if}
						{#if metadataReviewError}
							<p class="settings-error">{metadataReviewError}</p>
						{/if}
						<div class="metadata-review-list" data-testid="metadata-review-list">
							{#if isLoadingMetadataReview}
								<p class="settings-note">Loading metadata review items...</p>
							{:else if reviewItems.length === 0}
								<p class="settings-note">No metadata review items right now.</p>
							{:else}
								{#each reviewItems as item (reviewItemKey(item))}
									<div class="review-card">
										<div class="review-card__head">
											<div>
												<h4>{asText(item.title) || 'Metadata item'}</h4>
												<p>{reviewKindLabel(item.kind)} - {reviewReasonSummary(item.reviewReason)}</p>
											</div>
											<span class="status-pill status-pill--warn">Needs review</span>
										</div>
										<p class="review-card__reason">{reviewReasonDetail(item.reviewReason)}</p>
										<div class="status-actions review-card__actions">
											<XuvaButton variant="ghost" size="sm" onclick={() => void toggleReviewDetails(item)}>
												{reviewExpanded[reviewItemKey(item)] ? 'Hide records' : 'View records'}
											</XuvaButton>
											{#if canManageSettings}
												<XuvaButton
													variant="secondary"
													size="sm"
													disabled={Boolean(reviewRefreshState[reviewItemKey(item)])}
													onclick={() => void refreshReviewItem(item)}
												>
													{reviewRefreshState[reviewItemKey(item)] ? 'Refreshing...' : 'Refresh metadata'}
												</XuvaButton>
											{/if}
										</div>
										{#if reviewExpanded[reviewItemKey(item)]}
											<div class="review-card__details">
												{#if metadataRecordsLoading[reviewItemKey(item)]}
													<p class="settings-note">Loading records...</p>
												{:else}
													{#if metadataRecordsError[reviewItemKey(item)]}
														<p class="settings-error">{metadataRecordsError[reviewItemKey(item)]}</p>
													{/if}
													{#if reviewRecordsForItem(item).length > 0}
														<div class="review-records">
															{#each reviewRecordsForItem(item) as record, index (`${reviewItemKey(item)}-${asText(record.provider)}-${index}`)}
																<div class="review-record-card">
																	<div class="review-record-card__head">
																		<div>
																			<strong>{providerLabel(record.provider)}</strong>
																			<span>{asText(record.title) || 'Metadata record'}{#if Number(record.year || 0) > 0} ({record.year}){/if}</span>
																		</div>
																		{#if canManageSettings}
																			<XuvaButton
																				variant="ghost"
																				size="sm"
																				disabled={Boolean(reviewApplyState[reviewItemKey(item)])}
																				onclick={() => void applyRecordMatch(item, record)}
																			>
																				Apply match
																			</XuvaButton>
																		{/if}
																	</div>
																	{#if asText(record.overview)}
																		<p>{asText(record.overview)}</p>
																	{/if}
																</div>
															{/each}
														</div>
													{:else}
														<p class="settings-note">No metadata records yet. Try Refresh metadata to fetch another pass.</p>
													{/if}
												{/if}
												{#if canManageSettings}
													<form class="review-manual-form" onsubmit={(event) => { event.preventDefault(); void applyManualCorrection(item); }}>
														<div class="settings-subsection__head settings-subsection__head--tight">
															<div>
																<h5>Manual correction</h5>
																<p>Correct the title{#if reviewDraftSupportsYear(item.kind)} or year{/if}, then refresh or apply it directly.</p>
															</div>
														</div>
														<div class="review-manual-form__fields">
															<label class="settings-field">
																<span>Title</span>
																<input
																	type="text"
																	value={reviewDraftFor(item).title}
																	placeholder="Enter a corrected title"
																	oninput={(event) => {
																		reviewDraftFor(item).title = (event.currentTarget as HTMLInputElement).value;
																		reviewErrors[reviewItemKey(item)] = '';
																		reviewMessages[reviewItemKey(item)] = '';
																	}}
																/>
															</label>
															{#if reviewDraftSupportsYear(item.kind)}
																<label class="settings-field">
																	<span>Year</span>
																	<input
																	type="number"
																	min="1888"
																	max="2100"
																	inputmode="numeric"
																	value={reviewDraftFor(item).year}
																	placeholder="Optional"
																	oninput={(event) => {
																		reviewDraftFor(item).year = (event.currentTarget as HTMLInputElement).value;
																		reviewErrors[reviewItemKey(item)] = '';
																		reviewMessages[reviewItemKey(item)] = '';
																	}}
																/>
															</label>
															{/if}
														</div>
														<div class="status-actions">
															<XuvaButton
																variant="secondary"
																type="button"
																disabled={Boolean(reviewRefreshState[reviewItemKey(item)])}
																onclick={() => void refreshReviewItem(item)}
															>
																{reviewRefreshState[reviewItemKey(item)] ? 'Refreshing...' : 'Refresh with correction'}
															</XuvaButton>
															<XuvaButton
																variant="primary"
																type="submit"
																disabled={Boolean(reviewApplyState[reviewItemKey(item)])}
															>
																{reviewApplyState[reviewItemKey(item)] ? 'Applying...' : 'Apply correction'}
															</XuvaButton>
														</div>
													</form>
												{/if}
												{#if reviewMessages[reviewItemKey(item)]}
													<p class="settings-feedback">{reviewMessages[reviewItemKey(item)]}</p>
												{/if}
												{#if reviewErrors[reviewItemKey(item)]}
													<p class="settings-error">{reviewErrors[reviewItemKey(item)]}</p>
												{/if}
											</div>
										{/if}
									</div>
								{/each}
							{/if}
						</div>
					</div>

					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Version Groups</h3>
								<p>Multiple versions found. Review items where Xuva found more than one version.</p>
							</div>
						</div>
						<div class="metadata-version-groups" data-testid="metadata-version-groups">
							{#if versionGroups.length === 0}
								<p class="settings-note">No multiple-version groups right now.</p>
							{:else}
								{#each versionGroups as item (`${asText(item.kind)}-${asText(item.id)}`)}
									<div class="version-group-card">
										<div>
											<strong>{asText(item.title) || 'Version group'}</strong>
											<span>{reviewKindLabel(item.kind)} - {asCount(item.versionCount)} versions</span>
										</div>
										{#if versionGroupLink(item)}
											<XuvaButton variant="ghost" size="sm" href={versionGroupLink(item)}>
												Open item
											</XuvaButton>
										{/if}
									</div>
								{/each}
							{/if}
						</div>
					</div>

					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Refresh Metadata</h3>
								<p>Use these actions when movie or TV details need another pass.</p>
							</div>
						</div>
						<div class="status-actions">
							<XuvaButton variant="primary" onclick={refreshMovieMetadata} disabled={isRefreshingMovies || isRefreshingTV}>
								{isRefreshingMovies ? 'Refreshing Movies...' : 'Refresh Movies'}
							</XuvaButton>
							<XuvaButton variant="secondary" onclick={refreshTVMetadata} disabled={isRefreshingMovies || isRefreshingTV}>
								{isRefreshingTV ? 'Refreshing TV...' : 'Refresh TV'}
							</XuvaButton>
						</div>
						{#if requiresOwnerSignIn}
							<p class="settings-note">Sign in as an admin to update metadata.</p>
						{/if}
					</div>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'playback'}
			<section id="playback" class="settings-section" data-testid="settings-section-content" data-section="playback">
				<SettingsPanel title="Playback" description="Playback policy, compatibility status, and active playback sessions." status={playbackPanelStatus}>
					{#snippet actions()}
						{#if canManageSettings}
							<XuvaButton
								variant="primary"
								disabled={isSavingPlaybackPolicy || playbackPolicyDraft === normalizePlaybackPolicy(settings.config?.playbackPolicy)}
								onclick={savePlaybackPolicy}
							>
								{isSavingPlaybackPolicy ? 'Saving...' : 'Save Playback Policy'}
							</XuvaButton>
						{:else if !authDisabled}
							<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
						{/if}
					{/snippet}
					<div class="stat-grid stat-grid--compact">
						<XuvaStat
							label="Playback Policy"
							value={playbackPolicyDetails(settings.config?.playbackPolicy).label}
							meta={playbackPolicyDetails(settings.config?.playbackPolicy).description}
						/>
						<XuvaStat label="Compatibility Support" value={asText(performance.hardwareAcceleration?.status) || 'Unknown'} meta="Shows whether Xuva can help when a device needs a different playback format." />
						<XuvaStat label="Active Sessions" value={asCount(activeSessionCount)} meta={sessionsUnavailable ? 'Sign in to view current playback sessions.' : 'Current playback sessions.'} tone={activeSessionCount > 0 ? 'warn' : 'good'} />
					</div>
					{#if liveStatusError}
						<p class="settings-error">{liveStatusError}</p>
					{/if}
					{#if canManageSettings}
						<form class="playback-policy-form" data-testid="playback-policy-form" onsubmit={(event) => { event.preventDefault(); void savePlaybackPolicy(); }}>
							<div class="settings-subsection__head settings-subsection__head--tight">
								<div>
									<h3>Playback policy</h3>
									<p>Pick how Xuva balances original quality and device compatibility.</p>
								</div>
							</div>
							<div class="playback-policy-options">
								{#each playbackPolicyOptions as option (option.id)}
									<label class="playback-policy-option">
										<input
											type="radio"
											name="playbackPolicy"
											value={option.id}
											checked={playbackPolicyDraft === option.id}
											onchange={() => (playbackPolicyDraft = option.id)}
										/>
										<div>
											<strong>{option.label}</strong>
											<span>{option.description}</span>
										</div>
									</label>
								{/each}
							</div>
							{#if playbackSaveMessage}
								<p class="settings-feedback">{playbackSaveMessage}</p>
							{/if}
						</form>
					{:else if !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as an admin to manage Xuva settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
							</div>
						</div>
					{/if}
					<ActivityListShell title="Playback Sessions">
						<XuvaActionList
							items={sessionItems}
							emptyLabel={sessionsUnavailable ? 'Sign in to view current playback sessions.' : 'No active playback sessions right now.'}
						/>
					</ActivityListShell>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'storage' || activeSection === 'transcoding'}
			<section id={activeSection} class="settings-section" data-testid="settings-section-content" data-section={activeSection}>
				<SettingsPanel
					title={activeSection === 'transcoding' ? 'Transcoding' : 'Storage'}
					description={activeSection === 'transcoding'
						? 'Choose where Xuva keeps transcoding files and whether compatibility work uses software or hardware mode.'
						: 'Choose where Xuva keeps library artwork, cache, short-lived files, and local app data.'}
					status={storagePanelStatus}
				>
					{#snippet actions()}
						{#if canManageSettings}
							{#if activeSection === 'transcoding'}
								<XuvaButton
									variant="primary"
									disabled={isSavingTranscoding || !hasTranscodingChanges}
									onclick={saveTranscodingSettings}
								>
									{isSavingTranscoding ? 'Saving...' : 'Save Transcoding Settings'}
								</XuvaButton>
							{:else}
								<XuvaButton
									variant="primary"
									disabled={isSavingStorage || !hasStorageChanges}
									onclick={saveStorageSettings}
								>
									{isSavingStorage ? 'Saving...' : 'Save Storage Settings'}
								</XuvaButton>
							{/if}
						{:else if !authDisabled}
							<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
						{/if}
					{/snippet}
					<div class="stat-grid stat-grid--compact">
						<XuvaStat
							label="Folders configured"
							value={asCount(storageConfiguredCount)}
							meta="Saved folders Xuva can use for local storage."
							tone={storageConfiguredCount > 0 ? 'good' : 'neutral'}
						/>
						<XuvaStat
							label="Needs attention"
							value={asCount(storageNeedsAttentionCount)}
							meta={storageNeedsAttentionCount > 0 ? 'One or more folders may need review before restart.' : 'All checked folders look ready.'}
							tone={storageNeedsAttentionCount > 0 ? 'warn' : 'good'}
						/>
					</div>
					{#if liveStatusError}
						<p class="settings-error">{liveStatusError}</p>
					{/if}
					{#if canManageSettings}
						<form class="storage-settings-form" data-testid="storage-form" onsubmit={(event) => { event.preventDefault(); void saveStorageSettings(); }}>
							<div class="settings-subsection">
								<div class="settings-subsection__head">
									<div>
										<h3>{activeSection === 'transcoding' ? 'Transcoding folders' : 'Media processing folders'}</h3>
										<p>{activeSection === 'transcoding'
											? 'Choose where Xuva keeps temporary files while preparing playback and saved optimized versions.'
											: 'Choose where Xuva keeps temporary processing files and optimized versions.'}</p>
									</div>
								</div>
								<div class="storage-field-list">
									{#each storageFields.filter((field) => field.group === 'processing') as field (field.key)}
										<article class="storage-field-card" data-testid="storage-field">
											<div class="storage-field-card__head">
												<div>
													<h4>{field.label}</h4>
													<p>{field.helper}</p>
												</div>
												<span class:storage-status-pill--warn={field.readinessTone === 'warn'} class:storage-status-pill--neutral={field.readinessTone === 'neutral'} class="storage-status-pill">
													{field.readinessLabel}
												</span>
											</div>
											<label class="settings-field">
												<span>{field.label}</span>
												<div class="storage-input-row">
													<input bind:value={storageDraft[field.key]} readonly={!field.editable} />
													{#if field.browseAvailable}
														<XuvaButton
															variant="secondary"
															size="sm"
															type="button"
															onclick={() => browseStorageField(field.key)}
														>
															{isBrowsingStorage && activeStorageBrowseField === field.key ? 'Loading folders...' : 'Browse'}
														</XuvaButton>
													{/if}
												</div>
												<small>{field.readinessDetail}</small>
											</label>
											<dl class="storage-field-facts">
												<div>
													<dt>Readiness</dt>
													<dd>{field.readinessLabel}</dd>
												</div>
												{#if field.capacityLabel}
													<div>
														<dt>Capacity</dt>
														<dd>{field.capacityLabel}</dd>
													</div>
												{/if}
											</dl>
										</article>
									{/each}
								</div>
							</div>
							{#if activeSection === 'storage'}
							<div class="settings-subsection">
								<div class="settings-subsection__head">
									<div>
										<h3>Library data folders</h3>
										<p>Choose where Xuva keeps artwork, cache, and other short-lived working files.</p>
									</div>
								</div>
								<div class="storage-field-list">
									{#each storageFields.filter((field) => field.group === 'library') as field (field.key)}
										<article class="storage-field-card" data-testid="storage-field">
											<div class="storage-field-card__head">
												<div>
													<h4>{field.label}</h4>
													<p>{field.helper}</p>
												</div>
												<span class:storage-status-pill--warn={field.readinessTone === 'warn'} class:storage-status-pill--neutral={field.readinessTone === 'neutral'} class="storage-status-pill">
													{field.readinessLabel}
												</span>
											</div>
											<label class="settings-field">
												<span>{field.label}</span>
												<div class="storage-input-row">
													<input bind:value={storageDraft[field.key]} readonly={!field.editable} />
													{#if field.browseAvailable}
														<XuvaButton
															variant="secondary"
															size="sm"
															type="button"
															onclick={() => browseStorageField(field.key)}
														>
															{isBrowsingStorage && activeStorageBrowseField === field.key ? 'Loading folders...' : 'Browse'}
														</XuvaButton>
													{/if}
												</div>
												<small>{field.readinessDetail}</small>
											</label>
											<dl class="storage-field-facts">
												<div>
													<dt>Readiness</dt>
													<dd>{field.readinessLabel}</dd>
												</div>
												{#if field.capacityLabel}
													<div>
														<dt>Capacity</dt>
														<dd>{field.capacityLabel}</dd>
													</div>
												{/if}
											</dl>
										</article>
									{/each}
								</div>
							</div>
							<div class="settings-subsection settings-subsection--quiet">
								<div class="settings-subsection__head">
									<div>
										<h3>App data folder</h3>
										<p>This folder stores Xuva's main settings and local data.</p>
									</div>
								</div>
								{#each storageFields.filter((field) => field.group === 'app') as field (field.key)}
									<article class="storage-field-card storage-field-card--readonly" data-testid="storage-field">
										<div class="storage-field-card__head">
											<div>
												<h4>{field.label}</h4>
												<p>{field.helper}</p>
											</div>
											<span class:storage-status-pill--warn={field.readinessTone === 'warn'} class:storage-status-pill--neutral={field.readinessTone === 'neutral'} class="storage-status-pill">
												{field.readinessLabel}
											</span>
										</div>
										<label class="settings-field">
											<span>{field.label}</span>
											<input bind:value={storageDraft[field.key]} readonly />
											<small>Changing this folder can move where Xuva expects its settings and local data after restart. It stays read-only in this build.</small>
										</label>
										<dl class="storage-field-facts">
											<div>
												<dt>Readiness</dt>
												<dd>{field.readinessLabel}</dd>
											</div>
											{#if field.capacityLabel}
												<div>
													<dt>Capacity</dt>
													<dd>{field.capacityLabel}</dd>
												</div>
											{/if}
										</dl>
									</article>
								{/each}
							</div>
							{/if}
							{#if activeSection === 'transcoding'}
								<div class="settings-subsection settings-subsection--quiet">
									<div class="settings-subsection__head">
										<div>
											<h3>Transcoding mode</h3>
											<p>Software mode always works. Hardware mode uses FFmpeg hardware encoders when available and unlocked.</p>
										</div>
									</div>
									<div class="settings-field">
										<span>Compatibility mode</span>
										<select bind:value={hardwareModeDraft}>
											<option value="software">Software transcoding</option>
											<option value="hardware">Hardware transcoding</option>
										</select>
										<small>
											Detected status: {asText(performance.hardwareAcceleration?.status) || 'unknown'}.
											{#if asText(performance.hardwareAcceleration?.unlockState)}
												Unlock state: {asText(performance.hardwareAcceleration?.unlockState)}.
											{/if}
										</small>
									</div>
									<div class="status-actions">
										<XuvaButton variant="secondary" size="sm" type="button" disabled={isTestingHardware} onclick={runTranscodingHardwareTest}>
											{isTestingHardware ? 'Testing hardware...' : 'Test Hardware Support'}
										</XuvaButton>
									</div>
									{#if hardwareTestResult}
										<p class="settings-note">
											Hardware test {asText(hardwareTestResult.status) || 'finished'}.
											Working encoders: {asCount(Number(hardwareTestResult.working || 0))} of {asCount(Number(hardwareTestResult.tested || 0))}.
										</p>
										{#if asText(hardwareTestResult.error)}
											<p class="settings-error">{asText(hardwareTestResult.error)}</p>
										{/if}
									{/if}
								</div>
								{#if transcodingSaveMessage}
									<p class="settings-feedback">{transcodingSaveMessage}</p>
								{/if}
								{#if transcodingSettingsError}
									<p class="settings-error">{transcodingSettingsError}</p>
								{/if}
							{:else}
								{#if storageSaveMessage}
									<p class="settings-feedback">{storageSaveMessage}</p>
								{/if}
								{#if storageSettingsError}
									<p class="settings-error">{storageSettingsError}</p>
								{/if}
							{/if}
						</form>
						<FolderBrowserPanel
							browser={storageFolderBrowse}
							title="Folder browser"
							subtitle={storageFolderBrowse?.path || (activeStorageBrowseFieldLabel ? `Select a folder for ${activeStorageBrowseFieldLabel.toLowerCase()}.` : 'Select a folder path.')}
							onBrowse={browseActiveStoragePath}
							onUsePath={useStorageBrowsePath}
						/>
					{:else if !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as an admin to manage Xuva settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
							</div>
						</div>
					{/if}
				</SettingsPanel>
			</section>

			{:else if activeSection === 'admin-access' || activeSection === 'pairing' || activeSection === 'approved-devices'}
			<section id={activeSection} class="settings-section" data-testid="settings-section-content" data-section={activeSection}>
				<SettingsPanel
					title={activeSection === 'pairing' ? 'Pairing' : activeSection === 'approved-devices' ? 'Approved Devices' : 'Users'}
					description={activeSection === 'pairing'
						? 'Approve or deny pending device pairing requests.'
						: activeSection === 'approved-devices'
							? 'Review and revoke persisted approved devices.'
						: 'User accounts and admin session state.'}
					status={user ? 'healthy' : 'idle'}
				>
					{#snippet actions()}
						{#if activeSection === 'admin-access' && canManageUsers}
							<XuvaButton variant="primary" href="#admin-access-add-user">Create User</XuvaButton>
						{:else if activeSection === 'admin-access' && canShowSignIn}
							<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
						{:else if activeSection === 'pairing'}
							<XuvaButton variant="primary" onclick={() => void loadSettingsSurface(false)}>Refresh Pairing</XuvaButton>
						{:else if activeSection === 'approved-devices'}
							<XuvaButton variant="primary" onclick={() => void loadSettingsSurface(false)}>Refresh Devices</XuvaButton>
						{/if}
					{/snippet}
					{#if activeSection === 'admin-access'}
						<div class="stat-grid stat-grid--compact users-overview-grid">
							<XuvaStat label="Account" value={accessAccountValue} meta={accessAccountMeta} tone={user ? 'good' : 'neutral'} />
							<XuvaStat label="Session" value={accessSessionValue} meta={accessSessionMeta} tone={user ? 'good' : 'neutral'} />
						</div>
						{#if user && !authDisabled}
							<div class="settings-subsection users-subsection">
								<div class="settings-subsection__head">
									<div>
										<h3>Your account</h3>
										<p>Signed in as {userDisplayName}.</p>
									</div>
									<span class="storage-status-pill storage-status-pill--neutral">{currentUserRoleLabel}</span>
								</div>
								<form class="settings-stack users-password-form" onsubmit={(event) => { event.preventDefault(); void changeCurrentPassword(); }}>
									<label class="settings-field">
										<span>New password</span>
										<input type="password" bind:value={passwordDraft} minlength="10" autocomplete="new-password" />
									</label>
									<label class="settings-field">
										<span>Confirm password</span>
										<input type="password" bind:value={confirmPasswordDraft} minlength="10" autocomplete="new-password" />
									</label>
									<div class="status-actions">
										<XuvaButton
											variant="primary"
											disabled={isSavingPassword || !canSubmitPasswordChange}
											onclick={changeCurrentPassword}
										>
											{isSavingPassword ? 'Updating...' : 'Change Password'}
										</XuvaButton>
										<XuvaButton variant="secondary" disabled={isSigningOut} onclick={signOut}>
											{isSigningOut ? 'Signing out...' : 'Sign Out'}
										</XuvaButton>
									</div>
									<p class="settings-note">Passwords require at least 10 characters.</p>
								</form>
							</div>
						{/if}
						{#if canManageUsers}
							<div class="settings-subsection users-subsection" data-testid="users-account-list">
								<div class="settings-subsection__head">
									<div>
										<h3>User accounts</h3>
										<p>{usersCountLabel} with access to this server.</p>
									</div>
								</div>
								<p class="settings-note">Roles: Admin can manage server settings and accounts. User can browse and play media.</p>
								{#if users.length > 0}
									<div class="pairing-request-list approved-device-list users-account-list">
										{#each users as account (account.id)}
											<article class="pairing-request-card approved-device-card users-account-card">
												<div class="pairing-request-card__head">
													<div>
														<h4>{asText(account.displayName) || asText(account.username) || 'User'}</h4>
														<p>@{asText(account.username)}</p>
													</div>
													<span class="storage-status-pill storage-status-pill--neutral">{accountTypeLabel(account.role)}</span>
												</div>
												{#if asText(account.createdAt)}
													<p class="review-card__reason">Created {formatDateTime(account.createdAt)}</p>
												{/if}
												{#if asText(account.id) !== asText(user?.id)}
													<div class="status-actions">
														<XuvaButton
															variant="danger"
															size="sm"
															disabled={isDeletingUser && activeUserID === asText(account.id)}
															onclick={() => removeUserAccount(account)}
														>
															{isDeletingUser && activeUserID === asText(account.id) ? 'Removing...' : 'Remove'}
														</XuvaButton>
													</div>
												{/if}
											</article>
										{/each}
									</div>
								{:else}
									<p class="settings-note">No user accounts found yet.</p>
								{/if}
							</div>
							<div id="admin-access-add-user" class="settings-subsection users-subsection">
								<div class="settings-subsection__head">
									<div>
										<h3>Add user</h3>
										<p>Create another account for shared viewing.</p>
									</div>
								</div>
								<form class="settings-stack users-create-form" onsubmit={(event) => { event.preventDefault(); void createUserAccount(); }}>
									<label class="settings-field">
										<span>Display name</span>
										<input bind:value={newUserDisplayNameDraft} maxlength="80" placeholder="Family Room" />
									</label>
									<label class="settings-field">
										<span>Username</span>
										<input bind:value={newUserUsernameDraft} maxlength="40" required />
									</label>
									<label class="settings-field">
										<span>Password</span>
										<input type="password" bind:value={newUserPasswordDraft} minlength="10" autocomplete="new-password" required />
									</label>
									<label class="settings-field">
										<span>Role</span>
										<select bind:value={newUserRoleDraft}>
											<option value="standard">User</option>
											<option value="admin">Admin</option>
										</select>
									</label>
									<div class="status-actions">
										<XuvaButton
											variant="primary"
											disabled={isCreatingUser || !canSubmitNewUser}
											onclick={createUserAccount}
										>
											{isCreatingUser ? 'Creating...' : 'Create User'}
										</XuvaButton>
									</div>
								</form>
							</div>
						{:else if !authDisabled}
							<div class="settings-auth-callout">
								<p class="settings-note">Sign in with your admin account to view and manage user accounts.</p>
								<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
								<div class="status-actions">
									<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
								</div>
							</div>
						{/if}
						{#if usersActionMessage}
							<p class="settings-feedback">{usersActionMessage}</p>
						{/if}
						{#if usersError}
							<p class="settings-error">{usersError}</p>
						{/if}
					{:else if activeSection === 'pairing'}
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Device Pairing</h3>
								<p>Approve devices that ask to connect to this Xuva server.</p>
							</div>
						</div>
						<p class="settings-note">Devices still need admin approval before they can connect.</p>
						{#if canManageSettings || authDisabled}
							{#if isLoadingPairingRequests}
								<p class="settings-note">Loading pairing requests...</p>
							{:else if pairingRequests.length > 0}
								<div class="pairing-request-list" data-testid="pairing-request-list">
									{#each pairingRequests as request (request.id)}
										<article class="pairing-request-card">
											<div class="pairing-request-card__head">
												<div>
													<h4>{pairingDeviceName(request)}</h4>
													<p>{pairingProfileLabel(request.clientProfile)}</p>
												</div>
												<span
													class="storage-status-pill"
													class:storage-status-pill--warn={asText(request.status).toLowerCase() === 'expired'}
													class:storage-status-pill--neutral={asText(request.status).toLowerCase() !== 'pending' && asText(request.status).toLowerCase() !== 'expired'}
												>
													{humanStatus(request.status)}
												</span>
											</div>
											<p class="review-card__reason">{pairingStatusSummary(request)}</p>
											<dl class="pairing-request-facts">
												{#if asText(request.code)}
													<div>
														<dt>Pairing code</dt>
														<dd>{asText(request.code)}</dd>
													</div>
												{/if}
												{#if asText(request.expiresAt)}
													<div>
														<dt>Expires</dt>
														<dd>{formatDateTime(request.expiresAt)}</dd>
													</div>
												{/if}
												{#if asText(request.updatedAt) && asText(request.status).toLowerCase() !== 'pending'}
													<div>
														<dt>Updated</dt>
														<dd>{formatDateTime(request.updatedAt)}</dd>
													</div>
												{/if}
											</dl>
											{#if canUpdatePairingRequest(request)}
												<div class="status-actions">
													<XuvaButton
														variant="primary"
														size="sm"
														disabled={activePairingRequestID === asText(request.id)}
														onclick={() => updatePairingRequest(request, 'approve')}
													>
														{activePairingRequestID === asText(request.id) && activePairingActionKind === 'approve' ? 'Approving...' : 'Approve'}
													</XuvaButton>
													<XuvaButton
														variant="danger"
														size="sm"
														disabled={activePairingRequestID === asText(request.id)}
														onclick={() => updatePairingRequest(request, 'deny')}
													>
														{activePairingRequestID === asText(request.id) && activePairingActionKind === 'deny' ? 'Denying...' : 'Deny'}
													</XuvaButton>
												</div>
											{/if}
										</article>
									{/each}
								</div>
							{:else}
								<p class="settings-note">No pairing requests right now.</p>
							{/if}
							{#if pairingActionMessage}
								<p class="settings-feedback">{pairingActionMessage}</p>
							{/if}
							{#if pairingRequestsError}
								<p class="settings-error">{pairingRequestsError}</p>
							{/if}
						{:else}
							<div class="settings-auth-callout">
								<p class="settings-note">Sign in as an admin to update device pairing.</p>
								<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
								<div class="status-actions">
									<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
								</div>
							</div>
						{/if}
					</div>
					{:else if activeSection === 'approved-devices'}
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Approved Devices</h3>
								<p>Approved devices can connect to this Xuva server.</p>
							</div>
						</div>
						<p class="settings-note">Recent activity is not tracked yet.</p>
						{#if canManageSettings || authDisabled}
							{#if isLoadingApprovedDevices}
								<p class="settings-note">Loading approved devices...</p>
							{:else if approvedDevices.length > 0}
								<div class="pairing-request-list approved-device-list" data-testid="approved-device-list">
									{#each approvedDevices as device (device.id)}
										<article class="pairing-request-card approved-device-card">
											<div class="pairing-request-card__head">
												<div>
													<h4>{approvedDeviceName(device)}</h4>
													<p>{pairingProfileLabel(device.clientProfile)}</p>
												</div>
												<span class="storage-status-pill">Approved</span>
											</div>
											<p class="review-card__reason">{approvedDeviceSummary(device)}</p>
											<dl class="pairing-request-facts">
												{#if asText(device.approvedAt)}
													<div>
														<dt>Approved</dt>
														<dd>{formatDateTime(device.approvedAt)}</dd>
													</div>
												{/if}
												{#if asText(device.updatedAt)}
													<div>
														<dt>Updated</dt>
														<dd>{formatDateTime(device.updatedAt)}</dd>
													</div>
												{/if}
											</dl>
											{#if canManageSettings}
												<div class="status-actions">
													<XuvaButton
														variant="danger"
														size="sm"
														disabled={activeApprovedDeviceID === asText(device.id)}
														onclick={() => revokeApprovedDeviceEntry(device)}
													>
														{activeApprovedDeviceID === asText(device.id) ? 'Removing...' : 'Remove'}
													</XuvaButton>
												</div>
											{/if}
										</article>
									{/each}
								</div>
							{:else}
								<p class="settings-note">No approved devices yet.</p>
							{/if}
							{#if approvedDevicesActionMessage}
								<p class="settings-feedback">{approvedDevicesActionMessage}</p>
							{/if}
							{#if approvedDevicesError}
								<p class="settings-error">{approvedDevicesError}</p>
							{/if}
						{:else}
							<div class="settings-auth-callout">
								<p class="settings-note">Sign in as an admin to manage approved devices.</p>
								<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
								<div class="status-actions">
									<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
								</div>
							</div>
						{/if}
					</div>
					{/if}
				</SettingsPanel>
			</section>

			{:else if activeSection === 'planned-tools'}
			<section id="planned-tools" class="settings-section" data-testid="settings-section-content" data-section="planned-tools">
				<SettingsPanel
					title="Planned (Placeholder)"
					description="Deferred items that are intentionally placeholder-only."
					status="healthy"
				>
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Placeholder items</h3>
								<p>These items are parked until final scope and backend behavior are confirmed.</p>
							</div>
						</div>
						<div class="placeholder-grid">
							<article class="placeholder-item">
								<strong>Live TV</strong>
								<p>Placeholder only. Channel, tuner, and guide workflows are not implemented yet.</p>
							</article>
							<article class="placeholder-item">
								<strong>Database maintenance</strong>
								<p>Placeholder only. Cleanup and repair actions are not implemented yet.</p>
							</article>
							<article class="placeholder-item">
								<strong>Scheduled tasks</strong>
								<p>Placeholder only. Extended task scheduling is not implemented yet.</p>
							</article>
							<article class="placeholder-item">
								<strong>Server logs</strong>
								<p>Placeholder only. A user-facing log viewer is not implemented yet.</p>
							</article>
							<article class="placeholder-item">
								<strong>Plugins</strong>
								<p>Placeholder only. Plugin lifecycle controls are not implemented yet.</p>
							</article>
							<article class="placeholder-item">
								<strong>Integration credentials</strong>
								<p>Placeholder only. Client credential management is not implemented yet.</p>
							</article>
						</div>
					</div>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'general' || activeSection === 'network' || activeSection === 'discovery' || activeSection === 'about'}
			<section id={activeSection} class="settings-section" data-testid="settings-section-content" data-section={activeSection}>
				<SettingsPanel
					title={activeSection === 'general' ? 'General' : activeSection === 'network' ? 'Network' : activeSection === 'discovery' ? 'Discovery' : 'About'}
					description={activeSection === 'general'
						? 'Server identity and admin-facing basics.'
						: activeSection === 'network'
							? 'Local discovery status and network visibility.'
							: activeSection === 'discovery'
								? 'mDNS / Bonjour service identity and status.'
								: 'Xuva identity and build details.'}
					status={discoveryStatusError ? 'warning' : 'healthy'}
				>
					{#snippet actions()}
						{#if activeSection === 'general' && canManageSettings}
							<XuvaButton variant="primary" disabled={isSavingServerName} onclick={saveServerName}>
								{isSavingServerName ? 'Saving...' : 'Save Server Name'}
							</XuvaButton>
						{:else if activeSection === 'general' && !authDisabled}
							<XuvaButton variant="primary" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
						{:else if activeSection === 'network'}
							<XuvaButton variant="primary" href="#discovery">Open Discovery</XuvaButton>
						{:else if activeSection === 'discovery'}
							<XuvaButton variant="primary" onclick={() => void loadSettingsSurface(false)}>Refresh Status</XuvaButton>
						{:else if activeSection === 'about'}
							<XuvaButton variant="primary" href="#general">Open General</XuvaButton>
						{/if}
					{/snippet}
					<div class="stat-grid stat-grid--compact">
						<XuvaStat label="App" value="Xuva" meta="Local-first personal media library." />
						<XuvaStat label="Server Name" value={serverDisplayName} meta="Shown in the browser title and advertised on your home network when local discovery is running." />
						<XuvaStat label="Build" value={asText(buildInfo?.buildID) || 'Local build'} meta={asText(buildInfo?.publishedAt) || 'Build details are quiet in this view.'} />
						<XuvaStat label="Mode" value="Local" meta="No cloud account or vendor relay required." />
					</div>
					{#if activeSection === 'general'}
					<div class="settings-subsection settings-subsection--quiet">
						<div class="settings-subsection__head">
							<div>
								<h3>Server identity</h3>
								<p>{serverIdentityHelpText}</p>
							</div>
						</div>
					</div>
					{/if}
					{#if activeSection === 'network' || activeSection === 'discovery'}
					<div class="settings-subsection" data-testid="discovery-status-card">
						<div class="settings-subsection__head">
							<div>
								<h3>Local discovery</h3>
								<p>{discoveryStatusMessage}</p>
							</div>
							<span class={`status-pill status-pill--${discoveryStatusTone}`}>{discoveryStatusLabel}</span>
						</div>
						<div class="stat-grid stat-grid--compact">
							<XuvaStat
								label="Service name"
								value={discoveryServiceName}
								meta="Uses your configured Server Name."
							/>
							<XuvaStat
								label="Protocol"
								value="mDNS / Bonjour"
								meta={asText(discoveryStatus.serviceType) || '_xuva._tcp.local.'}
							/>
						</div>
						{#if discoveryStatusError}
							<p class="settings-error">{discoveryStatusError}</p>
						{/if}
						<p class:status-copy={true} class:status-copy--warn={discoveryStatusTone === 'warn'}>
							{discoveryStatusDetail}
						</p>
					</div>
					{/if}
					{#if activeSection === 'general' && canManageSettings}
						<form class="server-name-form" data-testid="server-name-form" onsubmit={(event) => { event.preventDefault(); void saveServerName(); }}>
							<label class="settings-field">
								<span>Server name</span>
								<input
									bind:value={serverNameDraft}
									maxlength="50"
									required
									placeholder="Xuva"
									aria-describedby="settings-server-name-help"
								/>
								<small id="settings-server-name-help">{serverIdentityHelpText}</small>
							</label>
							<div class="status-actions">
								<XuvaButton variant="primary" disabled={isSavingServerName} onclick={saveServerName}>
									{isSavingServerName ? 'Saving...' : 'Save Server Name'}
								</XuvaButton>
							</div>
							{#if serverNameSaveMessage}
								<p class="settings-feedback">{serverNameSaveMessage}</p>
							{/if}
							{#if serverNameError}
								<p class="settings-error">{serverNameError}</p>
							{/if}
						</form>
					{:else if activeSection === 'general' && !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as an admin to manage Xuva settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<XuvaButton variant="primary" size="sm" href={ownerActionHref}>{ownerActionLabel}</XuvaButton>
							</div>
						</div>
					{/if}
				</SettingsPanel>
			</section>
			{/if}
		{/if}
	</div>
</ServerShell>

<style>
	.settings-page {
		--settings-accent: var(--xuva-settings-accent, #6a74d9);
		--settings-accent-soft: var(--xuva-settings-accent-soft, rgb(106 116 217 / 10%));
		--settings-accent-border: var(--xuva-settings-accent-border, rgb(106 116 217 / 24%));
		--settings-surface: color-mix(in srgb, var(--xuva-color-bg-panel, #f5f7fb) 92%, white 8%);
		--settings-surface-elevated: color-mix(in srgb, var(--xuva-color-bg-panel-elevated, #fff) 96%, #eef2f8 4%);
		--settings-divider: color-mix(in srgb, var(--settings-accent-border) 32%, var(--xuva-color-border-soft));
		--settings-input-bg: color-mix(in srgb, var(--settings-surface-elevated) 96%, #e9eef7 4%);
		--settings-input-readonly-bg: color-mix(in srgb, var(--settings-surface) 92%, #e6ecf6 8%);
		display: grid;
		gap: 20px;
		padding-bottom: var(--xuva-space-8);
		min-width: 0;
		scroll-behavior: smooth;
	}

	.settings-head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
	}

	.settings-head h1 {
		margin: 0;
		font-family: var(--xuva-font-display);
		font-size: clamp(1.8rem, 1.6vw + 1rem, 2.45rem);
		letter-spacing: -0.03em;
	}

	.settings-head__eyebrow {
		margin: 0 0 7px;
		color: color-mix(in srgb, var(--settings-accent) 72%, white 28%);
		font-size: 0.72rem;
		font-weight: 800;
		letter-spacing: 0.12em;
		text-transform: uppercase;
	}

	.settings-head__server-name {
		display: block;
		margin-bottom: 4px;
		color: var(--xuva-color-text);
		font-size: clamp(1.55rem, 1.2vw + 1rem, 2rem);
		font-weight: 820;
		letter-spacing: -0.02em;
		line-height: 1.08;
	}

	.settings-head p {
		margin: 6px 0 0;
		color: color-mix(in srgb, var(--xuva-color-text-muted) 84%, transparent);
		font-size: 0.9rem;
		line-height: 1.42;
		max-width: 860px;
	}

	.settings-head__meta {
		color: color-mix(in srgb, var(--xuva-color-text-soft) 90%, transparent);
		font-size: 0.8rem;
		white-space: nowrap;
	}

	.settings-dashboard {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 0 16px;
	}

	.settings-dashboard-card {
		display: grid;
		gap: 7px;
		min-height: 0;
		align-content: start;
		padding: 11px 0;
		border-top: 1px solid var(--settings-divider);
		border-radius: 0;
		background: transparent;
		color: var(--xuva-color-text);
	}

	.settings-dashboard > :first-child {
		border-top: 0;
		padding-top: 0;
	}

	.settings-dashboard-card span {
		color: color-mix(in srgb, var(--xuva-color-text-muted) 92%, transparent);
		font-size: 0.78rem;
		font-weight: 760;
		letter-spacing: 0.07em;
		text-transform: uppercase;
	}

	.settings-dashboard-card strong {
		overflow: hidden;
		text-overflow: ellipsis;
		color: var(--xuva-color-text);
		font-size: clamp(1.02rem, 0.36vw + 0.95rem, 1.28rem);
		font-weight: 820;
		line-height: 1.2;
		overflow-wrap: anywhere;
	}

	.settings-dashboard-card small {
		color: var(--xuva-color-text-soft);
		font-size: 0.78rem;
		line-height: 1.35;
	}

	.settings-dashboard-card--identity {
		grid-column: span 2;
	}

	.settings-dashboard-card--quiet {
		min-height: 0;
	}

	.dashboard-card-actions {
		display: flex;
		flex-wrap: wrap;
		gap: 7px;
		align-items: center;
	}

	.dashboard-warning-list {
		display: grid;
		gap: 4px;
		margin: 0;
		padding: 0;
		list-style: none;
		color: var(--xuva-color-text-muted);
		font-size: 0.8rem;
		line-height: 1.35;
	}

	.settings-section {
		scroll-margin-top: 18px;
	}

	.settings-subsection {
		display: grid;
		gap: 14px;
		padding: 12px 0;
		border-top: 1px solid var(--settings-divider);
		border-radius: 0;
		background: transparent;
	}

	.settings-subsection--quiet {
		background: transparent;
	}

	.settings-subsection--empty {
		align-content: center;
		min-height: 140px;
	}

	.settings-subsection__head {
		display: grid;
		gap: 4px;
	}

	.settings-subsection__head--tight {
		margin-bottom: -2px;
	}

	.settings-subsection__head h3 {
		margin: 0;
		font-size: 0.98rem;
		font-weight: 760;
	}

	.settings-subsection__head p {
		margin: 0;
		color: var(--xuva-color-text-soft);
		font-size: 0.84rem;
		line-height: 1.45;
	}

	:global([data-shell='server'] .settings-panel) {
		border-left: 0;
	}

	:global([data-shell='server'] .settings-panel h2) {
		font-size: 1.13rem;
	}

	.stat-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 10px;
	}

	.stat-grid--compact {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.status-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--xuva-space-2);
	}

	.settings-feedback {
		margin: 0;
		color: color-mix(in srgb, #b7ffd4 78%, white 22%);
		font-size: 0.82rem;
		line-height: 1.42;
	}

	.library-list {
		display: grid;
		gap: 0;
	}

	.library-card {
		display: grid;
		gap: 12px;
		padding: 14px 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.library-list > :first-child {
		padding-top: 0;
		border-top: 0;
	}

	.library-card__body {
		display: grid;
		gap: 10px;
		min-width: 0;
	}

	.library-card__heading {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
	}

	.library-card__heading h3 {
		margin: 0;
		font-size: 1rem;
		font-weight: 760;
	}

	.library-card__heading p,
	.library-card__note,
	.settings-note {
		margin: 0;
		color: var(--xuva-color-text-soft);
		font-size: 0.82rem;
		line-height: 1.4;
	}

	.settings-note--inline {
		margin-bottom: 4px;
	}

	.library-card__status {
		display: inline-flex;
		align-items: center;
		padding: 4px 8px;
		border-radius: 4px;
		background: rgb(154 167 255 / 10%);
		color: color-mix(in srgb, var(--settings-accent) 72%, white 28%);
		font-size: 0.74rem;
		font-weight: 760;
	}

	.library-card__facts {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 10px;
		margin: 0;
	}

	.library-card__facts div {
		display: grid;
		gap: 4px;
		min-width: 0;
	}

	.library-card__facts dt {
		color: var(--xuva-color-text-muted);
		font-size: 0.76rem;
		font-weight: 700;
	}

	.library-card__facts dd {
		margin: 0;
		color: var(--xuva-color-text);
		font-size: 0.88rem;
		line-height: 1.35;
		overflow-wrap: anywhere;
	}

	.library-card__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		align-items: center;
	}

	.settings-auth-callout {
		display: grid;
		gap: 8px;
		padding: 10px 0 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.settings-auth-callout__detail {
		margin: 0;
		color: var(--xuva-color-text-muted);
		font-size: 0.88rem;
		line-height: 1.45;
	}

	.scanning-automation-form,
	.metadata-source-form,
	.playback-policy-form,
	.storage-settings-form {
		display: grid;
		gap: 12px;
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid var(--settings-divider);
	}

	.settings-advanced {
		padding: 12px 0 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.settings-advanced summary {
		cursor: pointer;
		color: var(--xuva-color-text);
		font-size: 0.9rem;
		font-weight: 760;
	}

	.settings-advanced__content {
		display: grid;
		gap: 10px;
		margin-top: 12px;
	}

	.playback-policy-options {
		display: grid;
		gap: 0;
	}

	.metadata-source-groups {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 12px;
	}

	.metadata-source-group {
		display: grid;
		gap: 10px;
	}

	.metadata-source-list {
		display: grid;
		gap: 0;
	}

	.metadata-source-card {
		display: grid;
		gap: 10px;
		padding: 12px 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.metadata-source-card--disabled {
		background: linear-gradient(90deg, rgb(255 255 255 / 1%), transparent);
	}

	.metadata-source-list > :first-child {
		padding-top: 0;
		border-top: 0;
	}

	.metadata-source-card__main {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
	}

	.metadata-source-card__toggle {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		gap: 10px;
		align-items: flex-start;
		min-width: 0;
	}

	.metadata-source-card__toggle input {
		margin-top: 3px;
	}

	.metadata-source-card__toggle div {
		display: grid;
		gap: 4px;
	}

	.metadata-source-card__toggle strong {
		font-size: 0.92rem;
		font-weight: 760;
	}

	.metadata-source-card__toggle span,
	.metadata-source-card__meta small {
		color: var(--xuva-color-text-soft);
		font-size: 0.8rem;
		line-height: 1.4;
	}

	.metadata-source-card__meta {
		display: grid;
		gap: 4px;
		justify-items: end;
		text-align: right;
	}

	.metadata-source-card__actions {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		justify-content: flex-end;
	}

	.status-pill {
		display: inline-flex;
		align-items: center;
		padding: 4px 8px;
		border-radius: 4px;
		background: rgb(154 167 255 / 10%);
		color: color-mix(in srgb, var(--settings-accent) 72%, white 28%);
		font-size: 0.74rem;
		font-weight: 760;
	}

	.status-pill--good {
		background: rgb(65 143 101 / 18%);
		color: color-mix(in srgb, #b7ffd4 75%, white 25%);
	}

	.status-pill--warn {
		background: rgb(255 178 102 / 15%);
		color: color-mix(in srgb, #ffc897 78%, white 22%);
	}

	.status-pill--neutral {
		background: rgb(154 167 255 / 10%);
		color: color-mix(in srgb, var(--settings-accent) 72%, white 28%);
	}

	.status-copy {
		margin: 0;
		font-size: 0.82rem;
		line-height: 1.4;
	}

	.status-copy--warn {
		color: var(--xuva-color-danger, #ff9f9f);
	}

	.playback-policy-option {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: flex-start;
		gap: 10px;
		padding: 12px 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.playback-policy-options > :first-child {
		padding-top: 0;
		border-top: 0;
	}

	.playback-policy-option input {
		margin-top: 3px;
	}

	.playback-policy-option div {
		display: grid;
		gap: 4px;
	}

	.playback-policy-option strong {
		font-size: 0.94rem;
		font-weight: 760;
	}

	.playback-policy-option span {
		color: var(--xuva-color-text-soft);
		font-size: 0.82rem;
		line-height: 1.4;
	}

	.metadata-review-list,
	.metadata-version-groups,
	.review-records,
	.pairing-request-list {
		display: grid;
		gap: 0;
	}

	.review-card,
	.review-record-card,
	.version-group-card,
	.pairing-request-card {
		display: grid;
		gap: 10px;
		padding: 12px 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.metadata-review-list > :first-child,
	.metadata-version-groups > :first-child,
	.review-records > :first-child,
	.pairing-request-list > :first-child {
		padding-top: 0;
		border-top: 0;
	}

	.review-card__head,
	.review-record-card__head,
	.version-group-card,
	.pairing-request-card__head {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
	}

	.review-card__head h4,
	.review-record-card__head strong,
	.pairing-request-card__head h4 {
		margin: 0;
		font-size: 0.94rem;
		font-weight: 760;
	}

	.review-card__head p,
	.review-record-card__head span,
	.version-group-card span,
	.pairing-request-card__head p {
		margin: 4px 0 0;
		color: var(--xuva-color-text-soft);
		font-size: 0.8rem;
		line-height: 1.4;
	}

	.review-card__reason,
	.review-record-card p,
	.pairing-request-card > p {
		margin: 0;
		color: var(--xuva-color-text-soft);
		font-size: 0.82rem;
		line-height: 1.45;
	}

	.pairing-request-facts {
		display: grid;
		grid-template-columns: repeat(auto-fit, minmax(120px, 1fr));
		gap: 10px;
		margin: 0;
	}

	.pairing-request-facts div {
		display: grid;
		gap: 4px;
	}

	.pairing-request-facts dt {
		color: var(--xuva-color-text-muted);
		font-size: 0.72rem;
		font-weight: 700;
		letter-spacing: 0;
		text-transform: uppercase;
	}

	.pairing-request-facts dd {
		margin: 0;
		color: var(--xuva-color-text);
		font-size: 0.82rem;
	}

	.review-card__details {
		display: grid;
		gap: 12px;
		padding-top: 12px;
		border-top: 1px solid var(--settings-divider);
	}

	.review-manual-form {
		display: grid;
		gap: 10px;
		padding: 12px 0 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.review-manual-form__fields {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}

	@media (max-width: 1120px) {
		.settings-dashboard {
			grid-template-columns: 1fr;
		}

		.library-card__facts {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.metadata-source-groups {
			grid-template-columns: 1fr;
		}
	}

	.server-name-form {
		display: grid;
		gap: 12px;
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid var(--settings-divider);
	}

	.settings-field {
		display: grid;
		gap: 6px;
	}

	.settings-field span {
		color: var(--xuva-color-text-muted);
		font-size: 0.8rem;
		font-weight: 700;
	}

	.settings-field input {
		min-height: 42px;
		border: 1px solid var(--settings-divider);
		border-radius: 4px;
		background: var(--settings-input-bg);
		color: var(--xuva-color-text);
		font: inherit;
		padding: 0 12px;
		width: 100%;
	}

	.settings-field select {
		min-height: 42px;
		border: 1px solid var(--settings-divider);
		border-radius: 4px;
		background: var(--settings-input-bg);
		color: var(--xuva-color-text);
		font: inherit;
		padding: 0 12px;
		width: 100%;
	}

	.settings-field input[readonly] {
		background: var(--settings-input-readonly-bg);
		color: color-mix(in srgb, var(--xuva-color-text) 86%, transparent);
	}

	.settings-stack {
		display: grid;
		gap: 10px;
		padding-top: 10px;
		border-top: 1px solid var(--settings-divider);
	}

	.placeholder-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
	}

	.placeholder-item {
		display: grid;
		gap: 6px;
		padding: 10px;
		border: 1px solid var(--settings-divider);
		border-radius: 6px;
		background: var(--settings-surface-elevated);
	}

	.placeholder-item strong {
		font-size: 0.9rem;
	}

	.placeholder-item p {
		margin: 0;
		color: var(--xuva-color-text-soft);
		font-size: 0.82rem;
		line-height: 1.4;
	}

	.users-overview-grid {
		margin-bottom: 4px;
	}

	.users-subsection {
		gap: 10px;
	}

	.users-account-list {
		display: grid;
		gap: 8px;
	}

	.users-account-card {
		padding-top: 10px;
		padding-bottom: 10px;
	}

	.users-password-form,
	.users-create-form {
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px 12px;
	}

	.users-password-form .status-actions,
	.users-password-form .settings-note,
	.users-create-form .status-actions {
		grid-column: 1 / -1;
	}

	.users-password-form .settings-note {
		margin-top: -2px;
	}

	.storage-field-list {
		display: grid;
		gap: 0;
	}

	.storage-field-card {
		display: grid;
		gap: 12px;
		padding: 14px 0;
		border-top: 1px solid var(--settings-divider);
		background: transparent;
	}

	.storage-field-card--readonly {
		background: linear-gradient(90deg, rgb(255 255 255 / 1%), transparent);
	}

	.storage-field-list > :first-child,
	.settings-subsection--quiet > .storage-field-card:first-of-type {
		padding-top: 0;
		border-top: 0;
	}

	.storage-field-card__head {
		display: flex;
		flex-wrap: wrap;
		align-items: flex-start;
		justify-content: space-between;
		gap: 10px;
	}

	.storage-field-card__head h4 {
		margin: 0;
		font-size: 0.98rem;
		font-weight: 760;
	}

	.storage-field-card__head p {
		margin: 4px 0 0;
		color: var(--xuva-color-text-soft);
		font-size: 0.82rem;
		line-height: 1.4;
	}

	.storage-status-pill {
		display: inline-flex;
		align-items: center;
		padding: 4px 8px;
		border-radius: 4px;
		background: rgb(65 143 101 / 18%);
		color: color-mix(in srgb, #b7ffd4 75%, white 25%);
		font-size: 0.74rem;
		font-weight: 760;
	}

	.storage-status-pill--warn {
		background: rgb(255 178 102 / 15%);
		color: color-mix(in srgb, #ffc897 78%, white 22%);
	}

	.storage-status-pill--neutral {
		background: rgb(154 167 255 / 10%);
		color: color-mix(in srgb, var(--settings-accent) 72%, white 28%);
	}

	.storage-input-row {
		display: flex;
		flex-wrap: wrap;
		gap: 8px;
		align-items: center;
	}

	.storage-input-row :global(.button) {
		flex: 0 0 auto;
	}

	.storage-input-row input {
		flex: 1 1 280px;
		min-width: 0;
	}

	.storage-field-facts {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 10px;
		margin: 0;
	}

	.storage-field-facts div {
		display: grid;
		gap: 3px;
		min-width: 0;
	}

	.storage-field-facts dt {
		color: var(--xuva-color-text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.storage-field-facts dd {
		margin: 0;
		color: var(--xuva-color-text);
		font-size: 0.84rem;
		line-height: 1.38;
		overflow-wrap: anywhere;
	}

	.settings-field small,
	.settings-error {
		margin: 0;
		font-size: 0.82rem;
		line-height: 1.38;
	}

	.settings-field small {
		color: var(--xuva-color-text-soft);
	}

	.settings-error {
		color: var(--xuva-color-danger, #ff9f9f);
	}

	@media (max-width: 820px) {
		.settings-dashboard {
			grid-template-columns: 1fr;
		}

		.settings-dashboard-card--identity {
			grid-column: span 1;
		}

		.settings-head {
			flex-direction: column;
			align-items: flex-start;
		}

		.settings-head__meta {
			white-space: normal;
		}

		.stat-grid,
		.stat-grid--compact {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.library-card__facts {
			grid-template-columns: 1fr;
		}

		.storage-field-facts {
			grid-template-columns: 1fr;
		}

		.metadata-source-card__main {
			display: grid;
		}

		.metadata-source-card__meta,
		.metadata-source-card__actions {
			justify-items: start;
			justify-content: flex-start;
			text-align: left;
		}

		.settings-subsection {
			padding-bottom: 13px;
		}

		.users-password-form,
		.users-create-form {
			grid-template-columns: 1fr;
		}

		.placeholder-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 560px) {
		.settings-dashboard {
			grid-template-columns: 1fr;
		}

		.settings-dashboard-card--identity {
			grid-column: span 1;
		}

		.stat-grid,
		.stat-grid--compact {
			grid-template-columns: 1fr;
		}

		.review-manual-form__fields {
			grid-template-columns: 1fr;
		}
	}
</style>
