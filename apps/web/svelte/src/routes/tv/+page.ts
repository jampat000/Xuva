import { getSeries } from '$lib/api/browse';
import { seriesToMedia } from '$lib/api/adapters';
import type { Media } from '$lib/mock-data';

export interface TVPageData {
	itemsPromise: Promise<Media[]>;
	loadErrorPromise: Promise<string | null>;
}

/**
 * Non-blocking load — see routes/movies/+page.ts for the full rationale.
 * Single fetch of the entire series list; the page mounts immediately
 * against an unresolved promise. The two-stage fetch the movies page used
 * to have (and this one mirrored) is gone now that the tv_series_list_view
 * snapshot (PR #356) makes /api/series fast for any limit.
 */
export function load(): TVPageData {
	const fetched = getSeries(undefined, 0).then(
		(resp) => ({ items: (resp.series ?? []).map(seriesToMedia), error: null as string | null }),
		(e: unknown) => ({ items: [] as Media[], error: e instanceof Error ? e.message : 'Failed to load TV shows' }),
	);
	return {
		itemsPromise: fetched.then((r) => r.items),
		loadErrorPromise: fetched.then((r) => r.error),
	};
}
