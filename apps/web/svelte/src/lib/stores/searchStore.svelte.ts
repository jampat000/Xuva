/**
 * Search store — calls the backend `/api/client/search` endpoint with a
 * short debounce.  The store caches the latest response so Header and
 * `/search` share the same results when both render concurrently.
 *
 * Exported API kept stable for callers that pre-date the backend search:
 *   - primeSearchCatalogue() is a no-op (kept for compatibility; the
 *     server-side search needs no warmup)
 *   - searchCatalogue(q, limit) returns synchronous cached results for
 *     the last completed query.  Components that want fresh results for
 *     a new query should call `runSearch(q, limit)` to kick off the
 *     fetch, then re-read `searchCatalogue` once `isSearchLoading()` is
 *     false.
 *   - isSearchLoading() reports whether a fetch is in flight.
 *   - getCatalogueSize() is best-effort: it returns the total result
 *     count of the last completed query (or 0).
 */

import { searchLibrary, type SearchHit, type SearchResponse } from '$lib/api/browse';

// ---------------------------------------------------------------------------
// Reactive state
// ---------------------------------------------------------------------------

let lastQuery = $state('');
let lastResponse = $state<SearchResponse | null>(null);
let loading = $state(false);

let debounceTimer: ReturnType<typeof setTimeout> | null = null;
let activeFetchId = 0;

// ---------------------------------------------------------------------------
// Internal: fire the actual fetch
// ---------------------------------------------------------------------------

async function performFetch(q: string, limit: number) {
	const fetchId = ++activeFetchId;
	loading = true;
	try {
		const resp = await searchLibrary(q, limit);
		// Ignore stale responses if a newer fetch has started.
		if (fetchId !== activeFetchId) return;
		lastQuery = q;
		lastResponse = resp;
	} catch {
		if (fetchId !== activeFetchId) return;
		lastQuery = q;
		lastResponse = { query: q, movies: [], series: [], people: [], collections: [] };
	} finally {
		if (fetchId === activeFetchId) {
			loading = false;
		}
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * No-op: kept so existing callers (Header, /search) compile unchanged.
 * The backend search does not need a warmup.
 */
export function primeSearchCatalogue() {
	// nothing to do
}

/**
 * Kick off a backend search for `query`. Debounced by 150ms to avoid a
 * fetch per keystroke. Returns immediately; consumers re-read
 * `getSearchResults` once `isSearchLoading()` flips to false.
 */
export function runSearch(query: string, limit = 8) {
	const q = query.trim();
	if (!q) {
		lastQuery = '';
		lastResponse = null;
		loading = false;
		if (debounceTimer) {
			clearTimeout(debounceTimer);
			debounceTimer = null;
		}
		return;
	}
	if (debounceTimer) clearTimeout(debounceTimer);
	debounceTimer = setTimeout(() => {
		debounceTimer = null;
		performFetch(q, limit);
	}, 150);
}

/**
 * Synchronous accessor: returns the most recent completed search
 * response, or null if no search has been run.  Reactive thanks to
 * Svelte 5 runes.
 */
export function getSearchResults(): SearchResponse | null {
	return lastResponse;
}

/**
 * Return a flat list of hits (movies + series + people + collections)
 * for the most recent completed search, when its query matches `query`.
 * Otherwise returns an empty array and (if the query is non-empty)
 * triggers a backend fetch.
 *
 * Kept for source-compat with the previous client-side store; new
 * callers should prefer `getSearchResults()` so they can render each
 * type with the right card.
 */
export function searchCatalogue(query: string, limit = 8): SearchHit[] {
	const q = query.trim();
	if (!q) return [];
	if (lastQuery !== q) {
		// Trigger a fetch for the new query; current render returns [].
		runSearch(q, limit);
		return [];
	}
	if (!lastResponse) return [];
	const all: SearchHit[] = [
		...lastResponse.movies.slice(0, limit),
		...lastResponse.series.slice(0, limit),
		...lastResponse.people.slice(0, limit),
		...lastResponse.collections.slice(0, limit),
	];
	return all.slice(0, limit);
}

/** True while a fetch is in flight. */
export function isSearchLoading(): boolean {
	return loading;
}

/** Total result count of the last completed query, or 0. */
export function getCatalogueSize(): number {
	if (!lastResponse) return 0;
	return (
		lastResponse.movies.length +
		lastResponse.series.length +
		lastResponse.people.length +
		lastResponse.collections.length
	);
}
