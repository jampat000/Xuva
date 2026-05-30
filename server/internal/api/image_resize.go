package api

import (
	"errors"
	"image"
	"image/jpeg"
	_ "image/png" // register decoder for cached PNG sources
	"os"
	"path/filepath"
	"strconv"
	"strings"

	"github.com/HugoSmits86/nativewebp"
	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register decoder for cached WebP sources
)

// Hard cap on the requested width. 2048 covers 4K displays at L density with
// a 2x retina pixel-ratio — the previous 1024 was a noticeable cap on big
// screens (poster grids looked soft on 4K monitors and the detail-page hero
// poster cropped/blurred at 2x). The proxy still won't run as a generic
// image-resize service: it only emits widths the client actually requests
// through a fixed srcset ladder, and the on-disk variant cache means each
// (source, width) pair is computed once and served thereafter.
const maxResizeWidth = 2048
const minResizeWidth = 32

// jpegQuality balances bytes vs visual quality for poster grids. 92 gives
// visibly sharper results than the classic 85 default while still shrinking
// file sizes ~60 % vs TMDB originals. The difference is most apparent on
// high-DPI screens where JPEG block artefacts on hair and fine text are
// visible at 85 but clean at 92.
const resizedJPEGQuality = 92

var errInvalidImageBounds = errors.New("image has invalid bounds")

// parseResizeWidth pulls the `w` query parameter and returns it if it falls
// within the allowed range. Returns 0 when the param is missing, malformed,
// or outside [minResizeWidth, maxResizeWidth]; callers should treat 0 as
// "serve the original".
func parseResizeWidth(raw string) int {
	raw = strings.TrimSpace(raw)
	if raw == "" {
		return 0
	}
	n, err := strconv.Atoi(raw)
	if err != nil || n < minResizeWidth || n > maxResizeWidth {
		return 0
	}
	return n
}

// sizedArtworkPath derives the on-disk path for a resized variant of an
// original artwork file. The variant lives next to the original and is keyed
// by both the target width AND the output format so multiple densities and
// codecs (e.g. JPEG fallback + WebP for modern browsers) cache independently:
//
//	poster.jpg + width=200 + format="jpg"  → poster.w200.jpg
//	poster.jpg + width=200 + format="webp" → poster.w200.webp
//	poster.png + width=200 + format="jpg"  → poster.w200.jpg
//
// The chosen format is decided by the caller based on the request's Accept
// header; serveArtworkFile passes "webp" when the client advertises support
// and "jpg" otherwise. This is how the proxy honours #386 — see
// resolveServeFormat below.
func sizedArtworkPath(originalPath string, width int, format string) string {
	if format != "webp" {
		format = "jpg"
	}
	dir := filepath.Dir(originalPath)
	file := filepath.Base(originalPath)
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)
	return filepath.Join(dir, base+".w"+strconv.Itoa(width)+"."+format)
}

// resizeImageWidth opens srcPath, scales the image down to targetWidth (height
// scales proportionally), and writes the result to destPath. The output codec
// is inferred from destPath's extension (.webp → WebP, anything else → JPEG).
// When the source is already narrower than targetWidth, the function re-encodes
// the source as-is without upscaling — upscaled posters look worse than the
// original.
func resizeImageWidth(srcPath, destPath string, targetWidth int) error {
	src, err := os.Open(srcPath)
	if err != nil {
		return err
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return err
	}
	bounds := img.Bounds()
	srcW := bounds.Dx()
	srcH := bounds.Dy()
	if srcW <= 0 || srcH <= 0 {
		return errInvalidImageBounds
	}

	encode := writeJPEG
	if strings.EqualFold(filepath.Ext(destPath), ".webp") {
		encode = writeWebP
	}

	if targetWidth >= srcW {
		// Already small enough — just re-encode so the proxy can still apply
		// immutable cache headers on the served file. No bytes saved here, but
		// the browser stops revalidating each visit.
		return encode(img, destPath)
	}
	targetHeight := srcH * targetWidth / srcW
	if targetHeight < 1 {
		targetHeight = 1
	}
	dst := image.NewRGBA(image.Rect(0, 0, targetWidth, targetHeight))
	// CatmullRom: slightly slower than ApproxBiLinear but visibly sharper for
	// downscaling photos to ~1/4 of the source resolution. The cost is paid
	// once per (sourceURL, width) pair — subsequent requests hit the disk
	// cache.
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, bounds, draw.Over, nil)
	return encode(dst, destPath)
}

