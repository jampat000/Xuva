package api

import (
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWriteCachedJSON_EmitsETagAnd200(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
	writeCachedJSON(rr, req, http.StatusOK, map[string]any{"movies": []string{"a", "b"}})

	if rr.Code != http.StatusOK {
		t.Fatalf("status = %d, want 200", rr.Code)
	}
	if got := rr.Header().Get("Content-Type"); got != "application/json" {
		t.Errorf("Content-Type = %q, want application/json", got)
	}
	etag := rr.Header().Get("ETag")
	if etag == "" || !strings.HasPrefix(etag, `W/"`) || !strings.HasSuffix(etag, `"`) {
		t.Errorf("ETag = %q, want a weak-quoted value", etag)
	}
	// Body should be valid JSON containing our payload.
	var decoded map[string]any
	if err := json.Unmarshal(rr.Body.Bytes(), &decoded); err != nil {
		t.Fatalf("decode body: %v", err)
	}
	if len(decoded["movies"].([]any)) != 2 {
		t.Errorf("movies length = %d, want 2", len(decoded["movies"].([]any)))
	}
}

func TestWriteCachedJSON_DeterministicETagAcrossCalls(t *testing.T) {
	// Same payload → same ETag. Repeated calls must not produce drift.
	payload := map[string]any{"movies": []map[string]any{
		{"id": "a", "title": "A"},
		{"id": "b", "title": "B"},
	}}
	rrA := httptest.NewRecorder()
	rrB := httptest.NewRecorder()
	writeCachedJSON(rrA, httptest.NewRequest(http.MethodGet, "/api/movies", nil), http.StatusOK, payload)
	writeCachedJSON(rrB, httptest.NewRequest(http.MethodGet, "/api/movies", nil), http.StatusOK, payload)
	if a, b := rrA.Header().Get("ETag"), rrB.Header().Get("ETag"); a != b {
		t.Errorf("ETag drift: %q vs %q", a, b)
	}
}

func TestWriteCachedJSON_DifferentPayloadsDifferentETags(t *testing.T) {
	rrA := httptest.NewRecorder()
	rrB := httptest.NewRecorder()
	writeCachedJSON(rrA, httptest.NewRequest(http.MethodGet, "/api/movies", nil), http.StatusOK, map[string]any{"movies": []string{"a"}})
	writeCachedJSON(rrB, httptest.NewRequest(http.MethodGet, "/api/movies", nil), http.StatusOK, map[string]any{"movies": []string{"b"}})
	if a, b := rrA.Header().Get("ETag"), rrB.Header().Get("ETag"); a == b {
		t.Errorf("ETags collide for different payloads: %q", a)
	}
}

func TestWriteCachedJSON_Returns304OnMatch(t *testing.T) {
	payload := map[string]any{"movies": []string{"a", "b"}}

	// First request: capture the ETag.
	rrA := httptest.NewRecorder()
	writeCachedJSON(rrA, httptest.NewRequest(http.MethodGet, "/api/movies", nil), http.StatusOK, payload)
	etag := rrA.Header().Get("ETag")
	if etag == "" {
		t.Fatal("missing ETag on initial response")
	}

	// Second request: send If-None-Match → expect 304 + empty body.
	rrB := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/movies", nil)
	req.Header.Set("If-None-Match", etag)
	writeCachedJSON(rrB, req, http.StatusOK, payload)

	if rrB.Code != http.StatusNotModified {
		t.Errorf("status = %d, want 304", rrB.Code)
	}
	if got := rrB.Header().Get("ETag"); got != etag {
		t.Errorf("304 response missing matching ETag: got %q want %q", got, etag)
	}
	body, _ := io.ReadAll(rrB.Body)
	if len(body) != 0 {
		t.Errorf("304 body should be empty, got %q", body)
	}
}

func TestWriteCachedJSON_MultipleETagsInMatch(t *testing.T) {
	payload := map[string]any{"x": 1}
	rr := httptest.NewRecorder()
	writeCachedJSON(rr, httptest.NewRequest(http.MethodGet, "/x", nil), http.StatusOK, payload)
	etag := rr.Header().Get("ETag")

	// Client sends multiple candidates, including ours: still a 304.
	rr2 := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("If-None-Match", `W/"deadbeef", `+etag+`, W/"00000000"`)
	writeCachedJSON(rr2, req, http.StatusOK, payload)
	if rr2.Code != http.StatusNotModified {
		t.Errorf("status = %d, want 304 when our ETag is in the list", rr2.Code)
	}
}

func TestWriteCachedJSON_WildcardMatchReturns304(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("If-None-Match", "*")
	writeCachedJSON(rr, req, http.StatusOK, map[string]any{"x": 1})
	if rr.Code != http.StatusNotModified {
		t.Errorf("status = %d, want 304 for wildcard match", rr.Code)
	}
}

func TestWriteCachedJSON_NoMatchReturns200(t *testing.T) {
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/x", nil)
	req.Header.Set("If-None-Match", `W/"not-our-etag"`)
	writeCachedJSON(rr, req, http.StatusOK, map[string]any{"x": 1})
	if rr.Code != http.StatusOK {
		t.Errorf("status = %d, want 200 when ETag doesn't match", rr.Code)
	}
	if rr.Body.Len() == 0 {
		t.Error("body should be present on 200")
	}
}
