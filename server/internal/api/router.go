package api

import (
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"io"
	"net"
	"net/http"
	"path/filepath"
	"strings"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/config"
	"github.com/vyrdenhq/vyrden/server/internal/devices"
	"github.com/vyrdenhq/vyrden/server/internal/downloads"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/playback"
	"github.com/vyrdenhq/vyrden/server/internal/playstate"
	"github.com/vyrdenhq/vyrden/server/internal/probe"
	"github.com/vyrdenhq/vyrden/server/internal/probes"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/scans"
	"github.com/vyrdenhq/vyrden/server/internal/sessions"
	"github.com/vyrdenhq/vyrden/server/internal/subtitles"
	"github.com/vyrdenhq/vyrden/server/internal/systemstats"
	"github.com/vyrdenhq/vyrden/server/internal/transcode"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
	"github.com/vyrdenhq/vyrden/server/internal/webapp"
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
	Probes    *probes.Service
	Playback  *playback.Service
	PlayState *playstate.Service
	Transcode *transcode.Service
	Downloads *downloads.Service
	Devices   *devices.Service
	Sessions  *sessions.Service
}

func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandler(deps))
	mux.HandleFunc("GET /api/events", eventsHandler(deps))
	mux.HandleFunc("GET /api/architecture", architectureHandler(deps))
	mux.HandleFunc("GET /api/libraries", librariesHandler(deps))
	mux.HandleFunc("POST /api/libraries", librarySaveHandler(deps))
	mux.HandleFunc("DELETE /api/libraries/{id}", libraryDeleteHandler(deps))
	mux.HandleFunc("POST /api/libraries/{id}/scan", libraryScanByIDHandler(deps))
	mux.HandleFunc("GET /api/catalog/summary", catalogSummaryHandler(deps))
	mux.HandleFunc("GET /api/catalog/health", catalogHealthHandler(deps))
	mux.HandleFunc("GET /api/movies", moviesHandler(deps))
	mux.HandleFunc("GET /api/movies/{id}", movieDetailHandler(deps))
	mux.HandleFunc("GET /api/series", seriesHandler(deps))
	mux.HandleFunc("GET /api/series/{id}", seriesDetailHandler(deps))
	mux.HandleFunc("GET /api/review", reviewHandler(deps))
	mux.HandleFunc("GET /api/metadata/suggestions", metadataSuggestionsHandler(deps))
	mux.HandleFunc("GET /api/metadata/{kind}/{id}", metadataRecordsHandler(deps))
	mux.HandleFunc("PUT /api/metadata/match", metadataMatchHandler(deps))
	mux.HandleFunc("GET /api/artwork/{kind}/{id}", artworkHandler(deps))
	mux.HandleFunc("GET /api/versions", versionsHandler(deps))
	mux.HandleFunc("GET /api/settings/performance", performanceSettingsHandler(deps))
	mux.HandleFunc("GET /api/settings", settingsHandler(deps))
	mux.HandleFunc("PUT /api/settings", settingsUpdateHandler(deps))
	mux.HandleFunc("GET /api/system/status", systemStatusHandler(deps))
	mux.HandleFunc("GET /api/remote/access", remoteAccessHandler(deps))
	mux.HandleFunc("POST /api/remote/wan", wanAddressHandler(deps))
	mux.HandleFunc("GET /api/media-sources", mediaSourcesHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}", mediaSourceDetailHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}/stream", mediaSourceStreamHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}/tracks", mediaSourceTracksHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}/subtitles", mediaSourceSubtitlesHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}/subtitles/{index}", mediaSourceSubtitleStreamHandler(deps))
	mux.HandleFunc("POST /api/media-sources/{id}/probe", mediaSourceProbeHandler(deps))
	mux.HandleFunc("GET /api/probes", probesHandler(deps))
	mux.HandleFunc("GET /api/probes/{id}", probeJobHandler(deps))
	mux.HandleFunc("POST /api/probes", probeStartHandler(deps))
	mux.HandleFunc("GET /api/work", workHandler(deps))
	mux.HandleFunc("POST /api/work", workStartHandler(deps))
	mux.HandleFunc("GET /api/downloads", downloadsHandler(deps))
	mux.HandleFunc("POST /api/downloads", downloadStartHandler(deps))
	mux.HandleFunc("GET /api/downloads/{id}", downloadJobHandler(deps))
	mux.HandleFunc("GET /api/downloads/{id}/file", downloadFileHandler(deps))
	mux.HandleFunc("GET /api/devices/profiles", deviceProfilesHandler(deps))
	mux.HandleFunc("GET /api/sessions", sessionsHandler(deps))
	mux.HandleFunc("POST /api/sessions", sessionStartHandler(deps))
	mux.HandleFunc("PATCH /api/sessions/{id}", sessionUpdateHandler(deps))
	mux.HandleFunc("DELETE /api/sessions/{id}", sessionStopHandler(deps))
	mux.HandleFunc("GET /api/playback/recent", playbackRecentHandler(deps))
	mux.HandleFunc("GET /api/playback/state/{id}", playbackStateGetHandler(deps))
	mux.HandleFunc("PUT /api/playback/state/{id}", playbackStateSetHandler(deps))
	mux.HandleFunc("GET /api/scans", scansHandler(deps))
	mux.HandleFunc("GET /api/scans/{id}", scanJobHandler(deps))
	mux.HandleFunc("POST /api/libraries/movies/scan", movieScanHandler(deps))
	mux.HandleFunc("POST /api/libraries/tv/scan", tvScanHandler(deps))
	mux.HandleFunc("POST /api/libraries/scan", allLibrariesScanHandler(deps))
	mux.HandleFunc("GET /api/playback/decision", playbackDecisionHandler(deps))
	mux.HandleFunc("GET /play/{id}", playerHandler(deps))
	mux.Handle("GET /", webapp.Handler())
	return mux
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

func librarySaveHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request libraries.Library
		if !decodeJSON(w, r, &request) {
			return
		}
		library, err := deps.Catalog.SaveLibrary(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		deps.Libraries.Set(library)
		deps.Events.Publish("library.updated", library)
		writeJSON(w, http.StatusOK, library)
	}
}

func libraryDeleteHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		if err := deps.Catalog.DeleteLibrary(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "library delete failed")
			return
		}
		deps.Libraries.Delete(id)
		deps.Events.Publish("library.deleted", map[string]string{"id": id})
		writeJSON(w, http.StatusOK, map[string]string{"id": id})
	}
}

func libraryScanByIDHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		library, ok := deps.Libraries.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "library not found")
			return
		}
		kind := scans.KindMovies
		if library.Kind == libraries.KindTV {
			kind = scans.KindTV
		}
		job, err := deps.Scans.Start(r.Context(), scans.Request{Kind: kind, Path: library.Path})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
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

func catalogHealthHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		health, err := deps.Catalog.Health(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "catalog health failed")
			return
		}
		writeJSON(w, http.StatusOK, health)
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

func metadataSuggestionsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.Catalog.ReviewItems(r.Context(), queryInt(r, "limit", 100))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "metadata suggestions failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"suggestions": items,
			"providers":   metadataProviders(),
			"strategy":    "filename and manual records are local-first; online providers layer on top only when configured",
		})
	}
}

func metadataRecordsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := r.PathValue("kind")
		id := r.PathValue("id")
		records, err := deps.Catalog.ListMetadataRecords(r.Context(), kind, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "metadata records failed")
			return
		}
		best, ok, err := deps.Catalog.GetBestMetadata(r.Context(), kind, id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "metadata records failed")
			return
		}
		var selected any
		if ok {
			selected = best
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"best":      selected,
			"records":   records,
			"providers": metadataProviders(),
		})
	}
}

func metadataMatchHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request catalog.MetadataUpdate
		if !decodeJSON(w, r, &request) {
			return
		}
		if err := deps.Catalog.ApplyMetadata(r.Context(), request); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		records, err := deps.Catalog.ListMetadataRecords(r.Context(), request.Kind, request.ID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "metadata records failed")
			return
		}
		deps.Events.Publish("metadata.updated", request)
		writeJSON(w, http.StatusOK, map[string]any{
			"match":   request,
			"records": records,
		})
	}
}

func metadataProviders() []map[string]any {
	return []map[string]any{
		{"id": "filename", "name": "Filename and folders", "status": "active", "local": true},
		{"id": "manual", "name": "Manual correction", "status": "active", "local": true},
		{"id": "nfo", "name": "Local NFO", "status": "planned", "local": true},
		{"id": "tmdb", "name": "TMDB", "status": "planned-configurable", "local": false},
		{"id": "tvdb", "name": "TVDB", "status": "planned-configurable", "local": false},
		{"id": "omdb", "name": "OMDb", "status": "planned-configurable", "local": false},
	}
}

func artworkHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := r.PathValue("kind")
		id := r.PathValue("id")
		title := id
		switch kind {
		case "movie":
			if item, ok, err := deps.Catalog.GetMovie(r.Context(), id); err == nil && ok {
				title = item.Title
			}
		case "series":
			if item, ok, err := deps.Catalog.GetSeries(r.Context(), id); err == nil && ok {
				title = item.Title
			}
		}
		w.Header().Set("Content-Type", "image/svg+xml")
		w.Header().Set("Cache-Control", "no-cache")
		_, _ = fmt.Fprintf(w, `<svg xmlns="http://www.w3.org/2000/svg" width="600" height="900" viewBox="0 0 600 900">
<defs><linearGradient id="g" x1="0" x2="1" y1="0" y2="1"><stop stop-color="#171815"/><stop offset="1" stop-color="#0c0d0c"/></linearGradient></defs>
<rect width="600" height="900" fill="#070807"/>
<rect x="24" y="24" width="552" height="852" rx="22" fill="url(#g)" stroke="#2a2925" stroke-width="2"/>
<rect x="54" y="54" width="492" height="792" rx="12" fill="none" stroke="#f5f0e7" stroke-opacity="0.08"/>
<path d="M96 702h408v26H96zM96 762h278v18H96z" fill="#f5f0e7" opacity="0.18"/>
<text x="96" y="660" fill="#f5f0e7" fill-opacity="0.82" font-family="Inter,Segoe UI,sans-serif" font-size="42" font-weight="800">%s</text>
</svg>`, html.EscapeString(truncate(title, 20)))
	}
}

func versionsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.Catalog.VersionGroups(r.Context(), queryInt(r, "limit", 100))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "version groups failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"versions": items})
	}
}

func performanceSettingsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		libraryList := deps.Libraries.List()
		profile := "local"
		for _, library := range libraryList {
			if library.StorageType == libraries.StorageNetwork || library.StorageType == libraries.StorageRemovable || library.StorageType == libraries.StorageMounted {
				profile = "conservative"
				break
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"profile":   profile,
			"limits":    deps.Resources.Limits(),
			"queues":    deps.Jobs.Snapshot(),
			"libraries": libraryList,
			"recommendations": []string{
				"Keep scan concurrency low for network/removable storage",
				"Probe jobs are isolated from scan and transcode queues",
				"Playback-critical work is separated from background jobs",
			},
		})
	}
}

func settingsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"config":       settingsPayload(deps.Config),
			"runtimePaths": runtimePaths(deps.Config),
			"libraries":    deps.Libraries.List(),
		})
	}
}

func settingsUpdateHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request config.Config
		if !decodeJSON(w, r, &request) {
			return
		}
		updated := deps.Config
		mergeString(&updated.HTTPAddr, request.HTTPAddr)
		mergeString(&updated.DataDir, request.DataDir)
		mergeString(&updated.TranscodeDir, request.TranscodeDir)
		mergeString(&updated.DownloadsDir, request.DownloadsDir)
		mergeString(&updated.MetadataDir, request.MetadataDir)
		mergeString(&updated.CacheDir, request.CacheDir)
		mergeString(&updated.TempDir, request.TempDir)
		mergeString(&updated.FFmpegPath, request.FFmpegPath)
		mergeString(&updated.FFprobePath, request.FFprobePath)
		mergeInt(&updated.ScanWorkers, request.ScanWorkers)
		mergeInt(&updated.ProbeWorkers, request.ProbeWorkers)
		mergeInt(&updated.TranscodeWorkers, request.TranscodeWorkers)
		mergeInt(&updated.GPUWorkers, request.GPUWorkers)
		if err := config.SaveFile(deps.Config.DataDir, updated); err != nil {
			writeError(w, http.StatusInternalServerError, "settings save failed")
			return
		}
		if err := deps.Catalog.SaveSettings(r.Context(), catalog.RuntimeSettings{
			HTTPAddr:         updated.HTTPAddr,
			DataDir:          updated.DataDir,
			TranscodeDir:     updated.TranscodeDir,
			DownloadsDir:     updated.DownloadsDir,
			MetadataDir:      updated.MetadataDir,
			CacheDir:         updated.CacheDir,
			TempDir:          updated.TempDir,
			FFmpegPath:       updated.FFmpegPath,
			FFprobePath:      updated.FFprobePath,
			ScanWorkers:      updated.ScanWorkers,
			ProbeWorkers:     updated.ProbeWorkers,
			TranscodeWorkers: updated.TranscodeWorkers,
			GPUWorkers:       updated.GPUWorkers,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "settings save failed")
			return
		}
		deps.Events.Publish("settings.updated", settingsPayload(updated))
		writeJSON(w, http.StatusOK, map[string]any{
			"config":          settingsPayload(updated),
			"runtimePaths":    runtimePaths(updated),
			"restartRequired": true,
		})
	}
}

func systemStatusHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, systemstats.Collect(runtimePaths(deps.Config)))
	}
}

func settingsPayload(cfg config.Config) map[string]any {
	return map[string]any{
		"httpAddr":         cfg.HTTPAddr,
		"dataDir":          cfg.DataDir,
		"transcodeDir":     cfg.TranscodeDir,
		"downloadsDir":     cfg.DownloadsDir,
		"metadataDir":      cfg.MetadataDir,
		"cacheDir":         cfg.CacheDir,
		"tempDir":          cfg.TempDir,
		"ffmpegPath":       cfg.FFmpegPath,
		"ffprobePath":      cfg.FFprobePath,
		"scanWorkers":      cfg.ScanWorkers,
		"probeWorkers":     cfg.ProbeWorkers,
		"transcodeWorkers": cfg.TranscodeWorkers,
		"gpuWorkers":       cfg.GPUWorkers,
	}
}

