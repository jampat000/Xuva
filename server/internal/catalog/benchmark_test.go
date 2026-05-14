package catalog

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/jampat000/Xuva/server/internal/database"
	"github.com/jampat000/Xuva/server/internal/libraries"
	"github.com/jampat000/Xuva/server/internal/movies"
	"github.com/jampat000/Xuva/server/internal/scanner"
)

func BenchmarkCatalogMovieBrowse(b *testing.B) {
	ctx := context.Background()
	service := newBenchmarkService(b)
	library := libraries.Library{ID: "movies", Name: "Movies", Path: `X:\Movies`, Kind: libraries.KindMovies}
	files := make([]scanner.FileCandidate, 0, 1000)
	candidates := make([]movies.Candidate, 0, 1000)
	for i := 0; i < 1000; i++ {
		title := fmt.Sprintf("Movie %04d", i)
		file := scanner.FileCandidate{
			Path:       filepath.Clean(fmt.Sprintf(`X:\Movies\%s (2024)\%s.2024.mkv`, title, title)),
			RelPath:    filepath.Clean(fmt.Sprintf(`%s (2024)\%s.2024.mkv`, title, title)),
			Name:       title + ".2024.mkv",
			Extension:  ".mkv",
			Size:       1024,
			ModifiedAt: time.Now().UTC(),
			Changed:    true,
		}
		files = append(files, file)
		candidates = append(candidates, movies.Candidate{Title: title, Year: 2024, QualityLabel: "1080p", Media: file})
	}
	if _, err := service.SaveMovieScan(ctx, library, scanResult(scanner.KindMovies, library.Path, files), candidates); err != nil {
		b.Fatalf("save scan: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.ListMovies(ctx, 100); err != nil {
			b.Fatalf("list movies: %v", err)
		}
	}
}

func BenchmarkCatalogMovieScanPersistFull(b *testing.B) {
	ctx := context.Background()
	library, files, candidates := benchmarkMovieCandidates(1000)
	service := newBenchmarkService(b)
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.SaveMovieScan(ctx, library, scanResult(scanner.KindMovies, library.Path, files), candidates); err != nil {
			b.Fatalf("save full scan: %v", err)
		}
	}
}

func BenchmarkCatalogMovieScanPersistIncrementalLowChange(b *testing.B) {
	ctx := context.Background()
	library, files, candidates := benchmarkMovieCandidates(1000)
	service := newBenchmarkService(b)
	if _, err := service.SaveMovieScan(ctx, library, scanResult(scanner.KindMovies, library.Path, files), candidates); err != nil {
		b.Fatalf("seed full scan: %v", err)
	}
	for index := range files {
		files[index].Changed = false
	}
	files[0].Changed = true
	candidates = []movies.Candidate{{Title: "Movie 0000", Year: 2024, QualityLabel: "1080p", Media: files[0]}}
	result := scanResult(scanner.KindMovies, library.Path, files)
	result.Files = []scanner.FileCandidate{files[0]}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.SaveMovieScan(ctx, library, result, candidates); err != nil {
			b.Fatalf("save incremental scan: %v", err)
		}
	}
}

func benchmarkMovieCandidates(count int) (libraries.Library, []scanner.FileCandidate, []movies.Candidate) {
	library := libraries.Library{ID: "movies", Name: "Movies", Path: `X:\Movies`, Kind: libraries.KindMovies}
	files := make([]scanner.FileCandidate, 0, count)
	candidates := make([]movies.Candidate, 0, count)
	for i := 0; i < count; i++ {
		title := fmt.Sprintf("Movie %04d", i)
		file := scanner.FileCandidate{
			Path:       filepath.Clean(fmt.Sprintf(`X:\Movies\%s (2024)\%s.2024.mkv`, title, title)),
			RelPath:    filepath.Clean(fmt.Sprintf(`%s (2024)\%s.2024.mkv`, title, title)),
			Name:       title + ".2024.mkv",
			Extension:  ".mkv",
			Size:       1024,
			ModifiedAt: time.Now().UTC(),
			Changed:    true,
		}
		files = append(files, file)
		candidates = append(candidates, movies.Candidate{Title: title, Year: 2024, QualityLabel: "1080p", Media: file})
	}
	return library, files, candidates
}

func newBenchmarkService(b *testing.B) *Service {
	b.Helper()
	db, err := database.Open(context.Background(), b.TempDir())
	if err != nil {
		b.Fatalf("open database: %v", err)
	}
	b.Cleanup(func() { _ = db.Close() })
	return NewService(db)
}
