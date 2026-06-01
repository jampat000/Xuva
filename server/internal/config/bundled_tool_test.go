package config

import (
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestResolveBundledTool(t *testing.T) {
	name := "ffmpeg"
	binName := name
	if runtime.GOOS == "windows" {
		binName += ".exe"
	}

	// Absent: falls back to the bare name (PATH lookup).
	dir := t.TempDir()
	if got := resolveBundledTool(dir, name); got != name {
		t.Errorf("absent: got %q, want %q", got, name)
	}

	// Present as a file: returns the bundled path.
	binDir := filepath.Join(dir, "bin")
	if err := os.MkdirAll(binDir, 0o755); err != nil {
		t.Fatal(err)
	}
	bundled := filepath.Join(binDir, binName)
	if err := os.WriteFile(bundled, []byte("stub"), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := resolveBundledTool(dir, name); got != bundled {
		t.Errorf("present: got %q, want %q", got, bundled)
	}

	// A directory with the tool's name is ignored (must be a regular file).
	dir2 := t.TempDir()
	if err := os.MkdirAll(filepath.Join(dir2, "bin", binName), 0o755); err != nil {
		t.Fatal(err)
	}
	if got := resolveBundledTool(dir2, name); got != name {
		t.Errorf("dir-not-file: got %q, want %q", got, name)
	}
}
