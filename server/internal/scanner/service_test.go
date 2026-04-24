package scanner

import (
	"context"
	"os"
	"path/filepath"
	"testing"
)

func TestScanFindsMediaAndIgnoresNoise(t *testing.T) {
	root := t.TempDir()
	writeFile(t, filepath.Join(root, "Movie (2024)", "Movie.2024.mkv"))
	writeFile(t, filepath.Join(root, "Movie (2024)", "poster.jpg"))
	writeFile(t, filepath.Join(root, ".hidden", "Hidden.2024.mp4"))
	writeFile(t, filepath.Join(root, "@eaDir", "Synology.2024.mp4"))

	result, err := NewService().Scan(context.Background(), Request{Kind: KindMovies, Root: root})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}

	if result.MediaFiles != 1 {
		t.Fatalf("expected one media file, got %d", result.MediaFiles)
	}
	if result.IgnoredFiles != 1 {
		t.Fatalf("expected one ignored non-media file, got %d", result.IgnoredFiles)
	}
	if len(result.Files) != 1 || result.Files[0].Extension != ".mkv" {
		t.Fatalf("expected mkv candidate, got %#v", result.Files)
	}
}

func TestScanRequiresDirectoryRoot(t *testing.T) {
	_, err := NewService().Scan(context.Background(), Request{Kind: KindMovies})
	if err != ErrMissingRoot {
		t.Fatalf("expected ErrMissingRoot, got %v", err)
	}
}

func writeFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