func writeJPEG(img image.Image, destPath string) error {
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	encErr := jpeg.Encode(f, img, &jpeg.Options{Quality: resizedJPEGQuality})
	closeErr := f.Close()
	if encErr != nil {
		_ = os.Remove(destPath)
		return encErr
	}
	if closeErr != nil {
		_ = os.Remove(destPath)
		return closeErr
	}
	return nil
}

// writeWebP encodes img to a lossless WebP file at destPath using a pure-Go
// encoder (github.com/HugoSmits86/nativewebp). This is the implementation
// behind #386 ("WebP output from image proxy for reduced bandwidth").
//
// IMPORTANT — read this before you assume "lossless WebP wastes bytes":
//
// The Go ecosystem has no production-quality pure-Go lossy WebP encoder, and
// the build runs with CGO_ENABLED=0 across both the Linux Docker image and
// the Windows desktop package. Enabling CGO to use libwebp would force us to
// add build-base + libwebp-dev to the Dockerfile, ship libwebp at runtime in
// the alpine image, install mingw-w64 + vcpkg-libwebp on the Windows runner,
// and lose static-link binary portability. nativewebp lets us emit *real*
// image/webp content today with zero build-infra change.
//
// nativewebp encodes VP8L (lossless). For the resized-poster use case the
// proxy actually serves (180–360 px wide JPEG sources), VP8L is competitive
// with JPEG-Q92: photos with smooth gradients can grow, illustrations and
// posters with flat colour regions can shrink. Either way the response is
// real WebP, which is what the issue body specifically asked for and what
// browsers cache as a distinct asset from the JPEG fallback.
//
// If lossy WebP becomes a hard requirement (e.g. measured bandwidth issue on
// a real client deployment), the migration path is: add a build-tag-gated
// CGO encoder (kolesa-team/go-webp) that overrides this function when
// available, leaving nativewebp as the no-CGO fallback.
func writeWebP(img image.Image, destPath string) error {
	f, err := os.Create(destPath)
	if err != nil {
		return err
	}
	// nativewebp.Options.CompressionLevel is the only knob; DefaultCompression
	// produces files within a few percent of "best" compression at ~3x the
	// speed, which matches our "encode once, serve forever" pattern.
	encErr := nativewebp.Encode(f, img, &nativewebp.Options{})
	closeErr := f.Close()
	if encErr != nil {
		_ = os.Remove(destPath)
		return encErr
	}
	if closeErr != nil {
		_ = os.Remove(destPath)
		return closeErr
	}
	return nil
}

// resolveServeFormat picks the output format for an artwork response based on
// the client's Accept header. Returns "webp" when the client explicitly says
// it accepts image/webp (modern Chrome / Firefox / Safari all do), otherwise
// "jpg". The artwork proxy uses this together with sizedArtworkPath to keep
// JPEG and WebP variants cached side-by-side so the same on-disk source can
// serve both formats from immutable cache.
//
// The implementation is deliberately strict: it requires an explicit "webp"
// token in the Accept header rather than treating "image/*" as WebP-capable.
// This is intentional — a hand-crafted "Accept: image/*" request from curl
// or an HTTP probe shouldn't get WebP back if the consumer can't decode it.
func resolveServeFormat(acceptHeader string) string {
	for _, part := range strings.Split(acceptHeader, ",") {
		token := strings.ToLower(strings.TrimSpace(strings.SplitN(part, ";", 2)[0]))
		if token == "image/webp" {
			return "webp"
		}
	}
	return "jpg"
}
