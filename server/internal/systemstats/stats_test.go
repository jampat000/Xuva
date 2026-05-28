package systemstats

import (
	"context"
	"sync"
	"testing"
	"time"
)

// resetSamplerState wipes the package-level cache between tests. Tests touch
// the same global state in series, so without this they bleed into each
// other depending on order.
func resetSamplerState() {
	samplerState.mu.Lock()
	samplerState.snap = Snapshot{}
	samplerState.at = time.Time{}
	samplerState.hasValue = false
	samplerState.paths = nil
	samplerState.getPaths = nil
	samplerState.mu.Unlock()
}

func TestCollect_SamplesSynchronouslyOnFirstCall(t *testing.T) {
	resetSamplerState()
	snap := Collect(nil)
	if snap.CollectedAt == "" {
		t.Fatal("expected first Collect to return a populated snapshot")
	}
}

func TestCollect_ReturnsCachedAfterSample(t *testing.T) {
	resetSamplerState()
	first := Collect(nil)
	second := Collect(nil)
	if second.CollectedAt != first.CollectedAt {
		t.Errorf("second Collect should return cached snapshot (same CollectedAt); got %q vs %q",
			second.CollectedAt, first.CollectedAt)
	}
}

func TestCollect_RefreshesIfStale(t *testing.T) {
	resetSamplerState()
	_ = Collect(nil)
	samplerState.mu.Lock()
	originalAt := samplerState.at
	samplerState.at = time.Now().Add(-2 * staleSampleAge)
	samplerState.mu.Unlock()

	_ = Collect(nil)

	samplerState.mu.Lock()
	refreshedAt := samplerState.at
	samplerState.mu.Unlock()

	if !refreshedAt.After(originalAt) {
		t.Errorf("expected stale Collect to re-sample; got %v then %v", originalAt, refreshedAt)
	}
}

func TestCollect_ConcurrentCallsAreSafe(t *testing.T) {
	// Smoke test for the mutex: a swarm of concurrent callers must not race
	// or panic, and they should all observe a non-empty snapshot once the
	// first sync sample finishes.
	resetSamplerState()
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

func TestStartSampler_PrimesSnapshotAndRefreshes(t *testing.T) {
	resetSamplerState()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	dataDir := t.TempDir()
	StartSampler(ctx, func() map[string]string { return map[string]string{"data": dataDir} })

	// First snapshot must already be populated (StartSampler primes
	// synchronously before launching the goroutine).
	samplerState.mu.RLock()
	primed := samplerState.hasValue
	samplerState.mu.RUnlock()
	if !primed {
		t.Fatal("StartSampler did not prime the initial snapshot")
	}

	// And Collect should return that snapshot in microseconds, never blocking,
	// when the caller's paths match what the sampler captured.
	start := time.Now()
	_ = Collect(map[string]string{"data": dataDir})
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("Collect with matching paths should be sub-50ms; took %v", elapsed)
	}
}

func TestCollect_ResamplesWhenPathsChange(t *testing.T) {
	// Tests/handlers may call Collect with new paths after a settings change.
	// The cache should be invalidated for the new paths instead of returning
	// the snapshot built for the old ones — otherwise the dashboard would
	// keep showing the old data-dir until the next background tick.
	resetSamplerState()
	a := Collect(map[string]string{"data": t.TempDir()})

	b := Collect(map[string]string{"data": t.TempDir(), "transcode": t.TempDir()})
	// Different paths → different disk lists → different snapshot.
	if len(a.Disks) == len(b.Disks) {
		t.Errorf("expected disk list to grow after adding 'transcode' path; got %d vs %d",
			len(a.Disks), len(b.Disks))
	}
}
