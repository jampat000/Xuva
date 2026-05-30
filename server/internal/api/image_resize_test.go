package api

import (
	"image"
	"image/color"
	"image/jpeg"
	"image/png"
	"os"
	"path/filepath"
	"testing"
)

func writeTestImage(t *testing.T, dir, name string, width, height int) string {
	t.Helper()
	img := image.NewRGBA(image.Rect(0, 0, width, height))
	for y := 0; y < height; y++ {
		for x := 0; x < width; x++ {
			img.Set(x, y, color.RGBA{R: uint8(x % 255), G: uint8(y % 255), B: 80, A: 255})
		}
	}
	path := filepath.Join(dir, name)
	f, err := os.Create(path)
	if err != nil {
		t.Fatalf("create test image %s: %v", path, err)
	}
	defer f.Close()
	ext := filepath.Ext(name)
	switch ext {
	case ".jpg", ".jpeg":
		if err := jpeg.Encode(f, img, &jpeg.Options{Quality: 90}); err != nil {
			t.Fatalf("encode jpeg %s: %v", path, err)
		}
	case ".png":
		if err := png.Encode(f, img); err != nil {
			t.Fatalf("encode png %s: %v", path, err)
		}
	default:
		t.Fatalf("unsupported test extension %s", ext)
	}
	return path
}

func decodeDimensions(t *testing.T, path string) (int, int) {
	t.Helper()
	f, err := os.Open(path)
	if err != nil {
		t.Fatalf("open %s: %v", path, err)
	}
	defer f.Close()
	cfg, _, err := image.DecodeConfig(f)
	if err != nil {
		t.Fatalf("decode config %s: %v", path, err)
	}
	return cfg.Width, cfg.Height
}

func TestParseResizeWidth(t *testing.T) {
	cases := []struct {
		raw  string
		want int
	}{
		{"", 0},
		{"abc", 0},
		{"0", 0},
		{"-5", 0},
		{"31", 0},          // below minResizeWidth
		{"32", 32},         // exactly min
		{"200", 200},       // typical poster width
		{"1024", 1024},     // mid-range — still valid
		{"2048", 2048},     // exactly max
		{"2049", 0},        // over max
		{"  300  ", 300},   // trimming
	}
	for _, tc := range cases {
		if got := parseResizeWidth(tc.raw); got != tc.want {
			t.Errorf("parseResizeWidth(%q) = %d, want %d", tc.raw, got, tc.want)
		}
	}
}

func TestSizedArtworkPath(t *testing.T) {
	cases := []struct {
		in     string
		width  int
		format string
		want   string
	}{
		// filepath.Join uses OS separators, so we let Join's behavior decide.
		{filepath.Join("a", "b", "poster.jpg"), 200, "jpg", filepath.Join("a", "b", "poster.w200.jpg")},
		{filepath.Join("a", "b", "poster.png"), 400, "jpg", filepath.Join("a", "b", "poster.w400.jpg")},
		{"poster.webp", 100, "jpg", "poster.w100.jpg"},
		// WebP format keeps the source intact, derives a .webp sibling so the
		// JPEG and WebP variants of the same source cache independently.
		{filepath.Join("a", "b", "poster.jpg"), 200, "webp", filepath.Join("a", "b", "poster.w200.webp")},
		{"hero.png", 360, "webp", "hero.w360.webp"},
		// Unknown format defaults to JPEG — defensive against typos / future
		// formats slipping in via query string.
		{"poster.jpg", 200, "avif", "poster.w200.jpg"},
		{"poster.jpg", 200, "", "poster.w200.jpg"},
	}
	for _, tc := range cases {
		if got := sizedArtworkPath(tc.in, tc.width, tc.format); got != tc.want {
			t.Errorf("sizedArtworkPath(%q, %d, %q) = %q, want %q", tc.in, tc.width, tc.format, got, tc.want)
		}
	}
}

func TestResolveServeFormat(t *testing.T) {
	cases := []struct {
		accept string
		want   string
	}{
		{"", "jpg"},
		{"*/*", "jpg"},
		{"image/jpeg", "jpg"},
		{"image/*", "jpg"}, // strict: wildcard image alone isn't WebP-capable
		{"image/webp", "webp"},
		{"image/avif,image/webp,image/apng,image/*,*/*;q=0.8", "webp"},
		{"image/webp;q=0.9", "webp"},
		{"  image/webp  ,  */*  ", "webp"},
		{"image/avif", "jpg"}, // avif support alone shouldn't get WebP
		{"text/html,application/xhtml+xml,image/webp", "webp"},
	}
	for _, tc := range cases {
		if got := resolveServeFormat(tc.accept); got != tc.want {
			t.Errorf("resolveServeFormat(%q) = %q, want %q", tc.accept, got, tc.want)
		}
	}
}

