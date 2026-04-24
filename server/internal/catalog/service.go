package catalog

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
)

type Service struct {
	db *sql.DB
}

type PersistSummary struct {
	LibraryID    string `json:"libraryId"`
	ScanRunID    int64  `json:"scanRunId"`
	MediaSources int    `json:"mediaSources"`
	Movies       int    `json:"movies,omitempty"`
	Episodes     int    `json:"episodes,omitempty"`
}

type Summary struct {
	Libraries    int `json:"libraries"`
	MediaSources int `json:"mediaSources"`
	Movies       int `json:"movies"`
	Series       int `json:"series"`
	Episodes     int `json:"episodes"`
	ScanRuns     int `json:"scanRuns"`
}

func NewService(database *database.Service) *Service {
	return &Service{db: database.DB()}
}

func (s *Service) Summary(ctx context.Context) (Summary, error) {
	var summary Summary
	counts := []struct {
		table string
		out   *int
	}{
		{"libraries", &summary.Libraries},
		{"media_sources", &summary.MediaSources},
		{"movies", &summary.Movies},
		{"tv_series", &summary.Series},
		{"tv_episodes", &summary.Episodes},
		{"scan_runs", &summary.ScanRuns},
	}
	for _, count := range counts {
		if err := s.db.QueryRowContext(ctx, "SELECT count(*) FROM "+count.table).Scan(count.out); err != nil {
			return Summary{}, err
		}
	}
	return summary, nil
}

func (s *Service) SaveMovieScan(ctx context.Context, library libraries.Library, result scanner.Result, candidates []movies.Candidate) (PersistSummary, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistSummary{}, err
	}
	defer tx.Rollback()

	if err := upsertLibrary(ctx, tx, library); err != nil {
		return PersistSummary{}, err
	}
	scanRunID, err := insertScanRun(ctx, tx, result.Summary)
	if err != nil {
		return PersistSummary{}, err
	}

	now := timestamp(time.Now())
	movieIDs := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		sourceID := sourceID(candidate.Media.Path)
		if err := upsertMediaSource(ctx, tx, library.ID, media.KindMovie, sourceID, candidate.Media, now); err != nil {
			return PersistSummary{}, err
		}
		movieID := movieRecordID(candidate, sourceID)
		movieIDs[movieID] = struct{}{}
		if err := upsertMovie(ctx, tx, movieID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertMovieVersion(ctx, tx, movieID, sourceID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return PersistSummary{}, err
	}
	return PersistSummary{
		LibraryID:    library.ID,
		ScanRunID:    scanRunID,
		MediaSources: len(candidates),
		Movies:       len(movieIDs),
	}, nil
}

