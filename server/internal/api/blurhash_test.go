package api

import (
	"image"
	"image/color"
	"image/jpeg"
	"os"
	"path/filepath"
	"testing"
)

func writeBlurhashTestImage(t *testing.T, dir, name string, width, height int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8((x * 255) / width), G: uint8((y * 255) / height), B: 80, A: 255})
		}
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create %s: %v", path, err)
	}
	defer f.Close()
	if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
		t.Fatalf("encode %s: %v", path, err)
	}
	return path
}

func TestEnsureBlurhash_ComputesAndPersists(t *testing.T) {
	dir := t.TempDir()
	src := writeBlurhashTestImage(t, dir, "poster.jpg", 200, 300)
	hash := ensureBlurhash(src)
	if hash == "" {
		t.Fatal("expected a non-empty blurhash")
	}
	// Sidecar should now exist.
	sidecar := blurhashSidecarPath(src)
	data, err := os.ReadFile(sidecar)
	if err != nil {
		t.Fatalf("sidecar file should exist after compute: %v", err)
	}
	if string(data) != hash {
		t.Errorf("sidecar content %q != computed hash %q", string(data), hash)
	}
}

func TestEnsureBlurhash_ReadsCachedSidecar(t *testing.T) {
	dir := t.TempDir()
	src := writeBlurhashTestImage(t, dir, "poster.jpg", 200, 300)

	// Pre-populate the sidecar with a known marker that isn't a real hash.
	// ensureBlurhash should return whatever's on disk without recomputing,
	// so we'd see this marker back, not a freshly-computed hash.
	const marker = "L00000000000000000000000000"
	if err := writeBlurhashSidecar(src, marker); err != nil {
		t.Fatalf("seed sidecar: %v", err)
	}

	hash := ensureBlurhash(src)
	if hash != marker {
		t.Errorf("expected cached value %q, got recomputed hash %q", marker, hash)
	}
}

func TestEnsureBlurhash_ReturnsEmptyOnMissingFile(t *testing.T) {
	dir := t.TempDir()
	hash := ensureBlurhash(filepath.Join(dir, "does-not-exist.jpg"))
	if hash != "" {
		t.Errorf("missing file should return empty hash, got %q", hash)
	}
}

func TestReadBlurhashSidecar_AbsenceIsNotError(t *testing.T) {
	dir := t.TempDir()
	src := writeBlurhashTestImage(t, dir, "poster.jpg", 100, 150)
	hash, err := readBlurhashSidecar(src)
	if err != nil {
		t.Fatalf("absent sidecar should not error, got %v", err)
	}
	if hash != "" {
		t.Errorf("absent sidecar should return empty hash, got %q", hash)
	}
}

func TestComputeBlurhash_ProducesValidPrefix(t *testing.T) {
	dir := t.TempDir()
	src := writeBlurhashTestImage(t, dir, "poster.jpg", 200, 300)
	hash, err := computeBlurhash(src)
	if err != nil {
		t.Fatalf("computeBlurhash: %v", err)
	}
	if len(hash) < 6 {
		t.Errorf("hash too short: %q", hash)
	}
	// First char encodes components: 4×3 → (3 + 4*9) = 39 → base83 → 'P'.
	// We just sanity-check the hash starts with a letter in the expected
	// neighborhood; exact value depends on go-blurhash version.
	if hash[0] < 'A' || hash[0] > 'z' {
		t.Errorf("unexpected leading char in hash %q", hash)
	}
}
