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
	LibraryID    string `json:"libraryId"`
	ScanRunID    int64  `json:"scanRunId"`
	MediaSources int    `json:"mediaSources"`
	Movies       int    `json:"movies,omitempty"`
	Episodes     int    `json:"episodes,omitempty"`
}

type RuntimeSettings struct {
	HTTPAddr         string `json:"httpAddr"`
	DataDir          string `json:"dataDir"`
	FFmpegPath       string `json:"ffmpegPath"`
	FFprobePath      string `json:"ffprobePath"`
	ScanWorkers      int    `json:"scanWorkers"`
	ProbeWorkers     int    `json:"probeWorkers"`
	TranscodeWorkers int    `json:"transcodeWorkers"`
	GPUWorkers       int    `json:"gpuWorkers"`
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
	ID           string `json:"id"`
	Title        string `json:"title"`
	Year         int    `json:"year"`
	SortTitle    string `json:"sortTitle"`
	NeedsReview  bool   `json:"needsReview"`
	VersionCount int    `json:"versionCount"`
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
	ID           string `json:"id"`
	Title        string `json:"title"`
	SortTitle    string `json:"sortTitle"`
	SeasonCount  int    `json:"seasonCount"`
	EpisodeCount int    `json:"episodeCount"`
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

type VersionGroup struct {
	Kind         string `json:"kind"`
	ID           string `json:"id"`
	Title        string `json:"title"`
	VersionCount int    `json:"versionCount"`
}

type MetadataUpdate struct {
	Kind   string `json:"kind"`
	ID     string `json:"id"`
	Title  string `json:"title"`
	Year   int    `json:"year,omitempty"`
	Review bool   `json:"review"`
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
		"httpAddr":         settings.HTTPAddr,
		"dataDir":          settings.DataDir,
		"ffmpegPath":       settings.FFmpegPath,
		"ffprobePath":      settings.FFprobePath,
		"scanWorkers":      intString(settings.ScanWorkers),
		"probeWorkers":     intString(settings.ProbeWorkers),
		"transcodeWorkers": intString(settings.TranscodeWorkers),
		"gpuWorkers":       intString(settings.GPUWorkers),
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
	case "episode":
		_, err = tx.ExecContext(ctx, `
			UPDATE tv_episodes
			SET title = ?, needs_review = ?, review_reason = '', updated_at = ?
			WHERE id = ?
		`, update.Title, boolInt(update.Review), now, update.ID)
	default:
		return errors.New("metadata kind must be movie or episode")
	}
	if err != nil {
		return err
	}
	_, err = tx.ExecContext(ctx, `
		INSERT INTO metadata_overrides(kind, item_id, title, year, review_resolved, updated_at)
		VALUES(?, ?, ?, ?, ?, ?)
		ON CONFLICT(kind, item_id) DO UPDATE SET
			title = excluded.title,
			year = excluded.year,
			review_resolved = excluded.review_resolved,
			updated_at = excluded.updated_at
	`, update.Kind, update.ID, update.Title, update.Year, boolInt(!update.Review), now)
	if err != nil {
		return err
	}
	return tx.Commit()
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
	sourceIDs := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		sourceID := sourceID(candidate.Media.Path)
		sourceIDs = append(sourceIDs, sourceID)
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
	if err := deleteMissingSources(ctx, tx, library.ID, sourceIDs); err != nil {
		return PersistSummary{}, err
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
	sourceIDs := make([]string, 0, len(candidates))
	for _, candidate := range candidates {
		sourceID := sourceID(candidate.Media.Path)
		sourceIDs = append(sourceIDs, sourceID)
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
	if err := deleteMissingSources(ctx, tx, library.ID, sourceIDs); err != nil {
		return PersistSummary{}, err
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
