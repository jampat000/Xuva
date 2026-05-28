package systemstats

import (
	"sync"
	"testing"
	"time"
)

// resetCollectCache wipes the package-level cache between tests. Tests touch
// the same global state in series, so without this they bleed into each
// other depending on order.
func resetCollectCache() {
	collectCache.mu.Lock()
	collectCache.snap = Snapshot{}
	collectCache.at = time.Time{}
	collectCache.hasValue = false
	collectCache.mu.Unlock()
}

func TestCollect_CachesWithinTTL(t *testing.T) {
	resetCollectCache()
	first := Collect(nil)
	if first.CollectedAt == "" {
		t.Fatal("expected fresh snapshot to have CollectedAt")
	}
	// Second call inside the TTL should return the identical snapshot.
	second := Collect(nil)
	if second.CollectedAt != first.CollectedAt {
		t.Errorf("expected cached snapshot to keep the same CollectedAt; got %q vs %q",
			second.CollectedAt, first.CollectedAt)
	}
}

func TestCollect_RefreshesAfterTTL(t *testing.T) {
	resetCollectCache()
	_ = Collect(nil)

	// Capture the cache's stored timestamp before the second call. We can't
	// rely on Snapshot.CollectedAt to differ — it's formatted to RFC3339
	// (whole seconds) and the entire test runs sub-second on modern hardware.
	collectCache.mu.Lock()
	originalAt := collectCache.at
	collectCache.at = time.Now().Add(-2 * collectCacheTTL) // force "stale"
	collectCache.mu.Unlock()

	_ = Collect(nil)

	collectCache.mu.Lock()
	refreshedAt := collectCache.at
	collectCache.mu.Unlock()

	if !refreshedAt.After(originalAt) {
		t.Errorf("expected post-TTL Collect to refresh cache.at; got %v then %v", originalAt, refreshedAt)
	}
}

func TestCollect_ConcurrentCallsAreSafe(t *testing.T) {
	// Smoke test for the mutex: a swarm of concurrent callers must not race or
	// panic, and they should all observe a non-empty snapshot. The cache layer
	// doesn't guarantee single-flight (two cold callers may both compute), but
	// it must not corrupt the shared state.
	resetCollectCache()
	const n = 32
	var wg sync.WaitGroup
	wg.Add(n)
	for i := 0; i < n; i++ {
		go func() {
			defer wg.Done()
			if Collect(nil).CollectedAt == "" {
				t.Errorf("concurrent Collect returned empty snapshot")
			}
		}()
	}
	wg.Wait()
}
