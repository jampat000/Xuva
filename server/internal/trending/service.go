// Package trending fetches TMDB weekly trending lists for a given region,
// cross-references them against the local catalog, and returns the ranked
// subset the user actually owns. Results are cached for 24 hours.
package trending

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	tmdbBase  = "https://api.themoviedb.org/3"
	cacheTTL  = 24 * time.Hour
	fetchLimit = 50 // how many trending titles to fetch per type

	// samplerInterval is how often the background goroutine refreshes the
	// trending cache for whichever region(s) the home handler has asked about.
	// 12 h keeps the data current enough for a "this week's trending" row
	// without hammering TMDB.
	samplerInterval = 12 * time.Hour
)

// Catalog is the subset of catalog.Service this package needs.
type Catalog interface {
	FindByExternalID(ctx context.Context, kind string, provider string, externalID string) (string, bool, error)
}

// Item is one trending title that exists in the local library.
type Item struct {
	CatalogID   string `json:"catalogId"`
	Kind        string `json:"kind"` // "movie" or "series"
	TMDBId      int    `json:"tmdbId"`
	Title       string `json:"title"`
	Overview    string `json:"overview,omitempty"`
	PosterPath  string `json:"posterPath,omitempty"`
	BackdropPath string `json:"backdropPath,omitempty"`
	VoteAverage float64 `json:"voteAverage,omitempty"`
	Year        int    `json:"year,omitempty"`
	Rank        int    `json:"rank"`
}

type cacheEntry struct {
	items     []Item
	fetchedAt time.Time
}

// Service fetches and caches trending data from TMDB.
//
// The cache is read-only from the request path — Trending() never blocks on
// a TMDB roundtrip. A background goroutine (see StartSampler) refreshes the
// cache on a fixed interval, and on first-request-for-a-new-region a one-shot
// fetch is kicked off async so subsequent calls warm up automatically.
//
// This matters because clientHomeHandler calls Trending() synchronously, and
// TMDB roundtrips routinely take 500-800 ms. Before this change, the first
// home request after every server restart paid that latency in full.
type Service struct {
	apiKey  string
	catalog Catalog
	client  *http.Client

	mu       sync.Mutex
	cache    map[string]cacheEntry // key = region
	inflight map[string]struct{}   // regions with a fetch already running
}

func NewService(apiKey string, catalog Catalog) *Service {
	return &Service{
		apiKey:   apiKey,
		catalog:  catalog,
		client:   &http.Client{Timeout: 10 * time.Second},
		cache:    map[string]cacheEntry{},
		inflight: map[string]struct{}{},
	}
}

// Trending returns up to limit items from the cached weekly-trending list
// for the region. Region is an ISO 3166-1 alpha-2 code (e.g. "AU"); an empty
// region falls back to "US".
//
// This call NEVER blocks on a TMDB roundtrip. If the cache is cold or stale,
// a background goroutine is kicked off to refresh it (one per region at a
// time) and the current best-known value — empty on first sight — is
// returned immediately. The home handler in turn falls back to a local
// "highest rated" spotlight when the trending row is empty, so the user
// always sees something useful even on the very first call after a restart.
func (s *Service) Trending(ctx context.Context, region string, limit int) ([]Item, error) {
	if s == nil || s.apiKey == "" {
		return nil, nil
	}
	if limit <= 0 {
		limit = 10
	}
	if region == "" {
		region = "US"
	}

	cacheKey := region
	s.mu.Lock()
	entry, ok := s.cache[cacheKey]
	s.mu.Unlock()

	// Stale or missing → kick off an async refresh, but don't wait for it.
	// The current call returns whatever we have (possibly nothing) so the
	// home response stays sub-50 ms regardless of TMDB latency.
	if !ok || time.Since(entry.fetchedAt) >= cacheTTL {
		s.refreshAsync(region)
	}

	if !ok {
		return nil, nil
	}
	if len(entry.items) > limit {
		return entry.items[:limit], nil
	}
	return entry.items, nil
}

// refreshAsync kicks off a single in-flight refresh per region. If one is
// already running for that region, this call is a no-op so the home handler
// can't accidentally storm TMDB by serving lots of simultaneous cold loads.
func (s *Service) refreshAsync(region string) {
	s.mu.Lock()
	if _, running := s.inflight[region]; running {
		s.mu.Unlock()
		return
	}
	s.inflight[region] = struct{}{}
	s.mu.Unlock()

	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.inflight, region)
			s.mu.Unlock()
		}()
		// Use a fresh context with a hard timeout — the original request
		// is long gone by the time we run, and we don't want a stuck TMDB
		// call to leak the goroutine.
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		items, err := s.fetch(ctx, region, fetchLimit)
		if err != nil {
			return // keep prior cache (if any) on fetch failure
		}
		s.mu.Lock()
		s.cache[region] = cacheEntry{items: items, fetchedAt: time.Now()}
		s.mu.Unlock()
	}()
}

