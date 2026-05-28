import { getMovies } from '$lib/api/browse';
import { movieToMedia } from '$lib/api/adapters';
import type { Media } from '$lib/mock-data';

const FIRST_PAGE = 60;

export interface MoviesPageData {
	/**
	 * Resolves to the first page of movies. We deliberately do NOT await this
	 * in load() — SvelteKit awaits returned values before mounting the page,
	 * which blocks navigation for the duration of the API call. On a
	 * 4000-item library the limit=60 query alone can take several seconds,
	 * during which the previous page stays on screen ("nothing happens").
	 *
	 * Returning the unresolved promise lets +page.svelte mount immediately
	 * with a skeleton; the grid fills in once the promise lands.
	 */
	itemsPromise: Promise<Media[]>;

	/**
	 * Same pattern for surface errors so the component can react with a
	 * retry UI rather than crashing the load.
	 */
	loadErrorPromise: Promise<string | null>;
}

/**
 * Non-blocking load — see comments on MoviesPageData. The TTL/SWR cache in
 * browse.ts means second-visit and beyond resolve in <1ms anyway, so this
 * change is purely about preventing the first cold visit from feeling like
 * a hang.
 */
export function load(): MoviesPageData {
	const fetched = getMovies(undefined, FIRST_PAGE).then(
		(resp) => ({ items: (resp.movies ?? []).map(movieToMedia), error: null as string | null }),
		(e: unknown) => ({ items: [] as Media[], error: e instanceof Error ? e.message : 'Failed to load movies' }),
	);
	return {
		itemsPromise: fetched.then((r) => r.items),
		loadErrorPromise: fetched.then((r) => r.error),
	};
}