func TestResizeImageWidth_EncodesWebPWhenDestEndsInWebP(t *testing.T) {
	dir := t.TempDir()
	src := writeTestImage(t, dir, "poster.jpg", 800, 1200)
	dst := filepath.Join(dir, "poster.w200.webp")
	if err := resizeImageWidth(src, dst, 200); err != nil {
		t.Fatalf("resizeImageWidth → webp: %v", err)
	}
	// Read first 12 bytes — WebP files start with RIFF....WEBP.
	f, err := os.Open(dst)
	if err != nil {
		t.Fatalf("open webp output: %v", err)
	}
	defer f.Close()
	head := make([]byte, 12)
	if _, err := f.Read(head); err != nil {
		t.Fatalf("read webp head: %v", err)
	}
	if string(head[0:4]) != "RIFF" || string(head[8:12]) != "WEBP" {
		t.Fatalf("output is not WebP — header was %q (expected RIFF..WEBP)", head)
	}
	w, h := decodeDimensions(t, dst)
	if w != 200 {
		t.Errorf("webp output width = %d, want 200", w)
	}
	if h != 300 {
		t.Errorf("webp output height = %d, want 300 (aspect-preserved)", h)
	}
}

func TestResizeImageWidth_Downscales(t *testing.T) {
	dir := t.TempDir()
	src := writeTestImage(t, dir, "poster.jpg", 800, 1200)
	dst := filepath.Join(dir, "poster.w200.jpg")
	if err := resizeImageWidth(src, dst, 200); err != nil {
		t.Fatalf("resizeImageWidth: %v", err)
	}
	w, h := decodeDimensions(t, dst)
	if w != 200 {
		t.Errorf("output width = %d, want 200", w)
	}
	// Aspect ratio preserved: 200 * (1200 / 800) = 300.
	if h != 300 {
		t.Errorf("output height = %d, want 300 (aspect-preserved)", h)
	}
}

func TestResizeImageWidth_DoesNotUpscale(t *testing.T) {
	dir := t.TempDir()
	src := writeTestImage(t, dir, "tiny.jpg", 100, 150)
	dst := filepath.Join(dir, "tiny.w400.jpg")
	if err := resizeImageWidth(src, dst, 400); err != nil {
		t.Fatalf("resizeImageWidth: %v", err)
	}
	w, h := decodeDimensions(t, dst)
	// targetWidth >= srcW path: the source is re-encoded at its own size,
	// not upscaled to 400. JPEG round-tripping may not preserve exact pixel
	// dimensions for tiny images via decode->encode, but the bounds should
	// stay within ±2 of the input.
	if w < 98 || w > 102 {
		t.Errorf("expected output width close to source 100, got %d", w)
	}
	if h < 148 || h > 152 {
		t.Errorf("expected output height close to source 150, got %d", h)
	}
}

func TestResizeImageWidth_AcceptsPNGSource(t *testing.T) {
	dir := t.TempDir()
	src := writeTestImage(t, dir, "logo.png", 600, 200)
	dst := filepath.Join(dir, "logo.w150.jpg")
	if err := resizeImageWidth(src, dst, 150); err != nil {
		t.Fatalf("resizeImageWidth (png source): %v", err)
	}
	w, _ := decodeDimensions(t, dst)
	if w != 150 {
		t.Errorf("png→jpg resize width = %d, want 150", w)
	}
}

func TestResizeImageWidth_RejectsCorruptInput(t *testing.T) {
	dir := t.TempDir()
	src := filepath.Join(dir, "broken.jpg")
	if err := os.WriteFile(src, []byte("not a real image"), 0o644); err != nil {
		t.Fatalf("write broken file: %v", err)
	}
	dst := filepath.Join(dir, "broken.w200.jpg")
	if err := resizeImageWidth(src, dst, 200); err == nil {
		t.Fatal("expected error decoding broken input, got nil")
	}
	if _, err := os.Stat(dst); err == nil {
		t.Error("destination file should not exist after a failed resize")
	}
}
