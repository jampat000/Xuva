import type { LibraryRecord } from '$lib/api/home';

export type MediaNavRoute =
	| 'home'
	| 'movies'
	| 'tv'
	| 'collections'
	| 'watchlist'
	| 'continue-watching'
	| 'recently-added';

export interface MediaNavItem {
	id: MediaNavRoute;
	label: string;
	href: string;
}

function normalizeKind(value: unknown): string {
	return String(value ?? '')
		.trim()
		.toLowerCase();
}

export function buildMediaNavItems(libraries: LibraryRecord[] = []): MediaNavItem[] {
	const items: MediaNavItem[] = [{ id: 'home', label: 'Home', href: '/' }];
	const hasMovies = libraries.some((library) => normalizeKind(library.kind) === 'movies');
	const hasTV = libraries.some((library) => {
		const kind = normalizeKind(library.kind);
		return kind === 'tv' || kind === 'series';
	});
	if (hasMovies) items.push({ id: 'movies', label: 'Movies', href: '/movies' });
	if (hasTV) items.push({ id: 'tv', label: 'TV', href: '/tv' });
	return items;
}
