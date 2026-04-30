package scanner

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"testing"
)

func BenchmarkScanFullLibrary(b *testing.B) {
	root := benchmarkLibrary(b, 1000)
	service := NewService()
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Scan(context.Background(), Request{Kind: KindMovies, Root: root}); err != nil {
			b.Fatalf("scan: %v", err)
		}
	}
}

func BenchmarkScanIncrementalLowChangeLibrary(b *testing.B) {
	root := benchmarkLibrary(b, 1000)
	service := NewService()
	initial, err := service.Scan(context.Background(), Request{Kind: KindMovies, Root: root})
	if err != nil {
		b.Fatalf("initial scan: %v", err)
	}
	known := make(map[string]FileSignature, len(initial.Files))
	for _, file := range initial.Files {
		known[file.RelPath] = FileSignature{Size: file.Size, ModifiedAt: file.ModifiedAt}
	}
	changedPath := filepath.Join(root, "Movie 0001 (2024)", "Movie.0001.2024.mkv")
	if err := os.WriteFile(changedPath, []byte("changed"), 0o644); err != nil {
		b.Fatalf("change file: %v", err)
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		if _, err := service.Scan(context.Background(), Request{Kind: KindMovies, Root: root, KnownFiles: known, SkipUnchanged: true}); err != nil {
			b.Fatalf("incremental scan: %v", err)
		}
	}
}

func benchmarkLibrary(b *testing.B, files int) string {
	b.Helper()
	root := b.TempDir()
	for i := 0; i < files; i++ {
		dir := filepath.Join(root, fmt.Sprintf("Movie %04d (2024)", i))
		if err := os.MkdirAll(dir, 0o755); err != nil {
			b.Fatalf("mkdir: %v", err)
		}
		path := filepath.Join(dir, fmt.Sprintf("Movie.%04d.2024.mkv", i))
		if err := os.WriteFile(path, []byte("video"), 0o644); err != nil {
			b.Fatalf("write: %v", err)
		}
	}
	return root
}
