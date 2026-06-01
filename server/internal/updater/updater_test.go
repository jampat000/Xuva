package updater

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
)

func TestClassifyAsset(t *testing.T) {
	cases := []struct {
		name string
		want string
	}{
		// MSI (Windows Service install) — the SYSTEM updater targets these.
		{"xuva-server-v1.2.3.msi", "windows-msi"},
		{"xuva-server-v1.2.3.msi.sha256", "windows-msi-checksum"},
		{"XUVA-SERVER-V1.2.3.MSI", "windows-msi"},
		// NSIS installer (desktop/tray) — the launcher applies these.
		{"xuva-v1.2.3-win-x64.exe", "windows-installer"},
		{"xuva-v1.2.3-win-x64.exe.sha256", "windows-installer-checksum"},
		// Portable zip.
		{"xuva-v1.2.3-win-x64.zip", "windows-portable"},
		{"xuva-v1.2.3-win-x64.zip.sha256", "windows-portable-checksum"},
		// Unknown.
		{"checksums.txt", "asset"},
		{"", "asset"},
		{"  xuva-server-v1.2.3.msi  ", "windows-msi"},
	}
	for _, tc := range cases {
		if got := ClassifyAsset(tc.name); got != tc.want {
			t.Errorf("ClassifyAsset(%q) = %q, want %q", tc.name, got, tc.want)
		}
	}
}

// The .msi/.exe checksum suffixes must win over the bare .msi/.exe cases, or a
// sidecar would be mistaken for the installer itself and handed to the applier.
func TestClassifyAssetChecksumPrecedence(t *testing.T) {
	if got := ClassifyAsset("foo.msi.sha256"); got == "windows-msi" {
		t.Fatalf("foo.msi.sha256 classified as the installer, not its checksum")
	}
}

func TestCompareVersions(t *testing.T) {
	cases := []struct {
		a, b string
		want int
	}{
		{"v1.2.4", "v1.2.3", 1},
		{"v1.2.3", "v1.2.4", -1},
		{"v1.2.3", "v1.2.3", 0},
		{"v1.3.0", "v1.2.9", 1},
		{"v2.0.0", "v1.9.9", 1},
		// "dev" (the un-ldflag'd build) is unparseable — never an upgrade target.
		{"v1.2.3", "dev", 0},
		{"dev", "v1.2.3", 0},
		{"1.2.4", "1.2.3", 1}, // tolerant of a missing leading "v"
		{"garbage", "v1.0.0", 0},
		// Pre-release tags are unparseable by the strict pattern → never upgrades.
		{"v1.2.4-beta.1", "v1.2.3", 0},
	}
	for _, tc := range cases {
		if got := CompareVersions(tc.a, tc.b); got != tc.want {
			t.Errorf("CompareVersions(%q, %q) = %d, want %d", tc.a, tc.b, got, tc.want)
		}
	}
}

func TestFindAsset(t *testing.T) {
	assets := []Asset{
		{Name: "a.exe", PackageType: "windows-installer"},
		{Name: "a.msi", PackageType: "windows-msi"},
	}
	if got, ok := FindAsset(assets, "windows-msi"); !ok || got.Name != "a.msi" {
		t.Fatalf("FindAsset(windows-msi) = %+v, %v", got, ok)
	}
	if _, ok := FindAsset(assets, "windows-portable"); ok {
		t.Fatalf("FindAsset(windows-portable) unexpectedly found a match")
	}
}

func TestExpectedSHA(t *testing.T) {
	if got := ExpectedSHA(Asset{Digest: "sha256:ABCDEF"}); got != "ABCDEF" {
		t.Errorf("ExpectedSHA stripped prefix wrong: %q", got)
	}
	if got := ExpectedSHA(Asset{Digest: "  deadbeef  "}); got != "deadbeef" {
		t.Errorf("ExpectedSHA trim wrong: %q", got)
	}
	if got := ExpectedSHA(Asset{}); got != "" {
		t.Errorf("ExpectedSHA of empty digest = %q, want empty", got)
	}
}

