<script lang="ts">
	import { onMount } from 'svelte';
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
		AdminPanel,
		ServerShell,
		LorivoActionList,
		LorivoButton,
		LorivoEmptyState,
		LorivoPanel,
		LorivoStat
	} from '$lib/components';

	let isLoading = $state(true);
	let loadError = $state('');
	let searchValue = $state('');
	let lastUpdatedLabel = $state('');
	let sessionsUnavailable = $state(false);

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

	let refreshTimer: ReturnType<typeof setTimeout> | null = null;

	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
	const queueItems = $derived.by(() => [
		...scans.map((item) => queueListItem('scan', item.id, item.status, item.libraryId || item.kind)),
		...probes.map((item) => queueListItem('probe', item.id, item.status)),
		...work.map((item) => queueListItem('work', item.id, item.status, item.mode)),
		...downloads.map((item) =>
			queueListItem('download', item.id, item.status, item.targetProfile || item.mediaSourceId)
		)
	]);
	const activeQueueCount = $derived.by(
		() => [...scans, ...probes, ...work, ...downloads].filter((item) => isActiveStatus(item.status)).length
	);
	const activeSessionCount = $derived.by(() => sessions.filter((item) => Boolean(asText(item.id))).length);
	const serverStatus = $derived.by(() => {
		const cpu = Number(system.cpu?.percent || 0);
		const memory = Number(system.memory?.usedPercent || 0);
		if (cpu >= 90 || memory >= 92) return 'critical';
		if (cpu >= 75 || memory >= 85) return 'warning';
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
		(settings.libraries || []).map((item) => ({
			id: asText(item.id) || asText(item.path) || asText(item.name),
			label: asText(item.name) || 'Library',
			description: [asText(item.kind), asText(item.storageType), asText(item.path)].filter(Boolean).join(' - ')
		}))
	);

	onMount(() => {
		void loadSettingsSurface();
		const stream = createEventStream();
		stream.connect();
		const unsubscribe = stream.subscribeAny(({ type }) => {
			if (!shouldRefreshForEvent(type)) return;
			queueSilentRefresh();
		});
		return () => {
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
			id: `${kind}-${asText(id) || cryptoSafeId(kind)}`,
			label: `${capitalize(kind)} ${asText(id).slice(0, 8) || 'queue item'}`,
			description: context ? `Context: ${context}` : 'Background queue task.',
			status: normalizedStatus
		};
	}

	function managedStateLabel(item: { configured?: boolean; note?: string }): string {
		if (item.configured === true) return 'Configured';
		if (item.configured === false) return 'Not provisioned';
		return asText(item.note) || 'Available';
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

	function initialsForName(name: string): string {
		const words = asText(name).split(/\s+/).filter(Boolean);
		if (words.length === 0) return 'V';
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
		return asText(value).replace(/\bLorivo\b/g, 'Lorivo');
	}

	function cryptoSafeId(prefix: string): string {
		return `${prefix}-${Math.random().toString(36).slice(2, 8)}`;
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
</script>

<ServerShell
	active="settings"
	bind:searchValue
	{userDisplayName}
	userRole={user?.role || 'Local Account'}
	{userInitials}
>
	<div class="settings-page">
		{#if isLoading}
			<LorivoPanel title="Loading Settings" subtitle="Reading server configuration and operator runtime data." />
		{:else if loadError}
			<LorivoPanel title="Settings could not load" subtitle={loadError}>
				<div class="status-actions">
					<LorivoButton variant="secondary" onclick={() => loadSettingsSurface(false)}>Retry</LorivoButton>
					<LorivoButton variant="ghost" href="/">Back to Home</LorivoButton>
				</div>
			</LorivoPanel>
		{:else}
			<header class="settings-head">
				<div>
					<h1>Settings</h1>
					<p>Review live server state and configuration snapshot. Settings editing is not available yet.</p>
				</div>
				<div class="settings-head__meta">
					<span>Updated {lastUpdatedLabel || '--'}</span>
				</div>
			</header>

			<section id="operator">
			<AdminPanel
				title="Operator Snapshot"
				description="Live runtime telemetry and queue state from existing server APIs."
				status={serverStatus}
			>
				<div class="stat-grid">
					<LorivoStat label="Active Sessions" value={asCount(activeSessionCount)} meta={sessionsUnavailable ? 'Requires authenticated session API access.' : 'Current playback sessions.'} tone={activeSessionCount > 0 ? 'warn' : 'good'} />
					<LorivoStat label="Queue Jobs" value={asCount(activeQueueCount)} meta="Scans, probes, work, and downloads in progress." tone={activeQueueCount > 0 ? 'warn' : 'good'} />
					<LorivoStat label="CPU" value={asPercent(system.cpu?.percent)} meta={`${asCount(system.cpu?.cores)} cores`} tone={Number(system.cpu?.percent || 0) >= 75 ? 'warn' : 'neutral'} />
					<LorivoStat label="Memory" value={asPercent(system.memory?.usedPercent)} meta={`${formatBytes(system.memory?.usedBytes)} used`} tone={Number(system.memory?.usedPercent || 0) >= 85 ? 'warn' : 'neutral'} />
					<LorivoStat label="Catalog" value={`${asCount(summary.movies)} movies`} meta={`${asCount(summary.series)} shows / ${asCount(summary.episodes)} episodes`} />
					<LorivoStat label="Review Needed" value={asCount(health.needsReview)} meta={`${asCount(health.unprobed)} unprobed`} tone={Number(health.needsReview || 0) > 0 ? 'warn' : 'good'} />
				</div>
			</AdminPanel>
			</section>

			<div class="operator-grid">
				<AdminPanel title="Queue Activity" description="Read-only queue stream. Controls are read-only in this build." status={activeQueueCount > 0 ? 'warning' : 'healthy'}>
					<ActivityListShell title="Scans / Probes / Work / Downloads">
						<LorivoActionList items={queueItems.slice(0, 16)} emptyLabel="No queue activity right now." />
					</ActivityListShell>
				</AdminPanel>

				<AdminPanel title="Configuration Snapshot" description="Current runtime profile and playback policy." status="healthy">
					<div class="stat-grid stat-grid--compact">
						<LorivoStat label="Server Name" value={asText(settings.config?.serverName) || 'My Server'} />
						<LorivoStat label="Sync Mode" value={asText(settings.config?.librarySyncMode) || 'unknown'} meta={`Every ${asCount(settings.config?.syncIntervalMins)} minutes`} />
						<LorivoStat
							label="Playback Policy"
							value={displayText(performance.playbackPolicy?.label) || displayText(settings.config?.playbackPolicy) || 'unknown'}
							meta={displayText(performance.playbackPolicy?.description)}
						/>
						<LorivoStat label="Hardware" value={asText(performance.hardwareAcceleration?.status) || 'unknown'} meta={`GPU workers: ${asCount(settings.config?.gpuWorkers)}`} />
					</div>
					<ActivityListShell title="Libraries">
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
				</AdminPanel>
			</div>

			<AdminPanel title="Write Controls" description="This page is read-only in this build." status="idle">
				<div class="status-actions">
					<LorivoButton variant="primary" href="/">Back to Home</LorivoButton>
					<LorivoButton variant="secondary" disabled>Settings editing is not available yet.</LorivoButton>
				</div>
				<LorivoEmptyState
					title="Read-only operator surface"
					message="This settings surface is intentionally read-only in this build. Existing backend APIs and settings contracts remain unchanged."
				/>
			</AdminPanel>
		{/if}
	</div>
</ServerShell>

<style>
	.settings-page {
		display: grid;
		gap: 18px;
		padding-bottom: var(--lorivo-space-8);
		min-width: 0;
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
		font-size: clamp(1.5rem, 1.4vw + 1rem, 2rem);
		letter-spacing: -0.03em;
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

	.operator-grid {
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
		.operator-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 820px) {
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
		.stat-grid,
		.stat-grid--compact {
			grid-template-columns: 1fr;
		}
	}
</style>
