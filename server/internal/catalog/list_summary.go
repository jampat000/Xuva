package catalog

// list_summary.go — slim list-view types and fetchers for /api/movies and
// /api/series. The browser grid card only needs ~6 metadata fields out of the
// 30+ in MetadataRecord. Sending the full blob — which includes cast arrays,
// crew, ratings, external IDs, and the raw TMDB JSON — cost 77 MB of JSON
// on a 4000-item library (17 KB per movie × 4008 = ~68 MB just in metadata).
// This file provides:
//   - MetadataListSummary   – 6-field slim struct (poster, genres, etc.)
//   - MovieListSummary      – MovieListItem counterpart using the slim type
//   - SeriesListSummary     – SeriesListItem counterpart using the slim type
//   - listMetadataBatch     – single query using json_extract; skips
//                             full details_json parse, ratings table, and
//                             (for movies) external-IDs table
//   - ListMoviesSummary     – drop-in for moviesHandler
//   - ListSeriesSummary     – drop-in for seriesHandler

import (
	"context"
	"database/sql"
	"encoding/json"
	"sort"
	"strings"
)

// MetadataListSummary is the metadata subset delivered by list endpoints.
// Detail endpoints (GET /api/movies/{id}) still return the full MetadataRecord.
type MetadataListSummary struct {
	Title         string   `json:"title,omitempty"`
	Year          int      `json:"year,omitempty"`
	PosterURL     string   `json:"posterUrl,omitempty"`
	BackdropURL   string   `json:"backdropUrl,omitempty"`
	Overview      string   `json:"overview,omitempty"` // used by "missing meta" filter
	Genres        []string `json:"genres,omitempty"`
	ContentRating string   `json:"contentRating,omitempty"`
	Studios       []string `json:"studios,omitempty"`

	// ExternalIDs is used internally to collapse series that share a
	// provider identity (tmdb/tvdb/imdb). Not sent to the client.
	ExternalIDs map[string]string `json:"-"`
}

// MovieListSummary is the slim list-view counterpart to MovieListItem.
type MovieListSummary struct {
	ID           string               `json:"id"`
	Title        string               `json:"title"`
	Year         int                  `json:"year"`
	SortTitle    string               `json:"sortTitle"`
	NeedsReview  bool                 `json:"needsReview"`
	Probed       bool                 `json:"probed"`
	VersionCount int                  `json:"versionCount"`
	AddedAt      string               `json:"addedAt,omitempty"`
	Watched      bool                 `json:"watched,omitempty"`
	Metadata     *MetadataListSummary `json:"metadata,omitempty"`
}

// SeriesListSummary is the slim list-view counterpart to SeriesListItem.
type SeriesListSummary struct {
	ID             string               `json:"id"`
	Title          string               `json:"title"`
	SortTitle      string               `json:"sortTitle"`
	SeasonCount    int                  `json:"seasonCount"`
	EpisodeCount   int                  `json:"episodeCount"`
	UnwatchedCount int                  `json:"unwatchedCount,omitempty"`
	AddedAt        string               `json:"addedAt,omitempty"`
	Watched        bool                 `json:"watched,omitempty"`
	NeedsReview    bool                 `json:"needsReview,omitempty"`
	Metadata       *MetadataListSummary `json:"metadata,omitempty"`
}

