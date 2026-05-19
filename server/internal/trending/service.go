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
type Service struct {
	apiKey  string
	catalog Catalog
	client  *http.Client

	mu    sync.Mutex
	cache map[string]cacheEntry // key = "region:week"
}

func NewService(apiKey string, catalog Catalog) *Service {
	return &Service{
		apiKey:  apiKey,
		catalog: catalog,
		client:  &http.Client{Timeout: 10 * time.Second},
		cache:   map[string]cacheEntry{},
	}
}

// Trending returns up to limit items from the TMDB weekly trending list that
// exist in the local library, in rank order. Region is an ISO 3166-1 alpha-2
// code (e.g. "AU"). An empty region falls back to global trending.
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
	if ok && time.Since(entry.fetchedAt) < cacheTTL {
		if len(entry.items) > limit {
			return entry.items[:limit], nil
		}
		return entry.items, nil
	}

	items, err := s.fetch(ctx, region, limit)
	if err != nil {
		// On fetch error return the stale cache rather than nothing.
		if ok {
			return entry.items, nil
		}
		return nil, err
	}

	s.mu.Lock()
	s.cache[cacheKey] = cacheEntry{items: items, fetchedAt: time.Now()}
	s.mu.Unlock()

	if len(items) > limit {
		return items[:limit], nil
	}
	return items, nil
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
