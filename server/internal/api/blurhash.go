package api

import (
	"errors"
	"image"
	"io/fs"
	"os"
	"strings"

	"github.com/buckket/go-blurhash"
)

/*
 * Blurhash placeholder helper.
 *
 * For every artwork file the resize proxy writes to disk, we compute a
 * blurhash string (~30 chars) and persist it next to the image as
 * `{type}.blurhash`. The frontend fetches the hash inline via the artwork
 * list response and renders a smooth blurred placeholder while the real
 * JPEG downloads — so the user sees colors and shape immediately instead
 * of a flat gradient.
 *
 * Why 4×3 components: this is the value the original blurhash paper
 * recommends for typical posters (taller than wide). It produces a
 * ~28-character string that's small enough to inline in API responses
 * without bloating them.
 *
 * Why lazy compute: blurhash takes 5-15 ms per image on modern CPUs.
 * Doing it eagerly at artwork-download time would block a 60-poster grid
 * by ~600 ms-1 s. Lazy compute happens once per image and is then cached
 * forever — typically by the time the user comes back to the page, every
 * poster they care about has a hash.
 */

const blurhashComponentsX = 4
const blurhashComponentsY = 3

// blurhashSidecarPath returns the on-disk path where the blurhash for an
// image file is cached. Mirrors the sized-variant pattern: sits next to
// the source with a stable suffix.
func blurhashSidecarPath(imagePath string) string {
	return imagePath + ".blurhash"
}

// readBlurhashSidecar returns the cached hash for an image, if one exists.
// Returns empty string + nil error when the sidecar file is absent — this
// is the common cold-state case, not an error.
func readBlurhashSidecar(imagePath string) (string, error) {
	data, err := os.ReadFile(blurhashSidecarPath(imagePath))
	if err != nil {
		if errors.Is(err, fs.ErrNotExist) {
			return "", nil
		}
		return "", err
	}
	hash := strings.TrimSpace(string(data))
	return hash, nil
}

// computeBlurhash decodes the image at imagePath and produces a 4×3
// blurhash. Callers should cache the result via writeBlurhashSidecar to
// avoid recomputing on every request.
func computeBlurhash(imagePath string) (string, error) {
	f, err := os.Open(imagePath)
	if err != nil {
		return "", err
	}
	defer f.Close()
	img, _, err := image.Decode(f)
	if err != nil {
		return "", err
	}
	hash, err := blurhash.Encode(blurhashComponentsX, blurhashComponentsY, img)
	if err != nil {
		return "", err
	}
	return hash, nil
}

// writeBlurhashSidecar stores the hash so future requests for the same
// image can read it without re-encoding. Best-effort: failures are
// logged-and-swallowed by callers because a missing sidecar just means
// "compute again next time" rather than a fatal condition.
func writeBlurhashSidecar(imagePath, hash string) error {
	return os.WriteFile(blurhashSidecarPath(imagePath), []byte(hash), 0o644)
}

// ensureBlurhash returns the hash for imagePath, computing and caching it
// on the first request and reading from the cache thereafter. Returns
// empty string on any failure (the frontend renders its existing palette
// gradient when blurhash is absent — graceful degradation).
func ensureBlurhash(imagePath string) string {
	if hash, err := readBlurhashSidecar(imagePath); err == nil && hash != "" {
		return hash
	}
	hash, err := computeBlurhash(imagePath)
	if err != nil {
		return ""
	}
	_ = writeBlurhashSidecar(imagePath, hash)
	return hash
}