// listMetadataBatch fetches only the fields needed for list-view grid cards
// in a single database round-trip. It uses json_extract to pull genres,
// contentRating and studios out of details_json without parsing the full blob.
//
// For kind="series", it also issues a second query to fetch external IDs so
// that collapseSeriesListSummaries can merge duplicate entries that share a
// TMDB/TVDB identity (e.g. the same show across two library locations).
//
// 10–30× faster than GetBestMetadataBatch on large libraries because:
//   - No details_json unmarshal (skips 4000 × 5–15 KB JSON decode)
//   - No metadata_ratings query
//   - No merge/sort logic per provider
func (s *Service) listMetadataBatch(ctx context.Context, kind string, ids []string) (map[string]MetadataListSummary, error) {
	if len(ids) == 0 {
		return map[string]MetadataListSummary{}, nil
	}

	ph := strings.Repeat("?,", len(ids))
	ph = ph[:len(ph)-1]
	args := make([]any, 0, 1+len(ids))
	args = append(args, kind)
	for _, id := range ids {
		args = append(args, id)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT item_id, title, year, poster_url, backdrop_url, overview,
		       json_extract(details_json, '$.genres')        AS genres_json,
		       json_extract(details_json, '$.contentRating') AS content_rating,
		       json_extract(details_json, '$.studios')       AS studios_json,
		       updated_at
		FROM metadata_records
		WHERE kind = ? AND item_id IN (`+ph+`)
		ORDER BY item_id, updated_at DESC
	`, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	result := make(map[string]MetadataListSummary, len(ids))
	for rows.Next() {
		var itemID, title, posterURL, backdropURL, overview, updatedAt string
		var year int
		var genresJSON, contentRating, studiosJSON sql.NullString
		if err := rows.Scan(
			&itemID, &title, &year, &posterURL, &backdropURL, &overview,
			&genresJSON, &contentRating, &studiosJSON, &updatedAt,
		); err != nil {
			return nil, err
		}
		if _, already := result[itemID]; already {
			continue // ORDER BY updated_at DESC → first row per item is best
		}
		var genres, studios []string
		if genresJSON.Valid && genresJSON.String != "" {
			_ = json.Unmarshal([]byte(genresJSON.String), &genres)
		}
		if studiosJSON.Valid && studiosJSON.String != "" {
			_ = json.Unmarshal([]byte(studiosJSON.String), &studios)
		}
		result[itemID] = MetadataListSummary{
			Title:         strings.TrimSpace(title),
			Year:          year,
			PosterURL:     strings.TrimSpace(posterURL),
			BackdropURL:   strings.TrimSpace(backdropURL),
			Overview:      strings.TrimSpace(overview),
			Genres:        compactStrings(genres),
			ContentRating: strings.TrimSpace(contentRating.String),
			Studios:       compactStrings(studios),
		}
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	// For series, attach external IDs so collapseSeriesListSummaries can merge
	// duplicate entries (same show in two library locations) by provider key.
	if kind == "series" {
		extRows, extErr := s.db.QueryContext(ctx, `
			SELECT item_id, provider, external_id
			FROM metadata_external_ids
			WHERE kind = ? AND item_id IN (`+ph+`)
			ORDER BY item_id, provider
		`, args...)
		if extErr == nil {
			defer extRows.Close()
			for extRows.Next() {
				var extItemID, extProvider, extID string
				if scanErr := extRows.Scan(&extItemID, &extProvider, &extID); scanErr == nil {
					if summary, ok := result[extItemID]; ok {
						if summary.ExternalIDs == nil {
							summary.ExternalIDs = map[string]string{}
						}
						summary.ExternalIDs[extProvider] = extID
						result[extItemID] = summary
					}
				}
			}
		}
	}

	return result, nil
}

// ListMoviesSummary is the slim variant of ListMovies used by /api/movies.
// The response payload is ~10× smaller because MetadataListSummary omits
// cast, crew, ratings, external IDs and all other fields not used by the
// browser grid card.
func (s *Service) ListMoviesSummary(ctx context.Context, limit int, maxRating string, userID string) ([]MovieListSummary, error) {
	sqlLimit := limit
	if sqlLimit <= 0 {
		sqlLimit = -1 // SQLite LIMIT -1 = no limit
	} else if maxRating != "" {
		sqlLimit = limit * 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT v.movie_id, v.title, v.year, v.sort_title, v.needs_review,
		       v.version_count, v.created_at,
		       COALESCE((SELECT MAX(CASE WHEN ps.watched != 0 THEN 1 ELSE 0 END)
		                 FROM movie_versions mv
		                 LEFT JOIN playback_states ps
		                     ON ps.media_source_id = mv.media_source_id AND ps.user_id = ?
		                 WHERE mv.movie_id = v.movie_id), 0) AS is_watched,
		       v.is_probed
		FROM movies_list_view v
		ORDER BY v.sort_title, v.year
		LIMIT ?
	`, userID, sqlLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type movieStub struct {
		id, title, sortTitle, addedAt string
		year, versionCount            int
		needsReview, isWatched        int
		isProbed                      int
	}
	var stubs []movieStub
	for rows.Next() {
		var st movieStub
		if err := rows.Scan(
			&st.id, &st.title, &st.year, &st.sortTitle, &st.needsReview,
			&st.versionCount, &st.addedAt, &st.isWatched, &st.isProbed,
		); err != nil {
			return nil, err
		}
		stubs = append(stubs, st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ids := make([]string, len(stubs))
	for i, st := range stubs {
		ids[i] = st.id
	}
	metaMap, err := s.listMetadataBatch(ctx, "movie", ids)
	if err != nil {
		return nil, err
	}

	output := make([]MovieListSummary, 0, len(stubs))
	for _, st := range stubs {
		item := MovieListSummary{
			ID:           st.id,
			Title:        st.title,
			Year:         st.year,
			SortTitle:    st.sortTitle,
			NeedsReview:  st.needsReview != 0,
			Probed:       st.isProbed != 0,
			VersionCount: st.versionCount,
			AddedAt:      st.addedAt,
			Watched:      st.isWatched != 0,
		}
		if summary, ok := metaMap[st.id]; ok {
			// Prefer the metadata title/year over the raw DB value when
			// available — same logic as applyMovieMetadata in the full path.
			if summary.Title != "" {
				item.Title = summary.Title
				item.SortTitle = sortTitle(summary.Title)
			}
			if summary.Year != 0 {
				item.Year = summary.Year
			}
			item.Metadata = &summary
		}
		if maxRating != "" {
			cr := ""
			if item.Metadata != nil {
				cr = item.Metadata.ContentRating
			}
			if !withinCeiling(cr, maxRating) {
				continue
			}
		}
		output = append(output, item)
		if limit > 0 && len(output) >= limit {
			break
		}
	}
	return output, nil
}

// ListSeriesSummary is the slim variant of ListSeries used by /api/series.
func (s *Service) ListSeriesSummary(ctx context.Context, limit int, maxRating string, userID string) ([]SeriesListSummary, error) {
	sqlLimit := limit
	if sqlLimit <= 0 {
		sqlLimit = -1
	} else if maxRating != "" {
		sqlLimit = limit * 10
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT v.series_id, v.title, v.sort_title,
		       v.season_count, v.episode_count, v.created_at,
		       COALESCE((SELECT MAX(CASE WHEN ps.watched != 0 THEN 1 ELSE 0 END)
		                 FROM tv_episodes e
		                 JOIN episode_versions ev ON ev.episode_id = e.id
		                 LEFT JOIN playback_states ps
		                     ON ps.media_source_id = ev.media_source_id AND ps.user_id = ?
		                 WHERE e.series_id = v.series_id), 0) AS is_watched,
		       COALESCE((SELECT COUNT(DISTINCT ev2.episode_id)
		                 FROM tv_episodes e2
		                 JOIN episode_versions ev2 ON ev2.episode_id = e2.id
		                 LEFT JOIN playback_states ps2
		                     ON ps2.media_source_id = ev2.media_source_id AND ps2.user_id = ?
		                 WHERE e2.series_id = v.series_id
		                   AND (ps2.watched IS NULL OR ps2.watched = 0)), 0) AS unwatched_count
		FROM tv_series_list_view v
		ORDER BY v.sort_title, v.series_id
		LIMIT ?
	`, userID, userID, sqlLimit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	type seriesStub struct {
		id, title, sortTitle, addedAt string
		seasonCount, episodeCount     int
		isWatched                     int
		unwatchedCount                int
	}
	var stubs []seriesStub
	for rows.Next() {
		var st seriesStub
		if err := rows.Scan(
			&st.id, &st.title, &st.sortTitle,
			&st.seasonCount, &st.episodeCount, &st.addedAt, &st.isWatched, &st.unwatchedCount,
		); err != nil {
			return nil, err
		}
		stubs = append(stubs, st)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}

	ids := make([]string, len(stubs))
	for i, st := range stubs {
		ids[i] = st.id
	}
	metaMap, err := s.listMetadataBatch(ctx, "series", ids)
	if err != nil {
		return nil, err
	}

	raw := make([]SeriesListSummary, 0, len(stubs))
	for _, st := range stubs {
		item := SeriesListSummary{
			ID:             st.id,
			Title:          st.title,
			SortTitle:      st.sortTitle,
			SeasonCount:    st.seasonCount,
			EpisodeCount:   st.episodeCount,
			UnwatchedCount: st.unwatchedCount,
			AddedAt:        st.addedAt,
			Watched:        st.isWatched != 0,
		}
		if summary, ok := metaMap[st.id]; ok {
			if summary.Title != "" {
				item.Title = summary.Title
				item.SortTitle = sortTitle(summary.Title)
			}
			item.Metadata = &summary
		}
		raw = append(raw, item)
	}

	collapsed := collapseSeriesListSummaries(raw)

	if maxRating == "" {
		if limit > 0 && len(collapsed) > limit {
			return collapsed[:limit], nil
		}
		return collapsed, nil
	}
	output := make([]SeriesListSummary, 0, len(collapsed))
	for _, item := range collapsed {
		cr := ""
		if item.Metadata != nil {
			cr = item.Metadata.ContentRating
		}
		if withinCeiling(cr, maxRating) {
			output = append(output, item)
			if limit > 0 && len(output) >= limit {
				break
			}
		}
	}
	return output, nil
}

// collapseSeriesListSummaries deduplicates series that share a provider
// identity (e.g. the same TMDB show split across two library paths).
// Mirrors the logic of collapseSeriesListItems for the slim summary type.
func collapseSeriesListSummaries(items []SeriesListSummary) []SeriesListSummary {
	if len(items) <= 1 {
		return items
	}
	output := make([]SeriesListSummary, 0, len(items))
	indexByKey := map[string]int{}
	for _, item := range items {
		key := seriesIdentityKeySummary(item.Metadata, item.ID)
		idx, ok := indexByKey[key]
		if !ok {
			indexByKey[key] = len(output)
			output = append(output, item)
			continue
		}
		existing := &output[idx]
		existing.SeasonCount += item.SeasonCount
		existing.EpisodeCount += item.EpisodeCount
		existing.UnwatchedCount += item.UnwatchedCount
		if shouldPreferSeriesSummary(item, *existing) {
			existing.ID = item.ID
			existing.Title = item.Title
			existing.SortTitle = item.SortTitle
			existing.Metadata = item.Metadata
		}
	}
	sort.SliceStable(output, func(i, j int) bool {
		if output[i].SortTitle != output[j].SortTitle {
			return output[i].SortTitle < output[j].SortTitle
		}
		return output[i].ID < output[j].ID
	})
	return output
}

func seriesIdentityKeySummary(record *MetadataListSummary, fallbackID string) string {
	if record != nil {
		for _, provider := range []string{"tmdb", "tvdb", "imdb"} {
			if value := strings.TrimSpace(record.ExternalIDs[provider]); value != "" {
				return provider + ":" + value
			}
		}
	}
	return "series:" + strings.TrimSpace(fallbackID)
}

func shouldPreferSeriesSummary(candidate, current SeriesListSummary) bool {
	switch {
	case current.Metadata == nil && candidate.Metadata != nil:
		return true
	case current.Metadata != nil && candidate.Metadata == nil:
		return false
	case candidate.EpisodeCount != current.EpisodeCount:
		return candidate.EpisodeCount > current.EpisodeCount
	case candidate.SeasonCount != current.SeasonCount:
		return candidate.SeasonCount > current.SeasonCount
	case strings.TrimSpace(candidate.Title) != strings.TrimSpace(current.Title):
		return strings.Compare(candidate.Title, current.Title) < 0
	default:
		return candidate.ID < current.ID
	}
}
