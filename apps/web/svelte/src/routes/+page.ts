import { getClientHome } from '$lib/api/home';

/**
 * Pre-warm the home page cache so that onMount gets an instant cache hit.
 * SvelteKit calls this during hover-prefetch (data-sveltekit-preload-data="hover"
 * on the body), meaning the fetch starts ~200ms before the user clicks the link.
 * Errors are swallowed — the component handles them via its own error state.
 */
export async function load() {
	try { await getClientHome(); } catch { /* component handles its own errors */ }
	return {};
}
