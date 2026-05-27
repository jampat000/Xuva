/// <reference types="@sveltejs/kit" />
/// <reference no-default-lib="true" />
/// <reference lib="esnext" />
/// <reference lib="webworker" />

import { build, files, version } from '$service-worker';

/**
 * App-shell service worker.
 *
 * SvelteKit auto-registers this on the client when the file exists at
 * src/service-worker.ts. The `$service-worker` virtual module exposes:
 *   - build:   hashed runtime bundles (JS/CSS) emitted by the static build
 *   - files:   static assets that ship under /static (favicons, manifest, …)
 *   - version: a unique build identifier — used to bust caches on upgrade
 *
 * Strategy:
 *   - App shell (build + files):        cache-first, fall back to network.
 *   - Navigations (HTML):               network-first, fall back to cached
 *                                       index.html so cold/offline opens still
 *                                       render. SvelteKit's adapter-static
 *                                       ships everything as an SPA off `/`.
 *   - API GETs to list/home endpoints:  stale-while-revalidate. The page-level
 *                                       SWR (lib/api/cache/swr-cache) still
 *                                       drives UI updates; this is a second
 *                                       belt-and-braces layer that survives
 *                                       full reloads and crashes.
 *   - Everything else:                  pass through unchanged.
 *
 * On version bump, the install handler precaches the new shell, the activate
 * handler deletes older caches, and `clients.claim()` immediately swaps the
 * controller so tabs don't need a refresh.
 */

declare const self: ServiceWorkerGlobalScope;

const CACHE_PREFIX = 'xuva';
const SHELL_CACHE = `${CACHE_PREFIX}-shell-${version}`;
const API_CACHE = `${CACHE_PREFIX}-api-${version}`;

const PRECACHE_URLS = [...build, ...files];

/** Paths whose GET responses we cache with stale-while-revalidate. */
const API_SWR_PATHS = new Set([
	'/api/movies',
	'/api/series',
	'/api/client/home',
]);

self.addEventListener('install', (event) => {
	event.waitUntil((async () => {
		const cache = await caches.open(SHELL_CACHE);
		// addAll() is atomic — if any asset 404s the whole install fails.
		// Add individually so a missing optional file doesn't kill the SW.
		await Promise.all(PRECACHE_URLS.map(async (url) => {
			try { await cache.add(url); } catch { /* skip — optional asset */ }
		}));
	})());
	self.skipWaiting();
});

self.addEventListener('activate', (event) => {
	event.waitUntil((async () => {
		const keys = await caches.keys();
		await Promise.all(keys.map((key) => {
			// Drop caches from older versions and any stray cache that doesn't
			// belong to this SW (left over from a previous deploy).
			if (!key.startsWith(CACHE_PREFIX)) return undefined;
			if (key === SHELL_CACHE || key === API_CACHE) return undefined;
			return caches.delete(key);
		}));
		await self.clients.claim();
	})());
});

self.addEventListener('fetch', (event) => {
	const { request } = event;
	if (request.method !== 'GET') return;
	const url = new URL(request.url);
	if (url.origin !== self.location.origin) return;

	// Precached static shell — cache first, network on miss.
	if (PRECACHE_URLS.includes(url.pathname)) {
		event.respondWith(cacheFirst(SHELL_CACHE, request));
		return;
	}

	// HTML navigations: try network, fall back to cached index.html so a
	// reload with no network still boots the app and lets the JS layer
	// take over rendering from IndexedDB.
	if (request.mode === 'navigate') {
		event.respondWith(networkFirst(SHELL_CACHE, request, '/index.html'));
		return;
	}

	// API GETs we want offline-capable: stale-while-revalidate. The user
	// sees a cached library instantly on hard refresh; the network refresh
	// fires in the background and updates the cache for the next read.
	if (API_SWR_PATHS.has(url.pathname)) {
		event.respondWith(staleWhileRevalidate(API_CACHE, request));
		return;
	}
});

async function cacheFirst(cacheName: string, request: Request): Promise<Response> {
	const cache = await caches.open(cacheName);
	const cached = await cache.match(request);
	if (cached) return cached;
	const response = await fetch(request);
	if (response.ok) cache.put(request, response.clone()).catch(() => {});
	return response;
}

async function networkFirst(cacheName: string, request: Request, fallbackPath: string): Promise<Response> {
	try {
		const response = await fetch(request);
		if (response.ok) {
			const cache = await caches.open(cacheName);
			cache.put(request, response.clone()).catch(() => {});
		}
		return response;
	} catch {
		const cache = await caches.open(cacheName);
		const cached = await cache.match(request) ?? await cache.match(fallbackPath);
		if (cached) return cached;
		return new Response('offline', { status: 503, statusText: 'Service Unavailable' });
	}
}

async function staleWhileRevalidate(cacheName: string, request: Request): Promise<Response> {
	const cache = await caches.open(cacheName);
	const cached = await cache.match(request);
	const networkPromise = fetch(request).then((response) => {
		// Only cache successful, non-304 responses. A 304 carries no body that
		// would be useful to store, and putting it in cache would mask future
		// 200s from updating the cache properly.
		if (response.ok && response.status !== 304) {
			cache.put(request, response.clone()).catch(() => {});
		}
		return response;
	}).catch(() => cached ?? Response.error());
	return cached ?? networkPromise;
}
