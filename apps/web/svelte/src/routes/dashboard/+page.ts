import {
	getSystemStatus,
	getCatalogHealth,
	getSessions,
	getJobs,
	getPerformanceSettings,
	type SystemStatusResponse,
	type CatalogHealthResponse,
	type SessionItem,
	type JobsStatusResponse,
	type PerformanceSettingsResponse,
} from '$lib/api/operator';

/**
 * Pre-load all dashboard data before mounting. Each call is independent;
 * failures are swallowed so partial data still renders.
 */
export async function load() {
	const [sysR, healthR, sessR, jobsR, perfR] = await Promise.allSettled([
		getSystemStatus(),
		getCatalogHealth(),
		getSessions(),
		getJobs(),
		getPerformanceSettings(),
	]);
	return {
		sys:      sysR.status    === 'fulfilled' ? sysR.value              : null as SystemStatusResponse | null,
		health:   healthR.status === 'fulfilled' ? healthR.value           : null as CatalogHealthResponse | null,
		sessions: sessR.status   === 'fulfilled' ? (sessR.value.sessions ?? []) : [] as SessionItem[],
		jobs:     jobsR.status   === 'fulfilled' ? jobsR.value             : null as JobsStatusResponse | null,
		perf:     perfR.status   === 'fulfilled' ? perfR.value             : null as PerformanceSettingsResponse | null,
	};
}
