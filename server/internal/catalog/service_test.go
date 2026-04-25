package catalog

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
)

func TestSaveMovieScanIsIdempotent(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "movies", Name: "Movies", Path: `X:\Movies`, Kind: libraries.KindMovies}
	result := scanResult(scanner.KindMovies, library.Path, []scanner.FileCandidate{fileCandidate(`X:\Movies\Heat (1995)\Heat.1995.1080p.mkv`, "Heat (1995)/Heat.1995.1080p.mkv")})
	candidates := []movies.Candidate{{
		Title:        "Heat",
		Year:         1995,
		QualityLabel: "1080p",
		Media:        result.Files[0],
	}}

	if _, err := service.SaveMovieScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save movie scan: %v", err)
	}
	if _, err := service.SaveMovieScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save movie scan again: %v", err)
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if summary.Libraries != 1 {
		t.Fatalf("expected one library, got %d", summary.Libraries)
	}
	if summary.MediaSources != 1 {
		t.Fatalf("expected one media source, got %d", summary.MediaSources)
	}
	if summary.Movies != 1 {
		t.Fatalf("expected one movie, got %d", summary.Movies)
	}
	if summary.ScanRuns != 2 {
		t.Fatalf("expected two scan runs, got %d", summary.ScanRuns)
	}

	movies, err := service.ListMovies(ctx, 10)
	if err != nil {
		t.Fatalf("list movies: %v", err)
	}
	if len(movies) != 1 || movies[0].Title != "Heat" || movies[0].VersionCount != 1 {
		t.Fatalf("unexpected movies: %#v", movies)
	}
	detail, ok, err := service.GetMovie(ctx, movies[0].ID)
	if err != nil {
		t.Fatalf("get movie: %v", err)
	}
	if !ok || len(detail.Versions) != 1 {
		t.Fatalf("unexpected movie detail: %#v", detail)
	}
	if detail.Metadata == nil || detail.Metadata.Provider != "filename" || detail.Metadata.Title != "Heat" || detail.Metadata.Year != 1995 {
		t.Fatalf("expected filename metadata on movie detail, got %#v", detail.Metadata)
	}
	records, err := service.ListMetadataRecords(ctx, "movie", movies[0].ID)
	if err != nil {
		t.Fatalf("list metadata records: %v", err)
	}
	if len(records) != 1 || records[0].Provider != "filename" {
		t.Fatalf("expected filename metadata record, got %#v", records)
	}
	if err := service.ApplyMetadata(ctx, MetadataUpdate{
		Kind:     "movie",
		ID:       movies[0].ID,
		Title:    "Heat",
		Year:     1995,
		Provider: "manual",
		Overview: "A professional thief weighs one last score.",
	}); err != nil {
		t.Fatalf("apply metadata: %v", err)
	}
	best, ok, err := service.GetBestMetadata(ctx, "movie", movies[0].ID)
	if err != nil {
		t.Fatalf("best metadata: %v", err)
	}
	if !ok || best.Provider != "manual" || best.Overview == "" {
		t.Fatalf("expected manual metadata to win, got %#v", best)
	}
	sources, err := service.ListMediaSources(ctx, 10, true)
	if err != nil {
		t.Fatalf("list media sources: %v", err)
	}
	if len(sources) != 1 || sources[0].Probed {
		t.Fatalf("expected one unprobed source, got %#v", sources)
	}
	if err := service.SaveProbe(ctx, sources[0].ID, ProbeResult{
		Container:       "matroska,webm",
		DurationSeconds: 120,
		Bitrate:         1000,
		VideoCodec:      "h264",
		Width:           1920,
		Height:          1080,
		AudioStreams:    1,
		SubtitleStreams: 1,
		RawJSON:         "{}",
	}); err != nil {
		t.Fatalf("save probe: %v", err)
	}
	source, ok, err := service.GetMediaSource(ctx, sources[0].ID)
	if err != nil {
		t.Fatalf("get media source: %v", err)
	}
	if !ok || !source.Probed || source.VideoCodec != "h264" {
		t.Fatalf("unexpected probed source: %#v", source)
	}
}

