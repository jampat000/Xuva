import { getMovies } from '$lib/api/browse';
import { movieToMedia } from '$lib/api/adapters';
import type { Media } from '$lib/mock-data';

export interface MoviesPageData {
	/**
	 * Resolves to the full library. Deliberately NOT awaited in load() —
	 * SvelteKit awaits returned values before mounting the page, which blocks
	 * navigation for the duration of the API call. Returning the unresolved
	 * promise lets +page.svelte mount immediately with a skeleton; the grid
	 * fills in once the promise lands.
	 *
	 * History: we used to do a two-step fetch — limit=60 first, then a
	 * background limit=0 — because /api/movies was 2.4s on a 4000-item
	 * library and the small first page felt snappier. After the
	 * movies_list_view snapshot (PR #355), limit=0 is ~50ms, so the staged
	 * fetch causes a useless double-paint on every repeat visit (60 items,
	 * then 4000) without buying any first-visit speed. One fetch + one
	 * paint is now strictly better.
	 */
	itemsPromise: Promise<Media[]>;

	/**
	 * Same pattern for surface errors so the component can react with a
	 * retry UI rather than crashing the load.
	 */
	loadErrorPromise: Promise<string | null>;
}

export function load(): MoviesPageData {
	const fetched = getMovies(undefined, 0).then(
		(resp) => ({ items: (resp.movies ?? []).map(movieToMedia), error: null as string | null }),
		(e: unknown) => ({ items: [] as Media[], error: e instanceof Error ? e.message : 'Failed to load movies' }),
	);
	return {
		itemsPromise: fetched.then((r) => r.items),
		loadErrorPromise: fetched.then((r) => r.error),
	};
}
