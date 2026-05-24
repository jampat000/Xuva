import { getMovies } from '$lib/api/browse';
import { movieToMedia } from '$lib/api/adapters';

/**
 * Load movie list. SvelteKit calls this before mounting the component, so the
 * page renders instantly with data on client-side navigation (no onMount flash).
 * The TTL cache in browse.ts means repeat visits resolve in <1ms.
 */
export async function load() {
	try {
		const resp = await getMovies();
		return {
			items: (resp.movies ?? []).map(movieToMedia),
			loadError: null as string | null,
		};
	} catch (e) {
		return {
			items: [] as ReturnType<typeof movieToMedia>[],
			loadError: e instanceof Error ? e.message : 'Failed to load movies',
		};
	}
}
