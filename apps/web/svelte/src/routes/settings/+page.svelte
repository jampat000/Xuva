	<script lang="ts">
	import { onMount } from 'svelte';
	import { scanMovies, scanTV, refreshMetadataBatch } from '$lib/api/browse';
	import {
		getAuthSession,
		getClientBootstrap,
		logout,
		type AuthSessionResponse,
		type AuthSessionUser,
		type ClientBootstrapResponse
	} from '$lib/api/auth';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import { browseFolder, deleteLibrary, startLibraryScan, type FolderBrowseResponse } from '$lib/api/setup';
	import {
		getCatalogHealth,
		getCatalogSummary,
		getDownloads,
		getPerformanceSettings,
		getProbes,
		getScans,
		getSessions,
		getSettings,
		getSystemStatus,
		getWork,
		updateSettings,
		type CatalogHealthResponse,
		type CatalogSummaryResponse,
		type DownloadJobItem,
		type PerformanceSettingsResponse,
		type ProbeJobItem,
		type ScanJobItem,
		type SessionItem,
		type SettingsResponse,
		type SystemStatusResponse,
		type UpdateSettingsRequest,
		type WorkQueueItem
	} from '$lib/api/operator';
	import { createEventStream } from '$lib/events/stream';
	import { lorivoTitle, normalizeServerName } from '$lib/server-name';
	import {
		ActivityListShell,
		ServerShell,
		LorivoActionList,
		LorivoButton,
		LorivoPanel,
		LorivoStat
	} from '$lib/components';
	import SettingsPanel from '$lib/components/operator/SettingsPanel.svelte';
	import FolderBrowserPanel from '$lib/components/operator/FolderBrowserPanel.svelte';

	type SettingsSection =
		| 'dashboard'
		| 'library'
		| 'scanning'
		| 'metadata'
		| 'playback'
		| 'storage'
		| 'access'
		| 'about';

	type LibraryActionKind = 'scan' | 'remove' | '';

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

	let isLoading = $state(true);
	let isScanningMovies = $state(false);
	let isScanningTV = $state(false);
	let isRefreshingMovies = $state(false);
	let isRefreshingTV = $state(false);
	let isSavingServerName = $state(false);
	let isSavingScanningAutomation = $state(false);
	let isSavingPlaybackPolicy = $state(false);
	let isSavingStorage = $state(false);
	let isSigningOut = $state(false);
	let isBrowsingStorage = $state(false);
	let loadError = $state('');
	let actionMessage = $state('');
	let scanningSettingsError = $state('');
	let serverNameError = $state('');
	let storageSettingsError = $state('');
	let lastUpdatedLabel = $state('');
	let sessionsUnavailable = $state(false);
	let authDisabled = $state(false);
	let devAuthBypass = $state(false);
	let devAuthBypassMessage = $state('');
	let activeSection = $state<SettingsSection>('dashboard');
	let activeLibraryActionID = $state('');
	let activeLibraryActionKind = $state<LibraryActionKind>('');

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
	let buildInfo = $state<BuildInfo | null>(null);
	let storageFolderBrowse = $state<FolderBrowseResponse | null>(null);
	let activeStorageBrowseField = $state<EditableStorageFieldKey | ''>('');
	let serverNameDraft = $state('Lorivo');
	let librarySyncModeDraft = $state<LibrarySyncModeOption['id']>('daily');
	let syncIntervalDraft = $state('1440');
	let watchDebounceDraft = $state('30');
	let probeBatchLimitDraft = $state('50');
	let playbackPolicyDraft = $state<'original_only' | 'light' | 'full' | 'cinema'>('original_only');
	let storageDraft = $state<Record<StorageFieldKey, string>>({
		dataDir: '',
		transcodeDir: '',
		downloadsDir: '',
		metadataDir: '',
		cacheDir: '',
		tempDir: ''
	});
	let sessionExpiresAt = $state('');

	let refreshTimer: ReturnType<typeof setTimeout> | null = null;

	const serverIdentityHelpText =
		'Lorivo uses this name in the browser title and shares it with local clients. Local network discovery is not available in this build yet.';

	const storageFieldDefinitions: StorageFieldDefinition[] = [
		{
			key: 'transcodeDir',
			diskName: 'transcode',
			label: 'Transcoding folder',
			helper: 'Where Lorivo stores temporary files while preparing playback.',
			group: 'processing',
			editable: true
		},
		{
			key: 'downloadsDir',
			diskName: 'downloads',
			label: 'Optimized versions folder',
			helper: 'Where Lorivo stores optimized versions created for playback or travel.',
			group: 'processing',
			editable: true
		},
		{
			key: 'metadataDir',
			diskName: 'metadata',
			label: 'Metadata folder',
			helper: 'Where Lorivo stores artwork and metadata it downloads for your library.',
			group: 'library',
			editable: true
		},
		{
			key: 'cacheDir',
			diskName: 'cache',
			label: 'Cache folder',
			helper: 'Where Lorivo stores temporary cached data.',
			group: 'library',
			editable: true
		},
		{
			key: 'tempDir',
			diskName: 'temp',
			label: 'Scratch/temp folder',
			helper: 'Where Lorivo stores short-lived working files.',
			group: 'library',
			editable: true
		},
		{
			key: 'dataDir',
			diskName: 'data',
			label: 'Data folder',
			helper: 'Where Lorivo stores its main settings and local data.',
			group: 'app',
			editable: false
		}
	];

	const librarySyncModeOptions: LibrarySyncModeOption[] = [
		{
			id: 'manual',
			label: 'Manual only',
			description: 'Lorivo scans your libraries only when you start a scan.'
		},
		{
			id: 'daily',
			label: 'Scheduled',
			description: 'Lorivo checks your libraries on a repeating schedule.'
		},
		{
			id: 'watch',
			label: 'Watch folders',
			description: 'Lorivo waits for folder changes, then starts a scan after a short delay.'
		}
	];

	const playbackPolicyOptions: PlaybackPolicyOption[] = [
		{
			id: 'original_only',
			label: 'Original files only',
			description: 'Keep playback as close to the original file as possible. If a device needs help, Lorivo offers fallback choices instead of converting automatically.'
		},
		{
			id: 'light',
			label: 'Direct play with audio fixes',
			description: 'Prefer the original video. Lorivo can repackage playback or convert audio when that is enough.'
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

	const devOwnerActive = $derived.by(() => devAuthBypass && Boolean(user));
	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userRoleLabel = $derived.by(() => (devOwnerActive ? 'Development Owner' : accountTypeLabel(user?.role)));
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
	const serverDisplayName = $derived.by(() => displayServerName(settings.config?.serverName));
	const canManageSettings = $derived.by(() => authDisabled || devOwnerActive || asText(user?.role).toLowerCase() === 'admin');
	const canShowSignIn = $derived.by(() => !authDisabled && !devAuthBypass && !user);
	const ownerSetupPending = $derived.by(() => !authDisabled && !devAuthBypass && !user && Boolean(clientBootstrap.auth?.bootstrapAllowed));
	const requiresOwnerSignIn = $derived.by(() => !authDisabled && !devAuthBypass && !canManageSettings);
	const ownerAccessMessage = $derived.by(() => 'Sign in as the owner to manage Lorivo settings.');
	const ownerActionLabel = $derived.by(() => (ownerSetupPending ? 'Create Owner Account' : 'Sign In'));
	const ownerActionDetail = $derived.by(() => {
		const defaultUsername = asText(clientBootstrap.auth?.defaultUsername);
		if (ownerSetupPending) {
			return defaultUsername
				? `This server still needs its first owner account. Open Sign In to create ${defaultUsername}.`
				: 'This server still needs its first owner account. Open Sign In to create it.';
		}
		return 'Open Sign In to continue with the owner account.';
	});
	const devAccessMessage = $derived.by(
		() =>
			asText(devAuthBypassMessage) ||
			'Development access is active. User management will be enabled before production.'
	);
	const accessCardLabel = $derived.by(() => {
		if (devOwnerActive) return 'Development owner';
		if (user) return 'Signed in';
		if (authDisabled) return 'Local access';
		if (ownerSetupPending) return 'Create owner account';
		return 'Sign in';
	});
	const accessCardDetail = $derived.by(() => {
		if (devOwnerActive) return 'owner controls are unlocked for local development';
		if (user) return userDisplayName;
		if (authDisabled) return 'sign-in is not required on this server';
		if (ownerSetupPending) return 'first owner account needed for changes';
		return 'owner account needed for changes';
	});
	const accessAccountValue = $derived.by(() => {
		if (user) return userDisplayName;
		if (authDisabled) return 'Local access';
		return 'Not signed in';
	});
	const accessAccountMeta = $derived.by(() => {
		if (devOwnerActive) return 'Development Owner';
		if (user) return userRoleLabel;
		if (authDisabled) return 'Sign-in is not required on this server.';
		return 'Sign in to change protected settings.';
	});
	const accessSessionValue = $derived.by(() => {
		if (devOwnerActive) return 'Development access';
		if (user) return 'Active';
		if (authDisabled) return 'Local';
		return 'Signed out';
	});
	const accessSessionMeta = $derived.by(() => {
		if (devOwnerActive) return 'Owner-only settings are unlocked on this device.';
		if (user && sessionExpiresAt) return `Current session expires ${formatDateTime(sessionExpiresAt)}.`;
		if (user) return 'This browser has an active Lorivo session.';
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
	const hasScanningAutomationChanges = $derived.by(
		() =>
			librarySyncModeDraft !== normalizeLibrarySyncMode(settings.config?.librarySyncMode) ||
			syncIntervalDraft !== stringDraft(settings.config?.syncIntervalMins, 1440) ||
			watchDebounceDraft !== stringDraft(settings.config?.watchDebounceSecs, 30) ||
			probeBatchLimitDraft !== stringDraft(settings.config?.probeBatchLimit, 50)
	);
	const hasStorageChanges = $derived.by(
		() =>
			asText(storageDraft.transcodeDir) !== asText(settings.config?.transcodeDir) ||
			asText(storageDraft.downloadsDir) !== asText(settings.config?.downloadsDir) ||
			asText(storageDraft.metadataDir) !== asText(settings.config?.metadataDir) ||
			asText(storageDraft.cacheDir) !== asText(settings.config?.cacheDir) ||
			asText(storageDraft.tempDir) !== asText(settings.config?.tempDir)
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
		if (libraryCards.length === 0) {
			output.push({
				id: 'warn-library',
				label: 'Library setup needed',
				description: 'Add a Movies or TV folder so Lorivo can build your media library.',
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
				description: `${asCount(health.unprobed)} files still need Lorivo's media check.`,
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
		if (storageNeedsAttentionCount > 0) return 'Needs attention';
		if (storageConfiguredCount > 0) return `${asCount(storageConfiguredCount)} folders configured`;
		return 'Folders not set yet';
	});
	const storageDashboardDetail = $derived.by(() => {
		if (storageNeedsAttentionCount > 0) {
			return `${asCount(storageNeedsAttentionCount)} folder${storageNeedsAttentionCount === 1 ? '' : 's'} need attention before restart.`;
		}
		return 'Review where Lorivo keeps its media processing, artwork, cache, and local data.';
	});
	const storagePanelStatus = $derived.by(() => {
		if (storageNeedsAttentionCount > 0) return 'warning';
		if (storageConfiguredCount > 0) return 'healthy';
		return 'idle';
	});
	const activeStorageBrowseFieldLabel = $derived.by(() => storageFieldLabel(activeStorageBrowseField));
	const scanningSaveMessage = $derived.by(() =>
		actionMessage === 'Scanning settings saved.' ||
		actionMessage === 'Saved. Restart Lorivo for this change to fully take effect.'
			? actionMessage
			: ''
	);
	const storageSaveMessage = $derived.by(() =>
		actionMessage === 'Storage settings saved.' ||
		actionMessage === 'Saved. Restart Lorivo for these folder changes to fully take effect.'
			? actionMessage
			: ''
	);
	const playbackSaveMessage = $derived.by(() =>
		actionMessage === 'Playback setting saved.' ||
		actionMessage === 'Playback setting saved. Restart Lorivo to apply it.'
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
		sessionsUnavailable = false;
		try {
			const [
				bootstrapPayload,
				sessionPayload,
				librariesPayload,
				summaryPayload,
				healthPayload,
				systemPayload,
				settingsPayload,
				performancePayload,
				scansPayload,
				probesPayload,
				workPayload,
				downloadsPayload,
				sessionsPayload
			] = await Promise.all([
				getClientBootstrap(apiClient).catch(() => ({} as ClientBootstrapResponse)),
				getAuthSession(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) return {} as AuthSessionResponse;
					throw error;
				}),
				getLibraries(apiClient),
				getCatalogSummary(apiClient),
				getCatalogHealth(apiClient),
				getSystemStatus(apiClient),
				getSettings(apiClient),
				getPerformanceSettings(apiClient),
				getScans(apiClient),
				getProbes(apiClient),
				getWork(apiClient),
				getDownloads(apiClient),
				getSessions(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) {
						sessionsUnavailable = true;
						return { sessions: [] };
					}
					throw error;
				})
			]);

			clientBootstrap = bootstrapPayload || {};
			authDisabled = Boolean(sessionPayload?.authDisabled);
			devAuthBypass = Boolean(sessionPayload?.devAuthBypass);
			devAuthBypassMessage = asText(sessionPayload?.devAuthBypassMessage);
			user = sessionPayload?.user || null;
			sessionExpiresAt = asText(sessionPayload?.session?.expiresAt);
			libraries = librariesPayload.libraries || [];
			summary = summaryPayload || {};
			health = healthPayload || {};
			system = systemPayload || {};
			settings = settingsPayload || {};
			serverNameDraft = displayServerName(settingsPayload.config?.serverName);
			syncStorageDraft(settingsPayload.config);
			librarySyncModeDraft = normalizeLibrarySyncMode(settingsPayload.config?.librarySyncMode);
			syncIntervalDraft = stringDraft(settingsPayload.config?.syncIntervalMins, 1440);
			watchDebounceDraft = stringDraft(settingsPayload.config?.watchDebounceSecs, 30);
			probeBatchLimitDraft = stringDraft(settingsPayload.config?.probeBatchLimit, 50);
			scanningSettingsError = '';
			storageSettingsError = '';
			playbackPolicyDraft = normalizePlaybackPolicy(settingsPayload.config?.playbackPolicy);
			performance = performancePayload || {};
			scans = scansPayload.scans || [];
			probes = probesPayload.probes || [];
			work = workPayload.work || [];
			downloads = downloadsPayload.downloads || [];
			sessions = sessionsPayload.sessions || [];
			lastUpdatedLabel = new Date().toLocaleTimeString();
		} catch (error) {
			loadError = formatLoadError(error);
		} finally {
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
				? 'Saved. Restart Lorivo for this change to fully take effect.'
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
				? 'Playback setting saved. Restart Lorivo to apply it.'
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

		const nextTranscodeDir = asText(storageDraft.transcodeDir);
		const nextDownloadsDir = asText(storageDraft.downloadsDir);
		const nextMetadataDir = asText(storageDraft.metadataDir);
		const nextCacheDir = asText(storageDraft.cacheDir);
		const nextTempDir = asText(storageDraft.tempDir);

		if (!nextTranscodeDir) {
			storageSettingsError = 'Choose a folder for the Transcoding folder.';
			return;
		}
		if (!nextDownloadsDir) {
			storageSettingsError = 'Choose a folder for the Optimized versions folder.';
			return;
		}
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
		if (nextTranscodeDir !== asText(settings.config?.transcodeDir)) payload.transcodeDir = nextTranscodeDir;
		if (nextDownloadsDir !== asText(settings.config?.downloadsDir)) payload.downloadsDir = nextDownloadsDir;
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
				? 'Saved. Restart Lorivo for these folder changes to fully take effect.'
				: 'Storage settings saved.';
			await loadSettingsSurface(true);
		} catch (error) {
			storageSettingsError = formatLoadError(error);
		} finally {
			isSavingStorage = false;
		}
	}

	async function scanLibraryItem(library: LibraryRecord): Promise<void> {
		const id = asText(library.id);
		if (!id || !canManageSettings) return;
		activeLibraryActionID = id;
		activeLibraryActionKind = 'scan';
		actionMessage = '';
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
		const confirmed = window.confirm('Remove this library from Lorivo? Media files are not deleted.');
		if (!confirmed) return;
		activeLibraryActionID = id;
		activeLibraryActionKind = 'remove';
		actionMessage = '';
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

	async function signOut(): Promise<void> {
		if (isSigningOut || authDisabled || devOwnerActive || !user) return;
		isSigningOut = true;
		actionMessage = '';
		try {
			await logout(apiClient);
			actionMessage = 'Signed out.';
			await loadSettingsSurface(true);
		} catch (error) {
			actionMessage = formatLoadError(error);
		} finally {
			isSigningOut = false;
		}
	}

	function syncSectionFromHash(): void {
		if (typeof window === 'undefined') return;
		const candidate = window.location.hash.replace(/^#/, '');
		activeSection = isSettingsSection(candidate) ? candidate : 'dashboard';
	}

	function isSettingsSection(value: string): value is SettingsSection {
		return ['dashboard', 'library', 'scanning', 'metadata', 'playback', 'storage', 'access', 'about'].includes(value);
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
		if (itemID === 'warn-library') return '#library';
		if (itemID === 'warn-unsupported') return '#playback';
		return '#metadata';
	}

	function warningItemActionLabel(itemID: string): string {
		if (itemID === 'warn-library') return 'Open Library';
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
				: 'Scans start only when you ask Lorivo to run them.';
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

	function displayText(value: unknown): string {
		return asText(value).replace(/_/g, ' ');
	}

	function displayServerName(value: unknown): string {
		return normalizeServerName(value);
	}

	function titleForServerName(value: unknown): string {
		return lorivoTitle(value);
	}

	function announceServerName(): void {
		if (typeof window === 'undefined') return;
		window.dispatchEvent(new CustomEvent('lorivo:server-name-changed', { detail: { serverName: serverDisplayName } }));
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
		if (!normalized) return 'Local Account';
		if (normalized === 'admin') return 'Owner account';
		if (normalized === 'standard') return 'Standard account';
		return 'Local Account';
	}

	function syncStorageDraft(configValues: SettingsResponse['config'] | undefined): void {
		storageDraft.dataDir = asText(configValues?.dataDir);
		storageDraft.transcodeDir = asText(configValues?.transcodeDir);
		storageDraft.downloadsDir = asText(configValues?.downloadsDir);
		storageDraft.metadataDir = asText(configValues?.metadataDir);
		storageDraft.cacheDir = asText(configValues?.cacheDir);
		storageDraft.tempDir = asText(configValues?.tempDir);
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
			return { label: 'Not checked', tone: 'neutral', detail: 'Lorivo has not checked this folder yet.' };
		}
		if (disk.error || disk.writable === false) {
			return {
				label: 'Needs attention',
				tone: 'warn',
				detail: asText(disk.error) || 'Lorivo could not confirm that this folder is writable.'
			};
		}
		return { label: 'Ready', tone: 'good', detail: 'Lorivo can use this folder.' };
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
			library: 'Library',
			scanning: 'Scanning',
			metadata: 'Metadata',
			playback: 'Playback',
			storage: 'Storage',
			access: 'Access',
			about: 'About'
		}[section];
	}

	function sectionDescription(section: SettingsSection): string {
		return {
			dashboard: 'Check whether your Lorivo library is ready and jump to the next useful setting.',
			library: 'Media folders, setup status, and the current Library Setup flow.',
			scanning: 'Choose how Lorivo checks libraries and review current scan activity.',
			metadata: 'Refresh movie and TV information and review items that need attention.',
			playback: 'Choose how Lorivo handles playback compatibility and review active sessions.',
			storage: 'Choose where Lorivo keeps media processing files, library artwork, cache, and local data.',
			access: 'See who is signed in and manage the current session.',
			about: 'Lorivo identity, build, and local-first details.'
		}[section];
	}
</script>

<svelte:head>
	<title>{titleForServerName(settings.config?.serverName)}</title>
</svelte:head>

<ServerShell
	active={activeSection}
	showStorage={canManageSettings}
	{userDisplayName}
	userRole={userRoleLabel}
	{userInitials}
>
	<div class="settings-page">
		{#if isLoading}
			<LorivoPanel title="Loading Settings" subtitle="Checking your Lorivo library." />
		{:else if loadError}
			<LorivoPanel title="Settings could not load" subtitle={loadError}>
				<div class="status-actions">
					<LorivoButton variant="secondary" onclick={() => loadSettingsSurface(false)}>Retry</LorivoButton>
					<LorivoButton variant="ghost" href="/">Back to Media</LorivoButton>
				</div>
			</LorivoPanel>
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
				<LorivoPanel title="Latest action" subtitle={actionMessage} />
			{/if}

			{#if requiresOwnerSignIn}
				<LorivoPanel title="Owner sign-in required" subtitle={ownerAccessMessage}>
					<p class="settings-note settings-note--inline">{ownerActionDetail}</p>
					<div class="status-actions">
						<LorivoButton variant="primary" href="/signin">{ownerActionLabel}</LorivoButton>
						<LorivoButton variant="ghost" href="#access">Open Access</LorivoButton>
					</div>
				</LorivoPanel>
			{/if}

			{#if activeSection === 'dashboard'}
				<section class="settings-dashboard" aria-label="Settings dashboard" data-testid="settings-dashboard">
					<article class="settings-dashboard-card settings-dashboard-card--identity">
						<span>Server name</span>
						<strong>{serverDisplayName}</strong>
						<small>{serverIdentityHelpText}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="secondary" size="sm" href="#about">Edit in About</LorivoButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Library</span>
						<strong>{libraryCards.length > 0 ? `${asCount(libraryCards.length)} folders ready` : 'Library setup needed'}</strong>
						<small>{libraryCards.length > 0 ? 'Review folders, run scans, or remove a library.' : 'Add a Movies or TV folder so Lorivo can start building your library.'}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="secondary" size="sm" href="#library">Review Library</LorivoButton>
							<LorivoButton variant="ghost" size="sm" href="/setup">Library Setup</LorivoButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Media</span>
						<strong>{Number(summary.movies || 0) + Number(summary.series || 0) > 0 ? `${asCount(summary.movies)} movies` : 'No media found yet'}</strong>
						<small>{Number(summary.movies || 0) + Number(summary.series || 0) > 0 ? `${asCount(summary.series)} shows and ${asCount(summary.episodes)} episodes ready to browse.` : 'Add a library and run a scan to populate Lorivo.'}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="ghost" size="sm" href="/movies">Movies</LorivoButton>
							<LorivoButton variant="ghost" size="sm" href="/tv">TV</LorivoButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Scanning</span>
						<strong>{scanningModeSummary()}</strong>
						<small>{scanningModeMeta()}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="secondary" size="sm" href="#scanning">Open Scanning</LorivoButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Metadata</span>
						<strong>{Number(health.needsReview || 0) > 0 ? 'Review' : 'Ready'}</strong>
						<small>{Number(health.needsReview || 0) > 0 ? `${asCount(health.needsReview)} items need attention.` : 'Movie and TV details look up to date.'}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="secondary" size="sm" href="#metadata">Open Metadata</LorivoButton>
						</div>
					</article>
					<article class="settings-dashboard-card">
						<span>Playback</span>
						<strong>{playbackPolicyDetails(settings.config?.playbackPolicy).label}</strong>
						<small>{activeSessionCount > 0 ? `${asCount(activeSessionCount)} active sessions right now.` : 'Playback is idle right now.'}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="secondary" size="sm" href="#playback">Open Playback</LorivoButton>
						</div>
					</article>
					{#if canManageSettings}
						<article class="settings-dashboard-card">
							<span>Storage</span>
							<strong>{storageDashboardLabel}</strong>
							<small>{storageDashboardDetail}</small>
							<div class="dashboard-card-actions">
								<LorivoButton variant="secondary" size="sm" href="#storage">Open Storage</LorivoButton>
							</div>
						</article>
					{/if}
					<article class="settings-dashboard-card">
						<span>Access</span>
						<strong>{accessCardLabel}</strong>
						<small>{accessCardDetail}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="secondary" size="sm" href="#access">Open Access</LorivoButton>
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
							<LorivoButton
								variant="ghost"
								size="sm"
								href={primaryWarningItem ? warningItemHref(primaryWarningItem.id) : '#dashboard'}
							>
								{primaryWarningItem ? warningItemActionLabel(primaryWarningItem.id) : 'Review Settings'}
							</LorivoButton>
						</div>
					</article>
					<article class="settings-dashboard-card settings-dashboard-card--quiet">
						<span>About</span>
						<strong>Lorivo</strong>
						<small>{asText(buildInfo?.buildID) || 'Local build details'}</small>
						<div class="dashboard-card-actions">
							<LorivoButton variant="ghost" size="sm" href="#about">Open About</LorivoButton>
						</div>
					</article>
				</section>
			{:else if activeSection === 'library'}
				<section id="library" class="settings-section" data-testid="settings-section-content" data-section="library">
				<SettingsPanel title="Library" description="Media folders and library setup." status={libraryCards.length > 0 ? 'healthy' : 'idle'}>
					{#snippet actions()}
						<LorivoButton variant="primary" href="/setup">Library Setup</LorivoButton>
					{/snippet}
					<div class="stat-grid stat-grid--compact">
						<LorivoStat label="Libraries" value={asCount(libraryCards.length)} meta={libraryCards.length > 0 ? 'Configured folders' : 'Add a library to begin'} tone={libraryCards.length > 0 ? 'good' : 'warn'} />
						<LorivoStat label="Movies" value={asCount(summary.movies)} meta="Current movie catalog count." />
						<LorivoStat label="Shows" value={asCount(summary.series)} meta={`${asCount(summary.episodes)} episodes in the current TV catalog.`} />
					</div>
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
											<LorivoButton
												variant="secondary"
												size="sm"
												disabled={activeLibraryActionID === library.id}
												onclick={() => scanLibraryItem(configuredLibraries.find((item) => asText(item.id) === library.id) || {})}
											>
												{activeLibraryActionID === library.id && activeLibraryActionKind === 'scan' ? 'Scanning...' : 'Scan'}
											</LorivoButton>
											<LorivoButton
												variant="danger"
												size="sm"
												disabled={activeLibraryActionID === library.id}
												onclick={() => removeLibraryItem(configuredLibraries.find((item) => asText(item.id) === library.id) || {})}
											>
												{activeLibraryActionID === library.id && activeLibraryActionKind === 'remove' ? 'Removing...' : 'Remove'}
											</LorivoButton>
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
									<h3>Set up your first library</h3>
									<p>Add a Movies or TV folder so Lorivo can scan it and build your media library.</p>
								</div>
							</div>
							<div class="status-actions">
								<LorivoButton variant="primary" href="/setup">Library Setup</LorivoButton>
							</div>
						</div>
					{/if}
					{#if !canManageSettings && !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as the owner to manage Lorivo settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<LorivoButton variant="primary" size="sm" href="/signin">{ownerActionLabel}</LorivoButton>
								<LorivoButton variant="ghost" size="sm" href="#access">Open Access</LorivoButton>
							</div>
						</div>
					{/if}
				</SettingsPanel>
			</section>

			{:else if activeSection === 'scanning'}
			<section id="scanning" class="settings-section" data-testid="settings-section-content" data-section="scanning">
				<SettingsPanel title="Scanning" description="Start real library scans, choose how Lorivo checks libraries, and review current scan activity." status={activeQueueCount > 0 ? 'warning' : 'healthy'}>
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Manual scans</h3>
								<p>Run a scan when you want Lorivo to check movies or TV right now.</p>
							</div>
						</div>
						<div class="status-actions">
							<LorivoButton variant="primary" onclick={startMovieScan} disabled={isScanningMovies || isScanningTV}>
								{isScanningMovies ? 'Scanning Movies...' : 'Scan Movies'}
							</LorivoButton>
							<LorivoButton variant="secondary" onclick={startTVScan} disabled={isScanningMovies || isScanningTV}>
								{isScanningTV ? 'Scanning TV...' : 'Scan TV'}
							</LorivoButton>
						</div>
					</div>
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Automation</h3>
								<p>Choose how Lorivo checks your libraries for new or changed media.</p>
							</div>
						</div>
						<div class="stat-grid stat-grid--compact">
						<LorivoStat
							label="Library Scan Mode"
							value={scanningModeDetails(settings.config?.librarySyncMode).label}
							meta={scanningModeDetails(settings.config?.librarySyncMode).description}
						/>
						{#if normalizeLibrarySyncMode(settings.config?.librarySyncMode) === 'watch'}
							<LorivoStat
								label="Folder Watch Delay"
								value={friendlySeconds(settings.config?.watchDebounceSecs, 30)}
								meta="How long Lorivo waits after a folder change before starting a scan."
							/>
						{:else}
							<LorivoStat
								label="Scan Interval"
								value={normalizeLibrarySyncMode(settings.config?.librarySyncMode) === 'manual' ? 'Manual' : friendlyMinutes(settings.config?.syncIntervalMins, 1440)}
								meta={normalizeLibrarySyncMode(settings.config?.librarySyncMode) === 'manual' ? 'Scans start only when you ask Lorivo to run them.' : 'Time between scheduled library scans.'}
							/>
						{/if}
						<LorivoStat
							label="Media check batch size"
							value={asCount(settings.config?.probeBatchLimit || 50)}
							meta="Items Lorivo checks at a time after a library scan."
						/>
						</div>
					</div>
					{#if canManageSettings}
						<form class="scanning-automation-form" data-testid="scanning-automation-form" onsubmit={(event) => { event.preventDefault(); void saveScanningAutomation(); }}>
							<div class="settings-field">
								<span>Library scan mode</span>
								<small>Choose how Lorivo checks your libraries for new or changed media.</small>
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
									<small>How often Lorivo runs a scheduled library scan. Minimum 15 minutes.</small>
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
									<small>How long Lorivo waits after a folder change before starting a scan.</small>
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
										<small>How many items Lorivo checks in one batch during library scanning.</small>
									</label>
								</div>
							</details>
							<div class="status-actions">
								<LorivoButton
									variant="primary"
									disabled={isSavingScanningAutomation || !hasScanningAutomationChanges}
									onclick={saveScanningAutomation}
								>
									{isSavingScanningAutomation ? 'Saving...' : 'Save Scanning Settings'}
								</LorivoButton>
							</div>
							{#if scanningSaveMessage}
								<p class="settings-feedback">{scanningSaveMessage}</p>
							{/if}
							{#if scanningSettingsError}
								<p class="settings-error">{scanningSettingsError}</p>
							{/if}
						</form>
					{:else if !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as the owner to manage Lorivo settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<LorivoButton variant="primary" size="sm" href="/signin">{ownerActionLabel}</LorivoButton>
								<LorivoButton variant="ghost" size="sm" href="#access">Open Access</LorivoButton>
							</div>
						</div>
					{/if}
					<ActivityListShell title="Recent Scans">
						<LorivoActionList
							items={scanItems}
							emptyLabel="No scan activity right now."
						/>
					</ActivityListShell>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'metadata'}
			<section id="metadata" class="settings-section" data-testid="settings-section-content" data-section="metadata">
				<SettingsPanel title="Metadata" description="Refresh movie and TV information." status={Number(health.needsReview || 0) > 0 ? 'warning' : 'healthy'}>
					<div class="settings-subsection">
						<div class="settings-subsection__head">
							<div>
								<h3>Refresh media details</h3>
								<p>Use these actions when movie or TV details need another pass.</p>
							</div>
						</div>
						<div class="status-actions">
							<LorivoButton variant="primary" onclick={refreshMovieMetadata} disabled={isRefreshingMovies || isRefreshingTV}>
								{isRefreshingMovies ? 'Refreshing Movies...' : 'Refresh Movies'}
							</LorivoButton>
							<LorivoButton variant="secondary" onclick={refreshTVMetadata} disabled={isRefreshingMovies || isRefreshingTV}>
								{isRefreshingTV ? 'Refreshing TV...' : 'Refresh TV'}
							</LorivoButton>
						</div>
					</div>
					<div class="stat-grid stat-grid--compact">
						<LorivoStat label="Review Needed" value={asCount(health.needsReview)} meta={`${asCount(health.unprobed)} files pending media checks`} tone={Number(health.needsReview || 0) > 0 ? 'warn' : 'good'} />
						<LorivoStat label="Subtitles Found" value={asCount(health.withSubtitles)} meta="Available for playback when supported." />
					</div>
					<div class="settings-subsection settings-subsection--quiet">
						<div class="settings-subsection__head">
							<div>
								<h3>Title and artwork lookup</h3>
								<p>Lorivo manages artwork and title lookups in this build. There are no lookup settings to manage here.</p>
							</div>
						</div>
					</div>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'playback'}
			<section id="playback" class="settings-section" data-testid="settings-section-content" data-section="playback">
				<SettingsPanel title="Playback" description="Playback policy, compatibility status, and active playback sessions." status={activeSessionCount > 0 ? 'warning' : 'healthy'}>
					<div class="stat-grid stat-grid--compact">
						<LorivoStat
							label="Playback Policy"
							value={playbackPolicyDetails(settings.config?.playbackPolicy).label}
							meta={playbackPolicyDetails(settings.config?.playbackPolicy).description}
						/>
						<LorivoStat label="Compatibility Support" value={asText(performance.hardwareAcceleration?.status) || 'Unknown'} meta="Shows whether Lorivo can help when a device needs a different playback format." />
						<LorivoStat label="Active Sessions" value={asCount(activeSessionCount)} meta={sessionsUnavailable ? 'Sign in to view current playback sessions.' : 'Current playback sessions.'} tone={activeSessionCount > 0 ? 'warn' : 'good'} />
					</div>
					{#if canManageSettings}
						<form class="playback-policy-form" data-testid="playback-policy-form" onsubmit={(event) => { event.preventDefault(); void savePlaybackPolicy(); }}>
							<div class="settings-subsection__head settings-subsection__head--tight">
								<div>
									<h3>Playback policy</h3>
									<p>Pick how Lorivo balances original quality and device compatibility.</p>
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
							<div class="status-actions">
								<LorivoButton
									variant="primary"
									disabled={isSavingPlaybackPolicy || playbackPolicyDraft === normalizePlaybackPolicy(settings.config?.playbackPolicy)}
									onclick={savePlaybackPolicy}
								>
									{isSavingPlaybackPolicy ? 'Saving...' : 'Save Playback Policy'}
								</LorivoButton>
							</div>
							{#if playbackSaveMessage}
								<p class="settings-feedback">{playbackSaveMessage}</p>
							{/if}
						</form>
					{:else if !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as the owner to manage Lorivo settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<LorivoButton variant="primary" size="sm" href="/signin">{ownerActionLabel}</LorivoButton>
								<LorivoButton variant="ghost" size="sm" href="#access">Open Access</LorivoButton>
							</div>
						</div>
					{/if}
					<ActivityListShell title="Playback Sessions">
						<LorivoActionList
							items={sessionItems}
							emptyLabel={sessionsUnavailable ? 'Sign in to view current playback sessions.' : 'No active playback sessions right now.'}
						/>
					</ActivityListShell>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'storage'}
			<section id="storage" class="settings-section" data-testid="settings-section-content" data-section="storage">
				<SettingsPanel
					title="Storage"
					description="Choose where Lorivo keeps media processing files, library artwork, cache, and local data."
					status={storagePanelStatus}
				>
					<div class="stat-grid stat-grid--compact">
						<LorivoStat
							label="Folders configured"
							value={asCount(storageConfiguredCount)}
							meta="Saved folders Lorivo can use for local storage."
							tone={storageConfiguredCount > 0 ? 'good' : 'neutral'}
						/>
						<LorivoStat
							label="Needs attention"
							value={asCount(storageNeedsAttentionCount)}
							meta={storageNeedsAttentionCount > 0 ? 'One or more folders may need review before restart.' : 'All checked folders look ready.'}
							tone={storageNeedsAttentionCount > 0 ? 'warn' : 'good'}
						/>
					</div>
					{#if canManageSettings}
						<form class="storage-settings-form" data-testid="storage-form" onsubmit={(event) => { event.preventDefault(); void saveStorageSettings(); }}>
							<div class="settings-subsection">
								<div class="settings-subsection__head">
									<div>
										<h3>Media processing folders</h3>
										<p>Choose where Lorivo keeps temporary processing files and optimized versions.</p>
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
														<LorivoButton
															variant="secondary"
															size="sm"
															type="button"
															onclick={() => browseStorageField(field.key)}
														>
															{isBrowsingStorage && activeStorageBrowseField === field.key ? 'Loading folders...' : 'Browse'}
														</LorivoButton>
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
							<div class="settings-subsection">
								<div class="settings-subsection__head">
									<div>
										<h3>Library data folders</h3>
										<p>Choose where Lorivo keeps artwork, cache, and other short-lived working files.</p>
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
														<LorivoButton
															variant="secondary"
															size="sm"
															type="button"
															onclick={() => browseStorageField(field.key)}
														>
															{isBrowsingStorage && activeStorageBrowseField === field.key ? 'Loading folders...' : 'Browse'}
														</LorivoButton>
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
										<p>This folder stores Lorivo's main settings and local data.</p>
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
											<small>Changing this folder can move where Lorivo expects its settings and local data after restart. It stays read-only in this build.</small>
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
							<div class="status-actions">
								<LorivoButton
									variant="primary"
									disabled={isSavingStorage || !hasStorageChanges}
									onclick={saveStorageSettings}
								>
									{isSavingStorage ? 'Saving...' : 'Save Storage Settings'}
								</LorivoButton>
							</div>
							{#if storageSaveMessage}
								<p class="settings-feedback">{storageSaveMessage}</p>
							{/if}
							{#if storageSettingsError}
								<p class="settings-error">{storageSettingsError}</p>
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
							<p class="settings-note">Sign in as the owner to manage Lorivo settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<LorivoButton variant="primary" size="sm" href="/signin">{ownerActionLabel}</LorivoButton>
								<LorivoButton variant="ghost" size="sm" href="#access">Open Access</LorivoButton>
							</div>
						</div>
					{/if}
				</SettingsPanel>
			</section>

			{:else if activeSection === 'access'}
			<section id="access" class="settings-section" data-testid="settings-section-content" data-section="access">
				<SettingsPanel title="Access" description="Current account and session." status={user ? 'healthy' : 'idle'}>
					{#snippet actions()}
						{#if user && !authDisabled && !devOwnerActive}
							<LorivoButton variant="secondary" disabled={isSigningOut} onclick={signOut}>
								{isSigningOut ? 'Signing out...' : 'Sign Out'}
							</LorivoButton>
						{:else if canShowSignIn}
							<LorivoButton variant="primary" href="/signin">{ownerActionLabel}</LorivoButton>
						{/if}
					{/snippet}
					<div class="stat-grid stat-grid--compact">
						<LorivoStat label="Account" value={accessAccountValue} meta={accessAccountMeta} tone={user ? 'good' : 'neutral'} />
						<LorivoStat label="Session" value={accessSessionValue} meta={accessSessionMeta} tone={user ? 'good' : 'neutral'} />
					</div>
					<div class="settings-auth-callout">
						<p class="settings-note">{devOwnerActive ? devAccessMessage : 'Access is limited to the current owner session in this build.'}</p>
						<ul class="settings-placeholder-list">
							<li>User management is not available yet.</li>
							<li>Device pairing will appear here when client pairing is implemented.</li>
						</ul>
					</div>
					{#if !user && !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as the owner to manage Lorivo settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
						</div>
					{/if}
				</SettingsPanel>
			</section>

			{:else if activeSection === 'about'}
			<section id="about" class="settings-section" data-testid="settings-section-content" data-section="about">
				<SettingsPanel title="About" description="Lorivo identity and local server details." status="healthy">
					<div class="stat-grid stat-grid--compact">
						<LorivoStat label="App" value="Lorivo" meta="Local-first personal media library." />
						<LorivoStat label="Server Name" value={serverDisplayName} meta="Shown in the browser title and shared with local clients." />
						<LorivoStat label="Build" value={asText(buildInfo?.buildID) || 'Local build'} meta={asText(buildInfo?.publishedAt) || 'Build details are quiet in this view.'} />
						<LorivoStat label="Mode" value="Local" meta="No cloud account or vendor relay required." />
					</div>
					<div class="settings-subsection settings-subsection--quiet">
						<div class="settings-subsection__head">
							<div>
								<h3>Server identity</h3>
								<p>{serverIdentityHelpText}</p>
							</div>
						</div>
					</div>
					{#if canManageSettings}
						<form class="server-name-form" data-testid="server-name-form" onsubmit={(event) => { event.preventDefault(); void saveServerName(); }}>
							<label class="settings-field">
								<span>Server name</span>
								<input
									bind:value={serverNameDraft}
									maxlength="50"
									required
									placeholder="Lorivo"
									aria-describedby="settings-server-name-help"
								/>
								<small id="settings-server-name-help">{serverIdentityHelpText}</small>
							</label>
							<div class="status-actions">
								<LorivoButton variant="primary" disabled={isSavingServerName} onclick={saveServerName}>
									{isSavingServerName ? 'Saving...' : 'Save Server Name'}
								</LorivoButton>
							</div>
							{#if serverNameSaveMessage}
								<p class="settings-feedback">{serverNameSaveMessage}</p>
							{/if}
							{#if serverNameError}
								<p class="settings-error">{serverNameError}</p>
							{/if}
						</form>
					{:else if !authDisabled}
						<div class="settings-auth-callout">
							<p class="settings-note">Sign in as the owner to manage Lorivo settings.</p>
							<p class="settings-auth-callout__detail">{ownerActionDetail}</p>
							<div class="status-actions">
								<LorivoButton variant="primary" size="sm" href="/signin">{ownerActionLabel}</LorivoButton>
								<LorivoButton variant="ghost" size="sm" href="#access">Open Access</LorivoButton>
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
		--settings-accent: #9aa7ff;
		--settings-accent-soft: rgb(154 167 255 / 9%);
		--settings-accent-border: rgb(154 167 255 / 18%);
		display: grid;
		gap: 20px;
		padding-bottom: var(--lorivo-space-8);
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
		font-family: var(--lorivo-font-display);
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
		color: var(--lorivo-color-text);
		font-size: clamp(1.55rem, 1.2vw + 1rem, 2rem);
		font-weight: 820;
		letter-spacing: -0.02em;
		line-height: 1.08;
	}

	.settings-head p {
		margin: 6px 0 0;
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 84%, transparent);
		font-size: 0.9rem;
		line-height: 1.42;
		max-width: 860px;
	}

	.settings-head__meta {
		color: color-mix(in srgb, var(--lorivo-color-text-soft) 90%, transparent);
		font-size: 0.8rem;
		white-space: nowrap;
	}

	.settings-dashboard {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 12px;
	}

	.settings-dashboard-card {
		display: grid;
		gap: 7px;
		min-height: 156px;
		align-content: space-between;
		padding: 15px;
		border: 1px solid color-mix(in srgb, var(--settings-accent-border) 42%, var(--lorivo-color-border-soft));
		border-radius: 12px;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), rgb(255 255 255 / 2%)),
			color-mix(in srgb, var(--lorivo-color-surface-elevated) 96%, #111827 4%);
		color: var(--lorivo-color-text);
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 7%),
			0 14px 32px rgb(0 0 0 / 14%);
	}

	.settings-dashboard-card span {
		color: color-mix(in srgb, var(--lorivo-color-text-muted) 92%, transparent);
		font-size: 0.78rem;
		font-weight: 760;
		letter-spacing: 0.07em;
		text-transform: uppercase;
	}

	.settings-dashboard-card strong {
		overflow: hidden;
		text-overflow: ellipsis;
		color: var(--lorivo-color-text);
		font-size: clamp(1.2rem, 0.6vw + 1rem, 1.6rem);
		font-weight: 820;
		line-height: 1.08;
		overflow-wrap: anywhere;
	}

	.settings-dashboard-card small {
		color: var(--lorivo-color-text-soft);
		font-size: 0.78rem;
		line-height: 1.35;
	}

	.settings-dashboard-card--identity {
		grid-column: span 2;
	}

	.settings-dashboard-card--quiet {
		min-height: 132px;
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
		color: var(--lorivo-color-text-muted);
		font-size: 0.8rem;
		line-height: 1.35;
	}

	.settings-section {
		scroll-margin-top: 18px;
	}

	.settings-subsection {
		display: grid;
		gap: 12px;
		padding: 14px;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: color-mix(in srgb, var(--lorivo-color-surface-elevated) 95%, #101827 5%);
	}

	.settings-subsection--quiet {
		background: color-mix(in srgb, var(--settings-accent-soft) 42%, transparent);
	}

	.settings-subsection--empty {
		align-content: center;
		min-height: 180px;
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
		color: var(--lorivo-color-text-soft);
		font-size: 0.84rem;
		line-height: 1.45;
	}

	:global([data-shell='server'] .settings-panel) {
		border-left: 2px solid rgb(154 167 255 / 18%);
	}

	:global([data-shell='server'] .settings-panel h2) {
		font-size: 1.13rem;
	}

	.stat-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 11px;
	}

	.stat-grid--compact {
		grid-template-columns: repeat(2, minmax(0, 1fr));
	}

	.status-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--lorivo-space-2);
	}

	.settings-feedback {
		margin: 0;
		color: color-mix(in srgb, #b7ffd4 78%, white 22%);
		font-size: 0.82rem;
		line-height: 1.42;
	}

	.library-list {
		display: grid;
		gap: 12px;
	}

	.library-card {
		display: grid;
		gap: 14px;
		padding: 14px;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: color-mix(in srgb, var(--lorivo-color-surface-elevated) 95%, #101827 5%);
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
		color: var(--lorivo-color-text-soft);
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
		border-radius: 999px;
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
		color: var(--lorivo-color-text-muted);
		font-size: 0.76rem;
		font-weight: 700;
	}

	.library-card__facts dd {
		margin: 0;
		color: var(--lorivo-color-text);
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
		padding: 12px 14px;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: color-mix(in srgb, var(--settings-accent-soft) 52%, transparent);
	}

	.settings-auth-callout__detail {
		margin: 0;
		color: var(--lorivo-color-text-muted);
		font-size: 0.88rem;
		line-height: 1.45;
	}

	.settings-placeholder-list {
		display: grid;
		gap: 4px;
		margin: 0;
		padding-left: 18px;
		color: var(--lorivo-color-text-muted);
		font-size: 0.82rem;
		line-height: 1.45;
	}

	.scanning-automation-form,
	.playback-policy-form,
	.storage-settings-form {
		display: grid;
		gap: 12px;
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid var(--lorivo-color-border-soft);
	}

	.settings-advanced {
		padding: 12px;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: color-mix(in srgb, var(--lorivo-color-surface-elevated) 95%, #101827 5%);
	}

	.settings-advanced summary {
		cursor: pointer;
		color: var(--lorivo-color-text);
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
		gap: 10px;
	}

	.playback-policy-option {
		display: grid;
		grid-template-columns: auto minmax(0, 1fr);
		align-items: flex-start;
		gap: 10px;
		padding: 12px;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: color-mix(in srgb, var(--lorivo-color-surface-elevated) 95%, #101827 5%);
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
		color: var(--lorivo-color-text-soft);
		font-size: 0.82rem;
		line-height: 1.4;
	}

	@media (max-width: 1120px) {
		.settings-dashboard {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}

		.library-card__facts {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	.server-name-form {
		display: grid;
		gap: 12px;
		margin-top: 14px;
		padding-top: 14px;
		border-top: 1px solid var(--lorivo-color-border-soft);
	}

	.settings-field {
		display: grid;
		gap: 6px;
	}

	.settings-field span {
		color: var(--lorivo-color-text-muted);
		font-size: 0.8rem;
		font-weight: 700;
	}

	.settings-field input {
		min-height: 42px;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: var(--lorivo-color-surface-elevated);
		color: var(--lorivo-color-text);
		font: inherit;
		padding: 0 12px;
		width: 100%;
	}

	.settings-field input[readonly] {
		background: color-mix(in srgb, var(--lorivo-color-surface-elevated) 92%, #0f172a 8%);
		color: color-mix(in srgb, var(--lorivo-color-text) 86%, transparent);
	}

	.storage-field-list {
		display: grid;
		gap: 12px;
	}

	.storage-field-card {
		display: grid;
		gap: 12px;
		padding: 14px;
		border: 1px solid var(--lorivo-color-border-soft);
		border-radius: var(--lorivo-radius-md);
		background: color-mix(in srgb, var(--lorivo-color-surface-elevated) 95%, #101827 5%);
	}

	.storage-field-card--readonly {
		background: color-mix(in srgb, var(--settings-accent-soft) 36%, transparent);
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
		color: var(--lorivo-color-text-soft);
		font-size: 0.82rem;
		line-height: 1.4;
	}

	.storage-status-pill {
		display: inline-flex;
		align-items: center;
		padding: 4px 8px;
		border-radius: 999px;
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
		color: var(--lorivo-color-text-muted);
		font-size: 0.75rem;
		font-weight: 700;
		text-transform: uppercase;
		letter-spacing: 0.06em;
	}

	.storage-field-facts dd {
		margin: 0;
		color: var(--lorivo-color-text);
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
		color: var(--lorivo-color-text-soft);
	}

	.settings-error {
		color: var(--lorivo-color-danger, #ff9f9f);
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

		.settings-subsection {
			padding: 13px;
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
	}
</style>
