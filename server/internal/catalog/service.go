package catalog

import (
	"context"
	"crypto/sha1"
	"database/sql"
	"encoding/hex"
	"errors"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/probe"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
)

type Service struct {
	db *sql.DB
}

type PersistSummary struct {
	LibraryID        string `json:"libraryId"`
	ScanRunID        int64  `json:"scanRunId"`
	MediaSources     int    `json:"mediaSources"`
	ChangedSources   int    `json:"changedSources"`
	UnchangedSources int    `json:"unchangedSources"`
	Movies           int    `json:"movies,omitempty"`
	Episodes         int    `json:"episodes,omitempty"`
}

type RuntimeSettings struct {
	HTTPAddr          string `json:"httpAddr"`
	DataDir           string `json:"dataDir"`
	TranscodeDir      string `json:"transcodeDir"`
	DownloadsDir      string `json:"downloadsDir"`
	MetadataDir       string `json:"metadataDir"`
	CacheDir          string `json:"cacheDir"`
	TempDir           string `json:"tempDir"`
	FFmpegPath        string `json:"ffmpegPath"`
	FFprobePath       string `json:"ffprobePath"`
	ScanWorkers       int    `json:"scanWorkers"`
	ProbeWorkers      int    `json:"probeWorkers"`
	TranscodeWorkers  int    `json:"transcodeWorkers"`
	GPUWorkers        int    `json:"gpuWorkers"`
	LibrarySyncMode   string `json:"librarySyncMode"`
	SyncIntervalMins  int    `json:"syncIntervalMins"`
	WatchDebounceSecs int    `json:"watchDebounceSecs"`
	ProbeBatchLimit   int    `json:"probeBatchLimit"`
}

type Summary struct {
	Libraries    int `json:"libraries"`
	MediaSources int `json:"mediaSources"`
	Movies       int `json:"movies"`
	Series       int `json:"series"`
	Episodes     int `json:"episodes"`
	ScanRuns     int `json:"scanRuns"`
	Unprobed     int `json:"unprobed"`
}

type Health struct {
	Summary       Summary `json:"summary"`
	NeedsReview   int     `json:"needsReview"`
	Unprobed      int     `json:"unprobed"`
	Unsupported   int     `json:"unsupported"`
	HighBitrate   int     `json:"highBitrate"`
	WithSubtitles int     `json:"withSubtitles"`
}

type MovieListItem struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	Year         int             `json:"year"`
	SortTitle    string          `json:"sortTitle"`
	NeedsReview  bool            `json:"needsReview"`
	VersionCount int             `json:"versionCount"`
	Metadata     *MetadataRecord `json:"metadata,omitempty"`
}

type MovieDetail struct {
	MovieListItem
	Versions []MovieVersion `json:"versions"`
}

type MovieVersion struct {
	MediaSourceID string `json:"mediaSourceId"`
	Path          string `json:"path"`
	RelPath       string `json:"relPath"`
	Edition       string `json:"edition,omitempty"`
	QualityLabel  string `json:"qualityLabel,omitempty"`
	SizeBytes     int64  `json:"sizeBytes"`
	ModifiedAt    string `json:"modifiedAt"`
}

type SeriesListItem struct {
	ID           string          `json:"id"`
	Title        string          `json:"title"`
	SortTitle    string          `json:"sortTitle"`
	SeasonCount  int             `json:"seasonCount"`
	EpisodeCount int             `json:"episodeCount"`
	Metadata     *MetadataRecord `json:"metadata,omitempty"`
}

type SeriesDetail struct {
	SeriesListItem
	Seasons []SeasonDetail `json:"seasons"`
}

type SeasonDetail struct {
	ID           string         `json:"id"`
	SeasonNumber int            `json:"seasonNumber"`
	Episodes     []EpisodeBrief `json:"episodes"`
}

type EpisodeBrief struct {
	ID            string         `json:"id"`
	SeasonNumber  int            `json:"seasonNumber"`
	EpisodeNumber int            `json:"episodeNumber"`
	EpisodeEnd    int            `json:"episodeEnd,omitempty"`
	Title         string         `json:"title,omitempty"`
	NeedsReview   bool           `json:"needsReview"`
	VersionCount  int            `json:"versionCount"`
	Versions      []MovieVersion `json:"versions,omitempty"`
}

type ReviewItem struct {
	Kind         string `json:"kind"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	ReviewReason string `json:"reviewReason"`
}

type MetadataRecord struct {
	Kind        string            `json:"kind"`
	ItemID      string            `json:"itemId"`
	Provider    string            `json:"provider"`
	ExternalID  string            `json:"externalId,omitempty"`
	Title       string            `json:"title"`
	Year        int               `json:"year,omitempty"`
	Overview    string            `json:"overview,omitempty"`
	PosterURL   string            `json:"posterUrl,omitempty"`
	BackdropURL string            `json:"backdropUrl,omitempty"`
	Confidence  float64           `json:"confidence"`
	RawJSON     string            `json:"rawJson,omitempty"`
	FetchedAt   string            `json:"fetchedAt"`
	UpdatedAt   string            `json:"updatedAt"`
	Ratings     Ratings           `json:"ratings,omitempty"`
	ExternalIDs map[string]string `json:"externalIds,omitempty"`
}

type Rating struct {
	Kind         string  `json:"kind"`
	ItemID       string  `json:"itemId"`
	Provider     string  `json:"provider"`
	RatingType   string  `json:"ratingType"`
	Value        float64 `json:"value"`
	DisplayValue string  `json:"displayValue"`
	Scale        float64 `json:"scale,omitempty"`
	Votes        int     `json:"votes,omitempty"`
	SourceURL    string  `json:"sourceUrl,omitempty"`
	FetchedAt    string  `json:"fetchedAt"`
	UpdatedAt    string  `json:"updatedAt"`
}

type Ratings map[string]any

type ExternalID struct {
	Kind       string `json:"kind"`
	ItemID     string `json:"itemId"`
	Provider   string `json:"provider"`
	ExternalID string `json:"externalId"`
	UpdatedAt  string `json:"updatedAt"`
}

type VersionGroup struct {
	Kind         string `json:"kind"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	VersionCount int    `json:"versionCount"`
}

type MetadataUpdate struct {
	Kind        string `json:"kind"`
	ID          string `json:"id"`
	Title       string `json:"title"`
	Year        int    `json:"year,omitempty"`
	Overview    string `json:"overview,omitempty"`
	Provider    string `json:"provider,omitempty"`
	ExternalID  string `json:"externalId,omitempty"`
	PosterURL   string `json:"posterUrl,omitempty"`
	BackdropURL string `json:"backdropUrl,omitempty"`
	Review      bool   `json:"review"`
}

