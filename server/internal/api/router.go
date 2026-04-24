package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/config"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/playback"
	"github.com/vyrdenhq/vyrden/server/internal/probe"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/scans"
	"github.com/vyrdenhq/vyrden/server/internal/sessions"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
)

type Deps struct {
	Config    config.Config
	StartedAt time.Time
	Events    *events.Bus
	Resources *resources.Manager
	Jobs      *jobs.Registry
	Libraries *libraries.Service
	Scanner   *scanner.Service
	Scans     *scans.Service
	Catalog   *catalog.Service
	Media     *media.Service
	Movies    *movies.Service
	TV        *tv.Service
	Probe     *probe.Service
	Playback  *playback.Service
	Sessions  *sessions.Service
}

func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /", devConsoleHandler(deps))
	mux.HandleFunc("GET /api/health", healthHandler(deps))
	mux.HandleFunc("GET /api/events", eventsHandler(deps))
	mux.HandleFunc("GET /api/architecture", architectureHandler(deps))
	mux.HandleFunc("GET /api/libraries", librariesHandler(deps))
	mux.HandleFunc("GET /api/catalog/summary", catalogSummaryHandler(deps))
	mux.HandleFunc("GET /api/movies", moviesHandler(deps))
	mux.HandleFunc("GET /api/movies/{id}", movieDetailHandler(deps))
	mux.HandleFunc("GET /api/series", seriesHandler(deps))
	mux.HandleFunc("GET /api/series/{id}", seriesDetailHandler(deps))
	mux.HandleFunc("GET /api/review", reviewHandler(deps))
	mux.HandleFunc("GET /api/media-sources", mediaSourcesHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}", mediaSourceDetailHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}/stream", mediaSourceStreamHandler(deps))
	mux.HandleFunc("POST /api/media-sources/{id}/probe", mediaSourceProbeHandler(deps))
	mux.HandleFunc("GET /api/scans", scansHandler(deps))
	mux.HandleFunc("GET /api/scans/{id}", scanJobHandler(deps))
	mux.HandleFunc("POST /api/libraries/movies/scan", movieScanHandler(deps))
	mux.HandleFunc("POST /api/libraries/tv/scan", tvScanHandler(deps))
	mux.HandleFunc("POST /api/libraries/scan", allLibrariesScanHandler(deps))
	mux.HandleFunc("GET /api/playback/decision", playbackDecisionHandler(deps))
	return mux
}

func devConsoleHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(devConsoleHTML))
	}
}

func healthHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		startedAt := deps.StartedAt
		if startedAt.IsZero() {
			startedAt = time.Now().UTC()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "vyrden-server",
			"startedAt": startedAt.UTC().Format(time.RFC3339),
			"httpAddr":  deps.Config.HTTPAddr,
			"libraries": map[string]string{
				"movies": deps.Config.MovieLibraryPath,
				"tv":     deps.Config.TVLibraryPath,
			},
		})
	}
}

func architectureHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"transports": []string{"HTTP/JSON", "SSE", "WebSocket later for two-way control"},
			"domains":    []string{"libraries", "scanner", "media", "movies", "tv", "playback", "subtitles", "devices", "sessions", "downloads"},
			"workloads":  deps.Resources.Classes(),
			"queues":     deps.Jobs.Snapshot(),
		})
	}
}

func librariesHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"libraries": deps.Libraries.List(),
		})
	}
}

func catalogSummaryHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		summary, err := deps.Catalog.Summary(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "catalog summary failed")
			return
		}
		writeJSON(w, http.StatusOK, summary)
	}
}

func moviesHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.Catalog.ListMovies(r.Context(), queryInt(r, "limit", 100))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "movie list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"movies": items})
	}
}

func movieDetailHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		detail, ok, err := deps.Catalog.GetMovie(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "movie detail failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "movie not found")
			return
		}
		writeJSON(w, http.StatusOK, detail)
	}
}

func seriesHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.Catalog.ListSeries(r.Context(), queryInt(r, "limit", 100))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "series list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"series": items})
	}
}

func seriesDetailHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		detail, ok, err := deps.Catalog.GetSeries(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "series detail failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "series not found")
			return
		}
		writeJSON(w, http.StatusOK, detail)
	}
}

func reviewHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.Catalog.ReviewItems(r.Context(), queryInt(r, "limit", 100))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "review list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"items": items})
	}
}

func mediaSourcesHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.Catalog.ListMediaSources(r.Context(), queryInt(r, "limit", 100), r.URL.Query().Get("unprobed") == "true")
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"mediaSources": items})
	}
}

func mediaSourceDetailHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, ok, err := deps.Catalog.GetMediaSource(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source detail failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

func mediaSourceProbeHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, ok, err := deps.Catalog.GetMediaSource(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		result, err := deps.Probe.Probe(r.Context(), item.Path)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "ffprobe failed")
			return
		}
		if err := deps.Catalog.SaveProbe(r.Context(), item.ID, catalog.ProbeResult{
			Container:       result.Container,
			DurationSeconds: result.DurationSeconds,
			Bitrate:         result.Bitrate,
			VideoCodec:      result.VideoCodec,
			Width:           result.Width,
			Height:          result.Height,
			AudioStreams:    result.AudioStreams,
			SubtitleStreams: result.SubtitleStreams,
			RawJSON:         result.RawJSON,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "probe persistence failed")
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func mediaSourceStreamHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		item, ok, err := deps.Catalog.GetMediaSource(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		http.ServeFile(w, r, item.Path)
	}
}

func movieScanHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request libraryScanRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		path := firstNonEmpty(request.Path, request.MoviesPath, deps.Config.MovieLibraryPath)
		if path == "" {
			writeError(w, http.StatusBadRequest, "movie library path is required")
			return
		}
		job, err := deps.Scans.Start(r.Context(), scans.Request{Kind: scans.KindMovies, Path: path, SampleLimit: request.SampleLimit})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
	}
}

func tvScanHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request libraryScanRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		path := firstNonEmpty(request.Path, request.TVPath, deps.Config.TVLibraryPath)
		if path == "" {
			writeError(w, http.StatusBadRequest, "tv library path is required")
			return
		}
		job, err := deps.Scans.Start(r.Context(), scans.Request{Kind: scans.KindTV, Path: path, SampleLimit: request.SampleLimit})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
	}
}

func allLibrariesScanHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request libraryScanRequest
		if !decodeJSON(w, r, &request) {
			return
		}

		moviesPath := firstNonEmpty(request.MoviesPath, deps.Config.MovieLibraryPath)
		tvPath := firstNonEmpty(request.TVPath, deps.Config.TVLibraryPath)
		if moviesPath == "" && tvPath == "" {
			writeError(w, http.StatusBadRequest, "at least one movie or tv library path is required")
			return
		}

		job, err := deps.Scans.Start(r.Context(), scans.Request{Kind: scans.KindAll, MoviesPath: moviesPath, TVPath: tvPath, SampleLimit: request.SampleLimit})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
	}
}

func scansHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"scans": deps.Scans.List()})
	}
}

func scanJobHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := deps.Scans.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "scan job not found")
			return
		}
		writeJSON(w, http.StatusOK, job)
	}
}

func playbackDecisionHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := playback.Request{
			MediaSourceID: r.URL.Query().Get("mediaSourceId"),
			ClientProfile: r.URL.Query().Get("clientProfile"),
		}
		decision := deps.Playback.Decide(r.Context(), request)
		if request.MediaSourceID != "" {
			source, ok, err := deps.Catalog.GetMediaSource(r.Context(), request.MediaSourceID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "media source lookup failed")
				return
			}
			if !ok {
				writeError(w, http.StatusNotFound, "media source not found")
				return
			}
			decision = deps.Playback.DecideSource(r.Context(), request, playback.SourceFacts{
				MediaSourceID:   source.ID,
				Container:       source.Container,
				VideoCodec:      source.VideoCodec,
				AudioStreams:    source.AudioStreams,
				SubtitleStreams: source.SubtitleStreams,
				Bitrate:         source.Bitrate,
				Probed:          source.Probed,
			})
		}
		writeJSON(w, http.StatusOK, decision)
	}
}

