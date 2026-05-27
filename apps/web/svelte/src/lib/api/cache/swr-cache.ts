/**
 * Stale-while-revalidate cache with IndexedDB persistence.
 *
 * Designed for our list-shaped GET endpoints (/api/movies, /api/series,
 * /api/client/home). Each path key has three states:
 *   • Fresh    — cached value younger than FRESH_TTL_MS, returned as-is.
 *   • Stale    — older than FRESH_TTL_MS but younger than STALE_MAX_AGE_MS:
 *                returned immediately and a background refresh is queued.
 *   • Expired  — older than STALE_MAX_AGE_MS: treated as a cache miss.
 *
 * Persistence: on first access for a key, the in-memory map is hydrated from
 * IndexedDB. Every fetch updates both layers. Subscribers (pages) can register
 * a callback that fires when a background refresh produces fresh data, so the
 * UI can update without the user navigating away and back.
 *
 * The cache is intentionally process-local. Across tabs / windows we let each
 * one keep its own copy; cross-tab sync is delegated to the SSE invalidation
 * stream shipped in a later PR.
 */

import { get as idbGet, set as idbSet, del as idbDel } from 'idb-keyval';

interface CacheEntry<T> {
	data: T;
	/** Wall-clock ms when this entry was last refreshed from the network. */
	fetchedAt: number;
}

export interface SwrOptions {
	/** Within this window, cache is returned without any refresh. */
	freshMs: number;
	/** After this window, cache is discarded and a fetch is awaited. */
	maxAgeMs: number;
}

const DEFAULT_OPTIONS: SwrOptions = {
	freshMs: 5 * 60_000,            // 5 minutes — matches the prior TTL
	maxAgeMs: 24 * 60 * 60_000,     // 24 hours — keep stale-but-usable for a day
};

const IDB_PREFIX = 'xuva:swr:';

const _memory = new Map<string, CacheEntry<unknown>>();
const _hydrated = new Set<string>();
const _refreshInflight = new Map<string, Promise<unknown>>();
const _listeners = new Map<string, Set<(data: unknown) => void>>();

function idbKey(path: string): string {
	return IDB_PREFIX + path;
}

async function hydrateFromIdb<T>(path: string): Promise<CacheEntry<T> | undefined> {
	if (_hydrated.has(path)) return _memory.get(path) as CacheEntry<T> | undefined;
	_hydrated.add(path);
	try {
		const entry = await idbGet<CacheEntry<T>>(idbKey(path));
		if (entry && typeof entry.fetchedAt === 'number') {
			_memory.set(path, entry as CacheEntry<unknown>);
			return entry;
		}
	} catch {
		// IDB unavailable (private mode, SSR, etc.) — fall through to network.
	}
	return undefined;
}

function notify(path: string, data: unknown): void {
	const set = _listeners.get(path);
	if (!set) return;
	for (const fn of set) {
		try { fn(data); } catch { /* listener crash should not break the cache */ }
	}
}

function persist<T>(path: string, entry: CacheEntry<T>): void {
	_memory.set(path, entry as CacheEntry<unknown>);
	void idbSet(idbKey(path), entry).catch(() => { /* IDB write failed — keep memory copy */ });
}

function startRefresh<T>(path: string, fetcher: () => Promise<T>): Promise<T> {
	const existing = _refreshInflight.get(path);
	if (existing) return existing as Promise<T>;
	const promise = fetcher()
		.then((data) => {
			persist(path, { data, fetchedAt: Date.now() });
			notify(path, data);
			return data;
		})
		.finally(() => {
			_refreshInflight.delete(path);
		});
	_refreshInflight.set(path, promise as Promise<unknown>);
	return promise;
}

/**
 * Fetch through the SWR cache. Returns cached data when available (fresh or
 * stale), otherwise awaits the network. Stale returns trigger a background
 * refresh that will fire the per-path listeners on success.
 *
 * @param path        Stable cache key, typically the URL with query string.
 * @param fetcher     Async function that returns the canonical fresh value.
 * @param options     Override TTLs per key.
 */
export async function swrFetch<T>(
	path: string,
	fetcher: () => Promise<T>,
	options: Partial<SwrOptions> = {},
): Promise<T> {
	const { freshMs, maxAgeMs } = { ...DEFAULT_OPTIONS, ...options };
	const now = Date.now();

	let entry = _memory.get(path) as CacheEntry<T> | undefined;
	if (!entry) entry = await hydrateFromIdb<T>(path);

	if (entry) {
		const age = now - entry.fetchedAt;
		if (age < freshMs) return entry.data;
		if (age < maxAgeMs) {
			void startRefresh(path, fetcher).catch(() => { /* surfaced via listeners only */ });
			return entry.data;
		}
		// Expired — fall through to a blocking fetch.
	}

	return startRefresh(path, fetcher);
}

/**
 * Subscribe to fresh data for a path. The callback fires every time a
 * background refresh (or initial fetch) succeeds for this key. Returns an
 * unsubscribe function — call it from onMount cleanup.
 */
export function subscribeSwr<T>(path: string, callback: (data: T) => void): () => void {
	let set = _listeners.get(path);
	if (!set) {
		set = new Set();
		_listeners.set(path, set);
	}
	set.add(callback as (data: unknown) => void);
	return () => {
		const current = _listeners.get(path);
		if (!current) return;
		current.delete(callback as (data: unknown) => void);
		if (current.size === 0) _listeners.delete(path);
	};
}

/**
 * Invalidate a single cache key (memory + IDB). The next swrFetch call will
 * miss and fetch fresh. Pre-empts the TTL and stale-window.
 */
export async function invalidateSwr(path: string): Promise<void> {
	_memory.delete(path);
	_hydrated.delete(path);
	try { await idbDel(idbKey(path)); } catch { /* ignore */ }
}

/**
 * Invalidate every key whose path starts with a prefix (e.g. "/api/movies").
 * Used after library mutations.
 */
export async function invalidateSwrPrefix(prefix: string): Promise<void> {
	const targets: string[] = [];
	for (const key of _memory.keys()) if (key.startsWith(prefix)) targets.push(key);
	for (const key of _hydrated) if (!targets.includes(key) && key.startsWith(prefix)) targets.push(key);
	for (const key of targets) {
		_memory.delete(key);
		_hydrated.delete(key);
		try { await idbDel(idbKey(key)); } catch { /* ignore */ }
	}
}
