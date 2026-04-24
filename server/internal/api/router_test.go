package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/config"
	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/playback"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/scans"
	"github.com/vyrdenhq/vyrden/server/internal/sessions"
	"github.com/vyrdenhq/vyrden/server/internal/tv"
)

func TestHealthUsesStableStartedAt(t *testing.T) {
	startedAt := time.Date(2026, 4, 24, 3, 4, 5, 0, time.UTC)
	router := NewRouter(testDeps(t, startedAt))

	first := getJSON(t, router, "/api/health")
	second := getJSON(t, router, "/api/health")

	if first["startedAt"] != startedAt.Format(time.RFC3339) {
		t.Fatalf("expected startedAt %q, got %q", startedAt.Format(time.RFC3339), first["startedAt"])
	}
	if first["startedAt"] != second["startedAt"] {
		t.Fatalf("expected stable startedAt, got %q then %q", first["startedAt"], second["startedAt"])
	}
}

func TestRootServesDevConsole(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", response.Code)
	}
	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatalf("read body: %v", err)
	}
	if !bytes.Contains(body, []byte("Vyrden Dev Console")) {
		t.Fatalf("expected dev console html, got %s", string(body))
	}
}

func TestArchitectureExposesSeparateQueuesAndWorkloads(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	payload := getJSON(t, router, "/api/architecture")

	queues, ok := payload["queues"].([]any)
	if !ok || len(queues) != 3 {
		t.Fatalf("expected scan, probe, and transcode queues, got %#v", payload["queues"])
	}
	workloads, ok := payload["workloads"].([]any)
	if !ok || len(workloads) != 3 {
		t.Fatalf("expected three workload classes, got %#v", payload["workloads"])
	}
}

func TestPlaybackDecisionEndpointIsExplicitlyDeferred(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	payload := getJSON(t, router, "/api/playback/decision?mediaSourceId=abc&clientProfile=android-tv")

	if payload["mode"] != string(playback.DecisionDeferred) {
		t.Fatalf("expected deferred decision mode, got %#v", payload["mode"])
	}
	if payload["mediaSourceId"] != "abc" {
		t.Fatalf("expected media source to round-trip, got %#v", payload["mediaSourceId"])
	}
	if payload["clientProfile"] != "android-tv" {
		t.Fatalf("expected client profile to round-trip, got %#v", payload["clientProfile"])
	}
}

func TestMovieScanEndpointUsesMovieClassifier(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Dune Part Two (2024)", "Dune.Part.Two.2024.2160p.REMUX.mkv"))
	writeTestFile(t, filepath.Join(root, "poster.jpg"))

	router := NewRouter(testDeps(t, time.Now()))
	payload := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{
		"path":        root,
		"sampleLimit": 10,
	})
	job := waitForScan(t, router, payload["id"].(string))
	result := job["result"].(map[string]any)
	moviesResult := result["movies"].(map[string]any)

	if moviesResult["moviesFound"] != float64(1) {
		t.Fatalf("expected one movie, got %#v", moviesResult["moviesFound"])
	}
	persisted := moviesResult["persisted"].(map[string]any)
	if persisted["mediaSources"] != float64(1) {
		t.Fatalf("expected one persisted media source, got %#v", persisted["mediaSources"])
	}
}

func TestTVScanEndpointUsesEpisodeClassifier(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "The Bear", "Season 02", "The.Bear.S02E03.Sundae.1080p.WEB-DL.mkv"))
	writeTestFile(t, filepath.Join(root, "The Bear", "Season 02", "notes.txt"))

	router := NewRouter(testDeps(t, time.Now()))
	payload := postJSON(t, router, "/api/libraries/tv/scan", map[string]any{
		"path":        root,
		"sampleLimit": 10,
	})
	job := waitForScan(t, router, payload["id"].(string))
	result := job["result"].(map[string]any)
	tvResult := result["tv"].(map[string]any)

	if tvResult["episodesFound"] != float64(1) {
		t.Fatalf("expected one episode, got %#v", tvResult["episodesFound"])
	}
}

