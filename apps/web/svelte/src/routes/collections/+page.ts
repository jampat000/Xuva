import { getCollections } from '$lib/api/home';

/**
 * Pre-warm the collections cache on hover-prefetch so onMount gets an instant hit.
 */
export async function load() {
	try { await getCollections(); } catch { /* component handles its own errors */ }
	return {};
}