// StartSampler primes the trending cache for the configured region at
// startup and keeps it warm with a periodic background refresh. The
// getRegion callback returns the LIVE region (read on each tick) so a
// settings change to the user's country flows in without a restart.
//
// Without this, the first home request after every server restart would pay
// the full TMDB roundtrip (~800 ms). With it, the cache is warm by the time
// the user lands on /home.
func (s *Service) StartSampler(ctx context.Context, getRegion func() string) {
	if s == nil || s.apiKey == "" {
		return
	}
	go func() {
		// Prime synchronously on boot so the first home request — which may
		// arrive seconds after startup — already has data to return.
		if region := getRegion(); region != "" {
			s.refreshAsync(region)
		}
		ticker := time.NewTicker(samplerInterval)
		defer ticker.Stop()
		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				if region := getRegion(); region != "" {
					s.refreshAsync(region)
				}
			}
		}
	}()
}

func (s *Service) fetch(ctx context.Context, region string, limit int) ([]Item, error) {
	movies, err := s.fetchPage(ctx, "movie", region)
	if err != nil {
		return nil, err
	}
	shows, err := s.fetchPage(ctx, "tv", region)
	if err != nil {
		return nil, err
	}

	// Interleave movies and shows in rank order so the list is mixed.
	// Both slices are already ranked 1..N from TMDB; we merge them by
	// alternating to keep rough parity between movies and series.
	merged := interleave(movies, shows)

	// Cross-reference against local catalog.
	var matched []Item
	rank := 1
	for _, t := range merged {
		if len(matched) >= limit {
			break
		}
		kind := "movie"
		if t.kind == "tv" {
			kind = "series"
		}
		catalogID, found, err := s.catalog.FindByExternalID(ctx, kind, "tmdb", fmt.Sprintf("%d", t.id))
		if err != nil || !found {
			continue
		}
		matched = append(matched, Item{
			CatalogID:    catalogID,
			Kind:         kind,
			TMDBId:       t.id,
			Title:        t.title,
			Overview:     t.overview,
			PosterPath:   t.posterPath,
			BackdropPath: t.backdropPath,
			VoteAverage:  t.voteAverage,
			Year:         t.year,
			Rank:         rank,
		})
		rank++
	}
	return matched, nil
}

type tmdbEntry struct {
	id           int
	kind         string // "movie" or "tv"
	title        string
	overview     string
	posterPath   string
	backdropPath string
	voteAverage  float64
	year         int
}

func (s *Service) fetchPage(ctx context.Context, mediaType string, region string) ([]tmdbEntry, error) {
	u, _ := url.Parse(fmt.Sprintf("%s/trending/%s/week", tmdbBase, mediaType))
	q := u.Query()
	q.Set("api_key", s.apiKey)
	q.Set("region", region)
	u.RawQuery = q.Encode()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u.String(), nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Accept", "application/json")

	resp, err := s.client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("tmdb trending: status %d", resp.StatusCode)
	}

	body, err := io.ReadAll(io.LimitReader(resp.Body, 512*1024))
	if err != nil {
		return nil, err
	}

	var payload struct {
		Results []struct {
			ID           int     `json:"id"`
			Title        string  `json:"title"`        // movies
			Name         string  `json:"name"`         // tv
			Overview     string  `json:"overview"`
			PosterPath   string  `json:"poster_path"`
			BackdropPath string  `json:"backdrop_path"`
			VoteAverage  float64 `json:"vote_average"`
			ReleaseDate  string  `json:"release_date"`  // movies
			FirstAirDate string  `json:"first_air_date"` // tv
		} `json:"results"`
	}
	if err := json.Unmarshal(body, &payload); err != nil {
		return nil, err
	}

	out := make([]tmdbEntry, 0, len(payload.Results))
	for _, r := range payload.Results {
		title := r.Title
		if title == "" {
			title = r.Name
		}
		date := r.ReleaseDate
		if date == "" {
			date = r.FirstAirDate
		}
		year := 0
		if len(date) >= 4 {
			fmt.Sscanf(date[:4], "%d", &year)
		}
		out = append(out, tmdbEntry{
			id:           r.ID,
			kind:         mediaType,
			title:        title,
			overview:     r.Overview,
			posterPath:   r.PosterPath,
			backdropPath: r.BackdropPath,
			voteAverage:  r.VoteAverage,
			year:         year,
		})
	}
	return out, nil
}

// interleave merges two slices by alternating elements, keeping relative rank.
func interleave(a, b []tmdbEntry) []tmdbEntry {
	out := make([]tmdbEntry, 0, len(a)+len(b))
	i, j := 0, 0
	for i < len(a) || j < len(b) {
		if i < len(a) {
			out = append(out, a[i])
			i++
		}
		if j < len(b) {
			out = append(out, b[j])
			j++
		}
	}
	return out
}