type MediaSourceItem struct {
	ID              string  `json:"id"`
	LibraryID       string  `json:"libraryId"`
	Kind            string  `json:"kind"`
	Path            string  `json:"path"`
	RelPath         string  `json:"relPath"`
	Name            string  `json:"name"`
	Extension       string  `json:"extension"`
	SizeBytes       int64   `json:"sizeBytes"`
	ModifiedAt      string  `json:"modifiedAt"`
	Probed          bool    `json:"probed"`
	Container       string  `json:"container,omitempty"`
	DurationSeconds float64 `json:"durationSeconds,omitempty"`
	Bitrate         int64   `json:"bitrate,omitempty"`
	VideoCodec      string  `json:"videoCodec,omitempty"`
	Width           int     `json:"width,omitempty"`
	Height          int     `json:"height,omitempty"`
	AudioStreams    int     `json:"audioStreams,omitempty"`
	SubtitleStreams int     `json:"subtitleStreams,omitempty"`
}

type MediaSourceDisplay struct {
	Kind         string `json:"kind"`
	ItemID       string `json:"itemId"`
	Title        string `json:"title"`
	ArtworkKind  string `json:"artworkKind"`
	ArtworkID    string `json:"artworkId"`
	QualityLabel string `json:"qualityLabel,omitempty"`
}

type MediaSourceTracks struct {
	AudioTracks    []probe.Track `json:"audioTracks"`
	SubtitleTracks []probe.Track `json:"subtitleTracks"`
}

type ProbeResult struct {
	Container       string  `json:"container"`
	DurationSeconds float64 `json:"durationSeconds"`
	Bitrate         int64   `json:"bitrate"`
	VideoCodec      string  `json:"videoCodec"`
	Width           int     `json:"width"`
	Height          int     `json:"height"`
	AudioStreams    int     `json:"audioStreams"`
	SubtitleStreams int     `json:"subtitleStreams"`
	RawJSON         string  `json:"rawJson"`
}

func NewService(database *database.Service) *Service {
	return &Service{db: database.DB()}
}

func (s *Service) ScanState(ctx context.Context, libraryID string) (map[string]scanner.FileSignature, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT rel_path, size_bytes, modified_at
		FROM scan_file_state
		WHERE library_id = ?
	`, libraryID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	output := map[string]scanner.FileSignature{}
	for rows.Next() {
		var relPath string
		var size int64
		var modifiedAt string
		if err := rows.Scan(&relPath, &size, &modifiedAt); err != nil {
			return nil, err
		}
		output[relPath] = scanner.FileSignature{Size: size, ModifiedAt: parseTimestamp(modifiedAt)}
	}
	return output, rows.Err()
}

func (s *Service) ListLibraries(ctx context.Context) ([]libraries.Library, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT id, kind, name, path, storage_type
		FROM libraries
		ORDER BY kind, name, path
	`)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	output := []libraries.Library{}
	for rows.Next() {
		var item libraries.Library
		if err := rows.Scan(&item.ID, &item.Kind, &item.Name, &item.Path, &item.StorageType); err != nil {
			return nil, err
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) SaveLibrary(ctx context.Context, library libraries.Library) (libraries.Library, error) {
	if strings.TrimSpace(library.Path) == "" {
		return libraries.Library{}, errors.New("library path is required")
	}
	if library.Kind != libraries.KindMovies && library.Kind != libraries.KindTV {
		return libraries.Library{}, errors.New("library kind must be movies or tv")
	}
	if strings.TrimSpace(library.Name) == "" {
		if library.Kind == libraries.KindMovies {
			library.Name = "Movies"
		} else {
			library.Name = "TV"
		}
	}
	if library.ID == "" {
		library.ID = libraries.IDFor(library.Kind, library.Path)
	}
	if library.StorageType == "" {
		library.StorageType = libraries.DetectStorageType(library.Path)
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return libraries.Library{}, err
	}
	defer tx.Rollback()
	if err := upsertLibrary(ctx, tx, library); err != nil {
		return libraries.Library{}, err
	}
	return library, tx.Commit()
}

func (s *Service) DeleteLibrary(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, "DELETE FROM libraries WHERE id = ?", id)
	return err
}