type libraryScanRequest struct {
	Path        string `json:"path"`
	MoviesPath  string `json:"moviesPath"`
	TVPath      string `json:"tvPath"`
	SampleLimit int    `json:"sampleLimit"`
}

type movieScanResponse struct {
	Kind        scanner.LibraryKind    `json:"kind"`
	Path        string                 `json:"path"`
	Summary     scanner.Summary        `json:"summary"`
	Persisted   catalog.PersistSummary `json:"persisted"`
	MoviesFound int                    `json:"moviesFound"`
	Movies      []movies.Candidate     `json:"movies"`
}

type tvScanResponse struct {
	Kind          scanner.LibraryKind    `json:"kind"`
	Path          string                 `json:"path"`
	Summary       scanner.Summary        `json:"summary"`
	Persisted     catalog.PersistSummary `json:"persisted"`
	EpisodesFound int                    `json:"episodesFound"`
	Episodes      []tv.EpisodeCandidate  `json:"episodes"`
}

type combinedScanResponse struct {
	Movies movieScanResponse `json:"movies,omitempty"`
	TV     tvScanResponse    `json:"tv,omitempty"`
}

func scanMovies(w http.ResponseWriter, r *http.Request, deps Deps, path string, sampleLimit int) (movieScanResponse, bool) {
	deps.Events.Publish("scan.started", map[string]string{"kind": string(scanner.KindMovies), "path": path})
	result, err := deps.Scanner.Scan(r.Context(), scanner.Request{Kind: scanner.KindMovies, Root: path})
	if err != nil {
		deps.Events.Publish("scan.failed", map[string]string{"kind": string(scanner.KindMovies), "path": path, "error": err.Error()})
		writeScanError(w, err)
		return movieScanResponse{}, false
	}
	candidates := deps.Movies.Classify(result.Files)
	library := libraries.Library{
		ID:   "movies",
		Name: "Movies",
		Path: result.Root,
		Kind: libraries.KindMovies,
	}
	deps.Libraries.Set(library)
	persisted, err := deps.Catalog.SaveMovieScan(r.Context(), library, result, candidates)
	if err != nil {
		deps.Events.Publish("scan.failed", map[string]string{"kind": string(scanner.KindMovies), "path": result.Root, "error": err.Error()})
		writeError(w, http.StatusInternalServerError, "movie catalog update failed")
		return movieScanResponse{}, false
	}
	response := movieScanResponse{
		Kind:        scanner.KindMovies,
		Path:        result.Root,
		Summary:     result.Summary,
		Persisted:   persisted,
		MoviesFound: len(candidates),
		Movies:      limitMovies(candidates, sampleLimit),
	}
	deps.Events.Publish("scan.completed", map[string]any{"kind": string(scanner.KindMovies), "path": result.Root, "mediaFiles": result.MediaFiles})
	return response, true
}

func scanTV(w http.ResponseWriter, r *http.Request, deps Deps, path string, sampleLimit int) (tvScanResponse, bool) {
	deps.Events.Publish("scan.started", map[string]string{"kind": string(scanner.KindTV), "path": path})
	result, err := deps.Scanner.Scan(r.Context(), scanner.Request{Kind: scanner.KindTV, Root: path})
	if err != nil {
		deps.Events.Publish("scan.failed", map[string]string{"kind": string(scanner.KindTV), "path": path, "error": err.Error()})
		writeScanError(w, err)
		return tvScanResponse{}, false
	}
	candidates := deps.TV.Classify(result.Files)
	library := libraries.Library{
		ID:   "tv",
		Name: "TV",
		Path: result.Root,
		Kind: libraries.KindTV,
	}
	deps.Libraries.Set(library)
	persisted, err := deps.Catalog.SaveTVScan(r.Context(), library, result, candidates)
	if err != nil {
		deps.Events.Publish("scan.failed", map[string]string{"kind": string(scanner.KindTV), "path": result.Root, "error": err.Error()})
		writeError(w, http.StatusInternalServerError, "tv catalog update failed")
		return tvScanResponse{}, false
	}
	response := tvScanResponse{
		Kind:          scanner.KindTV,
		Path:          result.Root,
		Summary:       result.Summary,
		Persisted:     persisted,
		EpisodesFound: len(candidates),
		Episodes:      limitEpisodes(candidates, sampleLimit),
	}
	deps.Events.Publish("scan.completed", map[string]any{"kind": string(scanner.KindTV), "path": result.Root, "mediaFiles": result.MediaFiles})
	return response, true
}

