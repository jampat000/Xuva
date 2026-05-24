import { getSeries } from '$lib/api/browse';
import { seriesToMedia } from '$lib/api/adapters';

/**
 * Load series list. SvelteKit calls this before mounting the component, so the
 * page renders instantly with data on client-side navigation (no onMount flash).
 * The TTL cache in browse.ts means repeat visits resolve in <1ms.
 */
export async function load() {
	try {
		const resp = await getSeries();
		return {
			items: (resp.series ?? []).map(seriesToMedia),
			loadError: null as string | null,
		};
	} catch (e) {
		return {
			items: [] as ReturnType<typeof seriesToMedia>[],
			loadError: e instanceof Error ? e.message : 'Failed to load TV shows',
		};
	}
}
