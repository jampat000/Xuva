import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

// Mock the cache and fetch helpers so the test can assert exactly which
// invalidations and refreshes the event stream triggers.
const invalidateListMock = vi.fn(() => Promise.resolve());
const invalidateHomeMock = vi.fn(() => Promise.resolve());
const getMoviesMock = vi.fn(() => Promise.resolve({ movies: [] }));
const getSeriesMock = vi.fn(() => Promise.resolve({ series: [] }));
const getHomeMock = vi.fn(() => Promise.resolve({}));

vi.mock('./browse', () => ({
	invalidateListCache: () => invalidateListMock(),
	getMovies: () => getMoviesMock(),
	getSeries: () => getSeriesMock(),
}));

vi.mock('./home', () => ({
	invalidateHomeCache: () => invalidateHomeMock(),
	getClientHome: () => getHomeMock(),
}));

// happy-dom doesn't ship EventSource; install a minimal capture stub.
class StubEventSource {
	url: string;
	onmessage: ((e: MessageEvent) => void) | null = null;
	onerror: ((e: Event) => void) | null = null;
	closed = false;
	constructor(url: string) {
		this.url = url;
		StubEventSource.instances.push(this);
	}
	close() { this.closed = true; }
	emit(data: unknown) {
		this.onmessage?.(new MessageEvent('message', { data: JSON.stringify(data) }));
	}
	static instances: StubEventSource[] = [];
	static reset() { StubEventSource.instances = []; }
}

beforeEach(() => {
	StubEventSource.reset();
	invalidateListMock.mockClear();
	invalidateHomeMock.mockClear();
	getMoviesMock.mockClear();
	getSeriesMock.mockClear();
	getHomeMock.mockClear();
	(globalThis as unknown as { EventSource: typeof StubEventSource }).EventSource = StubEventSource;
	vi.resetModules();
});

afterEach(() => {
	vi.useRealTimers();
});

async function loadModule() {
	return await import('./events.js');
}

describe('connectEventStream', () => {
	it('opens a single connection to /api/events', async () => {
		const { connectEventStream } = await loadModule();
		const cleanup = connectEventStream();
		expect(StubEventSource.instances.length).toBe(1);
		expect(StubEventSource.instances[0].url).toBe('/api/events');
		cleanup();
	});

	it('is idempotent — second call does not open another connection', async () => {
		const { connectEventStream } = await loadModule();
		const cleanupA = connectEventStream();
		connectEventStream();
		expect(StubEventSource.instances.length).toBe(1);
		cleanupA();
	});

	it('refreshes both list and home caches on library.updated', async () => {
		vi.useFakeTimers();
		const { connectEventStream } = await loadModule();
		const cleanup = connectEventStream();
		const src = StubEventSource.instances[0];
		src.emit({ id: 1, type: 'library.updated', createdAt: 'now' });
		// Debounce window — caches not invalidated synchronously.
		expect(invalidateListMock).not.toHaveBeenCalled();
		await vi.runAllTimersAsync();
		expect(invalidateListMock).toHaveBeenCalledTimes(1);
		expect(invalidateHomeMock).toHaveBeenCalledTimes(1);
		expect(getMoviesMock).toHaveBeenCalledTimes(1);
		expect(getSeriesMock).toHaveBeenCalledTimes(1);
		expect(getHomeMock).toHaveBeenCalledTimes(1);
		cleanup();
	});

	it('coalesces bursts of events into one refresh', async () => {
		vi.useFakeTimers();
		const { connectEventStream } = await loadModule();
		const cleanup = connectEventStream();
		const src = StubEventSource.instances[0];
		for (let i = 0; i < 5; i++) {
			src.emit({ id: i, type: 'metadata.updated', createdAt: 'now' });
		}
		await vi.runAllTimersAsync();
		expect(invalidateListMock).toHaveBeenCalledTimes(1);
		expect(getMoviesMock).toHaveBeenCalledTimes(1);
		cleanup();
	});

	it('refreshes only home cache on playback.state.updated', async () => {
		vi.useFakeTimers();
		const { connectEventStream } = await loadModule();
		const cleanup = connectEventStream();
		const src = StubEventSource.instances[0];
		src.emit({ id: 1, type: 'playback.state.updated', createdAt: 'now' });
		await vi.runAllTimersAsync();
		expect(invalidateHomeMock).toHaveBeenCalledTimes(1);
		expect(getHomeMock).toHaveBeenCalledTimes(1);
		expect(invalidateListMock).not.toHaveBeenCalled();
		expect(getMoviesMock).not.toHaveBeenCalled();
		cleanup();
	});

	it('ignores unrelated events (audit.route, automation.scan.started)', async () => {
		vi.useFakeTimers();
		const { connectEventStream } = await loadModule();
		const cleanup = connectEventStream();
		const src = StubEventSource.instances[0];
		src.emit({ id: 1, type: 'audit.route', createdAt: 'now' });
		src.emit({ id: 2, type: 'automation.scan.started', createdAt: 'now' });
		await vi.runAllTimersAsync();
		expect(invalidateListMock).not.toHaveBeenCalled();
		expect(invalidateHomeMock).not.toHaveBeenCalled();
		cleanup();
	});

	it('drops non-JSON payloads silently', async () => {
		vi.useFakeTimers();
		const { connectEventStream } = await loadModule();
		const cleanup = connectEventStream();
		const src = StubEventSource.instances[0];
		// Simulate the "event: ready\ndata: {...}" heartbeat that arrives as a
		// MessageEvent whose data isn't valid JSON for the ServerEvent shape.
		src.onmessage?.(new MessageEvent('message', { data: 'not-json' }));
		await vi.runAllTimersAsync();
		expect(invalidateListMock).not.toHaveBeenCalled();
		cleanup();
	});

	it('cleanup closes the EventSource', async () => {
		const { connectEventStream } = await loadModule();
		const cleanup = connectEventStream();
		const src = StubEventSource.instances[0];
		cleanup();
		expect(src.closed).toBe(true);
	});
});