func decodeJSON(w http.ResponseWriter, r *http.Request, destination any) bool {
	if r.Body == nil || r.Body == http.NoBody {
		return true
	}
	defer r.Body.Close()
	decoder := json.NewDecoder(r.Body)
	if err := decoder.Decode(destination); err != nil {
		if errors.Is(err, io.EOF) {
			return true
		}
		writeError(w, http.StatusBadRequest, "invalid json body")
		return false
	}
	return true
}

func writeScanError(w http.ResponseWriter, err error) {
	if errors.Is(err, scanner.ErrMissingRoot) {
		writeError(w, http.StatusBadRequest, err.Error())
		return
	}
	writeError(w, http.StatusBadRequest, err.Error())
}

func writeError(w http.ResponseWriter, status int, message string) {
	writeJSON(w, status, map[string]string{"error": message})
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value != "" {
			return value
		}
	}
	return ""
}

func queryInt(r *http.Request, key string, fallback int) int {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	var output int
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return fallback
		}
		output = output*10 + int(ch-'0')
	}
	if output == 0 {
		return fallback
	}
	return output
}

func limitMovies(candidates []movies.Candidate, limit int) []movies.Candidate {
	if limit <= 0 {
		limit = 100
	}
	if len(candidates) <= limit {
		return candidates
	}
	return candidates[:limit]
}

func limitEpisodes(candidates []tv.EpisodeCandidate, limit int) []tv.EpisodeCandidate {
	if limit <= 0 {
		limit = 100
	}
	if len(candidates) <= limit {
		return candidates
	}
	return candidates[:limit]
}

func eventsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		flusher, ok := w.(http.Flusher)
		if !ok {
			writeJSON(w, http.StatusInternalServerError, map[string]string{"error": "streaming unsupported"})
			return
		}

		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")

		ch, cancel := deps.Events.Subscribe(r.Context())
		defer cancel()

		fmt.Fprintf(w, "event: ready\ndata: {\"status\":\"connected\"}\n\n")
		flusher.Flush()

		for {
			select {
			case <-r.Context().Done():
				return
			case event, ok := <-ch:
				if !ok {
					return
				}
				payload, err := json.Marshal(event)
				if err != nil {
					continue
				}
				fmt.Fprintf(w, "event: %s\ndata: %s\n\n", event.Type, payload)
				flusher.Flush()
			}
		}
	}
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}

