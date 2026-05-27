package api

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"strings"
)

// writeCachedJSON serializes payload and emits an ETag derived from the
// response bytes. When the request carries a matching `If-None-Match`, returns
// 304 with no body so the browser substitutes its cached copy — the wire cost
// drops to ~150 bytes (one set of headers) and the client-side JSON.parse is
// skipped entirely.
//
// Use this for idempotent GET endpoints whose content is fully determined by
// the response itself (lists, snapshots, summaries). Don't use it for
// endpoints that return per-request data (timestamps, random samples) where
// every response is unique — ETag adds overhead without ever matching.
func writeCachedJSON(w http.ResponseWriter, r *http.Request, status int, payload any) {
	body, err := json.Marshal(payload)
	if err != nil {
		writeError(w, http.StatusInternalServerError, "encode failed")
		return
	}

	// SHA-256 is overkill for cache validation but it's a fixed cost vs
	// payload size and avoids any "did the hash collide" debate. We truncate
	// to 16 bytes (32 hex chars) — still 128 bits of collision resistance.
	sum := sha256.Sum256(body)
	etag := `W/"` + hex.EncodeToString(sum[:16]) + `"`

	// If-None-Match can list multiple ETags separated by commas. The wildcard
	// "*" always matches when the resource exists; treat it that way too.
	if match := strings.TrimSpace(r.Header.Get("If-None-Match")); match != "" {
		if match == "*" || etagMatches(match, etag) {
			w.Header().Set("ETag", etag)
			w.Header().Set("Cache-Control", "private, must-revalidate")
			w.WriteHeader(http.StatusNotModified)
			return
		}
	}

	w.Header().Set("ETag", etag)
	w.Header().Set("Content-Type", "application/json")
	// Force a conditional request every time — the SWR layer in front owns
	// freshness; we just want to skip the body when content hasn't changed.
	w.Header().Set("Cache-Control", "private, must-revalidate")
	w.WriteHeader(status)
	_, _ = w.Write(body)
}

func etagMatches(header, candidate string) bool {
	for _, token := range strings.Split(header, ",") {
		if strings.TrimSpace(token) == candidate {
			return true
		}
	}
	return false
}
