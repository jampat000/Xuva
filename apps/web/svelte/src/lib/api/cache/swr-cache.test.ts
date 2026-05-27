import { afterEach, beforeEach, describe, expect, it, vi } from 'vitest';

// idb-keyval imports indexedDB at module-evaluation time, which happy-dom
// does not provide. Stub the four entry points with an in-memory map so the
// SWR cache exercises its IDB hydration / persistence paths without needing a
// real database.
vi.mock('idb-keyval', () => {
	const store = new Map<string, unknown>();
	return {
		get: vi.fn((k: string) => Promise.resolve(store.get(k))),
		set: vi.fn((k: string, v: unknown) => { store.set(k, v); return Promise.resolve(); }),
		del: vi.fn((k: string) => { store.delete(k); return Promise.resolve(); }),
		clear: vi.fn(() => { store.clear(); return Promise.resolve(); }),
	};
});

import { clear as idbClear } from 'idb-keyval';
import { invalidateSwr, invalidateSwrPrefix, subscribeSwr, swrFetch } from './swr-cache.js';

async function flushMicrotasks() {
	for (let i = 0; i < 4; i++) await Promise.resolve();
}

describe('swrFetch', () => {
	beforeEach(async () => {
		await idbClear();
		await invalidateSwrPrefix('');
	});

	afterEach(() => {
		vi.useRealTimers();
	});

	it('returns the network value on first call and caches it', async () => {
		const fetcher = vi.fn().mockResolvedValue({ items: ['a', 'b'] });
		const result = await swrFetch('test:cold', fetcher);
		expect(result).toEqual({ items: ['a', 'b'] });
		expect(fetcher).toHaveBeenCalledTimes(1);

		// Second call within freshMs: should return cached, no new fetch.
		const second = await swrFetch('test:cold', fetcher, { freshMs: 100, maxAgeMs: 1_000 });
		expect(second).toEqual({ items: ['a', 'b'] });
		expect(fetcher).toHaveBeenCalledTimes(1);
	});

	it('returns stale cache immediately and refreshes in background', async () => {
		vi.useFakeTimers();
		const fetcher = vi.fn()
			.mockResolvedValueOnce({ v: 1 })
			.mockResolvedValueOnce({ v: 2 });

		const first = await swrFetch('test:stale', fetcher, { freshMs: 10, maxAgeMs: 10_000 });
		expect(first).toEqual({ v: 1 });

		// Advance past freshMs so the entry is stale but still usable.
		vi.advanceTimersByTime(50);

		const seen: Array<{ v: number }> = [];
		const unsub = subscribeSwr<{ v: number }>('test:stale', (d) => seen.push(d));

		const second = await swrFetch('test:stale', fetcher, { freshMs: 10, maxAgeMs: 10_000 });
		// Stale → returns OLD value synchronously.
		expect(second).toEqual({ v: 1 });

		// Let the background refresh resolve.
		await vi.runAllTimersAsync();
		await flushMicrotasks();

		expect(fetcher).toHaveBeenCalledTimes(2);
		expect(seen).toEqual([{ v: 2 }]);
		unsub();
	});

	it('treats cache as expired beyond maxAgeMs and awaits a fresh fetch', async () => {
		vi.useFakeTimers();
		const fetcher = vi.fn()
			.mockResolvedValueOnce({ v: 1 })
			.mockResolvedValueOnce({ v: 2 });

		await swrFetch('test:expired', fetcher, { freshMs: 10, maxAgeMs: 50 });
		vi.advanceTimersByTime(200);  // older than maxAgeMs

		const result = await swrFetch('test:expired', fetcher, { freshMs: 10, maxAgeMs: 50 });
		expect(result).toEqual({ v: 2 });
		expect(fetcher).toHaveBeenCalledTimes(2);
	});

	it('deduplicates parallel refresh requests', async () => {
		const slow = new Promise<{ ok: true }>((resolve) => setTimeout(() => resolve({ ok: true }), 10));
		const fetcher = vi.fn().mockReturnValue(slow);

		const [a, b] = await Promise.all([
			swrFetch('test:dedup', fetcher),
			swrFetch('test:dedup', fetcher),
		]);
		expect(a).toEqual({ ok: true });
		expect(b).toEqual({ ok: true });
		expect(fetcher).toHaveBeenCalledTimes(1);
	});

	it('invalidates a key and forces the next call to refetch', async () => {
		const fetcher = vi.fn()
			.mockResolvedValueOnce({ v: 1 })
			.mockResolvedValueOnce({ v: 2 });

		await swrFetch('test:inv', fetcher);
		await invalidateSwr('test:inv');
		const result = await swrFetch('test:inv', fetcher);

		expect(result).toEqual({ v: 2 });
		expect(fetcher).toHaveBeenCalledTimes(2);
	});

	it('invalidates by prefix', async () => {
		const fA = vi.fn().mockResolvedValueOnce('A1').mockResolvedValueOnce('A2');
		const fB = vi.fn().mockResolvedValueOnce('B1').mockResolvedValueOnce('B2');
		await swrFetch('test:pre:a', fA);
		await swrFetch('test:pre:b', fB);

		await invalidateSwrPrefix('test:pre:');

		expect(await swrFetch('test:pre:a', fA)).toBe('A2');
		expect(await swrFetch('test:pre:b', fB)).toBe('B2');
	});

	it('preserves cached data when a background refresh fails', async () => {
		vi.useFakeTimers();
		// First call succeeds, every subsequent refresh fails — caller should
		// keep seeing the cached value rather than crashing or returning null.
		const fetcher = vi.fn()
			.mockResolvedValueOnce({ v: 1 })
			.mockRejectedValue(new Error('boom'));

		await swrFetch('test:fail', fetcher, { freshMs: 10, maxAgeMs: 10_000 });
		vi.advanceTimersByTime(50);

		const result = await swrFetch('test:fail', fetcher, { freshMs: 10, maxAgeMs: 10_000 });
		expect(result).toEqual({ v: 1 });

		await vi.runAllTimersAsync();
		await flushMicrotasks();

		// Cached value is still served on the next call (refresh failed, didn't poison cache).
		const after = await swrFetch('test:fail', fetcher, { freshMs: 10, maxAgeMs: 10_000 });
		expect(after).toEqual({ v: 1 });
	});
});
