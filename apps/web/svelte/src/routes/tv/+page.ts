import { getSeries } from '$lib/api/browse';
import { seriesToMedia } from '$lib/api/adapters';

/**
 * Fast first-paint pattern: fetch the first FIRST_PAGE items synchronously so
 * SvelteKit can mount the page immediately. The rest of the library is
 * background-fetched in +page.svelte and merged into the grid when it lands.
 *
 * The TTL cache in browse.ts keys on the full URL (including `limit`), so the
 * first-page and full-list responses are cached independently. Repeat visits
 * hit both caches in <1ms.
 */
const FIRST_PAGE = 60;

export async function load() {
	try {
		const resp = await getSeries(undefined, FIRST_PAGE);
		const items = (resp.series ?? []).map(seriesToMedia);
		return {
			items,
			hasMore: items.length >= FIRST_PAGE,
			loadError: null as string | null,
		};
	} catch (e) {
		return {
			items: [] as ReturnType<typeof seriesToMedia>[],
			hasMore: false,
			loadError: e instanceof Error ? e.message : 'Failed to load TV shows',
		};
	}
}
