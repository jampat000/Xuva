<script lang="ts">
	import { onMount } from 'svelte';
	import { getAuthSession, type AuthSessionUser } from '$lib/api/auth';
	import { ApiClientError, apiClient } from '$lib/api/client';
	import { getLibraries, type LibraryRecord } from '$lib/api/home';
	import {
		getCatalogHealth,
		getCatalogSummary,
		getDownloads,
		getProbes,
		getScans,
		getSessions,
		getSystemStatus,
		getWork,
		type CatalogHealthResponse,
		type CatalogSummaryResponse,
		type DownloadJobItem,
		type ProbeJobItem,
		type ScanJobItem,
		type SessionItem,
		type SystemStatusResponse,
		type WorkQueueItem
	} from '$lib/api/operator';
	import { createEventStream } from '$lib/events/stream';
	import {
		ActivityListShell,
		AdminPanel,
		ServerShell,
		VyrdenActionList,
		VyrdenButton,
		VyrdenEmptyState,
		VyrdenPanel,
		VyrdenStat
	} from '$lib/components';

	let isLoading = $state(true);
	let loadError = $state('');
	let authMessage = $state('');
	let partialLoadMessage = $state('');
	let searchValue = $state('');
	let lastUpdatedLabel = $state('');
	let sessionsUnavailable = $state(false);

	let user = $state<AuthSessionUser | null>(null);
	let libraries = $state<LibraryRecord[]>([]);
	let summary = $state<CatalogSummaryResponse>({});
	let health = $state<CatalogHealthResponse>({});
	let system = $state<SystemStatusResponse>({});
	let scans = $state<ScanJobItem[]>([]);
	let probes = $state<ProbeJobItem[]>([]);
	let work = $state<WorkQueueItem[]>([]);
	let downloads = $state<DownloadJobItem[]>([]);
	let sessions = $state<SessionItem[]>([]);

	let refreshTimer: ReturnType<typeof setTimeout> | null = null;
	const ADMIN_REQUEST_TIMEOUT_MS = 7000;

	const userDisplayName = $derived.by(() => user?.displayName || user?.username || 'Local User');
	const userInitials = $derived.by(() => initialsForName(userDisplayName));
	const queueItems = $derived.by(() => [
		...scans.map((item) => queueListItem('Scan', item.id, item.status, item.libraryId || item.kind)),
		...probes.map((item) => queueListItem('Probe', item.id, item.status)),
		...work.map((item) => queueListItem('Work', item.id, item.status, item.mode)),
		...downloads.map((item) =>
			queueListItem('Download', item.id, item.status, item.targetProfile || item.mediaSourceId)
		)
	]);
	const activeQueueItems = $derived.by(() =>
		queueItems.filter((item) => isActiveStatus(item.rawStatus)).map((item) => ({
			id: item.id,
			label: item.label,
			description: item.description,
			status: item.status
		}))
	);
	const sessionItems = $derived.by(() =>
		sessions.map((item, index) => ({
			id: asText(item.id) || `session-${index}`,
			label: asText(item.title) || asText(item.sourceName) || 'Active playback',
			description: [asText(item.deviceId), asText(item.mode) || asText(item.route)]
				.filter(Boolean)
				.join(' - '),
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
		if (Number(system.cpu?.percent || 0) >= 85) {
			output.push({
				id: 'warn-cpu',
				label: 'High CPU load',
				description: `CPU is at ${asPercent(system.cpu?.percent)}.`,
				status: 'Warning'
			});
		}
		if (Number(system.memory?.usedPercent || 0) >= 90) {
			output.push({
				id: 'warn-memory',
				label: 'High memory pressure',
				description: `Memory is at ${asPercent(system.memory?.usedPercent)}.`,
				status: 'Warning'
			});
		}
		return output;
	});
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

	onMount(() => {
		void loadAdmin();
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

	async function loadAdmin(silent = false): Promise<void> {
		if (!silent) {
			isLoading = true;
			loadError = '';
			authMessage = '';
			partialLoadMessage = '';
		}
		sessionsUnavailable = false;
		try {
			const settled = await Promise.allSettled([
				withTimeout(
					getAuthSession(apiClient).catch((error: unknown) => {
						if (isApiStatus(error, 401)) return { user: null };
						throw error;
					}),
					'Session'
				),
				withTimeout(getLibraries(apiClient), 'Libraries'),
				withTimeout(getCatalogSummary(apiClient), 'Catalog summary'),
				withTimeout(getCatalogHealth(apiClient), 'Catalog health'),
				withTimeout(getSystemStatus(apiClient), 'System status'),
				withTimeout(getScans(apiClient), 'Scan queue'),
				withTimeout(getProbes(apiClient), 'Probe queue'),
				withTimeout(getWork(apiClient), 'Work queue'),
				withTimeout(getDownloads(apiClient), 'Download queue'),
				withTimeout(
					getSessions(apiClient).catch((error: unknown) => {
						if (isApiStatus(error, 401)) {
							sessionsUnavailable = true;
							return { sessions: [] };
						}
						throw error;
					}),
					'Sessions'
				)
			]);

			const sessionPayload = unwrapSettled(settled[0], { user: null }, 'Session');
			const librariesPayload = unwrapSettled(settled[1], { libraries: [] }, 'Libraries');
			const summaryPayload = unwrapSettled(settled[2], {}, 'Catalog summary');
			const healthPayload = unwrapSettled(settled[3], {}, 'Catalog health');
			const systemPayload = unwrapSettled(settled[4], {}, 'System status');
			const scansPayload = unwrapSettled(settled[5], { scans: [] }, 'Scan queue');
			const probesPayload = unwrapSettled(settled[6], { probes: [] }, 'Probe queue');
			const workPayload = unwrapSettled(settled[7], { work: [] }, 'Work queue');
			const downloadsPayload = unwrapSettled(settled[8], { downloads: [] }, 'Download queue');
			const sessionsPayload = unwrapSettled(settled[9], { sessions: [] }, 'Sessions');

			const authFailureCount = settled.reduce((count, result) => {
				if (result.status !== 'rejected') return count;
				return count + (isApiStatus(result.reason, 401) ? 1 : 0);
			}, 0);
			if (authFailureCount >= 3 && !sessionPayload?.user) {
				authMessage = 'Your session has expired. Sign in again to view live operations.';
				lastUpdatedLabel = '';
				return;
			}

			const failedCount = settled.filter((result) => result.status === 'rejected').length;
			partialLoadMessage =
				failedCount > 0
					? `${failedCount} dashboard request${failedCount === 1 ? '' : 's'} could not be loaded. Showing available data.`
					: '';

			user = sessionPayload?.user || null;
			libraries = librariesPayload.libraries || [];
			summary = summaryPayload || {};
			health = healthPayload || {};
			system = systemPayload || {};
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

	function withTimeout<T>(promise: Promise<T>, label: string, timeoutMs = ADMIN_REQUEST_TIMEOUT_MS): Promise<T> {
		return new Promise<T>((resolve, reject) => {
			const timer = setTimeout(() => {
				reject(new Error(`${label} request timed out.`));
			}, timeoutMs);
			promise.then(
				(value) => {
					clearTimeout(timer);
					resolve(value);
				},
				(error) => {
					clearTimeout(timer);
					reject(error);
				}
			);
		});
	}

	function unwrapSettled<T>(
		result: PromiseSettledResult<T>,
		fallback: T,
		label: string
	): T {
		if (result.status === 'fulfilled') return result.value;
		if (isApiStatus(result.reason, 401)) return fallback;
		console.warn(`[admin] ${label} unavailable`, result.reason);
		return fallback;
	}

	function queueSilentRefresh(): void {
		if (refreshTimer) clearTimeout(refreshTimer);
		refreshTimer = setTimeout(() => {
			refreshTimer = null;
			void loadAdmin(true);
		}, 200);
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
		labelPrefix: string,
		id: string | undefined,
		status: string | undefined,
		context = ''
	): { id: string; label: string; description: string; status: string; rawStatus: string } {
		const rawStatus = asText(status).toLowerCase();
		const prettyStatus = humanStatus(rawStatus);
		return {
			id: `${labelPrefix.toLowerCase()}-${asText(id) || randomId(labelPrefix)}`,
			label: `${labelPrefix} ${asText(id).slice(0, 8) || 'task'}`,
			description: context ? `Context: ${context}` : 'Queue entry',
			status: prettyStatus,
			rawStatus
		};
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

	function asText(value: unknown): string {
		return String(value ?? '').trim();
	}

	function randomId(prefix: string): string {
		return `${prefix.toLowerCase()}-${Math.random().toString(36).slice(2, 8)}`;
	}

	function capitalize(value: string): string {
		if (!value) return value;
		return `${value.slice(0, 1).toUpperCase()}${value.slice(1)}`;
	}

	function initialsForName(name: string): string {
		const words = asText(name).split(/\s+/).filter(Boolean);
		if (words.length === 0) return 'V';
		if (words.length === 1) return words[0].slice(0, 1).toUpperCase();
		return `${words[0][0] || ''}${words[1][0] || ''}`.toUpperCase();
	}

	function formatLoadError(error: unknown): string {
		if (error instanceof ApiClientError) return error.userMessage || error.message;
		if (isApiStatus(error, 401)) return 'Your session is no longer active. Sign in again to continue.';
		if (error instanceof Error) return error.message;
		return 'Admin dashboard could not load.';
	}

	function isApiStatus(error: unknown, expectedStatus: number): boolean {
		if (error instanceof ApiClientError) return error.status === expectedStatus;
		if (typeof error !== 'object' || !error) return false;
		const candidate = (error as { status?: unknown }).status;
		return Number(candidate) === expectedStatus;
	}
</script>

<ServerShell
	active="admin"
	bind:searchValue
	{userDisplayName}
	userRole={user?.role || 'Local Account'}
	{userInitials}
>
	<div class="admin-page">
		{#if isLoading}
			<VyrdenPanel title="Loading Admin Dashboard" subtitle="Reading live operations from the server APIs." />
		{:else if authMessage}
			<VyrdenPanel title="Sign in required" subtitle={authMessage}>
				<div class="status-actions">
					<VyrdenButton variant="secondary" onclick={() => loadAdmin(false)}>Retry</VyrdenButton>
					<VyrdenButton variant="ghost" href="/signin">Open Sign In</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else if loadError}
			<VyrdenPanel title="Admin dashboard could not load" subtitle={loadError}>
				<div class="status-actions">
					<VyrdenButton variant="secondary" onclick={() => loadAdmin(false)}>Retry</VyrdenButton>
					<VyrdenButton variant="ghost" href="/settings">Open Settings</VyrdenButton>
				</div>
			</VyrdenPanel>
		{:else}
			<header class="admin-head">
				<div>
					<h1>Admin Dashboard</h1>
					<p>Live operations, queue activity, session state, and server health.</p>
				</div>
				<div class="admin-head__meta">
					<span>Updated {lastUpdatedLabel || '--'}</span>
				</div>
			</header>

			{#if partialLoadMessage}
				<VyrdenPanel title="Partial data" subtitle={partialLoadMessage}>
					<p class="partial-copy">Live tiles continue to update as APIs recover.</p>
				</VyrdenPanel>
			{/if}

			<AdminPanel
				title="Live Server Status"
				description="Operational telemetry surface for operators. This route is read-only in this build."
				status={serverStatus}
			>
				<div class="stat-grid">
					<VyrdenStat label="Active Sessions" value={asCount(activeSessionCount)} meta={sessionsUnavailable ? 'Session API requires authenticated access.' : 'Current playback sessions.'} tone={activeSessionCount > 0 ? 'warn' : 'good'} />
					<VyrdenStat label="Active Queue Jobs" value={asCount(activeQueueCount)} meta="Scans, probes, work, and downloads." tone={activeQueueCount > 0 ? 'warn' : 'good'} />
					<VyrdenStat label="CPU" value={asPercent(system.cpu?.percent)} meta={`${asCount(system.cpu?.cores)} cores`} tone={Number(system.cpu?.percent || 0) >= 75 ? 'warn' : 'neutral'} />
					<VyrdenStat label="Memory" value={asPercent(system.memory?.usedPercent)} meta={`${asCount(system.memory?.usedBytes)} bytes used`} tone={Number(system.memory?.usedPercent || 0) >= 85 ? 'warn' : 'neutral'} />
					<VyrdenStat label="Catalog" value={`${asCount(summary.mediaSources)} sources`} meta={`${asCount(summary.movies)} movies / ${asCount(summary.series)} shows`} />
					<VyrdenStat label="Needs Review" value={asCount(health.needsReview)} meta={`${asCount(health.unprobed)} unprobed`} tone={Number(health.needsReview || 0) > 0 ? 'warn' : 'good'} />
				</div>
			</AdminPanel>

			<div class="ops-grid">
				<AdminPanel title="Current Playback Sessions" description="Active sessions and routes." status={activeSessionCount > 0 ? 'warning' : 'healthy'}>
					<ActivityListShell title="Sessions">
						<VyrdenActionList
							items={sessionItems}
							emptyLabel={sessionsUnavailable
								? 'Session endpoint is protected for this account.'
								: 'No active playback sessions right now.'}
						/>
					</ActivityListShell>
				</AdminPanel>

				<AdminPanel title="Queue Activity" description="Live queue stream from scan/probe/work/download APIs." status={activeQueueCount > 0 ? 'warning' : 'healthy'}>
					<ActivityListShell title="Active Queue Entries">
						<VyrdenActionList items={activeQueueItems} emptyLabel="No active queue jobs right now." />
					</ActivityListShell>
				</AdminPanel>
			</div>

			<div class="ops-grid">
				<AdminPanel title="Operational Warnings" description="Warnings from catalog and runtime signals." status={warningItems.length > 0 ? 'warning' : 'healthy'}>
					<ActivityListShell title="Warnings">
						<VyrdenActionList items={warningItems} emptyLabel="No operational warnings right now." />
					</ActivityListShell>
				</AdminPanel>

				<AdminPanel title="Admin Controls" description="Admin controls are read-only in this build." status="idle">
					<div class="status-actions">
						<VyrdenButton variant="primary" href="/settings">Open Settings</VyrdenButton>
						<VyrdenButton variant="secondary" disabled>Admin controls are read-only in this build.</VyrdenButton>
					</div>
					<VyrdenEmptyState
						title="Read-only operator view"
						message="Write operations remain intentionally unavailable in this build to keep backend behavior unchanged."
					/>
				</AdminPanel>
			</div>
		{/if}
	</div>
</ServerShell>

<style>
	.admin-page {
		display: grid;
		gap: 18px;
		padding-bottom: var(--vyrden-space-8);
		min-width: 0;
	}

	.admin-head {
		display: flex;
		align-items: flex-end;
		justify-content: space-between;
		gap: 12px;
	}

	.admin-head h1 {
		margin: 0;
		font-family: var(--vyrden-font-display);
		font-size: clamp(1.5rem, 1.4vw + 1rem, 2rem);
		letter-spacing: -0.03em;
	}

	.admin-head p {
		margin: 6px 0 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 84%, transparent);
		font-size: 0.9rem;
		line-height: 1.42;
	}

	.admin-head__meta {
		color: color-mix(in srgb, var(--vyrden-color-text-soft) 90%, transparent);
		font-size: 0.8rem;
		white-space: nowrap;
	}

	.stat-grid {
		display: grid;
		grid-template-columns: repeat(3, minmax(0, 1fr));
		gap: 11px;
	}

	.ops-grid {
		display: grid;
		grid-template-columns: repeat(2, minmax(0, 1fr));
		gap: 15px;
	}

	.status-actions {
		display: flex;
		flex-wrap: wrap;
		gap: var(--vyrden-space-2);
	}

	.partial-copy {
		margin: 0;
		color: color-mix(in srgb, var(--vyrden-color-text-muted) 84%, transparent);
		font-size: 0.86rem;
	}

	@media (max-width: 1120px) {
		.ops-grid {
			grid-template-columns: 1fr;
		}
	}

	@media (max-width: 820px) {
		.admin-head {
			flex-direction: column;
			align-items: flex-start;
		}

		.admin-head__meta {
			white-space: normal;
		}

		.stat-grid {
			grid-template-columns: repeat(2, minmax(0, 1fr));
		}
	}

	@media (max-width: 560px) {
		.stat-grid {
			grid-template-columns: 1fr;
		}
	}
</style>