func (s *Service) SaveTVScan(ctx context.Context, library libraries.Library, result scanner.Result, candidates []tv.EpisodeCandidate) (PersistSummary, error) {
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return PersistSummary{}, err
	}
	defer tx.Rollback()

	if err := upsertLibrary(ctx, tx, library); err != nil {
		return PersistSummary{}, err
	}
	scanRunID, err := insertScanRun(ctx, tx, result.Summary)
	if err != nil {
		return PersistSummary{}, err
	}

	now := timestamp(time.Now())
	episodeIDs := make(map[string]struct{}, len(candidates))
	for _, candidate := range candidates {
		sourceID := sourceID(candidate.Media.Path)
		if err := upsertMediaSource(ctx, tx, library.ID, media.KindEpisode, sourceID, candidate.Media, now); err != nil {
			return PersistSummary{}, err
		}
		seriesID := seriesID(candidate.SeriesTitle)
		seasonID := seasonID(seriesID, candidate.SeasonNumber)
		episodeID := episodeRecordID(seriesID, candidate, sourceID)
		episodeIDs[episodeID] = struct{}{}
		if err := upsertSeries(ctx, tx, seriesID, candidate.SeriesTitle, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertSeason(ctx, tx, seasonID, seriesID, candidate.SeasonNumber, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertEpisode(ctx, tx, episodeID, seriesID, seasonID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertEpisodeVersion(ctx, tx, episodeID, sourceID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
	}

	if err := tx.Commit(); err != nil {
		return PersistSummary{}, err
	}
	return PersistSummary{
		LibraryID:    library.ID,
		ScanRunID:    scanRunID,
		MediaSources: len(candidates),
		Episodes:     len(episodeIDs),
	}, nil
}

func upsertLibrary(ctx context.Context, tx *sql.Tx, library libraries.Library) error {
	now := timestamp(time.Now())
	_, err := tx.ExecContext(ctx, `
		INSERT INTO libraries(id, kind, name, path, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			kind = excluded.kind,
			name = excluded.name,
			path = excluded.path,
			updated_at = excluded.updated_at
	`, library.ID, library.Kind, library.Name, library.Path, now)
	return err
}

func upsertMediaSource(ctx context.Context, tx *sql.Tx, libraryID string, kind media.Kind, id string, file scanner.FileCandidate, now string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO media_sources(id, library_id, kind, path, rel_path, name, extension, size_bytes, modified_at, discovered_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(path) DO UPDATE SET
			library_id = excluded.library_id,
			kind = excluded.kind,
			rel_path = excluded.rel_path,
			name = excluded.name,
			extension = excluded.extension,
			size_bytes = excluded.size_bytes,
			modified_at = excluded.modified_at,
			updated_at = excluded.updated_at
	`, id, libraryID, kind, file.Path, file.RelPath, file.Name, file.Extension, file.Size, timestamp(file.ModifiedAt), now, now)
	return err
}

func upsertMovie(ctx context.Context, tx *sql.Tx, id string, candidate movies.Candidate, now string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO movies(id, title, year, sort_title, needs_review, review_reason, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			title = excluded.title,
			year = excluded.year,
			sort_title = excluded.sort_title,
			needs_review = excluded.needs_review,
			review_reason = excluded.review_reason,
			updated_at = excluded.updated_at
	`, id, candidate.Title, candidate.Year, sortTitle(candidate.Title), boolInt(candidate.NeedsReview), candidate.ReviewReason, now)
	return err
}

func upsertMovieVersion(ctx context.Context, tx *sql.Tx, movieID string, sourceID string, candidate movies.Candidate, now string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO movie_versions(movie_id, media_source_id, edition, quality_label, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(movie_id, media_source_id) DO UPDATE SET
			edition = excluded.edition,
			quality_label = excluded.quality_label,
			updated_at = excluded.updated_at
	`, movieID, sourceID, candidate.Edition, candidate.QualityLabel, now)
	return err
}

func upsertSeries(ctx context.Context, tx *sql.Tx, id string, title string, now string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tv_series(id, title, sort_title, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(title) DO UPDATE SET
			sort_title = excluded.sort_title,
			updated_at = excluded.updated_at
	`, id, title, sortTitle(title), now)
	return err
}

func upsertSeason(ctx context.Context, tx *sql.Tx, id string, seriesID string, seasonNumber int, now string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tv_seasons(id, series_id, season_number, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(series_id, season_number) DO UPDATE SET
			updated_at = excluded.updated_at
	`, id, seriesID, seasonNumber, now)
	return err
}

func upsertEpisode(ctx context.Context, tx *sql.Tx, id string, seriesID string, seasonID string, candidate tv.EpisodeCandidate, now string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO tv_episodes(id, series_id, season_id, season_number, episode_number, episode_end, title, needs_review, review_reason, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			series_id = excluded.series_id,
			season_id = excluded.season_id,
			season_number = excluded.season_number,
			episode_number = excluded.episode_number,
			episode_end = excluded.episode_end,
			title = excluded.title,
			needs_review = excluded.needs_review,
			review_reason = excluded.review_reason,
			updated_at = excluded.updated_at
	`, id, seriesID, seasonID, candidate.SeasonNumber, candidate.EpisodeNumber, candidate.EpisodeEnd, candidate.EpisodeTitle, boolInt(candidate.NeedsReview), candidate.ReviewReason, now)
	return err
}

func upsertEpisodeVersion(ctx context.Context, tx *sql.Tx, episodeID string, sourceID string, candidate tv.EpisodeCandidate, now string) error {
	_, err := tx.ExecContext(ctx, `
		INSERT INTO episode_versions(episode_id, media_source_id, quality_label, updated_at)
		VALUES(?, ?, ?, ?)
		ON CONFLICT(episode_id, media_source_id) DO UPDATE SET
			quality_label = excluded.quality_label,
			updated_at = excluded.updated_at
	`, episodeID, sourceID, candidate.QualityLabel, now)
	return err
}

func insertScanRun(ctx context.Context, tx *sql.Tx, summary scanner.Summary) (int64, error) {
	result, err := tx.ExecContext(ctx, `
		INSERT INTO scan_runs(kind, root, started_at, completed_at, duration_ms, total_files, media_files, ignored_files, error_count)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, summary.Kind, summary.Root, timestamp(summary.StartedAt), timestamp(summary.CompletedAt), summary.DurationMS, summary.TotalFiles, summary.MediaFiles, summary.IgnoredFiles, summary.ErrorCount)
	if err != nil {
		return 0, err
	}
	return result.LastInsertId()
}

func sourceID(path string) string {
	return idFor("source", filepath.Clean(path))
}

func movieRecordID(candidate movies.Candidate, sourceID string) string {
	if candidate.NeedsReview && candidate.Year == 0 {
		return idFor("movie-review", sourceID)
	}
	return movieID(candidate.Title, candidate.Year)
}

func movieID(title string, year int) string {
	return idFor("movie", normalize(title), intString(year))
}

func seriesID(title string) string {
	return idFor("series", normalize(title))
}

func seasonID(seriesID string, seasonNumber int) string {
	return idFor("season", seriesID, intString(seasonNumber))
}

func episodeRecordID(seriesID string, candidate tv.EpisodeCandidate, sourceID string) string {
	if candidate.NeedsReview && candidate.EpisodeNumber == 0 {
		return idFor("episode-review", sourceID)
	}
	return episodeID(seriesID, candidate.SeasonNumber, candidate.EpisodeNumber, candidate.EpisodeEnd)
}

func episodeID(seriesID string, seasonNumber int, episodeNumber int, episodeEnd int) string {
	return idFor("episode", seriesID, intString(seasonNumber), intString(episodeNumber), intString(episodeEnd))
}

func idFor(parts ...string) string {
	hash := sha1.New()
	for _, part := range parts {
		hash.Write([]byte(part))
		hash.Write([]byte{0})
	}
	return hex.EncodeToString(hash.Sum(nil))
}

func normalize(value string) string {
	return strings.ToLower(strings.Join(strings.Fields(value), " "))
}

func sortTitle(value string) string {
	return normalize(strings.TrimPrefix(strings.TrimPrefix(value, "The "), "A "))
}

func timestamp(value time.Time) string {
	if value.IsZero() {
		value = time.Now()
	}
	return value.UTC().Format(time.RFC3339Nano)
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}

func intString(value int) string {
	return strconv.Itoa(value)
}
