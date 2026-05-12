<script lang="ts">
	import { onMount } from 'svelte';
	import { scanMovies, scanTV, refreshMetadataBatch } from '$lib/api/browse';
	import { getAuthSession, type AuthSessionUser } from '$lib/api/auth';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
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
		type CatalogHealthResponse,
		type CatalogSummaryResponse,
		type DownloadJobItem,
		type PerformanceSettingsResponse,
		type ProbeJobItem,
		type ScanJobItem,
		type SessionItem,
		type SettingsResponse,
		type SystemStatusResponse,
		type WorkQueueItem
	} from '$lib/api/operator';
	import { createEventStream } from '$lib/events/stream';
	import {
		ActivityListShell,
		ServerShell,
		LorivoActionList,
		LorivoButton,
		LorivoEmptyState,
		LorivoPanel,
		LorivoStat
	} from '$lib/components';
	import SettingsPanel from '$lib/components/operator/SettingsPanel.svelte';

	type SettingsSection = 'library' | 'scanning' | 'metadata' | 'playback' | 'server' | 'about';
	type SettingsView = 'overview' | SettingsSection;

	interface BuildInfo {
		buildID?: string;
		publishedAt?: string;
		gitCommit?: string | null;
		sourceApp?: string;
	}

	let isLoading = $state(true);
	let isScanningMovies = $state(false);
	let isScanningTV = $state(false);
	let isRefreshingMovies = $state(false);
	let isRefreshingTV = $state(false);
	let loadError = $state('');
	let searchValue = $state('');
	let actionMessage = $state('');
	let lastUpdatedLabel = $state('');
	let sessionsUnavailable = $state(false);
	let activeSection = $state<SettingsView>('overview');

	let user = $state<AuthSessionUser | null>(null);
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

	let refreshTimer: ReturnType<typeof setTimeout> | null = null;

	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
	const activeQueueCount = $derived.by(
		() => [...scans, ...probes, ...work, ...downloads].filter((item) => isActiveStatus(item.status)).length
	);
	const activeSessionCount = $derived.by(() => sessions.filter((item) => Boolean(asText(item.id))).length);
	const serverStatus = $derived.by(() => {
		const cpu = Number(system.cpu?.percent || 0);
		const memory = Number(system.memory?.usedPercent || 0);
		if (cpu >= 90 || memory >= 92) return 'critical';
		if (cpu >= 75 || memory >= 85 || activeQueueCount > 0) return 'warning';
		if (cpu > 0 || memory > 0) return 'healthy';
		return 'idle';
	});
	const providerStates = $derived.by(() => {
		const automatic = settings.config?.metadataProviders?.automatic || [];
		const managed = settings.config?.metadataProviders?.managedOverrides || [];
		return [...automatic, ...managed].map((item, index) => ({
			id: asText(item.id) || `provider-${index}`,
			name: asText(item.name) || asText(item.id) || 'Provider',
			status: managedStateLabel(item)
		}));
	});
	const libraryRows = $derived.by(() =>
		(settings.libraries || libraries || []).map((item) => ({
			id: asText(item.id) || asText(item.path) || asText(item.name),
			label: asText(item.name) || libraryKindLabel(asText(item.kind)),
			description: [asText(item.kind), asText(item.storageType), asText(item.path)].filter(Boolean).join(' - ')
		}))
	);
	const queueItems = $derived.by(() => [
		...scans.map((item) => queueListItem('Scan', item.id, item.status, item.libraryId || item.kind)),
		...probes.map((item) => queueListItem('Probe', item.id, item.status)),
		...work.map((item) => queueListItem('Work', item.id, item.status, item.mode)),
		...downloads.map((item) =>
			queueListItem('Download', item.id, item.status, item.targetProfile || item.mediaSourceId)
		)
	]);
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
		if (Number(health.needsReview || 0) > 0) {
			output.push({
				id: 'warn-review',
				label: 'Metadata review pending',
				description: `${asCount(health.needsReview)} items need review in catalog health.`,
				status: 'Warning'
			});
		}
		if (Number(health.unprobed || 0) > 0) {
			output.push({
				id: 'warn-unprobed',
				label: 'Unprobed media sources',
				description: `${asCount(health.unprobed)} sources still need probe metadata.`,
				status: 'Warning'
			});
		}
		if (Number(health.unsupported || 0) > 0) {
			output.push({
				id: 'warn-unsupported',
				label: 'Playback compatibility risk',
				description: `${asCount(health.unsupported)} items are flagged unsupported.`,
				status: 'Critical'
			});
		}
		return output;
	});
	const selectedTitle = $derived.by(() => {
		if (activeSection === 'overview') return 'Settings';
		return sectionTitle(activeSection);
	});
	const selectedDescription = $derived.by(() => {
		if (activeSection === 'overview') {
			return 'Choose a settings category to manage Lorivo without leaving the control area.';
		}
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
				getAuthSession(apiClient).catch((error: unknown) => {
					if (isApiStatus(error, 401)) return { user: null };
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

			user = sessionPayload?.user || null;
			libraries = librariesPayload.libraries || [];
			summary = summaryPayload || {};
			health = healthPayload || {};
			system = systemPayload || {};
			settings = settingsPayload || {};
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

	function syncSectionFromHash(): void {
		if (typeof window === 'undefined') return;
		const candidate = window.location.hash.replace(/^#/, '');
		activeSection = isSettingsSection(candidate) ? candidate : 'overview';
	}

	function isSettingsSection(value: string): value is SettingsSection {
		return ['library', 'scanning', 'metadata', 'playback', 'server', 'about'].includes(value);
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

	function queueListItem(
		kind: string,
		id: string | undefined,
		status: string | undefined,
		context = ''
	): { id: string; label: string; description: string; status: string } {
		const normalizedStatus = humanStatus(status);
		return {
			id: `${kind.toLowerCase()}-${asText(id) || randomId(kind)}`,
			label: `${kind} ${asText(id).slice(0, 8) || 'task'}`,
			description: context ? `Context: ${context}` : 'Background queue task.',
			status: normalizedStatus
		};
	}

	function managedStateLabel(item: { configured?: boolean; note?: string }): string {
		if (item.configured === true) return 'Configured';
		if (item.configured === false) return 'Not provisioned';
		return asText(item.note) || 'Available';
	}

	function metadataRefreshMessage(warnings: string[] | undefined, fallback: string): string {
		if (Array.isArray(warnings) && warnings.length > 0) return warnings.slice(0, 3).join(' | ');
		return fallback;
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
		if (normalized === 'in_progress') return 'In Progress';
		return normalized
			.split('_')
			.map((part) => capitalize(part))
			.join(' ');
	}

	function asPercent(value: unknown): string {
		const parsed = Number(value || 0);
		if (!Number.isFinite(parsed)) return '0%';
		return `${Math.round(parsed)}%`;
	}

	function asCount(value: unknown): string {
		const parsed = Number(value || 0);
		if (!Number.isFinite(parsed)) return '0';
		return new Intl.NumberFormat().format(Math.max(0, Math.round(parsed)));
	}

	function formatBytes(value: unknown): string {
		const bytes = Number(value || 0);
		if (!Number.isFinite(bytes) || bytes <= 0) return '0 B';
		const units = ['B', 'KB', 'MB', 'GB', 'TB'];
		let size = bytes;
		let unitIndex = 0;
		while (size >= 1024 && unitIndex < units.length - 1) {
			size /= 1024;
			unitIndex += 1;
		}
		return `${size >= 100 ? Math.round(size) : size.toFixed(1)} ${units[unitIndex]}`;
	}

	function libraryKindLabel(kind: string): string {
		const normalized = asText(kind).toLowerCase();
		if (normalized === 'movies' || normalized === 'movie') return 'Movies';
		if (normalized === 'tv' || normalized === 'series') return 'TV';
		return 'Library';
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

	function sectionTitle(section: SettingsSection): string {
		return {
			library: 'Library',
			scanning: 'Scanning',
			metadata: 'Metadata',
			playback: 'Playback',
			server: 'Server',
			about: 'About'
		}[section];
	}

	function sectionDescription(section: SettingsSection): string {
		return {
			library: 'Media folders, setup status, and the current Library Setup flow.',
			scanning: 'Run library scans and review current scan queue state.',
			metadata: 'Refresh metadata and check provider availability.',
			playback: 'Review playback policy, hardware status, and active sessions.',
			server: 'Check server status, catalog health, queues, and runtime settings.',
			about: 'Lorivo identity, build, and local-first details.'
		}[section];
	}
</script>

<svelte:head>
	<title>Settings - Lorivo</title>
</svelte:head>

<ServerShell
	active={activeSection}
	bind:searchValue
	{userDisplayName}
	userRole={user?.role || 'Local Account'}
	{userInitials}
>
	<div class="settings-page">
		{#if isLoading}
			<LorivoPanel title="Loading Settings" subtitle="Reading server configuration and runtime data." />
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

			{#if activeSection === 'overview'}
				<section class="settings-dashboard" aria-label="Settings dashboard" data-testid="settings-dashboard">
					<a class="settings-dashboard-card" href="#library">
						<span>Library</span>
						<strong>{asCount(libraryRows.length)}</strong>
						<small>{libraryRows.length === 1 ? 'configured library' : 'configured libraries'}</small>
					</a>
					<a class="settings-dashboard-card" href="#scanning">
						<span>Scanning</span>
						<strong>{asCount(scans.length)}</strong>
						<small>{activeQueueCount > 0 ? `${asCount(activeQueueCount)} active tasks` : 'ready for scans'}</small>
					</a>
					<a class="settings-dashboard-card" href="#metadata">
						<span>Metadata</span>
						<strong>{asCount(providerStates.length)}</strong>
						<small>{Number(health.needsReview || 0) > 0 ? 'review needed' : 'providers listed'}</small>
					</a>
					<a class="settings-dashboard-card" href="#playback">
						<span>Playback</span>
						<strong>{asCount(activeSessionCount)}</strong>
						<small>{activeSessionCount === 1 ? 'active session' : 'active sessions'}</small>
					</a>
					<a class="settings-dashboard-card" href="#server">
						<span>Server</span>
						<strong>{capitalize(serverStatus)}</strong>
						<small>{activeQueueCount > 0 ? 'activity in progress' : 'status snapshot'}</small>
					</a>
					<a class="settings-dashboard-card" href="#about">
						<span>About</span>
						<strong>Lorivo</strong>
						<small>{asText(buildInfo?.buildID) || 'build details'}</small>
					</a>
				</section>
			{:else if activeSection === 'library'}
				<section id="library" class="settings-section" data-testid="settings-section-content" data-section="library">
				<SettingsPanel title="Library" description="Media folders and library setup." status={libraryRows.length > 0 ? 'healthy' : 'idle'}>
					{#snippet actions()}
						<LorivoButton variant="primary" href="/setup">Library Setup</LorivoButton>
					{/snippet}
					<ActivityListShell title="Configured Libraries">
						<LorivoActionList
							items={libraryRows.map((item) => ({
								id: item.id,
								label: item.label,
								description: item.description,
								status: 'Configured'
							}))}
							emptyLabel="No libraries configured."
						/>
					</ActivityListShell>
					<LorivoEmptyState
						title="Library folder management"
						message="Use Library Setup to add Movies or TV folders. Editing existing folders is not available here yet."
					/>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'scanning'}
			<section id="scanning" class="settings-section" data-testid="settings-section-content" data-section="scanning">
				<SettingsPanel title="Scanning" description="Start real library scans and review current scan queue state." status={activeQueueCount > 0 ? 'warning' : 'healthy'}>
					{#snippet actions()}
						<LorivoButton variant="primary" onclick={startMovieScan} disabled={isScanningMovies || isScanningTV}>
							{isScanningMovies ? 'Scanning Movies...' : 'Scan Movies'}
						</LorivoButton>
						<LorivoButton variant="secondary" onclick={startTVScan} disabled={isScanningMovies || isScanningTV}>
							{isScanningTV ? 'Scanning TV...' : 'Scan TV'}
						</LorivoButton>
					{/snippet}
					<ActivityListShell title="Scan Queue">
						<LorivoActionList
							items={scans.map((item) => queueListItem('Scan', item.id, item.status, item.libraryId || item.kind))}
							emptyLabel="No scan jobs right now."
						/>
					</ActivityListShell>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'metadata'}
			<section id="metadata" class="settings-section" data-testid="settings-section-content" data-section="metadata">
				<SettingsPanel title="Metadata" description="Refresh metadata and review provider availability." status={Number(health.needsReview || 0) > 0 ? 'warning' : 'healthy'}>
					{#snippet actions()}
						<LorivoButton variant="primary" onclick={refreshMovieMetadata} disabled={isRefreshingMovies || isRefreshingTV}>
							{isRefreshingMovies ? 'Refreshing Movies...' : 'Refresh Movie Metadata'}
						</LorivoButton>
						<LorivoButton variant="secondary" onclick={refreshTVMetadata} disabled={isRefreshingMovies || isRefreshingTV}>
							{isRefreshingTV ? 'Refreshing TV...' : 'Refresh TV Metadata'}
						</LorivoButton>
					{/snippet}
					<div class="stat-grid stat-grid--compact">
						<LorivoStat label="Review Needed" value={asCount(health.needsReview)} meta={`${asCount(health.unprobed)} unprobed`} tone={Number(health.needsReview || 0) > 0 ? 'warn' : 'good'} />
						<LorivoStat label="With Subtitles" value={asCount(health.withSubtitles)} meta="Detected by current catalog health payload." />
					</div>
					<ActivityListShell title="Metadata Providers">
						<LorivoActionList
							items={providerStates.map((provider) => ({
								id: provider.id,
								label: provider.name,
								status: provider.status
							}))}
							emptyLabel="No providers listed by current settings payload."
						/>
					</ActivityListShell>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'playback'}
			<section id="playback" class="settings-section" data-testid="settings-section-content" data-section="playback">
				<SettingsPanel title="Playback" description="Playback policy, hardware status, and active playback sessions." status={activeSessionCount > 0 ? 'warning' : 'healthy'}>
					<div class="stat-grid stat-grid--compact">
						<LorivoStat
							label="Playback Policy"
							value={displayText(performance.playbackPolicy?.label) || displayText(settings.config?.playbackPolicy) || 'Unknown'}
							meta={displayText(performance.playbackPolicy?.description) || 'Current policy from server settings.'}
						/>
						<LorivoStat label="Hardware" value={asText(performance.hardwareAcceleration?.status) || 'Unknown'} meta={`GPU workers: ${asCount(settings.config?.gpuWorkers)}`} />
						<LorivoStat label="Active Sessions" value={asCount(activeSessionCount)} meta={sessionsUnavailable ? 'Session endpoint requires authenticated access.' : 'Current playback sessions.'} tone={activeSessionCount > 0 ? 'warn' : 'good'} />
						<LorivoStat label="Transcode Workers" value={asCount(performance.limits?.transcodeWorkers || settings.config?.transcodeWorkers)} meta="Current runtime limit." />
					</div>
					<ActivityListShell title="Playback Sessions">
						<LorivoActionList
							items={sessionItems}
							emptyLabel={sessionsUnavailable ? 'Session endpoint is protected for this account.' : 'No active playback sessions right now.'}
						/>
					</ActivityListShell>
					<LorivoEmptyState
						title="Playback controls"
						message="Playback policy editing is not available in this build. Current policy and session state are shown from existing server APIs."
					/>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'server'}
			<section id="server" class="settings-section" data-testid="settings-section-content" data-section="server">
				<SettingsPanel title="Server" description="Server status, queue activity, and runtime configuration snapshot." status={serverStatus}>
					<div class="stat-grid">
						<LorivoStat label="Server Name" value={asText(settings.config?.serverName) || 'My Server'} />
						<LorivoStat label="CPU" value={asPercent(system.cpu?.percent)} meta={`${asCount(system.cpu?.cores)} cores`} tone={Number(system.cpu?.percent || 0) >= 75 ? 'warn' : 'neutral'} />
						<LorivoStat label="Memory" value={asPercent(system.memory?.usedPercent)} meta={`${formatBytes(system.memory?.usedBytes)} used`} tone={Number(system.memory?.usedPercent || 0) >= 85 ? 'warn' : 'neutral'} />
						<LorivoStat label="Catalog" value={`${asCount(summary.movies)} movies`} meta={`${asCount(summary.series)} shows / ${asCount(summary.episodes)} episodes`} />
						<LorivoStat label="Queue Jobs" value={asCount(activeQueueCount)} meta="Scans, probes, work, and downloads in progress." tone={activeQueueCount > 0 ? 'warn' : 'good'} />
						<LorivoStat label="Sync Mode" value={asText(settings.config?.librarySyncMode) || 'Unknown'} meta={`Every ${asCount(settings.config?.syncIntervalMins)} minutes`} />
					</div>
					<div class="settings-grid">
						<ActivityListShell title="Queue Activity">
							<LorivoActionList items={queueItems.slice(0, 16)} emptyLabel="No queue activity right now." />
						</ActivityListShell>
						<ActivityListShell title="Server Warnings">
							<LorivoActionList items={warningItems} emptyLabel="No server warnings right now." />
						</ActivityListShell>
					</div>
					<LorivoEmptyState
						title="Controls"
						message="Controls are not available in this build. Server status and queue data are read from existing APIs."
					/>
				</SettingsPanel>
			</section>

			{:else if activeSection === 'about'}
			<section id="about" class="settings-section" data-testid="settings-section-content" data-section="about">
				<SettingsPanel title="About" description="Lorivo build and application identity." status="healthy">
					<div class="stat-grid stat-grid--compact">
						<LorivoStat label="App" value="Lorivo" meta="Local-first personal media server." />
						<LorivoStat label="Build" value={asText(buildInfo?.buildID) || 'Unavailable'} meta={asText(buildInfo?.publishedAt) || 'Build metadata is not available.'} />
						<LorivoStat label="Commit" value={asText(buildInfo?.gitCommit) || 'Unavailable'} meta={asText(buildInfo?.sourceApp) || 'apps/web/svelte'} />
						<LorivoStat label="Mode" value="Local" meta="No cloud account or vendor relay required." />
					</div>
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
		grid-template-columns: repeat(6, minmax(0, 1fr));
		gap: 12px;
	}

	.settings-dashboard-card {
		display: grid;
		gap: 7px;
		min-height: 112px;
		align-content: space-between;
		padding: 15px;
		border: 1px solid color-mix(in srgb, var(--settings-accent-border) 42%, var(--lorivo-color-border-soft));
		border-radius: 12px;
		background:
			linear-gradient(180deg, rgb(255 255 255 / 5%), rgb(255 255 255 / 2%)),
			color-mix(in srgb, var(--lorivo-color-surface-elevated) 96%, #111827 4%);
		color: var(--lorivo-color-text);
		text-decoration: none;
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 7%),
			0 14px 32px rgb(0 0 0 / 14%);
		transition:
			border-color 160ms ease,
			background-color 160ms ease,
			transform 160ms ease,
			box-shadow 160ms ease;
	}

	.settings-dashboard-card:hover,
	.settings-dashboard-card:focus-visible {
		transform: translateY(-1px);
		border-color: var(--settings-accent-border);
		outline: none;
		box-shadow:
			inset 0 1px 0 rgb(255 255 255 / 8%),
			0 18px 42px rgb(30 35 62 / 16%);
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
		white-space: nowrap;
	}

	.settings-dashboard-card small {
		color: var(--lorivo-color-text-soft);
		font-size: 0.78rem;
		line-height: 1.35;
	}

	.settings-section {
		scroll-margin-top: 18px;
	}

	:global([data-shell='server'] .settings-panel) {
		border-left: 2px solid rgb(154 167 255 / 18%);
	}

	:global([data-shell='server'] .settings-panel h2) {
		font-size: 1.13rem;
	}

	.settings-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 15px;
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

	@media (max-width: 1120px) {
		.settings-dashboard {
			grid-template-columns: repeat(3, minmax(0, 1fr));
		}

		.settings-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 820px) {
		.settings-dashboard {
			grid-template-columns: repeat(2, minmax(0, 1fr));
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
	}

	@media (max-width: 560px) {
		.settings-dashboard {
			grid-template-columns: 1fr;
		}

		.stat-grid,
		.stat-grid--compact {
			grid-template-columns: 1fr;
		}
	}
</style>
