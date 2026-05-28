import { getSeries } from '$lib/api/browse';
import { seriesToMedia } from '$lib/api/adapters';
import type { Media } from '$lib/mock-data';

const FIRST_PAGE = 60;

export interface TVPageData {
	itemsPromise: Promise<Media[]>;
	loadErrorPromise: Promise<string | null>;
}

/**
 * Non-blocking load. See routes/movies/+page.ts for rationale — same
 * pattern (return unresolved promises so the page mounts instantly on
 * click; the grid fills in when the API responds).
 */
export function load(): TVPageData {
	const fetched = getSeries(undefined, FIRST_PAGE).then(
		(resp) => ({ items: (resp.series ?? []).map(seriesToMedia), error: null as string | null }),
		(e: unknown) => ({ items: [] as Media[], error: e instanceof Error ? e.message : 'Failed to load TV shows' }),
	);
	return {
		itemsPromise: fetched.then((r) => r.items),
		loadErrorPromise: fetched.then((r) => r.error),
	};
}
