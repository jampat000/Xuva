package smb

import (
	"errors"
	"os"
	"path/filepath"
	"runtime"
	"testing"
)

func TestIsUNC(t *testing.T) {
	cases := map[string]bool{
		`\\NAS\Media`:       true,
		`\\192.168.1.10\tv`: true,
		`//nas/media`:       true,
		`  \\nas\share  `:   true,
		`C:\Users\me`:       false,
		`/mnt/media`:        false,
		`relative\path`:     false,
		``:                  false,
		`\single`:           false,
	}
	for in, want := range cases {
		if got := IsUNC(in); got != want {
			t.Errorf("IsUNC(%q) = %v, want %v", in, got, want)
		}
	}
}

func TestNormalizeUNC(t *testing.T) {
	cases := map[string]string{
		`//nas/media/`:   `\\nas\media`,
		` \\nas\media\ `: `\\nas\media`,
		`\\nas\media`:    `\\nas\media`,
		`//host/a/b/`:    `\\host\a\b`,
	}
	for in, want := range cases {
		if got := normalizeUNC(in); got != want {
			t.Errorf("normalizeUNC(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestListDirs(t *testing.T) {
	root := t.TempDir()
	for _, name := range []string{"Zebra", "alpha", "Movies"} {
		if err := os.Mkdir(filepath.Join(root, name), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	if err := os.WriteFile(filepath.Join(root, "afile.txt"), []byte("x"), 0o644); err != nil {
		t.Fatal(err)
	}
	got, err := listDirs(root)
	if err != nil {
		t.Fatalf("listDirs: %v", err)
	}
	// directories only, sorted case-insensitively
	want := []string{"alpha", "Movies", "Zebra"}
	if len(got) != len(want) {
		t.Fatalf("got %d entries, want %d (%v)", len(got), len(want), got)
	}
	for i, e := range got {
		if e.Name != want[i] {
			t.Errorf("entry %d = %q, want %q", i, e.Name, want[i])
		}
		if !e.IsDir {
			t.Errorf("entry %q not marked IsDir", e.Name)
		}
	}
}

// On non-Windows, the credentialed entry points must cleanly report that the
// feature is unavailable rather than doing anything surprising.
func TestUnsupportedOnNonWindows(t *testing.T) {
	if runtime.GOOS == "windows" {
		t.Skip("Windows has a real implementation")
	}
	if err := Validate(`\\nas\media`, "user", "pass"); !errors.Is(err, ErrUnsupported) {
		t.Errorf("Validate err = %v, want ErrUnsupported", err)
	}
	if _, err := Browse(`\\nas\media`, "user", "pass"); !errors.Is(err, ErrUnsupported) {
		t.Errorf("Browse err = %v, want ErrUnsupported", err)
	}
}
