/**
 * Watchlist store — server-authoritative with optimistic local updates.
 * Falls back to localStorage when the server is unavailable.
 */

import {
	addToServerWatchlist,
	getServerWatchlist,
	removeFromServerWatchlist,
	type WatchlistAddRequest
} from '$lib/api/home';

export interface WatchlistItem {
	id: string;
	kind: 'movie' | 'series';
	title: string;
	year?: number;
	posterUrl?: string;
	backdropUrl?: string;
	genres?: string[];
	addedAt: string;
}

const STORAGE_KEY = 'xuva-watchlist';

function loadFromStorage(): WatchlistItem[] {
	if (typeof localStorage === 'undefined') return [];
	try {
		const raw = localStorage.getItem(STORAGE_KEY);
		return raw ? (JSON.parse(raw) as WatchlistItem[]) : [];
	} catch {
		return [];
	}
}

function persist(items: WatchlistItem[]) {
	if (typeof localStorage === 'undefined') return;
	try {
		localStorage.setItem(STORAGE_KEY, JSON.stringify(items));
	} catch {
		// Quota exceeded or private-browsing restriction — fail silently
	}
}

let items = $state<WatchlistItem[]>(loadFromStorage());
let synced = false;

// ── Server sync ───────────────────────────────────────────────────────────────

/** Load server watchlist once per session and replace the local cache. */
export async function syncWatchlistFromServer(): Promise<void> {
	if (synced) return;
	try {
		const { items: serverItems } = await getServerWatchlist();
		items = serverItems.map((i) => ({
			id: i.mediaId,
			kind: i.kind,
			title: i.title,
			year: i.year,
			posterUrl: i.posterUrl,
			backdropUrl: i.backdropUrl,
			genres: i.genres,
			addedAt: i.addedAt
		}));
		persist(items);
		synced = true;
	} catch {
		// Server unreachable — keep localStorage state
	}
}

// ── Public API ────────────────────────────────────────────────────────────────

/** Add an item with an optimistic local update, then sync to server. */
export function addToWatchlist(item: Omit<WatchlistItem, 'addedAt'>) {
	if (items.some((i) => i.id === item.id && i.kind === item.kind)) return;
	const newItem: WatchlistItem = { ...item, addedAt: new Date().toISOString() };
	items = [newItem, ...items];
	persist(items);

	const req: WatchlistAddRequest = {
		mediaId: item.id,
		kind: item.kind,
		title: item.title,
		year: item.year,
		posterUrl: item.posterUrl,
		backdropUrl: item.backdropUrl,
		genres: item.genres
	};
	addToServerWatchlist(req).catch(() => {
		// Server write failed — optimistic state stays locally
	});
}

/** Remove an item with an optimistic local update, then sync to server. */
export function removeFromWatchlist(id: string, kind: 'movie' | 'series') {
	items = items.filter((i) => !(i.id === id && i.kind === kind));
	persist(items);
	removeFromServerWatchlist(id, kind).catch(() => {
		// Server write failed — optimistic state stays locally
	});
}

/** Toggle: adds if absent, removes if present. Returns true when added. */
export function toggleWatchlist(item: Omit<WatchlistItem, 'addedAt'>): boolean {
	const exists = isInWatchlist(item.id, item.kind);
	if (exists) {
		removeFromWatchlist(item.id, item.kind);
	} else {
		addToWatchlist(item);
	}
	return !exists;
}

/** Reactive check — safe to call inside $derived / template expressions. */
export function isInWatchlist(id: string, kind: 'movie' | 'series'): boolean {
	return items.some((i) => i.id === id && i.kind === kind);
}

/** Reactive getter — returns the current watchlist (most-recently-added first). */
export function getWatchlist(): WatchlistItem[] {
	return items;
}

/** Reactive count. */
export function getWatchlistCount(): number {
	return items.length;
}
