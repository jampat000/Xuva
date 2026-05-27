import type { ClientHomeResponse } from './home';
import { clientHomeItemToMedia } from './adapters';
import type { Media } from '$lib/mock-data';

export interface HomeRows {
	slides: Media[];
	continueWatching: Media[];
	recentMovies: Media[];
	recentSeries: Media[];
	topTen: Media[];
	topRowTitle: string;
	topRowEyebrow: string;
}

/**
 * Bucket the heterogeneous /api/client/home payload into the named rows the
 * home component renders. Lives outside +page.ts so the SWR subscription can
 * import it without colliding with SvelteKit's strict export rules on route
 * files.
 */
export function normalizeClientHome(resp: ClientHomeResponse): HomeRows {
	const out: HomeRows = {
		slides: [],
		continueWatching: [],
		recentMovies: [],
		recentSeries: [],
		topTen: [],
		topRowTitle: 'Highest rated',
		topRowEyebrow: 'From your library',
	};

	if (resp.heroes?.length) out.slides = resp.heroes.map(clientHomeItemToMedia);

	for (const row of resp.rows ?? []) {
		const items = (row.items ?? []).map(clientHomeItemToMedia);
		const id = (row.id ?? '').toLowerCase();
		const t = (row.title ?? '').toLowerCase();
		if (id === 'continue' || t.includes('continue') || t.includes('watching')) {
			out.continueWatching = items;
		} else if (id === 'movies' || t.includes('movie')) {
			out.recentMovies = items;
		} else if (id === 'tv' || t.includes('series') || t.includes('show') || t.includes('episode')) {
			out.recentSeries = items;
		} else if (id === 'top' || t.includes('top') || t.includes('trend') || t.includes('rated')) {
			out.topTen = items;
			const r = row as Record<string, unknown>;
			if (r.title) out.topRowTitle = r.title as string;
			if (r.eyebrow) out.topRowEyebrow = r.eyebrow as string;
		}
	}

	return out;
}
