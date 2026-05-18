/**
 * Search store — lazily loads all movies + series on first use, then
 * provides fast client-side filtering.  Keeps a single shared copy of the
 * catalogue so navigation between Header and /search doesn't refetch.
 */

import { getMovies, getSeries } from '$lib/api/browse';
import { movieToMedia, seriesToMedia } from '$lib/api/adapters';
import type { Media } from '$lib/mock-data';

// ---------------------------------------------------------------------------
// Shared reactive state (Svelte 5 runes in a .svelte.ts module)
// ---------------------------------------------------------------------------

let catalogue = $state<Media[]>([]);
let catalogueLoaded = $state(false);
let catalogueLoading = $state(false);

async function ensureCatalogue() {
	if (catalogueLoaded || catalogueLoading) return;
	catalogueLoading = true;
	try {
		const [moviesResp, seriesResp] = await Promise.all([
			getMovies().catch(() => ({ movies: [] })),
			getSeries().catch(() => ({ series: [] })),
		]);
		const movies = (moviesResp.movies ?? []).map(movieToMedia);
		const series = (seriesResp.series ?? []).map(seriesToMedia);
		catalogue = [...movies, ...series].sort((a, b) => a.title.localeCompare(b.title));
		catalogueLoaded = true;
	} finally {
		catalogueLoading = false;
	}
}

// ---------------------------------------------------------------------------
// Public API
// ---------------------------------------------------------------------------

/**
 * Trigger catalogue load (idempotent — safe to call multiple times).
 * Call from onMount or when the search input is first focused.
 */
export function primeSearchCatalogue() {
	ensureCatalogue();
}

/**
 * Return search results for `query`.  Returns an empty array if the catalogue
 * is not yet loaded.  Scoring: exact title prefix > any word prefix > contains.
 */
export function searchCatalogue(query: string, limit = 8): Media[] {
	const q = query.trim().toLowerCase();
	if (!q || !catalogueLoaded) return [];

	type Scored = { media: Media; score: number };
	const scored: Scored[] = [];

	for (const m of catalogue) {
		const t = m.title.toLowerCase();
		let score = 0;
		if (t === q) score = 100;
		else if (t.startsWith(q)) score = 80;
		else if (t.split(/\s+/).some((w) => w.startsWith(q))) score = 60;
		else if (t.includes(q)) score = 40;
		else if (m.synopsis?.toLowerCase().includes(q)) score = 10;
		if (score > 0) scored.push({ media: m, score });
	}

	scored.sort((a, b) => b.score - a.score || a.media.title.localeCompare(b.media.title));
	return scored.slice(0, limit).map((s) => s.media);
}

/** Reactive getter — true while first load is in-flight */
export function isSearchLoading(): boolean {
	return catalogueLoading;
}

/** Total catalogue size (useful for the search page) */
export function getCatalogueSize(): number {
	return catalogue.length;
}