func TestSaveTVScanStoresSeriesEpisodeAndSource(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "tv", Name: "TV", Path: `X:\TV`, Kind: libraries.KindTV}
	result := scanResult(scanner.KindTV, library.Path, []scanner.FileCandidate{fileCandidate(`X:\TV\The Bear\Season 02\The.Bear.S02E03.Sundae.mkv`, "The Bear/Season 02/The.Bear.S02E03.Sundae.mkv")})
	candidates := []tv.EpisodeCandidate{{
		SeriesTitle:   "The Bear",
		SeasonNumber:  2,
		EpisodeNumber: 3,
		EpisodeTitle:  "Sundae",
		QualityLabel:  "1080p",
		Media:         result.Files[0],
	}}

	if _, err := service.SaveTVScan(ctx, library, result, candidates); err != nil {
		t.Fatalf("save tv scan: %v", err)
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}

	if summary.Series != 1 {
		t.Fatalf("expected one series, got %d", summary.Series)
	}
	if summary.Episodes != 1 {
		t.Fatalf("expected one episode, got %d", summary.Episodes)
	}
	if summary.MediaSources != 1 {
		t.Fatalf("expected one media source, got %d", summary.MediaSources)
	}

	series, err := service.ListSeries(ctx, 10)
	if err != nil {
		t.Fatalf("list series: %v", err)
	}
	if len(series) != 1 || series[0].Title != "The Bear" || series[0].EpisodeCount != 1 {
		t.Fatalf("unexpected series: %#v", series)
	}
	detail, ok, err := service.GetSeries(ctx, series[0].ID)
	if err != nil {
		t.Fatalf("get series: %v", err)
	}
	if !ok || len(detail.Seasons) != 1 || len(detail.Seasons[0].Episodes) != 1 {
		t.Fatalf("unexpected series detail: %#v", detail)
	}
	if detail.Metadata == nil || detail.Metadata.Provider != "filename" || detail.Metadata.Title != "The Bear" {
		t.Fatalf("expected filename metadata on series detail, got %#v", detail.Metadata)
	}
}

func TestSaveTVScanKeepsUnmatchedEpisodesSeparate(t *testing.T) {
	ctx := context.Background()
	service := newTestService(t)
	library := libraries.Library{ID: "tv", Name: "TV", Path: `X:\TV`, Kind: libraries.KindTV}
	files := []scanner.FileCandidate{
		fileCandidate(`X:\TV\Series\Special.Video.A.mkv`, "Series/Special.Video.A.mkv"),
		fileCandidate(`X:\TV\Series\Special.Video.B.mkv`, "Series/Special.Video.B.mkv"),
	}
	result := scanResult(scanner.KindTV, library.Path, files)
	candidates := []tv.EpisodeCandidate{
		{SeriesTitle: "Series", NeedsReview: true, ReviewReason: "unable to infer episode number", Media: files[0]},
		{SeriesTitle: "Series", NeedsReview: true, ReviewReason: "unable to infer episode number", Media: files[1]},
	}

	persisted, err := service.SaveTVScan(ctx, library, result, candidates)
	if err != nil {
		t.Fatalf("save tv scan: %v", err)
	}
	if persisted.Episodes != 2 {
		t.Fatalf("expected two separate review episodes, got %d", persisted.Episodes)
	}
	summary, err := service.Summary(ctx)
	if err != nil {
		t.Fatalf("summary: %v", err)
	}
	if summary.Episodes != 2 {
		t.Fatalf("expected two stored review episodes, got %d", summary.Episodes)
	}
	reviewItems, err := service.ReviewItems(ctx, 10)
	if err != nil {
		t.Fatalf("review items: %v", err)
	}
	if len(reviewItems) != 2 {
		t.Fatalf("expected two review items, got %#v", reviewItems)
	}
}

func newTestService(t *testing.T) *Service {
	t.Helper()

	db, err := database.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})
	return NewService(db)
}

func scanResult(kind scanner.LibraryKind, root string, files []scanner.FileCandidate) scanner.Result {
	startedAt := time.Now().UTC()
	return scanner.Result{
		Summary: scanner.Summary{
			Kind:        kind,
			Root:        root,
			StartedAt:   startedAt,
			CompletedAt: startedAt.Add(time.Millisecond),
			DurationMS:  1,
			TotalFiles:  len(files),
			MediaFiles:  len(files),
			Extensions:  map[string]int{".mkv": len(files)},
		},
		Files: files,
	}
}

func fileCandidate(path string, relPath string) scanner.FileCandidate {
	return scanner.FileCandidate{
		Path:       filepath.Clean(path),
		RelPath:    filepath.Clean(relPath),
		Name:       filepath.Base(path),
		Extension:  ".mkv",
		Size:       1024,
		ModifiedAt: time.Now().UTC(),
	}
}
