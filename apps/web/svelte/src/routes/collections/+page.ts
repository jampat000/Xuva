import { getCollections, type CollectionListItem } from '$lib/api/home';

export interface CollectionsPageData {
	/**
	 * Resolves to the full collection list. Not awaited in load() so SvelteKit
	 * mounts the page immediately with skeleton cards rather than blocking
	 * navigation for the duration of the API call.
	 */
	itemsPromise: Promise<CollectionListItem[]>;
	loadErrorPromise: Promise<string | null>;
}

export function load(): CollectionsPageData {
	const fetched = getCollections().then(
		(resp) => ({ items: resp.collections ?? [] as CollectionListItem[], error: null as string | null }),
		(e: unknown) => ({ items: [] as CollectionListItem[], error: e instanceof Error ? e.message : 'Failed to load collections' }),
	);
	return {
		itemsPromise: fetched.then((r) => r.items),
		loadErrorPromise: fetched.then((r) => r.error),
	};
}
