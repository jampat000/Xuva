import { getClientHome } from '$lib/api/home';
import { normalizeClientHome } from '$lib/api/home-normalize';

/**
 * Load home page data. SvelteKit calls this before mounting the component, so
 * the page renders instantly on client-side navigation (no onMount flash).
 * The SWR cache in home.ts means repeat visits resolve in <1ms; background-
 * refresh updates flow to the page via subscribeClientHome() in +page.svelte.
 */
export async function load() {
	try {
		const resp = await getClientHome();
		return normalizeClientHome(resp);
	} catch {
		// Components render gracefully with empty arrays
		return normalizeClientHome({});
	}
}
