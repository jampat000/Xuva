/**
 * Server-Sent Events bridge to the backend /api/events bus.
 *
 * The backend publishes every state-changing operation to a single event
 * stream (see server/internal/events/bus.go). This module subscribes once
 * per browser tab and translates relevant event types into SWR cache
 * invalidations + immediate refreshes — so when a scan completes or a movie's
 * metadata is updated on the server, every mounted page receives fresh data
 * without polling and without a manual reload.
 *
 * EventSource auto-reconnects on transient errors, so callers wire up the
 * connection once on layout mount and forget it. The cleanup function the
 * factory returns closes the stream and tears down any pending debounce
 * timers — call it from onMount's cleanup.
 */

import { getMovies, getSeries, invalidateListCache } from './browse';
import { getClientHome, invalidateHomeCache } from './home';

interface ServerEvent {
	id: number;
	type: string;
	data?: unknown;
	createdAt: string;
}

// Many events arrive in bursts (a single scan publishes several scan.completed
// signals across libraries; metadata.updated fires once per item during a
// refresh). Coalesce them so the client doesn't refetch the world on every
// signal — one fetch ~250ms after the last event covers the burst.
const DEBOUNCE_MS = 250;

let source: EventSource | null = null;
let listTimer: ReturnType<typeof setTimeout> | null = null;
let homeTimer: ReturnType<typeof setTimeout> | null = null;

function scheduleListRefresh(): void {
	if (listTimer) clearTimeout(listTimer);
	listTimer = setTimeout(() => {
		listTimer = null;
		void (async () => {
			await invalidateListCache();
			// Trigger SWR refresh — pages subscribed to movies/series will
			// receive fresh data via subscribeMovies / subscribeSeries.
			void getMovies();
			void getSeries();
		})();
	}, DEBOUNCE_MS);
}

function scheduleHomeRefresh(): void {
	if (homeTimer) clearTimeout(homeTimer);
	homeTimer = setTimeout(() => {
		homeTimer = null;
		void (async () => {
			await invalidateHomeCache();
			void getClientHome();
		})();
	}, DEBOUNCE_MS);
}

function handleServerEvent(event: ServerEvent): void {
	// Library mutations: movies/series lists change, and the home page's
	// "new in your library" rows change with them.
	const libraryEvents = new Set([
		'library.updated',
		'library.deleted',
		'metadata.updated',
		'scan.completed',
		'automation.scan.completed',
	]);
	if (libraryEvents.has(event.type)) {
		scheduleListRefresh();
		scheduleHomeRefresh();
		return;
	}
	// Playback state: continue-watching reshuffles. Lists are unchanged.
	if (event.type === 'playback.state.updated') {
		scheduleHomeRefresh();
		return;
	}
	// Note: pairing.request.* events are handled by the Settings page directly
	// via its own EventStream subscription (lib/events/stream.ts), not here —
	// the approvals list is an admin-only view and shouldn't fight a global
	// SWR-cache layer for ownership of "pending approvals" state.
}

/**
 * Connect to /api/events and start dispatching server events to cache
 * invalidations. Idempotent — calling twice in the same tab is a no-op.
 * Returns a cleanup function that closes the stream.
 */
export function connectEventStream(): () => void {
	if (typeof window === 'undefined' || typeof EventSource === 'undefined') {
		return () => { /* SSR / unsupported environment */ };
	}
	if (source) {
		return () => disconnect();
	}
	try {
		source = new EventSource('/api/events');
	} catch {
		return () => { /* construction failed (CSP, etc.) */ };
	}
	source.onmessage = (e) => {
		try {
			const event = JSON.parse(e.data) as ServerEvent;
			if (event && typeof event.type === 'string') handleServerEvent(event);
		} catch {
			// Non-JSON payload (e.g. the `event: ready` heartbeat) — ignore.
		}
	};
	source.onerror = () => {
		// EventSource auto-reconnects on transient errors. We only act on
		// CONNECTING state changes; nothing else useful to do here.
	};
	return () => disconnect();
}

function disconnect(): void {
	if (source) {
		source.close();
		source = null;
	}
	if (listTimer) { clearTimeout(listTimer); listTimer = null; }
	if (homeTimer) { clearTimeout(homeTimer); homeTimer = null; }
}