const devConsoleHTML = `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>Vyrden Dev Console</title>
  <style>
    :root {
      color-scheme: dark;
      --bg: #090a0d;
      --panel: #12151b;
      --panel-2: #171b22;
      --text: #f4f1ea;
      --muted: #a8a193;
      --line: #2b313b;
      --accent: #d8f36a;
      --accent-2: #70d6ff;
      --danger: #ff7b7b;
    }
    * { box-sizing: border-box; }
    body {
      margin: 0;
      min-height: 100vh;
      font-family: Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      background:
        radial-gradient(circle at 10% 10%, rgba(216, 243, 106, 0.08), transparent 28rem),
        radial-gradient(circle at 90% 0%, rgba(112, 214, 255, 0.08), transparent 26rem),
        var(--bg);
      color: var(--text);
    }
    main {
      width: min(1180px, calc(100vw - 40px));
      margin: 0 auto;
      padding: 42px 0;
    }
    header {
      display: flex;
      align-items: flex-end;
      justify-content: space-between;
      gap: 20px;
      margin-bottom: 28px;
    }
    h1 {
      margin: 0;
      font-size: clamp(34px, 5vw, 72px);
      line-height: 0.92;
      letter-spacing: 0;
    }
    .eyebrow {
      margin: 0 0 10px;
      color: var(--accent);
      font-size: 13px;
      font-weight: 700;
      text-transform: uppercase;
    }
    .status {
      min-width: 220px;
      padding: 14px 16px;
      border: 1px solid var(--line);
      background: rgba(18, 21, 27, 0.82);
      border-radius: 8px;
      color: var(--muted);
      text-align: right;
    }
    .status strong {
      display: block;
      color: var(--text);
      font-size: 22px;
    }
    section {
      margin-top: 18px;
      padding: 20px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: rgba(18, 21, 27, 0.88);
    }
    h2 {
      margin: 0 0 16px;
      font-size: 15px;
      text-transform: uppercase;
      letter-spacing: 0;
      color: var(--muted);
    }
    .grid {
      display: grid;
      grid-template-columns: repeat(6, minmax(0, 1fr));
      gap: 12px;
    }
    .metric {
      min-height: 96px;
      padding: 14px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel-2);
    }
    .metric span {
      display: block;
      color: var(--muted);
      font-size: 12px;
      text-transform: uppercase;
    }
    .metric strong {
      display: block;
      margin-top: 14px;
      font-size: 30px;
      line-height: 1;
    }
    .actions, .links {
      display: flex;
      flex-wrap: wrap;
      gap: 10px;
    }
    button, a.button {
      display: inline-flex;
      align-items: center;
      justify-content: center;
      min-height: 42px;
      padding: 0 15px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #1a1f27;
      color: var(--text);
      font: inherit;
      font-weight: 700;
      text-decoration: none;
      cursor: pointer;
    }
    button.primary {
      background: var(--accent);
      color: #15170d;
      border-color: transparent;
    }
    button:disabled {
      cursor: progress;
      opacity: 0.55;
    }
    code, pre {
      font-family: "Cascadia Code", "SFMono-Regular", Consolas, monospace;
    }
    pre {
      max-height: 300px;
      overflow: auto;
      padding: 14px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: #080a0e;
      color: #d9dde5;
      white-space: pre-wrap;
    }
    .paths {
      display: grid;
      grid-template-columns: repeat(2, minmax(0, 1fr));
      gap: 12px;
    }
    .path {
      padding: 14px;
      border: 1px solid var(--line);
      border-radius: 8px;
      background: var(--panel-2);
      overflow-wrap: anywhere;
    }
    .path span {
      display: block;
      margin-bottom: 8px;
      color: var(--muted);
      font-size: 12px;
      text-transform: uppercase;
    }
    .error { color: var(--danger); }
    table {
      width: 100%;
      margin-top: 14px;
      border-collapse: collapse;
      table-layout: fixed;
    }
    th, td {
      padding: 10px;
      border-bottom: 1px solid var(--line);
      text-align: left;
      vertical-align: top;
      overflow-wrap: anywhere;
    }
    th {
      color: var(--muted);
      font-size: 12px;
      text-transform: uppercase;
    }
    td a {
      color: var(--accent-2);
      text-decoration: none;
      font-weight: 700;
    }
    @media (max-width: 860px) {
      header { align-items: stretch; flex-direction: column; }
      .status { text-align: left; }
      .grid { grid-template-columns: repeat(2, minmax(0, 1fr)); }
      .paths { grid-template-columns: 1fr; }
      main { width: min(100vw - 24px, 1180px); padding: 24px 0; }
    }
  </style>
</head>
<body>
  <main>
    <header>
      <div>
        <p class="eyebrow">Local dev build</p>
        <h1>Vyrden</h1>
      </div>
      <div class="status">
        Server
        <strong id="serverStatus">Checking</strong>
      </div>
    </header>

    <section>
      <h2>Catalog</h2>
      <div class="grid">
        <div class="metric"><span>Libraries</span><strong id="libraries">0</strong></div>
        <div class="metric"><span>Media</span><strong id="mediaSources">0</strong></div>
        <div class="metric"><span>Movies</span><strong id="movies">0</strong></div>
        <div class="metric"><span>Series</span><strong id="series">0</strong></div>
        <div class="metric"><span>Episodes</span><strong id="episodes">0</strong></div>
        <div class="metric"><span>Scans</span><strong id="scanRuns">0</strong></div>
      </div>
    </section>

    <section>
      <h2>Libraries</h2>
      <div class="paths">
        <div class="path"><span>Movies</span><code id="moviesPath">Loading</code></div>
        <div class="path"><span>TV</span><code id="tvPath">Loading</code></div>
      </div>
    </section>

    <section>
      <h2>Actions</h2>
      <div class="actions">
        <button class="primary" id="scanAll">Scan Movies + TV</button>
        <button id="scanMovies">Scan Movies</button>
        <button id="scanTV">Scan TV</button>
        <button id="refresh">Refresh</button>
      </div>
    </section>

    <section>
      <h2>Browse</h2>
      <div class="actions">
        <button id="loadMovies">Movies</button>
        <button id="loadSeries">Series</button>
        <button id="loadReview">Needs Review</button>
        <button id="loadMedia">Media Sources</button>
      </div>
      <div id="browse"></div>
    </section>

    <section>
      <h2>API</h2>
      <div class="links">
        <a class="button" href="/api/health">Health</a>
        <a class="button" href="/api/libraries">Libraries</a>
        <a class="button" href="/api/movies">Movies</a>
        <a class="button" href="/api/series">Series</a>
        <a class="button" href="/api/review">Needs Review</a>
        <a class="button" href="/api/scans">Scans</a>
        <a class="button" href="/api/catalog/summary">Catalog Summary</a>
        <a class="button" href="/api/architecture">Architecture</a>
      </div>
    </section>

    <section>
      <h2>Output</h2>
      <pre id="output">Ready.</pre>
    </section>
  </main>

  <script>
    const output = document.getElementById("output");
    const buttons = [...document.querySelectorAll("button")];
    let activeScanId = "";

    function setBusy(busy) {
      for (const button of buttons) button.disabled = busy;
    }

    function show(value) {
      output.textContent = typeof value === "string" ? value : JSON.stringify(value, null, 2);
    }

    async function getJSON(path) {
      const response = await fetch(path);
      if (!response.ok) throw new Error(path + " returned " + response.status);
      return response.json();
    }

    async function refresh() {
      const [health, summary, scans] = await Promise.all([
        getJSON("/api/health"),
        getJSON("/api/catalog/summary"),
        getJSON("/api/scans")
      ]);
      document.getElementById("serverStatus").textContent = health.status || "unknown";
      document.getElementById("moviesPath").textContent = health.libraries?.movies || "Not configured";
      document.getElementById("tvPath").textContent = health.libraries?.tv || "Not configured";
      for (const key of ["libraries", "mediaSources", "movies", "series", "episodes", "scanRuns"]) {
        document.getElementById(key).textContent = summary[key] ?? 0;
      }
      if (!activeScanId && scans.scans?.length) {
        const latest = scans.scans[0];
        show({
          latestScan: latest.id,
          status: latest.status,
          mediaFiles: latest.mediaFiles,
          lastPath: latest.lastPath,
          result: latest.result
        });
      }
    }

    async function post(path) {
      setBusy(true);
      show("Starting " + path + " ...");
      try {
        const response = await fetch(path, {
          method: "POST",
          headers: { "Content-Type": "application/json" },
          body: JSON.stringify({ sampleLimit: 50 })
        });
        const payload = await response.json();
        if (!response.ok) throw new Error(payload.error || response.statusText);
        activeScanId = payload.id;
        show({
          scan: payload.id,
          status: payload.status,
          note: "Scan is running in the background. Progress will update here."
        });
      } catch (error) {
        output.innerHTML = "<span class=\"error\">" + error.message + "</span>";
        setBusy(false);
      }
    }

    async function loadScan(id) {
      const job = await getJSON("/api/scans/" + encodeURIComponent(id));
      show({
        scan: job.id,
        status: job.status,
        totalFiles: job.totalFiles,
        mediaFiles: job.mediaFiles,
        ignoredFiles: job.ignoredFiles,
        lastPath: job.lastPath,
        error: job.error,
        result: job.result
      });
      if (job.status === "completed" || job.status === "failed") {
        activeScanId = "";
        setBusy(false);
        await refresh();
      }
    }

    const events = new EventSource("/api/events");
    for (const eventName of ["scan.queued", "scan.started", "scan.progress", "scan.completed", "scan.failed"]) {
      events.addEventListener(eventName, async event => {
        const message = JSON.parse(event.data);
        const job = message.data || message;
        if (!activeScanId || job.id === activeScanId) {
          activeScanId = job.id;
          show({
            scan: job.id,
            event: eventName,
            status: job.status,
            totalFiles: job.totalFiles,
            mediaFiles: job.mediaFiles,
            ignoredFiles: job.ignoredFiles,
            lastPath: job.lastPath,
            error: job.error,
            result: job.result
          });
          if (eventName === "scan.completed" || eventName === "scan.failed") {
            await loadScan(job.id);
          }
        }
      });
    }

    document.getElementById("scanAll").addEventListener("click", () => post("/api/libraries/scan"));
    document.getElementById("scanMovies").addEventListener("click", () => post("/api/libraries/movies/scan"));
    document.getElementById("scanTV").addEventListener("click", () => post("/api/libraries/tv/scan"));
    document.getElementById("refresh").addEventListener("click", () => refresh().catch(error => show(error.message)));
    document.getElementById("loadMovies").addEventListener("click", () => loadBrowse("movies"));
    document.getElementById("loadSeries").addEventListener("click", () => loadBrowse("series"));
    document.getElementById("loadReview").addEventListener("click", () => loadBrowse("review"));
    document.getElementById("loadMedia").addEventListener("click", () => loadBrowse("media"));

    async function loadBrowse(kind) {
      const browse = document.getElementById("browse");
      browse.innerHTML = "<pre>Loading " + kind + "...</pre>";
      if (kind === "movies") {
        const payload = await getJSON("/api/movies?limit=50");
        browse.innerHTML = table(["Title", "Year", "Versions", "Review"], payload.movies.map(item => [
          "<a href=\"/api/movies/" + item.id + "\">" + escapeHTML(item.title) + "</a>",
          item.year || "",
          item.versionCount || 0,
          item.needsReview ? "Yes" : "No"
        ]));
      }
      if (kind === "series") {
        const payload = await getJSON("/api/series?limit=50");
        browse.innerHTML = table(["Series", "Seasons", "Episodes"], payload.series.map(item => [
          "<a href=\"/api/series/" + item.id + "\">" + escapeHTML(item.title) + "</a>",
          item.seasonCount || 0,
          item.episodeCount || 0
        ]));
      }
      if (kind === "review") {
        const payload = await getJSON("/api/review?limit=50");
        browse.innerHTML = table(["Kind", "Title", "Reason"], payload.items.map(item => [
          escapeHTML(item.kind),
          escapeHTML(item.title),
          escapeHTML(item.reviewReason)
        ]));
      }
      if (kind === "media") {
        const payload = await getJSON("/api/media-sources?limit=50");
        browse.innerHTML = table(["Name", "Kind", "Probed", "Codec", "Stream"], payload.mediaSources.map(item => [
          "<a href=\"/api/media-sources/" + item.id + "\">" + escapeHTML(item.name) + "</a>",
          escapeHTML(item.kind),
          item.probed ? "Yes" : "No",
          escapeHTML(item.videoCodec || ""),
          "<a href=\"/api/media-sources/" + item.id + "/stream\">Open</a>"
        ]));
      }
    }

    function table(headers, rows) {
      if (!rows.length) return "<pre>No items yet. Run a scan first.</pre>";
      return "<table><thead><tr>" + headers.map(header => "<th>" + escapeHTML(header) + "</th>").join("") + "</tr></thead><tbody>" +
        rows.map(row => "<tr>" + row.map(cell => "<td>" + cell + "</td>").join("") + "</tr>").join("") +
        "</tbody></table>";
    }

    function escapeHTML(value) {
      return String(value ?? "").replace(/[&<>"']/g, char => ({
        "&": "&amp;",
        "<": "&lt;",
        ">": "&gt;",
        "\"": "&quot;",
        "'": "&#039;"
      }[char]));
    }

    refresh().catch(error => {
      document.getElementById("serverStatus").textContent = "Error";
      show(error.message);
    });
  </script>
</body>
</html>`