func (s *Service) SaveSettings(ctx context.Context, settings RuntimeSettings) error {
	now := timestamp(time.Now())
	values := map[string]string{
		"httpAddr":          settings.HTTPAddr,
		"dataDir":           settings.DataDir,
		"transcodeDir":      settings.TranscodeDir,
		"downloadsDir":      settings.DownloadsDir,
		"metadataDir":       settings.MetadataDir,
		"cacheDir":          settings.CacheDir,
		"tempDir":           settings.TempDir,
		"ffmpegPath":        settings.FFmpegPath,
		"ffprobePath":       settings.FFprobePath,
		"scanWorkers":       intString(settings.ScanWorkers),
		"probeWorkers":      intString(settings.ProbeWorkers),
		"transcodeWorkers":  intString(settings.TranscodeWorkers),
		"gpuWorkers":        intString(settings.GPUWorkers),
		"librarySyncMode":   settings.LibrarySyncMode,
		"syncIntervalMins":  intString(settings.SyncIntervalMins),
		"watchDebounceSecs": intString(settings.WatchDebounceSecs),
		"probeBatchLimit":   intString(settings.ProbeBatchLimit),
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	for key, value := range values {
		if value == "" || value == "0" {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO app_settings(key, value, updated_at)
			VALUES(?, ?, ?)
			ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = excluded.updated_at
		`, key, value, now); err != nil {
			return err
		}
	}
	return tx.Commit()
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
	_ = s.db.QueryRowContext(ctx, "SELECT count(*) FROM media_sources ms LEFT JOIN media_probes mp ON mp.media_source_id = ms.id WHERE mp.media_source_id IS NULL").Scan(&summary.Unprobed)
	return summary, nil
}

func (s *Service) Health(ctx context.Context) (Health, error) {
	summary, err := s.Summary(ctx)
	if err != nil {
		return Health{}, err
	}
	health := Health{Summary: summary, Unprobed: summary.Unprobed}
	queries := []struct {
		query string
		out   *int
	}{
		{"SELECT (SELECT count(*) FROM movies WHERE needs_review = 1) + (SELECT count(*) FROM tv_episodes WHERE needs_review = 1)", &health.NeedsReview},
		{"SELECT count(*) FROM media_probes WHERE video_codec NOT IN ('h264', 'av1', 'vp9')", &health.Unsupported},
		{"SELECT count(*) FROM media_probes WHERE bitrate > 40000000", &health.HighBitrate},
		{"SELECT count(*) FROM media_probes WHERE subtitle_streams > 0", &health.WithSubtitles},
	}
	for _, item := range queries {
		if err := s.db.QueryRowContext(ctx, item.query).Scan(item.out); err != nil {
			return Health{}, err
		}
	}
	return health, nil
}

func (s *Service) ListMovies(ctx context.Context, limit int) ([]MovieListItem, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT m.id, m.title, m.year, m.sort_title, m.needs_review, count(mv.media_source_id) AS version_count
		FROM movies m
		LEFT JOIN movie_versions mv ON mv.movie_id = m.id
		GROUP BY m.id
		ORDER BY m.sort_title, m.year
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	output := []MovieListItem{}
	for rows.Next() {
		var item MovieListItem
		var needsReview int
		if err := rows.Scan(&item.ID, &item.Title, &item.Year, &item.SortTitle, &needsReview, &item.VersionCount); err != nil {
			return nil, err
		}
		item.NeedsReview = needsReview != 0
		if record, ok, err := s.GetBestMetadata(ctx, "movie", item.ID); err != nil {
			return nil, err
		} else if ok {
			item.Metadata = &record
			applyMovieMetadata(&item.Title, &item.Year, &item.SortTitle, record)
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) GetMovie(ctx context.Context, id string) (MovieDetail, bool, error) {
	var detail MovieDetail
	var needsReview int
	err := s.db.QueryRowContext(ctx, `
		SELECT m.id, m.title, m.year, m.sort_title, m.needs_review, count(mv.media_source_id) AS version_count
		FROM movies m
		LEFT JOIN movie_versions mv ON mv.movie_id = m.id
		WHERE m.id = ?
		GROUP BY m.id
	`, id).Scan(&detail.ID, &detail.Title, &detail.Year, &detail.SortTitle, &needsReview, &detail.VersionCount)
	if errors.Is(err, sql.ErrNoRows) {
		return MovieDetail{}, false, nil
	}
	if err != nil {
		return MovieDetail{}, false, err
	}
	detail.NeedsReview = needsReview != 0
	if record, ok, err := s.GetBestMetadata(ctx, "movie", id); err != nil {
		return MovieDetail{}, false, err
	} else if ok {
		detail.Metadata = &record
		applyMovieMetadata(&detail.Title, &detail.Year, &detail.SortTitle, record)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT ms.id, ms.path, ms.rel_path, mv.edition, mv.quality_label, ms.size_bytes, ms.modified_at
		FROM movie_versions mv
		JOIN media_sources ms ON ms.id = mv.media_source_id
		WHERE mv.movie_id = ?
		ORDER BY mv.quality_label DESC, ms.rel_path
	`, id)
	if err != nil {
		return MovieDetail{}, false, err
	}
	defer rows.Close()
	for rows.Next() {
		var version MovieVersion
		if err := rows.Scan(&version.MediaSourceID, &version.Path, &version.RelPath, &version.Edition, &version.QualityLabel, &version.SizeBytes, &version.ModifiedAt); err != nil {
			return MovieDetail{}, false, err
		}
		detail.Versions = append(detail.Versions, version)
	}
	return detail, true, rows.Err()
}

func (s *Service) ListSeries(ctx context.Context, limit int) ([]SeriesListItem, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT s.id, s.title, s.sort_title, count(DISTINCT seasons.id) AS season_count, count(DISTINCT e.id) AS episode_count
		FROM tv_series s
		LEFT JOIN tv_seasons seasons ON seasons.series_id = s.id
		LEFT JOIN tv_episodes e ON e.series_id = s.id
		GROUP BY s.id
		ORDER BY s.sort_title
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	output := []SeriesListItem{}
	for rows.Next() {
		var item SeriesListItem
		if err := rows.Scan(&item.ID, &item.Title, &item.SortTitle, &item.SeasonCount, &item.EpisodeCount); err != nil {
			return nil, err
		}
		if record, ok, err := s.GetBestMetadata(ctx, "series", item.ID); err != nil {
			return nil, err
		} else if ok {
			item.Metadata = &record
			applyTitleMetadata(&item.Title, &item.SortTitle, record)
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) GetSeries(ctx context.Context, id string) (SeriesDetail, bool, error) {
	var detail SeriesDetail
	err := s.db.QueryRowContext(ctx, `
		SELECT s.id, s.title, s.sort_title, count(DISTINCT seasons.id) AS season_count, count(DISTINCT e.id) AS episode_count
		FROM tv_series s
		LEFT JOIN tv_seasons seasons ON seasons.series_id = s.id
		LEFT JOIN tv_episodes e ON e.series_id = s.id
		WHERE s.id = ?
		GROUP BY s.id
	`, id).Scan(&detail.ID, &detail.Title, &detail.SortTitle, &detail.SeasonCount, &detail.EpisodeCount)
	if errors.Is(err, sql.ErrNoRows) {
		return SeriesDetail{}, false, nil
	}
	if err != nil {
		return SeriesDetail{}, false, err
	}
	if record, ok, err := s.GetBestMetadata(ctx, "series", id); err != nil {
		return SeriesDetail{}, false, err
	} else if ok {
		detail.Metadata = &record
		applyTitleMetadata(&detail.Title, &detail.SortTitle, record)
	}

	rows, err := s.db.QueryContext(ctx, `
		SELECT seasons.id, seasons.season_number, e.id, e.season_number, e.episode_number, e.episode_end, e.title, e.needs_review, count(ev.media_source_id) AS version_count
		FROM tv_seasons seasons
		LEFT JOIN tv_episodes e ON e.season_id = seasons.id
		LEFT JOIN episode_versions ev ON ev.episode_id = e.id
		WHERE seasons.series_id = ?
		GROUP BY seasons.id, e.id
		ORDER BY seasons.season_number, e.episode_number
	`, id)
	if err != nil {
		return SeriesDetail{}, false, err
	}
	defer rows.Close()

	seasonIndex := map[string]int{}
	episodeLocation := map[string][2]int{}
	for rows.Next() {
		var seasonID string
		var seasonNumber int
		var episode EpisodeBrief
		var needsReview int
		if err := rows.Scan(&seasonID, &seasonNumber, &episode.ID, &episode.SeasonNumber, &episode.EpisodeNumber, &episode.EpisodeEnd, &episode.Title, &needsReview, &episode.VersionCount); err != nil {
			return SeriesDetail{}, false, err
		}
		index, ok := seasonIndex[seasonID]
		if !ok {
			index = len(detail.Seasons)
			seasonIndex[seasonID] = index
			detail.Seasons = append(detail.Seasons, SeasonDetail{ID: seasonID, SeasonNumber: seasonNumber})
		}
		if episode.ID != "" {
			episode.NeedsReview = needsReview != 0
			detail.Seasons[index].Episodes = append(detail.Seasons[index].Episodes, episode)
			episodeLocation[episode.ID] = [2]int{index, len(detail.Seasons[index].Episodes) - 1}
		}
	}
	if err := rows.Err(); err != nil {
		return SeriesDetail{}, false, err
	}

	versionRows, err := s.db.QueryContext(ctx, `
		SELECT ev.episode_id, ms.id, ms.path, ms.rel_path, '' AS edition, ev.quality_label, ms.size_bytes, ms.modified_at
		FROM episode_versions ev
		JOIN media_sources ms ON ms.id = ev.media_source_id
		JOIN tv_episodes e ON e.id = ev.episode_id
		WHERE e.series_id = ?
		ORDER BY e.season_number, e.episode_number, ev.quality_label DESC
	`, id)
	if err != nil {
		return SeriesDetail{}, false, err
	}
	defer versionRows.Close()
	for versionRows.Next() {
		var episodeID string
		var version MovieVersion
		if err := versionRows.Scan(&episodeID, &version.MediaSourceID, &version.Path, &version.RelPath, &version.Edition, &version.QualityLabel, &version.SizeBytes, &version.ModifiedAt); err != nil {
			return SeriesDetail{}, false, err
		}
		location, ok := episodeLocation[episodeID]
		if ok {
			episode := &detail.Seasons[location[0]].Episodes[location[1]]
			episode.Versions = append(episode.Versions, version)
		}
	}
	return detail, true, versionRows.Err()
}

func (s *Service) ReviewItems(ctx context.Context, limit int) ([]ReviewItem, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT 'movie' AS kind, id, title, review_reason
		FROM movies
		WHERE needs_review = 1
		UNION ALL
		SELECT 'episode' AS kind, id, title, review_reason
		FROM tv_episodes
		WHERE needs_review = 1
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	output := []ReviewItem{}
	for rows.Next() {
		var item ReviewItem
		if err := rows.Scan(&item.Kind, &item.ID, &item.Title, &item.ReviewReason); err != nil {
			return nil, err
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) VersionGroups(ctx context.Context, limit int) ([]VersionGroup, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT 'movie', m.id, m.title, count(mv.media_source_id) AS version_count
		FROM movies m
		JOIN movie_versions mv ON mv.movie_id = m.id
		GROUP BY m.id
		HAVING version_count > 1
		UNION ALL
		SELECT 'episode', e.id, e.title, count(ev.media_source_id) AS version_count
		FROM tv_episodes e
		JOIN episode_versions ev ON ev.episode_id = e.id
		GROUP BY e.id
		HAVING version_count > 1
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	output := []VersionGroup{}
	for rows.Next() {
		var item VersionGroup
		if err := rows.Scan(&item.Kind, &item.ID, &item.Title, &item.VersionCount); err != nil {
			return nil, err
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) ApplyMetadata(ctx context.Context, update MetadataUpdate) error {
	update.Title = strings.TrimSpace(update.Title)
	if update.Kind == "" || update.ID == "" || update.Title == "" {
		return errors.New("kind, id, and title are required")
	}
	now := timestamp(time.Now())
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()

	switch update.Kind {
	case "movie":
		_, err = tx.ExecContext(ctx, `
			UPDATE movies
			SET title = ?, year = ?, sort_title = ?, needs_review = ?, review_reason = '', updated_at = ?
			WHERE id = ?
		`, update.Title, update.Year, sortTitle(update.Title), boolInt(update.Review), now, update.ID)
	case "series":
		_, err = tx.ExecContext(ctx, `
			UPDATE tv_series
			SET title = ?, sort_title = ?, updated_at = ?
			WHERE id = ?
		`, update.Title, sortTitle(update.Title), now, update.ID)
	case "episode":
		_, err = tx.ExecContext(ctx, `
			UPDATE tv_episodes
			SET title = ?, needs_review = ?, review_reason = '', updated_at = ?
			WHERE id = ?
		`, update.Title, boolInt(update.Review), now, update.ID)
	default:
		return errors.New("metadata kind must be movie, series, or episode")
	}
	if err != nil {
		return err
	}
	if _, err = tx.ExecContext(ctx, `
		INSERT INTO metadata_overrides(kind, item_id, title, year, review_resolved, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(kind, item_id) DO UPDATE SET
			title = excluded.title,
			year = excluded.year,
			review_resolved = excluded.review_resolved,
			updated_at = excluded.updated_at
	`, update.Kind, update.ID, update.Title, update.Year, boolInt(!update.Review), now); err != nil {
		return err
	}
	provider := update.Provider
	if provider == "" {
		provider = "manual"
	}
	if err := upsertMetadataRecord(ctx, tx, MetadataRecord{
		Kind:        update.Kind,
		ItemID:      update.ID,
		Provider:    provider,
		ExternalID:  update.ExternalID,
		Title:       update.Title,
		Year:        update.Year,
		Overview:    update.Overview,
		PosterURL:   update.PosterURL,
		BackdropURL: update.BackdropURL,
		Confidence:  1,
		FetchedAt:   now,
		UpdatedAt:   now,
	}); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) UpsertMetadataRecord(ctx context.Context, record MetadataRecord) error {
	switch record.Kind {
	case "movie", "series", "episode":
	default:
		return errors.New("metadata kind must be movie, series, or episode")
	}
	if strings.TrimSpace(record.ItemID) == "" || strings.TrimSpace(record.Title) == "" {
		return errors.New("metadata item id and title are required")
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	if err := upsertMetadataRecord(ctx, tx, record); err != nil {
		return err
	}
	return tx.Commit()
}

func (s *Service) GetBestMetadata(ctx context.Context, kind string, itemID string) (MetadataRecord, bool, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT kind, item_id, provider, external_id, title, year, overview, poster_url, backdrop_url, confidence, raw_json, fetched_at, updated_at
		FROM metadata_records
		WHERE kind = ? AND item_id = ?
		ORDER BY CASE provider WHEN 'manual' THEN 0 WHEN 'tmdb' THEN 1 WHEN 'tvdb' THEN 2 WHEN 'omdb' THEN 3 ELSE 9 END, confidence DESC, updated_at DESC
		LIMIT 1
	`, kind, itemID)
	if err != nil {
		return MetadataRecord{}, false, err
	}
	defer rows.Close()
	records, err := scanMetadataRecords(rows)
	if err != nil {
		return MetadataRecord{}, false, err
	}
	if len(records) == 0 {
		return MetadataRecord{}, false, nil
	}
	if err := s.AttachMetadataSignals(ctx, &records[0]); err != nil {
		return MetadataRecord{}, false, err
	}
	return records[0], true, nil
}

func (s *Service) ListMetadataRecords(ctx context.Context, kind string, itemID string) ([]MetadataRecord, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT kind, item_id, provider, external_id, title, year, overview, poster_url, backdrop_url, confidence, raw_json, fetched_at, updated_at
		FROM metadata_records
		WHERE kind = ? AND item_id = ?
		ORDER BY CASE provider WHEN 'manual' THEN 0 WHEN 'tmdb' THEN 1 WHEN 'tvdb' THEN 2 WHEN 'omdb' THEN 3 WHEN 'filename' THEN 8 ELSE 9 END, confidence DESC, updated_at DESC
	`, kind, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	records, err := scanMetadataRecords(rows)
	if err != nil {
		return nil, err
	}
	for index := range records {
		if err := s.AttachMetadataSignals(ctx, &records[index]); err != nil {
			return nil, err
		}
	}
	return records, nil
}

func (s *Service) AttachMetadataSignals(ctx context.Context, record *MetadataRecord) error {
	if record == nil || record.Kind == "" || record.ItemID == "" {
		return nil
	}
	ratings, err := s.ListRatings(ctx, record.Kind, record.ItemID)
	if err != nil {
		return err
	}
	record.Ratings = RatingMap(ratings)
	externalIDs, err := s.ListExternalIDs(ctx, record.Kind, record.ItemID)
	if err != nil {
		return err
	}
	record.ExternalIDs = map[string]string{}
	for _, item := range externalIDs {
		record.ExternalIDs[item.Provider] = item.ExternalID
	}
	return nil
}

func RatingMap(ratings []Rating) Ratings {
	output := Ratings{}
	for _, rating := range ratings {
		key := rating.Provider
		if rating.RatingType != "" && rating.RatingType != rating.Provider {
			key = rating.RatingType
		}
		if rating.DisplayValue != "" {
			output[key] = rating.DisplayValue
		} else {
			output[key] = rating.Value
		}
	}
	return output
}

func scanMetadataRecords(rows *sql.Rows) ([]MetadataRecord, error) {
	output := []MetadataRecord{}
	for rows.Next() {
		var item MetadataRecord
		if err := rows.Scan(
			&item.Kind,
			&item.ItemID,
			&item.Provider,
			&item.ExternalID,
			&item.Title,
			&item.Year,
			&item.Overview,
			&item.PosterURL,
			&item.BackdropURL,
			&item.Confidence,
			&item.RawJSON,
			&item.FetchedAt,
			&item.UpdatedAt,
		); err != nil {
			return nil, err
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) UpsertExternalID(ctx context.Context, item ExternalID) error {
	item.Kind = strings.TrimSpace(item.Kind)
	item.ItemID = strings.TrimSpace(item.ItemID)
	item.Provider = strings.TrimSpace(item.Provider)
	item.ExternalID = strings.TrimSpace(item.ExternalID)
	if item.Kind == "" || item.ItemID == "" || item.Provider == "" || item.ExternalID == "" {
		return errors.New("kind, item id, provider, and external id are required")
	}
	now := timestamp(time.Now())
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO metadata_external_ids(kind, item_id, provider, external_id, updated_at)
		VALUES(?, ?, ?, ?, ?)
		ON CONFLICT(kind, item_id, provider) DO UPDATE SET
			external_id = excluded.external_id,
			updated_at = excluded.updated_at
	`, item.Kind, item.ItemID, item.Provider, item.ExternalID, now)
	return err
}

func (s *Service) ListExternalIDs(ctx context.Context, kind string, itemID string) ([]ExternalID, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT kind, item_id, provider, external_id, updated_at
		FROM metadata_external_ids
		WHERE kind = ? AND item_id = ?
		ORDER BY provider
	`, kind, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	output := []ExternalID{}
	for rows.Next() {
		var item ExternalID
		if err := rows.Scan(&item.Kind, &item.ItemID, &item.Provider, &item.ExternalID, &item.UpdatedAt); err != nil {
			return nil, err
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) UpsertRatings(ctx context.Context, ratings []Rating) error {
	if len(ratings) == 0 {
		return nil
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	defer tx.Rollback()
	now := timestamp(time.Now())
	for _, rating := range ratings {
		rating.Kind = strings.TrimSpace(rating.Kind)
		rating.ItemID = strings.TrimSpace(rating.ItemID)
		rating.Provider = strings.TrimSpace(rating.Provider)
		rating.RatingType = strings.TrimSpace(rating.RatingType)
		if rating.Kind == "" || rating.ItemID == "" || rating.Provider == "" || rating.RatingType == "" {
			return errors.New("rating kind, item id, provider, and type are required")
		}
		fetchedAt := rating.FetchedAt
		if fetchedAt == "" {
			fetchedAt = now
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO metadata_ratings(kind, item_id, provider, rating_type, value, display_value, scale, votes, source_url, fetched_at, updated_at)
			VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
			ON CONFLICT(kind, item_id, provider, rating_type) DO UPDATE SET
				value = excluded.value,
				display_value = excluded.display_value,
				scale = excluded.scale,
				votes = excluded.votes,
				source_url = excluded.source_url,
				fetched_at = excluded.fetched_at,
				updated_at = excluded.updated_at
		`, rating.Kind, rating.ItemID, rating.Provider, rating.RatingType, rating.Value, rating.DisplayValue, rating.Scale, rating.Votes, rating.SourceURL, fetchedAt, now); err != nil {
			return err
		}
	}
	return tx.Commit()
}

func (s *Service) ListRatings(ctx context.Context, kind string, itemID string) ([]Rating, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT kind, item_id, provider, rating_type, value, display_value, scale, votes, source_url, fetched_at, updated_at
		FROM metadata_ratings
		WHERE kind = ? AND item_id = ?
		ORDER BY CASE rating_type
			WHEN 'imdb' THEN 0
			WHEN 'rottenTomatoesCritics' THEN 1
			WHEN 'rottenTomatoesAudience' THEN 2
			WHEN 'tmdb' THEN 3
			WHEN 'metacritic' THEN 4
			WHEN 'tvdb' THEN 5
			ELSE 9 END, provider
	`, kind, itemID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	output := []Rating{}
	for rows.Next() {
		var item Rating
		if err := rows.Scan(&item.Kind, &item.ItemID, &item.Provider, &item.RatingType, &item.Value, &item.DisplayValue, &item.Scale, &item.Votes, &item.SourceURL, &item.FetchedAt, &item.UpdatedAt); err != nil {
			return nil, err
		}
		output = append(output, item)
	}
	return output, rows.Err()
}

func (s *Service) ListMediaSources(ctx context.Context, limit int, unprobedOnly bool) ([]MediaSourceItem, error) {
	if limit <= 0 || limit > 500 {
		limit = 100
	}
	filter := ""
	if unprobedOnly {
		filter = "WHERE mp.media_source_id IS NULL"
	}
	rows, err := s.db.QueryContext(ctx, `
		SELECT ms.id, ms.library_id, ms.kind, ms.path, ms.rel_path, ms.name, ms.extension, ms.size_bytes, ms.modified_at,
			mp.container, mp.duration_seconds, mp.bitrate, mp.video_codec, mp.width, mp.height, mp.audio_streams, mp.subtitle_streams
		FROM media_sources ms
		LEFT JOIN media_probes mp ON mp.media_source_id = ms.id
		`+filter+`
		ORDER BY ms.rel_path
		LIMIT ?
	`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return scanMediaSources(rows)
}

func (s *Service) GetMediaSource(ctx context.Context, id string) (MediaSourceItem, bool, error) {
	rows, err := s.db.QueryContext(ctx, `
		SELECT ms.id, ms.library_id, ms.kind, ms.path, ms.rel_path, ms.name, ms.extension, ms.size_bytes, ms.modified_at,
			mp.container, mp.duration_seconds, mp.bitrate, mp.video_codec, mp.width, mp.height, mp.audio_streams, mp.subtitle_streams
		FROM media_sources ms
		LEFT JOIN media_probes mp ON mp.media_source_id = ms.id
		WHERE ms.id = ?
	`, id)
	if err != nil {
		return MediaSourceItem{}, false, err
	}
	defer rows.Close()
	items, err := scanMediaSources(rows)
	if err != nil {
		return MediaSourceItem{}, false, err
	}
	if len(items) == 0 {
		return MediaSourceItem{}, false, nil
	}
	return items[0], true, nil
}

func (s *Service) GetMediaSourceDisplay(ctx context.Context, id string) (MediaSourceDisplay, bool, error) {
	var movieID string
	var movieTitle string
	var movieYear int
	var movieQuality sql.NullString
	err := s.db.QueryRowContext(ctx, `
		SELECT m.id, m.title, m.year, mv.quality_label
		FROM movie_versions mv
		JOIN movies m ON m.id = mv.movie_id
		WHERE mv.media_source_id = ?
		LIMIT 1
	`, id).Scan(&movieID, &movieTitle, &movieYear, &movieQuality)
	if err != nil && !errors.Is(err, sql.ErrNoRows) {
		return MediaSourceDisplay{}, false, err
	}
	if err == nil {
		sortTitle := sortTitle(movieTitle)
		if record, ok, err := s.GetBestMetadata(ctx, "movie", movieID); err != nil {
			return MediaSourceDisplay{}, false, err
		} else if ok {
			applyMovieMetadata(&movieTitle, &movieYear, &sortTitle, record)
		}
		return MediaSourceDisplay{
			Kind:         "movie",
			ItemID:       movieID,
			Title:        titleWithYear(movieTitle, movieYear),
			ArtworkKind:  "movie",
			ArtworkID:    movieID,
			QualityLabel: movieQuality.String,
		}, true, nil
	}

	var episodeID string
	var episodeTitle string
	var episodeNumber int
	var episodeEnd int
	var seriesID string
	var seriesTitle string
	var seasonNumber int
	var episodeQuality sql.NullString
	err = s.db.QueryRowContext(ctx, `
		SELECT e.id, e.title, e.episode_number, e.episode_end, s.id, s.title, seasons.season_number, ev.quality_label
		FROM episode_versions ev
		JOIN tv_episodes e ON e.id = ev.episode_id
		JOIN tv_series s ON s.id = e.series_id
		JOIN tv_seasons seasons ON seasons.id = e.season_id
		WHERE ev.media_source_id = ?
		LIMIT 1
	`, id).Scan(&episodeID, &episodeTitle, &episodeNumber, &episodeEnd, &seriesID, &seriesTitle, &seasonNumber, &episodeQuality)
	if errors.Is(err, sql.ErrNoRows) {
		return MediaSourceDisplay{}, false, nil
	}
	if err != nil {
		return MediaSourceDisplay{}, false, err
	}
	seriesSortTitle := sortTitle(seriesTitle)
	if record, ok, err := s.GetBestMetadata(ctx, "series", seriesID); err != nil {
		return MediaSourceDisplay{}, false, err
	} else if ok {
		applyTitleMetadata(&seriesTitle, &seriesSortTitle, record)
	}
	return MediaSourceDisplay{
		Kind:         "episode",
		ItemID:       episodeID,
		Title:        episodeDisplayTitle(seriesTitle, seasonNumber, episodeNumber, episodeEnd, episodeTitle),
		ArtworkKind:  "series",
		ArtworkID:    seriesID,
		QualityLabel: episodeQuality.String,
	}, true, nil
}

func (s *Service) GetMediaSourceTracks(ctx context.Context, id string) (MediaSourceTracks, bool, error) {
	var raw string
	err := s.db.QueryRowContext(ctx, "SELECT raw_json FROM media_probes WHERE media_source_id = ?", id).Scan(&raw)
	if errors.Is(err, sql.ErrNoRows) {
		return MediaSourceTracks{}, false, nil
	}
	if err != nil {
		return MediaSourceTracks{}, false, err
	}
	result, err := probe.Parse([]byte(raw))
	if err != nil {
		return MediaSourceTracks{}, false, err
	}
	return MediaSourceTracks{AudioTracks: result.AudioTracks, SubtitleTracks: result.SubtitleTracks}, true, nil
}

func (s *Service) SaveProbe(ctx context.Context, mediaSourceID string, result ProbeResult) error {
	_, err := s.db.ExecContext(ctx, `
		INSERT INTO media_probes(media_source_id, container, duration_seconds, bitrate, video_codec, width, height, audio_streams, subtitle_streams, raw_json, probed_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(media_source_id) DO UPDATE SET
			container = excluded.container,
			duration_seconds = excluded.duration_seconds,
			bitrate = excluded.bitrate,
			video_codec = excluded.video_codec,
			width = excluded.width,
			height = excluded.height,
			audio_streams = excluded.audio_streams,
			subtitle_streams = excluded.subtitle_streams,
			raw_json = excluded.raw_json,
			probed_at = excluded.probed_at
	`, mediaSourceID, result.Container, result.DurationSeconds, result.Bitrate, result.VideoCodec, result.Width, result.Height, result.AudioStreams, result.SubtitleStreams, result.RawJSON, timestamp(time.Now()))
	return err
}

func scanMediaSources(rows *sql.Rows) ([]MediaSourceItem, error) {
	output := []MediaSourceItem{}
	for rows.Next() {
		var item MediaSourceItem
		var container sql.NullString
		var duration sql.NullFloat64
		var bitrate sql.NullInt64
		var videoCodec sql.NullString
		var width sql.NullInt64
		var height sql.NullInt64
		var audioStreams sql.NullInt64
		var subtitleStreams sql.NullInt64
		if err := rows.Scan(&item.ID, &item.LibraryID, &item.Kind, &item.Path, &item.RelPath, &item.Name, &item.Extension, &item.SizeBytes, &item.ModifiedAt, &container, &duration, &bitrate, &videoCodec, &width, &height, &audioStreams, &subtitleStreams); err != nil {
			return nil, err
		}
		item.Probed = container.Valid
		if container.Valid {
			item.Container = container.String
			item.DurationSeconds = duration.Float64
			item.Bitrate = bitrate.Int64
			item.VideoCodec = videoCodec.String
			item.Width = int(width.Int64)
			item.Height = int(height.Int64)
			item.AudioStreams = int(audioStreams.Int64)
			item.SubtitleStreams = int(subtitleStreams.Int64)
		}
		output = append(output, item)
	}
	return output, rows.Err()
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
	sourceIDs := observedSourceIDs(result)
	for _, candidate := range candidates {
		sourceID := sourceID(candidate.Media.Path)
		if err := upsertScanFileState(ctx, tx, library.ID, candidate.Media, now); err != nil {
			return PersistSummary{}, err
		}
		movieID := movieRecordID(candidate, sourceID)
		movieIDs[movieID] = struct{}{}
		if !candidate.Media.Changed {
			continue
		}
		if err := upsertMediaSource(ctx, tx, library.ID, media.KindMovie, sourceID, candidate.Media, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertMovie(ctx, tx, movieID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertMetadataRecord(ctx, tx, MetadataRecord{
			Kind:       "movie",
			ItemID:     movieID,
			Provider:   "filename",
			Title:      candidate.Title,
			Year:       candidate.Year,
			Confidence: metadataConfidence(candidate.NeedsReview),
			FetchedAt:  now,
			UpdatedAt:  now,
		}); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertMovieVersion(ctx, tx, movieID, sourceID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
	}
	if err := deleteMissingSources(ctx, tx, library.ID, sourceIDs); err != nil {
		return PersistSummary{}, err
	}
	if err := touchSeenScanState(ctx, tx, library.ID, result.SeenRelPaths, now); err != nil {
		return PersistSummary{}, err
	}
	if err := deleteMissingScanState(ctx, tx, library.ID, result.SeenRelPaths); err != nil {
		return PersistSummary{}, err
	}

	if err := tx.Commit(); err != nil {
		return PersistSummary{}, err
	}
	return PersistSummary{
		LibraryID:        library.ID,
		ScanRunID:        scanRunID,
		MediaSources:     len(result.SeenRelPaths),
		ChangedSources:   result.ChangedFiles,
		UnchangedSources: result.Unchanged,
		Movies:           len(movieIDs),
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
	sourceIDs := observedSourceIDs(result)
	for _, candidate := range candidates {
		sourceID := sourceID(candidate.Media.Path)
		if err := upsertScanFileState(ctx, tx, library.ID, candidate.Media, now); err != nil {
			return PersistSummary{}, err
		}
		seriesID := seriesID(candidate.SeriesTitle)
		seasonID := seasonID(seriesID, candidate.SeasonNumber)
		episodeID := episodeRecordID(seriesID, candidate, sourceID)
		episodeIDs[episodeID] = struct{}{}
		if !candidate.Media.Changed {
			continue
		}
		if err := upsertMediaSource(ctx, tx, library.ID, media.KindEpisode, sourceID, candidate.Media, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertSeries(ctx, tx, seriesID, candidate.SeriesTitle, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertMetadataRecord(ctx, tx, MetadataRecord{
			Kind:       "series",
			ItemID:     seriesID,
			Provider:   "filename",
			Title:      candidate.SeriesTitle,
			Confidence: metadataConfidence(candidate.SeriesTitle == ""),
			FetchedAt:  now,
			UpdatedAt:  now,
		}); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertSeason(ctx, tx, seasonID, seriesID, candidate.SeasonNumber, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertEpisode(ctx, tx, episodeID, seriesID, seasonID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertMetadataRecord(ctx, tx, MetadataRecord{
			Kind:       "episode",
			ItemID:     episodeID,
			Provider:   "filename",
			Title:      candidate.EpisodeTitle,
			Confidence: metadataConfidence(candidate.NeedsReview || candidate.EpisodeTitle == ""),
			FetchedAt:  now,
			UpdatedAt:  now,
		}); err != nil {
			return PersistSummary{}, err
		}
		if err := upsertEpisodeVersion(ctx, tx, episodeID, sourceID, candidate, now); err != nil {
			return PersistSummary{}, err
		}
	}
	if err := deleteMissingSources(ctx, tx, library.ID, sourceIDs); err != nil {
		return PersistSummary{}, err
	}
	if err := touchSeenScanState(ctx, tx, library.ID, result.SeenRelPaths, now); err != nil {
		return PersistSummary{}, err
	}
	if err := deleteMissingScanState(ctx, tx, library.ID, result.SeenRelPaths); err != nil {
		return PersistSummary{}, err
	}

	if err := tx.Commit(); err != nil {
		return PersistSummary{}, err
	}
	return PersistSummary{
		LibraryID:        library.ID,
		ScanRunID:        scanRunID,
		MediaSources:     len(result.SeenRelPaths),
		ChangedSources:   result.ChangedFiles,
		UnchangedSources: result.Unchanged,
		Episodes:         len(episodeIDs),
	}, nil
}

func upsertLibrary(ctx context.Context, tx *sql.Tx, library libraries.Library) error {
	now := timestamp(time.Now())
	if library.StorageType == "" {
		library.StorageType = libraries.DetectStorageType(library.Path)
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO libraries(id, kind, name, path, storage_type, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(id) DO UPDATE SET
			kind = excluded.kind,
			name = excluded.name,
			path = excluded.path,
			storage_type = excluded.storage_type,
			updated_at = excluded.updated_at
	`, library.ID, library.Kind, library.Name, library.Path, library.StorageType, now)
	return err
}

func upsertMetadataRecord(ctx context.Context, tx *sql.Tx, record MetadataRecord) error {
	record.Kind = strings.TrimSpace(record.Kind)
	record.ItemID = strings.TrimSpace(record.ItemID)
	record.Provider = strings.TrimSpace(record.Provider)
	record.Title = strings.TrimSpace(record.Title)
	if record.Provider == "" {
		record.Provider = "filename"
	}
	now := timestamp(time.Now())
	if record.FetchedAt == "" {
		record.FetchedAt = now
	}
	if record.UpdatedAt == "" {
		record.UpdatedAt = now
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO metadata_records(kind, item_id, provider, external_id, title, year, overview, poster_url, backdrop_url, confidence, raw_json, fetched_at, updated_at)
		VALUES(?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(kind, item_id, provider) DO UPDATE SET
			external_id = excluded.external_id,
			title = excluded.title,
			year = excluded.year,
			overview = excluded.overview,
			poster_url = excluded.poster_url,
			backdrop_url = excluded.backdrop_url,
			confidence = excluded.confidence,
			raw_json = excluded.raw_json,
			fetched_at = excluded.fetched_at,
			updated_at = excluded.updated_at
	`, record.Kind, record.ItemID, record.Provider, record.ExternalID, record.Title, record.Year, record.Overview, record.PosterURL, record.BackdropURL, record.Confidence, record.RawJSON, record.FetchedAt, record.UpdatedAt)
	return err
}

func metadataConfidence(needsReview bool) float64 {
	if needsReview {
		return 0.35
	}
	return 0.7
}

func applyMovieMetadata(title *string, year *int, sort *string, record MetadataRecord) {
	applyTitleMetadata(title, sort, record)
	if record.Year != 0 {
		*year = record.Year
	}
}

func applyTitleMetadata(title *string, sort *string, record MetadataRecord) {
	if record.Title == "" {
		return
	}
	*title = record.Title
	*sort = sortTitle(record.Title)
}

func titleWithYear(title string, year int) string {
	title = strings.TrimSpace(title)
	if year <= 0 {
		return title
	}
	return title + " (" + intString(year) + ")"
}

func episodeDisplayTitle(seriesTitle string, seasonNumber int, episodeNumber int, episodeEnd int, episodeTitle string) string {
	episodeCode := "S" + twoDigit(seasonNumber) + "E" + twoDigit(episodeNumber)
	if episodeEnd > episodeNumber {
		episodeCode += "-E" + twoDigit(episodeEnd)
	}
	parts := []string{strings.TrimSpace(seriesTitle), episodeCode}
	if strings.TrimSpace(episodeTitle) != "" {
		parts = append(parts, strings.TrimSpace(episodeTitle))
	}
	return strings.Join(parts, " - ")
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

func upsertScanFileState(ctx context.Context, tx *sql.Tx, libraryID string, file scanner.FileCandidate, now string) error {
	changedAt := now
	if !file.Changed {
		changedAt = timestamp(file.ModifiedAt)
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO scan_file_state(library_id, rel_path, size_bytes, modified_at, last_seen_at, changed_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(library_id, rel_path) DO UPDATE SET
			size_bytes = excluded.size_bytes,
			modified_at = excluded.modified_at,
			last_seen_at = excluded.last_seen_at,
			changed_at = CASE
				WHEN scan_file_state.size_bytes != excluded.size_bytes OR scan_file_state.modified_at != excluded.modified_at THEN excluded.changed_at
				ELSE scan_file_state.changed_at
			END
	`, libraryID, file.RelPath, file.Size, timestamp(file.ModifiedAt), now, changedAt)
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

func deleteMissingSources(ctx context.Context, tx *sql.Tx, libraryID string, sourceIDs []string) error {
	if len(sourceIDs) == 0 {
		_, err := tx.ExecContext(ctx, "DELETE FROM media_sources WHERE library_id = ?", libraryID)
		return err
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(sourceIDs)), ",")
	args := make([]any, 0, len(sourceIDs)+1)
	args = append(args, libraryID)
	for _, id := range sourceIDs {
		args = append(args, id)
	}
	_, err := tx.ExecContext(ctx, "DELETE FROM media_sources WHERE library_id = ? AND id NOT IN ("+placeholders+")", args...)
	return err
}

func observedSourceIDs(result scanner.Result) []string {
	relPaths := result.SeenRelPaths
	if len(relPaths) == 0 {
		relPaths = make([]string, 0, len(result.Files))
		for _, file := range result.Files {
			relPaths = append(relPaths, file.RelPath)
		}
	}
	fileIDs := make(map[string]string, len(result.Files))
	for _, file := range result.Files {
		fileIDs[file.RelPath] = sourceID(file.Path)
	}
	output := make([]string, 0, len(relPaths))
	for _, relPath := range relPaths {
		if id := fileIDs[relPath]; id != "" {
			output = append(output, id)
		} else {
			output = append(output, sourceID(filepath.Join(result.Root, relPath)))
		}
	}
	return output
}

func deleteMissingScanState(ctx context.Context, tx *sql.Tx, libraryID string, relPaths []string) error {
	if len(relPaths) == 0 {
		_, err := tx.ExecContext(ctx, "DELETE FROM scan_file_state WHERE library_id = ?", libraryID)
		return err
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(relPaths)), ",")
	args := make([]any, 0, len(relPaths)+1)
	args = append(args, libraryID)
	for _, relPath := range relPaths {
		args = append(args, relPath)
	}
	_, err := tx.ExecContext(ctx, "DELETE FROM scan_file_state WHERE library_id = ? AND rel_path NOT IN ("+placeholders+")", args...)
	return err
}

func touchSeenScanState(ctx context.Context, tx *sql.Tx, libraryID string, relPaths []string, now string) error {
	if len(relPaths) == 0 {
		return nil
	}
	placeholders := strings.TrimRight(strings.Repeat("?,", len(relPaths)), ",")
	args := make([]any, 0, len(relPaths)+2)
	args = append(args, now, libraryID)
	for _, relPath := range relPaths {
		args = append(args, relPath)
	}
	_, err := tx.ExecContext(ctx, "UPDATE scan_file_state SET last_seen_at = ? WHERE library_id = ? AND rel_path IN ("+placeholders+")", args...)
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

func parseTimestamp(value string) time.Time {
	parsed, err := time.Parse(time.RFC3339Nano, value)
	if err != nil {
		return time.Time{}
	}
	return parsed
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

func twoDigit(value int) string {
	if value < 10 {
		return "0" + intString(value)
	}
	return intString(value)
}
