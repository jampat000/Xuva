import { getCollections, type CollectionListItem } from '$lib/api/home';

/**
 * Load collections. SvelteKit calls this before mounting the component, so
 * the page renders instantly on client-side navigation (no onMount flash).
 * The TTL cache in home.ts means repeat visits resolve in <1ms.
 */
export async function load() {
	try {
		const resp = await getCollections();
		return {
			items: resp.collections ?? [] as CollectionListItem[],
			loadError: null as string | null,
		};
	} catch (e) {
		return {
			items: [] as CollectionListItem[],
			loadError: e instanceof Error ? e.message : 'Failed to load collections',
		};
	}
}
