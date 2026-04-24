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
