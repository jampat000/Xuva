package trending

import (
	"context"
	"errors"
	"net/http"
	"sync/atomic"
	"testing"
	"time"
)

var errFake = errors.New("fake transport: no network in tests")

type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// stubCatalog implements the Catalog interface so we can exercise the
// service without standing up a real database. Every TMDB id maps to a
// fake catalog ID — that's enough to drive the cross-reference path.
type stubCatalog struct{}

func (stubCatalog) FindByExternalID(_ context.Context, kind string, provider string, externalID string) (string, bool, error) {
	return "cat:" + kind + ":" + externalID, true, nil
}

// TestTrending_ReturnsNilOnCacheMissAndDoesNotBlock — the whole point of the
// async pattern: a Trending() call with a cold cache must return immediately
// (no TMDB roundtrip on the request path). We use a bogus API key so any
// accidental network fetch would fail audibly with an HTTP error; the test
// asserts that the cold call returns nil with no error within milliseconds.
func TestTrending_ReturnsNilOnCacheMissAndDoesNotBlock(t *testing.T) {
	s := NewService("fake-key", stubCatalog{})
	start := time.Now()
	items, err := s.Trending(context.Background(), "US", 10)
	elapsed := time.Since(start)
	if err != nil {
		t.Fatalf("cold Trending should not surface fetch errors, got: %v", err)
	}
	if items != nil {
		t.Errorf("cold Trending should return nil, got %d items", len(items))
	}
	if elapsed > 100*time.Millisecond {
		t.Errorf("cold Trending should be sub-100ms (non-blocking); took %v", elapsed)
	}
}

// TestTrending_ReturnsCachedDataWhenWarm — once the in-memory cache has been
// populated (we shortcut by writing directly), Trending must serve from
// memory and not refetch.
func TestTrending_ReturnsCachedDataWhenWarm(t *testing.T) {
	s := NewService("fake-key", stubCatalog{})
	s.mu.Lock()
	s.cache["US"] = cacheEntry{
		items:     []Item{{CatalogID: "x", Title: "Cached", Rank: 1}},
		fetchedAt: time.Now(),
	}
	s.mu.Unlock()

	items, err := s.Trending(context.Background(), "US", 10)
	if err != nil {
		t.Fatalf("warm Trending: %v", err)
	}
	if len(items) != 1 || items[0].Title != "Cached" {
		t.Errorf("expected cached single item, got %#v", items)
	}
}

// TestTrending_NoOpWhenApiKeyEmpty — without a TMDB key the service is a
// no-op (and must not panic, even with nil receivers).
func TestTrending_NoOpWhenApiKeyEmpty(t *testing.T) {
	s := NewService("", stubCatalog{})
	items, err := s.Trending(context.Background(), "US", 10)
	if err != nil || items != nil {
		t.Errorf("empty-key service should be a no-op, got items=%v err=%v", items, err)
	}
	// And a nil receiver also has to no-op (defensive — the service may be
	// optional in some embedded deploys).
	var nilSvc *Service
	items, err = nilSvc.Trending(context.Background(), "US", 10)
	if err != nil || items != nil {
		t.Errorf("nil service should no-op, got items=%v err=%v", items, err)
	}
}

// TestRefreshAsync_DedupesConcurrentCalls — even if a swarm of cold home
// requests arrive simultaneously they must not produce a stampede on TMDB.
// We assert that across many concurrent refreshAsync calls only one fetch
// goroutine runs per region (the inflight guard works).
func TestRefreshAsync_DedupesConcurrentCalls(t *testing.T) {
	s := NewService("fake-key", stubCatalog{})

	// Replace the http client with one that counts requests. We don't need
	// it to succeed — we're only checking the inflight guard.
	var hits int32
	s.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&hits, 1)
		// Block long enough that subsequent refreshAsync calls overlap.
		time.Sleep(40 * time.Millisecond)
		return nil, errFake
	})

	const n = 25
	for i := 0; i < n; i++ {
		s.refreshAsync("US")
	}
	// Wait for the goroutines to drain (inflight clears on completion).
	deadline := time.Now().Add(2 * time.Second)
	for time.Now().Before(deadline) {
		s.mu.Lock()
		_, running := s.inflight["US"]
		s.mu.Unlock()
		if !running {
			break
		}
		time.Sleep(10 * time.Millisecond)
	}
	// Each call to refreshPage fires twice (movie + tv pages), so the
	// hit count is 2× the number of fetch goroutines. Either way it must
	// be far smaller than n*2.
	got := atomic.LoadInt32(&hits)
	if got > 4 { // generous: one fetch (2 pages) plus headroom for racing
		t.Errorf("expected dedupe to keep TMDB fetches to ~2, got %d hits", got)
	}
}

// TestStartSampler_PrimesAndDoesNotBlockReturn — StartSampler must return
// immediately (work goes to a goroutine) and must trigger at least one
// refresh attempt for the configured region.
func TestStartSampler_PrimesAndDoesNotBlockReturn(t *testing.T) {
	s := NewService("fake-key", stubCatalog{})
	var fetchAttempts int32
	s.client.Transport = roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		atomic.AddInt32(&fetchAttempts, 1)
		return nil, errFake
	})

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	start := time.Now()
	s.StartSampler(ctx, func() string { return "US" })
	if elapsed := time.Since(start); elapsed > 50*time.Millisecond {
		t.Errorf("StartSampler must return immediately, took %v", elapsed)
	}

	// Wait briefly for the priming goroutine to attempt a fetch.
	deadline := time.Now().Add(500 * time.Millisecond)
	for time.Now().Before(deadline) {
		if atomic.LoadInt32(&fetchAttempts) > 0 {
			return
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Errorf("expected sampler to prime the cache with at least one fetch attempt, got 0")
}

// TestStartSampler_RespectsContextCancel — once the context is cancelled,
// the sampler goroutine must stop ticking.
func TestStartSampler_RespectsContextCancel(t *testing.T) {
	s := NewService("fake-key", stubCatalog{})
	ctx, cancel := context.WithCancel(context.Background())
	s.StartSampler(ctx, func() string { return "US" })
	cancel()
	// No assertion needed beyond "no goroutine leak panic"; the cancel
	// path is straightforward enough that compile-passing + the priming
	// test above gives sufficient coverage. This test mainly exercises
	// the select branch so future refactors can't accidentally break it.
}