func runtimePaths(cfg config.Config) map[string]string {
	return map[string]string{
		"data":      cfg.DataDir,
		"transcode": cfg.TranscodeDir,
		"downloads": cfg.DownloadsDir,
		"metadata":  cfg.MetadataDir,
		"cache":     cfg.CacheDir,
		"temp":      cfg.TempDir,
	}
}

func mergeString(target *string, value string) {
	value = strings.TrimSpace(value)
	if value != "" {
		*target = value
	}
}

func mergeInt(target *int, value int) {
	if value > 0 {
		*target = value
	}
}

func remoteAccessHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"httpAddr":       deps.Config.HTTPAddr,
			"lanAddresses":   lanAddresses(deps.Config.HTTPAddr),
			"wanAddress":     "",
			"wanLookup":      "available_on_request",
			"recommendation": "Use your own VPN, reverse proxy, or port-forwarding setup. Vyrden does not require hosted relay servers.",
		})
	}
}

func wanAddressHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		client := http.Client{Timeout: 5 * time.Second}
		resp, err := client.Get("https://api.ipify.org?format=json")
		if err != nil {
			writeError(w, http.StatusBadGateway, "wan lookup failed")
			return
		}
		defer resp.Body.Close()
		if resp.StatusCode != http.StatusOK {
			writeError(w, http.StatusBadGateway, "wan lookup failed")
			return
		}
		var payload map[string]string
		if err := json.NewDecoder(resp.Body).Decode(&payload); err != nil {
			writeError(w, http.StatusBadGateway, "wan lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"publicIp": payload["ip"],
			"source":   "https://api.ipify.org",
			"note":     "WAN address was requested explicitly and discovered using an external IP check service.",
		})
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

func mediaSourceTracksHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		tracks, ok, err := deps.Catalog.GetMediaSourceTracks(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source tracks failed")
			return
		}
		if !ok {
			writeJSON(w, http.StatusOK, catalog.MediaSourceTracks{})
			return
		}
		writeJSON(w, http.StatusOK, tracks)
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

func probesHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"probes": deps.Probes.List()})
	}
}

func probeJobHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := deps.Probes.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "probe job not found")
			return
		}
		writeJSON(w, http.StatusOK, job)
	}
}

func probeStartHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request probes.Request
		if !decodeJSON(w, r, &request) {
			return
		}
		job, err := deps.Probes.Start(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
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

func playerHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		id := r.PathValue("id")
		item, ok, err := deps.Catalog.GetMediaSource(r.Context(), id)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s - Vyrden</title>
  <style>
    :root {
      color-scheme: dark;
      --void:#050505;
      --panel:rgba(14,14,13,.78);
      --panel-strong:rgba(24,24,22,.9);
      --text:#f7f1e7;
      --soft:#d6cec0;
      --muted:#9d9487;
      --champagne:#d4b06f;
      --ice:#9adbd4;
      --green:#98d99e;
      --warn:#e0b86e;
      --line:rgba(245,240,231,.14);
      --shadow:0 28px 90px rgba(0,0,0,.56);
    }
    * { box-sizing: border-box; }
    body {
      margin:0;
      min-height:100vh;
      overflow:hidden;
      background:
        radial-gradient(circle at 16%% -10%%, rgba(212,176,111,.15), transparent 30rem),
        radial-gradient(circle at 86%% 4%%, rgba(154,219,212,.13), transparent 34rem),
        var(--void);
      color:var(--text);
      font-family:Inter, ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", sans-serif;
      letter-spacing:0;
    }
    .player-shell { min-height:100vh; display:grid; grid-template-rows:auto 1fr; background:#030303; }
    .topbar {
      position:fixed;
      z-index:5;
      inset:18px 18px auto;
      display:flex;
      align-items:center;
      justify-content:space-between;
      gap:16px;
      pointer-events:none;
      transition:opacity .16s ease, transform .16s ease;
    }
    .hud-toggle {
      position:fixed;
      z-index:6;
      right:clamp(18px, 2vw, 34px);
      bottom:clamp(18px, 2vw, 34px);
      opacity:0;
      transform:translateY(8px);
      transition:opacity .16s ease, transform .16s ease;
      pointer-events:none;
    }
    .brand, .status-pill {
      display:inline-flex;
      align-items:center;
      min-height:38px;
      padding:0 14px;
      border:1px solid var(--line);
      border-radius:999px;
      background:var(--panel);
      backdrop-filter:blur(20px);
      box-shadow:0 14px 48px rgba(0,0,0,.34);
      font-size:13px;
      font-weight:850;
      white-space:nowrap;
    }
    .brand b { color:var(--champagne); margin-right:8px; }
    .status-pill { color:var(--ice); }
    video {
      width:100vw;
      height:100vh;
      object-fit:contain;
      background:linear-gradient(180deg,#020202,#050505);
    }
    .overlay {
      position:fixed;
      z-index:4;
      left:clamp(18px, 2vw, 34px);
      right:clamp(18px, 2vw, 34px);
      bottom:clamp(18px, 2vw, 34px);
      display:grid;
      grid-template-columns:minmax(0, 1fr) clamp(360px, 28vw, 560px);
      gap:clamp(14px, 1.2vw, 22px);
      align-items:end;
      pointer-events:none;
      transition:opacity .18s ease, transform .18s ease;
    }
    body.is-idle .overlay, body.is-idle .topbar {
      opacity:0;
      transform:translateY(10px);
      pointer-events:none;
    }
    body.is-idle .hud-toggle {
      opacity:1;
      transform:none;
      pointer-events:auto;
    }
    .panel {
      border:1px solid var(--line);
      border-radius:12px;
      background:var(--panel);
      backdrop-filter:blur(22px);
      padding:clamp(16px, 1.4vw, 24px);
      box-shadow:var(--shadow);
      min-width:0;
    }
    .title-panel {
      display:grid;
      grid-template-columns:minmax(0,1fr);
      gap:18px;
      align-items:start;
      max-width:100%%;
    }
    .eyebrow {
      color:var(--champagne);
      font-size:12px;
      font-weight:900;
      text-transform:uppercase;
    }
    h1 {
      margin:5px 0 0;
      max-width:min(1180px, 100%%);
      font-size:clamp(30px, 2.55vw, 54px);
      line-height:1.02;
      font-weight:900;
      text-wrap:balance;
    }
    .meta {
      display:flex;
      flex-wrap:wrap;
      gap:8px;
      margin-top:12px;
    }
    .chip {
      display:inline-flex;
      align-items:center;
      min-height:30px;
      padding:0 11px;
      border:1px solid var(--line);
      border-radius:999px;
      background:rgba(245,240,231,.055);
      color:var(--soft);
      font-size:13px;
      font-weight:800;
      white-space:nowrap;
    }
    .chip.good { color:var(--green); border-color:rgba(152,217,158,.28); background:rgba(152,217,158,.08); }
    .chip.route { color:var(--ice); border-color:rgba(154,219,212,.3); background:rgba(154,219,212,.08); }
    .chip.warn { color:var(--warn); border-color:rgba(224,184,110,.3); background:rgba(224,184,110,.08); }
    .control-stack {
      display:flex;
      justify-content:flex-start;
      align-items:center;
      gap:10px;
      flex-wrap:wrap;
      padding-top:2px;
    }
    button, a.button {
      pointer-events:auto;
      min-height:42px;
      padding:0 16px;
      border:1px solid var(--line);
      border-radius:8px;
      background:rgba(245,240,231,.07);
      color:var(--text);
      font:inherit;
      font-size:14px;
      font-weight:850;
      text-decoration:none;
      cursor:pointer;
      white-space:nowrap;
    }
    button.primary { border-color:transparent; background:var(--champagne); color:#11100d; }
    .forecast {
      display:grid;
      gap:10px;
      align-self:stretch;
      min-height:100%%;
      align-content:start;
    }
    .forecast h2 {
      margin:0;
      font-size:14px;
      text-transform:uppercase;
    }
    .decision {
      display:grid;
      gap:6px;
      padding:13px;
      border:1px solid rgba(154,219,212,.32);
      border-radius:9px;
      background:rgba(154,219,212,.075);
    }
    .decision strong { color:var(--ice); font-size:22px; line-height:1.05; }
    .decision span { color:var(--soft); font-size:13px; line-height:1.35; }
    .kv { display:grid; gap:0; }
    .kv div {
      display:grid;
      grid-template-columns:max-content minmax(0,1fr);
      gap:16px;
      padding:10px 0;
      border-top:1px solid rgba(245,240,231,.09);
      font-size:13px;
    }
    .kv span:first-child { color:var(--muted); font-weight:800; }
    .kv span:last-child { color:var(--text); font-weight:850; text-align:right; }
    .hint { margin-top:8px; color:var(--muted); font-size:12px; font-weight:750; }
    @media (max-width: 900px) {
      .overlay { grid-template-columns:1fr; left:12px; right:12px; bottom:12px; }
      .title-panel { grid-template-columns:1fr; }
      .control-stack { justify-content:flex-start; }
      .forecast { display:none; }
      h1 { font-size:clamp(24px, 8vw, 38px); }
    }
  </style>
</head>
<body>
  <div class="player-shell">
    <div class="topbar">
      <div class="brand"><b>V</b> Vyrden Player</div>
      <div class="status-pill" id="sessionState">Starting</div>
    </div>
    <button class="hud-toggle" id="hudToggle" type="button">Controls</button>
    <video id="player" src="/api/media-sources/%s/stream" controls autoplay></video>
  </div>
  <div class="overlay">
    <section class="panel title-panel">
      <div>
        <div class="eyebrow">Now playing</div>
        <h1>%s</h1>
        <div class="meta">
          <span class="chip route" id="decisionMode">Preparing</span>
          <span class="chip" id="progressChip">0:00</span>
          <span class="chip" id="subtitleChip">Subtitles checking</span>
        </div>
        <div class="hint">Space toggles playback. Arrow keys seek. Progress saves automatically.</div>
      </div>
      <div class="control-stack">
        <button class="primary" id="playToggle" type="button">Pause</button>
        <button id="restartButton" type="button">Restart</button>
        <button id="markButton" type="button">Mark Watched</button>
        <a class="button" href="/">Dashboard</a>
      </div>
    </section>
    <aside class="panel forecast">
      <h2>Playback Forecast</h2>
      <div class="decision"><strong id="forecastMode">Checking</strong><span id="forecastReason">Inspecting selected source and client profile.</span></div>
      <div class="kv">
        <div><span>Client</span><span>Web</span></div>
        <div><span>Route</span><span id="forecastRoute">LAN direct</span></div>
        <div><span>Server</span><span id="forecastServer">Low impact</span></div>
      </div>
    </aside>
  </div>
  <script>
    const mediaSourceId = %q;
    const resumeEnabled = new URLSearchParams(location.search).get("start") !== "0";
    let sessionId = "";
    let saveInFlight = false;
    let lastMouseMove = Date.now();
    const player = document.getElementById("player");
    const sessionState = document.getElementById("sessionState");
    const playToggle = document.getElementById("playToggle");
    const restartButton = document.getElementById("restartButton");
    const markButton = document.getElementById("markButton");
    const hudToggle = document.getElementById("hudToggle");
    const decisionMode = document.getElementById("decisionMode");
    const progressChip = document.getElementById("progressChip");
    const subtitleChip = document.getElementById("subtitleChip");
    const forecastMode = document.getElementById("forecastMode");
    const forecastReason = document.getElementById("forecastReason");
    const forecastServer = document.getElementById("forecastServer");
    let idleTimer = 0;

    async function send(path, body, method = "POST", keepalive = false) {
      const response = await fetch(path, { method, keepalive, headers: { "Content-Type": "application/json" }, body: JSON.stringify(body || {}) });
      return response.ok ? response.json() : {};
    }
    async function getJSON(path, fallback = {}) {
      return fetch(path).then(r => r.ok ? r.json() : fallback).catch(() => fallback);
    }
    function formatTime(value) {
      value = Math.max(0, Math.floor(Number(value || 0)));
      const hours = Math.floor(value / 3600);
      const minutes = Math.floor((value %% 3600) / 60);
      const seconds = value %% 60;
      return hours ? hours + ":" + String(minutes).padStart(2, "0") + ":" + String(seconds).padStart(2, "0") : minutes + ":" + String(seconds).padStart(2, "0");
    }
    function progressBody(status) {
      return { progressSeconds: player.currentTime || 0, durationSeconds: Number.isFinite(player.duration) ? player.duration : 0, status: status || (player.paused ? "paused" : "playing") };
    }
    async function loadForecast() {
      const decision = await getJSON("/api/playback/decision?mediaSourceId=" + mediaSourceId + "&clientProfile=web");
      const mode = decision.mode || "direct";
      const label = mode.replaceAll("_", " ");
      decisionMode.textContent = label;
      decisionMode.className = "chip " + (mode === "direct" ? "good" : mode === "remux" ? "route" : "warn");
      forecastMode.textContent = label;
      forecastReason.textContent = decision.reason || "Direct file stream is available for this client.";
      forecastServer.textContent = mode === "direct" ? "Low impact" : "Server work required";
      return mode;
    }
    async function loadSubtitles() {
      const subtitles = await getJSON("/api/media-sources/" + mediaSourceId + "/subtitles", { sidecars: [] });
      const sidecars = subtitles.sidecars || [];
      subtitleChip.textContent = sidecars.length ? sidecars.length + " subtitle file" + (sidecars.length === 1 ? "" : "s") : "No sidecar subtitles";
      sidecars.forEach((item, index) => {
        if (item.format !== "vtt") return;
        const track = document.createElement("track");
        track.kind = "subtitles";
        track.label = item.language || item.relPath || "Subtitle";
        track.srclang = item.language || "und";
        track.src = "/api/media-sources/" + mediaSourceId + "/subtitles/" + index;
        player.appendChild(track);
      });
    }
    async function loadResumeState() {
      const state = await getJSON("/api/playback/state/" + mediaSourceId);
      player.addEventListener("loadedmetadata", () => {
        if (resumeEnabled && state.progressSeconds > 5 && state.progressSeconds < player.duration - 10) {
          player.currentTime = state.progressSeconds;
        }
      }, { once: true });
    }
    async function startSession(mode) {
      const session = await send("/api/sessions", { mediaSourceId, deviceId: "web", mode });
      sessionId = session.id || "";
      sessionState.textContent = sessionId ? "Live session" : "Local playback";
    }
    async function saveProgress(status) {
      if (saveInFlight) return;
      saveInFlight = true;
      const body = progressBody(status);
      try {
        if (sessionId) await send("/api/sessions/" + sessionId, body, "PATCH");
        await send("/api/playback/state/" + mediaSourceId, body, "PUT");
      } finally {
        saveInFlight = false;
      }
    }
    async function stopSession(status = "stopped") {
      const body = progressBody(status);
      if (sessionId) {
        await send("/api/sessions/" + sessionId, body, "PATCH").catch(() => {});
        await fetch("/api/sessions/" + sessionId, { method: "DELETE", keepalive: true }).catch(() => {});
        sessionId = "";
      }
      await send("/api/playback/state/" + mediaSourceId, body, "PUT", true).catch(() => {});
    }
    function refreshProgress() {
      const current = formatTime(player.currentTime || 0);
      const total = Number.isFinite(player.duration) && player.duration > 0 ? formatTime(player.duration) : "--:--";
      progressChip.textContent = current + " / " + total;
      playToggle.textContent = player.paused ? "Play" : "Pause";
      sessionState.textContent = player.paused ? "Paused" : "Playing";
    }
    function showHud(timeout = 2600) {
      lastMouseMove = Date.now();
      document.body.classList.remove("is-idle");
      clearTimeout(idleTimer);
      idleTimer = setTimeout(() => {
        if (!document.activeElement || document.activeElement === document.body || document.activeElement === player) {
          document.body.classList.add("is-idle");
        }
      }, timeout);
    }
    function hideHudSoon() {
      clearTimeout(idleTimer);
      idleTimer = setTimeout(() => document.body.classList.add("is-idle"), 900);
    }
    playToggle.addEventListener("click", () => player.paused ? player.play() : player.pause());
    restartButton.addEventListener("click", () => { player.currentTime = 0; player.play(); saveProgress("playing"); });
    markButton.addEventListener("click", async () => {
      await send("/api/playback/state/" + mediaSourceId, { watched: true, progressSeconds: player.duration || player.currentTime || 0, durationSeconds: player.duration || 0 }, "PUT");
      markButton.textContent = "Marked";
    });
    hudToggle.addEventListener("click", () => showHud(4200));
    player.addEventListener("timeupdate", refreshProgress);
    player.addEventListener("loadedmetadata", refreshProgress);
    player.addEventListener("play", () => { refreshProgress(); saveProgress("playing"); hideHudSoon(); });
    player.addEventListener("pause", () => { refreshProgress(); saveProgress("paused"); showHud(4200); });
    player.addEventListener("ended", () => stopSession("completed"));
    window.addEventListener("keydown", event => {
      if (event.target && ["INPUT", "TEXTAREA", "SELECT"].includes(event.target.tagName)) return;
      if (event.code === "Space") { event.preventDefault(); player.paused ? player.play() : player.pause(); }
      if (event.code === "ArrowRight") player.currentTime = Math.min((player.duration || 0), (player.currentTime || 0) + 10);
      if (event.code === "ArrowLeft") player.currentTime = Math.max(0, (player.currentTime || 0) - 10);
    });
    window.addEventListener("mousemove", () => {
      showHud();
    });
    window.addEventListener("pointerdown", event => {
      if (event.target === player || event.target === document.body) showHud();
    });
    setInterval(() => {
      refreshProgress();
      saveProgress();
    }, 10000);
    window.addEventListener("beforeunload", () => { stopSession("stopped"); });
    (async function boot() {
      await loadResumeState();
      const mode = await loadForecast();
      await loadSubtitles();
      await startSession(mode);
      refreshProgress();
      showHud(2400);
    })();
  </script>
</body>
</html>`, html.EscapeString(item.Name), id, html.EscapeString(item.Name), id)
	}
}

func mediaSourceSubtitlesHandler(deps Deps) http.HandlerFunc {
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
		writeJSON(w, http.StatusOK, map[string]any{"sidecars": subtitles.DiscoverSidecars(item.Path)})
	}
}

func mediaSourceSubtitleStreamHandler(deps Deps) http.HandlerFunc {
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
		index := parsePathInt(r.PathValue("index"), -1)
		sidecars := subtitles.DiscoverSidecars(item.Path)
		if index < 0 || index >= len(sidecars) {
			writeError(w, http.StatusNotFound, "subtitle not found")
			return
		}
		if sidecars[index].Format == "vtt" {
			w.Header().Set("Content-Type", "text/vtt; charset=utf-8")
		} else {
			w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		}
		http.ServeFile(w, r, sidecars[index].Path)
	}
}

func workHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"work": deps.Transcode.List()})
	}
}

func workStartHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request transcode.Request
		if !decodeJSON(w, r, &request) {
			return
		}
		if request.MediaSourceID != "" && request.SourcePath == "" {
			source, ok, err := deps.Catalog.GetMediaSource(r.Context(), request.MediaSourceID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "media source lookup failed")
				return
			}
			if !ok {
				writeError(w, http.StatusNotFound, "media source not found")
				return
			}
			request.SourcePath = source.Path
		}
		job, err := deps.Transcode.Start(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
	}
}

func downloadsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"downloads": deps.Downloads.List()})
	}
}

func downloadJobHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := deps.Downloads.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "download job not found")
			return
		}
		writeJSON(w, http.StatusOK, job)
	}
}

func downloadStartHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request downloads.Request
		if !decodeJSON(w, r, &request) {
			return
		}
		if request.MediaSourceID != "" && request.SourcePath == "" {
			source, ok, err := deps.Catalog.GetMediaSource(r.Context(), request.MediaSourceID)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "media source lookup failed")
				return
			}
			if !ok {
				writeError(w, http.StatusNotFound, "media source not found")
				return
			}
			request.SourcePath = source.Path
			request.SourceName = source.Name
		}
		job, err := deps.Downloads.Start(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, job)
	}
}

func downloadFileHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := deps.Downloads.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "download job not found")
			return
		}
		if job.Status != downloads.StatusCompleted {
			writeError(w, http.StatusConflict, "download is not ready")
			return
		}
		w.Header().Set("Content-Disposition", fmt.Sprintf("attachment; filename=%q", filepath.Base(job.OutputPath)))
		http.ServeFile(w, r, job.OutputPath)
	}
}

func deviceProfilesHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"profiles": deps.Devices.Profiles()})
	}
}

func sessionsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"sessions": deps.Sessions.List()})
	}
}

func sessionStartHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request sessions.StartRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		session, err := deps.Sessions.Start(request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, session)
	}
}

func sessionUpdateHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request sessions.UpdateRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		session, ok := deps.Sessions.Update(r.PathValue("id"), request)
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		_, _ = deps.PlayState.Set(r.Context(), session.MediaSourceID, playstate.Update{
			UserID:          session.UserID,
			ProgressSeconds: session.Progress,
			DurationSeconds: session.Duration,
		})
		writeJSON(w, http.StatusOK, session)
	}
}

func sessionStopHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		session, ok := deps.Sessions.Stop(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeJSON(w, http.StatusOK, session)
	}
}

func playbackRecentHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := deps.PlayState.Recent(r.Context(), r.URL.Query().Get("userId"), queryInt(r, "limit", 24))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "recent playback lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"recent": items})
	}
}

func playbackStateGetHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, _, err := deps.PlayState.Get(r.Context(), r.URL.Query().Get("userId"), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "playback state lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, state)
	}
}

func playbackStateSetHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request playstate.Update
		if !decodeJSON(w, r, &request) {
			return
		}
		state, err := deps.PlayState.Set(r.Context(), r.PathValue("id"), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, state)
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
				MediaSourceID:    source.ID,
				Container:        source.Container,
				VideoCodec:       source.VideoCodec,
				AudioStreams:     source.AudioStreams,
				SubtitleStreams:  source.SubtitleStreams,
				SidecarSubtitles: len(subtitles.DiscoverSidecars(source.Path)),
				Bitrate:          source.Bitrate,
				Probed:           source.Probed,
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

func lanAddresses(httpAddr string) []string {
	_, port, err := net.SplitHostPort(httpAddr)
	if err != nil || port == "" {
		port = "8097"
	}
	output := []string{}
	interfaces, err := net.Interfaces()
	if err != nil {
		return output
	}
	for _, iface := range interfaces {
		if iface.Flags&net.FlagUp == 0 || iface.Flags&net.FlagLoopback != 0 {
			continue
		}
		addrs, err := iface.Addrs()
		if err != nil {
			continue
		}
		for _, addr := range addrs {
			var ip net.IP
			switch value := addr.(type) {
			case *net.IPNet:
				ip = value.IP
			case *net.IPAddr:
				ip = value.IP
			}
			if ip == nil || ip.IsLoopback() {
				continue
			}
			if ipv4 := ip.To4(); ipv4 != nil {
				output = append(output, "http://"+ipv4.String()+":"+port)
			}
		}
	}
	return output
}

func truncate(value string, limit int) string {
	runes := []rune(value)
	if len(runes) <= limit {
		return value
	}
	return string(runes[:limit])
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

func parsePathInt(value string, fallback int) int {
	if value == "" {
		return fallback
	}
	output := 0
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return fallback
		}
		output = output*10 + int(ch-'0')
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
