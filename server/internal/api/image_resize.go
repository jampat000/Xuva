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

	"golang.org/x/image/draw"
	_ "golang.org/x/image/webp" // register decoder for cached WebP sources
)

// Hard cap on the requested width — generous for 2x retina posters but stops
// the proxy from being used as a generic image-resize service that could chew
// CPU. Posters never need more than ~600px wide rendered at 2x density.
const maxResizeWidth = 1024
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
// by the target width so multiple densities (e.g. 1x and 2x) cache
// independently. Output is always JPEG regardless of the source format.
//
//	poster.jpg → poster.w200.jpg
//	poster.png → poster.w200.jpg
func sizedArtworkPath(originalPath string, width int) string {
	dir := filepath.Dir(originalPath)
	file := filepath.Base(originalPath)
	ext := filepath.Ext(file)
	base := strings.TrimSuffix(file, ext)
	return filepath.Join(dir, base+".w"+strconv.Itoa(width)+".jpg")
}

// resizeImageWidth opens srcPath, scales the image down to targetWidth (height
// scales proportionally), and writes the result to destPath as a JPEG. When
// the source is already narrower than targetWidth, the function re-encodes the
// source as-is without upscaling — upscaled posters look worse than the
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

	if targetWidth >= srcW {
		// Already small enough — just re-encode so the proxy can still apply
		// immutable cache headers on the served file. No bytes saved here, but
		// the browser stops revalidating each visit.
		return writeJPEG(img, destPath)
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
	return writeJPEG(dst, destPath)
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
