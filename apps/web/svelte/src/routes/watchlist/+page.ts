import { getServerWatchlist } from '$lib/api/home';

/**
 * Pre-warm the server watchlist on hover-prefetch so onMount gets an instant hit.
 */
export async function load() {
	try { await getServerWatchlist(); } catch { /* component handles its own errors */ }
	return {};
}