func TestCatalogSummaryUpdatesAfterScans(t *testing.T) {
	movieRoot := t.TempDir()
	tvRoot := t.TempDir()
	writeTestFile(t, filepath.Join(movieRoot, "Heat (1995)", "Heat.1995.1080p.BluRay.mkv"))
	writeTestFile(t, filepath.Join(tvRoot, "The Bear", "Season 02", "The.Bear.S02E03.Sundae.1080p.WEB-DL.mkv"))

	router := NewRouter(testDeps(t, time.Now()))
	payload := postJSON(t, router, "/api/libraries/scan", map[string]any{
		"moviesPath":  movieRoot,
		"tvPath":      tvRoot,
		"sampleLimit": 10,
	})
	waitForScan(t, router, payload["id"].(string))
	summary := getJSON(t, router, "/api/catalog/summary")

	if summary["libraries"] != float64(2) {
		t.Fatalf("expected two libraries, got %#v", summary["libraries"])
	}
	if summary["mediaSources"] != float64(2) {
		t.Fatalf("expected two media sources, got %#v", summary["mediaSources"])
	}
	if summary["movies"] != float64(1) {
		t.Fatalf("expected one movie, got %#v", summary["movies"])
	}
	if summary["episodes"] != float64(1) {
		t.Fatalf("expected one episode, got %#v", summary["episodes"])
	}
}

func testDeps(t *testing.T, startedAt time.Time) Deps {
	t.Helper()

	cfg := config.Config{
		HTTPAddr:         "127.0.0.1:8097",
		EventBuffer:      2,
		ScanWorkers:      1,
		ProbeWorkers:     2,
		TranscodeWorkers: 1,
		GPUWorkers:       1,
	}
	manager := resources.NewManager(resources.Limits{
		ScanWorkers:      cfg.ScanWorkers,
		ProbeWorkers:     cfg.ProbeWorkers,
		TranscodeWorkers: cfg.TranscodeWorkers,
		GPUWorkers:       cfg.GPUWorkers,
	})
	registry := jobs.NewRegistry(manager)
	ctx, cancel := context.WithCancel(context.Background())
	registry.Start(ctx)
	t.Cleanup(cancel)
	db, err := database.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close database: %v", err)
		}
	})
	libraryService := libraries.NewService()
	scannerService := scanner.NewService()
	catalogService := catalog.NewService(db)
	movieService := movies.NewService()
	tvService := tv.NewService()
	eventBus := events.NewBus(cfg.EventBuffer)
	return Deps{
		Config:    cfg,
		StartedAt: startedAt,
		Events:    eventBus,
		Resources: manager,
		Jobs:      registry,
		Libraries: libraryService,
		Scanner:   scannerService,
		Scans:     scans.NewService(cfg, eventBus, registry.Scan, libraryService, scannerService, catalogService, movieService, tvService),
		Catalog:   catalogService,
		Media:     media.NewService(),
		Movies:    movieService,
		TV:        tvService,
		Playback:  playback.NewService(),
		Sessions:  sessions.NewService(),
	}
}

func getJSON(t *testing.T, router http.Handler, path string) map[string]any {
	t.Helper()

	request := httptest.NewRequest(http.MethodGet, path, nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK && response.Code != http.StatusAccepted {
		t.Fatalf("expected 200 or 202, got %d: %s", response.Code, response.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return payload
}

func waitForScan(t *testing.T, router http.Handler, id string) map[string]any {
	t.Helper()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		job := getJSON(t, router, "/api/scans/"+id)
		switch job["status"] {
		case string(scans.StatusCompleted):
			return job
		case string(scans.StatusFailed):
			t.Fatalf("scan failed: %#v", job)
		}
		time.Sleep(25 * time.Millisecond)
	}
	t.Fatalf("scan %s did not complete", id)
	return nil
}

func postJSON(t *testing.T, router http.Handler, path string, body any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	request := httptest.NewRequest(http.MethodPost, path, bytes.NewReader(raw))
	request.Header.Set("Content-Type", "application/json")
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK && response.Code != http.StatusAccepted {
		t.Fatalf("expected 200 or 202, got %d: %s", response.Code, response.Body.String())
	}

	var payload map[string]any
	if err := json.NewDecoder(response.Body).Decode(&payload); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return payload
}

func writeTestFile(t *testing.T, path string) {
	t.Helper()

	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(path, []byte("test"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
}
