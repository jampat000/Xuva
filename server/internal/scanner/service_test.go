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

func TestScanMarksKnownFilesUnchangedAndPrioritizesChanges(t *testing.T) {
	root := t.TempDir()
	unchanged := filepath.Join(root, "a-unchanged.mkv")
	changed := filepath.Join(root, "b-changed.mkv")
	writeFile(t, unchanged)
	writeFile(t, changed)
	info, err := os.Stat(unchanged)
	if err != nil {
		t.Fatalf("stat unchanged: %v", err)
	}

	result, err := NewService().Scan(context.Background(), Request{
		Kind: KindMovies,
		Root: root,
		KnownFiles: map[string]FileSignature{
			"a-unchanged.mkv": {Size: info.Size(), ModifiedAt: info.ModTime().UTC()},
		},
	})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if result.ChangedFiles != 1 || result.Unchanged != 1 {
		t.Fatalf("expected one changed and one unchanged file, got %#v", result.Summary)
	}
	if len(result.Files) != 2 || !result.Files[0].Changed {
		t.Fatalf("expected changed files first, got %#v", result.Files)
	}
}

func TestScanCanSkipUnchangedPayloads(t *testing.T) {
	root := t.TempDir()
	unchanged := filepath.Join(root, "unchanged.mkv")
	writeFile(t, unchanged)
	info, err := os.Stat(unchanged)
	if err != nil {
		t.Fatalf("stat unchanged: %v", err)
	}
	result, err := NewService().Scan(context.Background(), Request{
		Kind:          KindMovies,
		Root:          root,
		SkipUnchanged: true,
		KnownFiles: map[string]FileSignature{
			"unchanged.mkv": {Size: info.Size(), ModifiedAt: info.ModTime().UTC()},
		},
	})
	if err != nil {
		t.Fatalf("scan: %v", err)
	}
	if result.MediaFiles != 1 || result.Unchanged != 1 || len(result.Files) != 0 || len(result.SeenRelPaths) != 1 {
		t.Fatalf("expected unchanged file to be counted but omitted from payload, got %#v", result)
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
