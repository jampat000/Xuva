import { getClientHome } from '$lib/api/home';
import { clientHomeItemToMedia } from '$lib/api/adapters';
import type { Media } from '$lib/mock-data';

/**
 * Load home page data. SvelteKit calls this before mounting the component, so
 * the page renders instantly on client-side navigation (no onMount flash).
 * The TTL cache in home.ts means repeat visits resolve in <1ms.
 */
export async function load() {
	let slides: Media[] = [];
	let continueWatching: Media[] = [];
	let recentMovies: Media[] = [];
	let recentSeries: Media[] = [];
	let topTen: Media[] = [];
	let topRowTitle = 'Highest rated';
	let topRowEyebrow = 'From your library';

	try {
		const resp = await getClientHome();
		if (resp.heroes?.length) {
			slides = resp.heroes.map(clientHomeItemToMedia);
		}
		for (const row of resp.rows ?? []) {
			const items = (row.items ?? []).map(clientHomeItemToMedia);
			const id = (row.id ?? '').toLowerCase();
			const t = (row.title ?? '').toLowerCase();
			if (id === 'continue' || t.includes('continue') || t.includes('watching')) {
				continueWatching = items;
			} else if (id === 'movies' || t.includes('movie')) {
				recentMovies = items;
			} else if (id === 'tv' || t.includes('series') || t.includes('show') || t.includes('episode')) {
				recentSeries = items;
			} else if (id === 'top' || t.includes('top') || t.includes('trend') || t.includes('rated')) {
				topTen = items;
				const r = row as Record<string, unknown>;
				if (r.title) topRowTitle = r.title as string;
				if (r.eyebrow) topRowEyebrow = r.eyebrow as string;
			}
		}
	} catch {
		// Components render gracefully with empty arrays
	}

	return { slides, continueWatching, recentMovies, recentSeries, topTen, topRowTitle, topRowEyebrow };
}
