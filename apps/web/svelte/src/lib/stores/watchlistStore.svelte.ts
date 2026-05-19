/**
 * Watchlist store — persists to localStorage so saves survive page refreshes.
 * Usage:
 *   import { toggleWatchlist, isInWatchlist, getWatchlist } from '$lib/stores/watchlistStore.svelte';
 */

export interface WatchlistItem {
	id: string;
	kind: 'movie' | 'series';
	title: string;
	year?: number;
	posterUrl?: string;
	backdropUrl?: string;
	genres?: string[];
	addedAt: string; // ISO timestamp
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

// Shared reactive state
let items = $state<WatchlistItem[]>([]);
let initialized = $state(false);

function init() {
	if (initialized) return;
	items = loadFromStorage();
	initialized = true;
}

// ── Public API ────────────────────────────────────────────────────────────────

/** Add an item. No-op if already present. */
export function addToWatchlist(item: Omit<WatchlistItem, 'addedAt'>) {
	init();
	if (items.some((i) => i.id === item.id && i.kind === item.kind)) return;
	items = [{ ...item, addedAt: new Date().toISOString() }, ...items];
	persist(items);
}

/** Remove an item. No-op if not present. */
export function removeFromWatchlist(id: string, kind: 'movie' | 'series') {
	init();
	items = items.filter((i) => !(i.id === id && i.kind === kind));
	persist(items);
}

/** Toggle: adds if absent, removes if present. Returns true when added. */
export function toggleWatchlist(item: Omit<WatchlistItem, 'addedAt'>): boolean {
	init();
	const exists = isInWatchlist(item.id, item.kind);
	if (exists) {
		removeFromWatchlist(item.id, item.kind);
	} else {
		addToWatchlist(item);
	}
	return !exists;
}

/** Reactive check — use inside $derived / template for live updates. */
export function isInWatchlist(id: string, kind: 'movie' | 'series'): boolean {
	init();
	return items.some((i) => i.id === id && i.kind === kind);
}

/** Reactive getter — returns the current watchlist (most-recently-added first). */
export function getWatchlist(): WatchlistItem[] {
	init();
	return items;
}

/** Reactive count. */
export function getWatchlistCount(): number {
	init();
	return items.length;
}