func TestSafePathName(t *testing.T) {
	cases := map[string]string{
		"v1.2.3":     "v1.2.3",
		"":           "unknown",
		"  ":         "unknown",
		`a/b\c`:      "a-b-c",
		"a/../b":     "a-..-b",
	}
	for in, want := range cases {
		if got := SafePathName(in); got != want {
			t.Errorf("SafePathName(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestFetchLatestParsesAndClassifies(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if got := r.Header.Get("Accept"); got != "application/vnd.github+json" {
			t.Errorf("Accept header = %q", got)
		}
		_ = json.NewEncoder(w).Encode(map[string]any{
			"tag_name":     "v1.2.3",
			"html_url":     "https://example.test/releases/v1.2.3",
			"published_at": "2026-06-01T00:00:00Z",
			"assets": []map[string]any{
				{"name": "xuva-server-v1.2.3.msi", "browser_download_url": "https://example.test/x.msi", "size": 123, "digest": "sha256:abc"},
				{"name": "skip-me", "browser_download_url": ""}, // dropped: empty URL
			},
		})
	}))
	defer srv.Close()
	t.Setenv("XUVA_UPDATE_RELEASE_API", srv.URL)

	rel, err := FetchLatest(context.Background())
	if err != nil {
		t.Fatalf("FetchLatest: %v", err)
	}
	if rel.Version != "v1.2.3" || rel.URL != "https://example.test/releases/v1.2.3" {
		t.Fatalf("unexpected release meta: %+v", rel)
	}
	if len(rel.Assets) != 1 {
		t.Fatalf("expected 1 asset (empty-URL dropped), got %d: %+v", len(rel.Assets), rel.Assets)
	}
	if rel.Assets[0].PackageType != "windows-msi" {
		t.Errorf("asset classified as %q", rel.Assets[0].PackageType)
	}
}

func TestFetchLatestNon2xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "rate limited", http.StatusForbidden)
	}))
	defer srv.Close()
	t.Setenv("XUVA_UPDATE_RELEASE_API", srv.URL)

	if _, err := FetchLatest(context.Background()); err == nil {
		t.Fatal("expected an error on HTTP 403")
	}
}

func TestDownloadAndSHA256(t *testing.T) {
	body := []byte("hello xuva update")
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write(body)
	}))
	defer srv.Close()

	dest := filepath.Join(t.TempDir(), "nested", "asset.bin")
	if err := Download(context.Background(), srv.URL, dest); err != nil {
		t.Fatalf("Download: %v", err)
	}
	got, err := os.ReadFile(dest)
	if err != nil {
		t.Fatalf("read downloaded file: %v", err)
	}
	if string(got) != string(body) {
		t.Fatalf("downloaded content mismatch: %q", got)
	}
	// No leftover .tmp after a successful atomic rename.
	if _, err := os.Stat(dest + ".tmp"); !os.IsNotExist(err) {
		t.Errorf("temp file was not cleaned up")
	}
	// Known SHA-256 of the body.
	sum, err := SHA256File(dest)
	if err != nil {
		t.Fatalf("SHA256File: %v", err)
	}
	if len(sum) != 64 {
		t.Errorf("unexpected sha length: %q", sum)
	}
}

func TestParseSHA256File(t *testing.T) {
	dir := t.TempDir()
	good := filepath.Join(dir, "good.sha256")
	hash := "c9500a9754fc779c157ef9e67be663ebc43018f76bb87c229cfadc7179575b42"
	if err := os.WriteFile(good, []byte(hash+"  xuva-server-v1.2.3.msi\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ParseSHA256File(good); got != hash {
		t.Errorf("ParseSHA256File = %q, want %q", got, hash)
	}
	// Malformed (too short) → "".
	bad := filepath.Join(dir, "bad.sha256")
	if err := os.WriteFile(bad, []byte("deadbeef  x.msi"), 0o644); err != nil {
		t.Fatal(err)
	}
	if got := ParseSHA256File(bad); got != "" {
		t.Errorf("ParseSHA256File(malformed) = %q, want empty", got)
	}
	// Missing file → "".
	if got := ParseSHA256File(filepath.Join(dir, "nope.sha256")); got != "" {
		t.Errorf("ParseSHA256File(missing) = %q, want empty", got)
	}
}
