package api

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/vyrdenhq/vyrden/server/internal/auth"
	"github.com/vyrdenhq/vyrden/server/internal/catalog"
	"github.com/vyrdenhq/vyrden/server/internal/config"
	"github.com/vyrdenhq/vyrden/server/internal/database"
	"github.com/vyrdenhq/vyrden/server/internal/devices"
	"github.com/vyrdenhq/vyrden/server/internal/downloads"
	"github.com/vyrdenhq/vyrden/server/internal/events"
	"github.com/vyrdenhq/vyrden/server/internal/jobs"
	"github.com/vyrdenhq/vyrden/server/internal/libraries"
	"github.com/vyrdenhq/vyrden/server/internal/media"
	metaprovider "github.com/vyrdenhq/vyrden/server/internal/metadata"
	"github.com/vyrdenhq/vyrden/server/internal/migration"
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/observability"
	"github.com/vyrdenhq/vyrden/server/internal/pairing"
	"github.com/vyrdenhq/vyrden/server/internal/playback"
	"github.com/vyrdenhq/vyrden/server/internal/playstate"
	"github.com/vyrdenhq/vyrden/server/internal/probe"
	"github.com/vyrdenhq/vyrden/server/internal/probes"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/scans"
	"github.com/vyrdenhq/vyrden/server/internal/sessions"
	"github.com/vyrdenhq/vyrden/server/internal/streaming"
	"github.com/vyrdenhq/vyrden/server/internal/subtitles"
	"github.com/vyrdenhq/vyrden/server/internal/transcode"
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

func TestRootServesWebApp(t *testing.T) {
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
	if !bytes.Contains(body, []byte("Vyrden")) || !bytes.Contains(body, []byte("/app.js")) {
		t.Fatalf("expected web app html, got %s", string(body))
	}
}

func TestSecurityHeadersPresentOnAPIAndWebResponses(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	for _, path := range []string{"/", "/api/health"} {
		request := httptest.NewRequest(http.MethodGet, path, nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusOK {
			t.Fatalf("expected %s 200, got %d", path, response.Code)
		}
		for header, want := range map[string]string{
			"X-Content-Type-Options": "nosniff",
			"X-Frame-Options":        "DENY",
			"Referrer-Policy":        "no-referrer",
		} {
			if got := response.Header().Get(header); got != want {
				t.Fatalf("expected %s header %s=%q, got %q", path, header, want, got)
			}
		}
		if got := response.Header().Get("Content-Security-Policy"); !strings.Contains(got, "default-src 'self'") {
			t.Fatalf("expected CSP on %s, got %q", path, got)
		}
	}
}

func TestDisallowedOriginBlockedByCORS(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	request.Header.Set("Origin", "https://evil.example")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected disallowed origin 403, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "" {
		t.Fatalf("expected no allow-origin header, got %q", got)
	}
}

func TestLocalOriginAllowedByCORS(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	request := httptest.NewRequest(http.MethodOptions, "/api/health", nil)
	request.Header.Set("Origin", "http://localhost:8097")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusNoContent {
		t.Fatalf("expected preflight 204, got %d", response.Code)
	}
	if got := response.Header().Get("Access-Control-Allow-Origin"); got != "http://localhost:8097" {
		t.Fatalf("expected local allow-origin, got %q", got)
	}
}

func TestArtworkPathTraversalFailsSafely(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	for _, target := range []string{
		"/api/artwork/movie/example?type=..",
		"/api/artwork/movie/example?type=poster/../../secret",
		"/api/artwork/%2e%2e/example",
	} {
		request := httptest.NewRequest(http.MethodGet, target, nil)
		response := httptest.NewRecorder()

		router.ServeHTTP(response, request)

		if response.Code != http.StatusBadRequest {
			t.Fatalf("expected traversal request %s to fail with 400, got %d", target, response.Code)
		}
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

func TestRequestCorrelationAndMetrics(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	request := httptest.NewRequest(http.MethodGet, "/api/health", nil)
	request.Header.Set("X-Request-ID", "trace-test-1")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected health 200, got %d", response.Code)
	}
	if got := response.Header().Get("X-Request-ID"); got != "trace-test-1" {
		t.Fatalf("expected propagated request id, got %q", got)
	}
	metrics := getJSON(t, router, "/api/metrics")
	requests := metrics["requests"].([]any)
	if len(requests) == 0 {
		t.Fatalf("expected request metrics, got %#v", metrics)
	}
	var found bool
	for _, item := range requests {
		metric := item.(map[string]any)
		if metric["path"] == "/api/health" && metric["lastCorrelationId"] == "trace-test-1" {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected health metric with correlation id, got %#v", requests)
	}
	queues := metrics["queues"].([]any)
	if len(queues) != 3 {
		t.Fatalf("expected queue metrics, got %#v", metrics["queues"])
	}
	firstQueue := queues[0].(map[string]any)
	if _, ok := firstQueue["workerUtilization"]; !ok {
		t.Fatalf("expected worker utilization in queue metric, got %#v", firstQueue)
	}
}

func TestClientBootstrapDefaultsToAppleTVContract(t *testing.T) {
	startedAt := time.Date(2026, 4, 30, 4, 5, 6, 0, time.UTC)
	router := NewRouter(testDeps(t, startedAt))
	request := httptest.NewRequest(http.MethodGet, "/api/client/bootstrap", nil)
	request.Host = "vyrden.local:8097"
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected bootstrap 200, got %d", response.Code)
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	server := payload["server"].(map[string]any)
	if server["baseUrl"] != "http://vyrden.local:8097" || server["startedAt"] != startedAt.Format(time.RFC3339) {
		t.Fatalf("expected server identity, got %#v", server)
	}
	client := payload["client"].(map[string]any)
	profile := client["profile"].(map[string]any)
	if client["requestedProfile"] != "apple-tv" || profile["id"] != "apple-tv" || profile["supportsHls"] != true {
		t.Fatalf("expected apple-tv HLS profile, got %#v", client)
	}
	authPayload := payload["auth"].(map[string]any)
	if authPayload["required"] != false {
		t.Fatalf("expected auth disabled bootstrap to report auth not required, got %#v", authPayload)
	}
	features := payload["features"].(map[string]any)
	if features["vendorRelay"] != false || features["hlsAdaptive"] != true {
		t.Fatalf("expected local-first playback features, got %#v", features)
	}
	endpoints := payload["endpoints"].(map[string]any)
	if endpoints["adaptiveMaster"] != "/api/media-sources/{id}/adaptive/master.m3u8" || endpoints["sessions"] != "/api/sessions" {
		t.Fatalf("expected tv playback endpoints, got %#v", endpoints)
	}
}

func TestClientBootstrapIsReadableBeforeLogin(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	request := httptest.NewRequest(http.MethodGet, "/api/client/bootstrap?clientProfile=ios", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusOK {
		t.Fatalf("expected unauthenticated bootstrap 200, got %d: %s", response.Code, response.Body.String())
	}
	var payload map[string]any
	if err := json.Unmarshal(response.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode bootstrap: %v", err)
	}
	authPayload := payload["auth"].(map[string]any)
	if authPayload["required"] != true {
		t.Fatalf("expected auth-enabled bootstrap to report auth required, got %#v", authPayload)
	}
	client := payload["client"].(map[string]any)
	profile := client["profile"].(map[string]any)
	if client["requestedProfile"] != "ios" || profile["id"] != "ios" {
		t.Fatalf("expected requested ios profile, got %#v", client)
	}
}

func TestPairingRequestCreateStatusAndAdminApprove(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))

	create := requestJSON(t, router, http.MethodPost, "/api/pairing/requests", map[string]any{
		"deviceName":    "Living Room Apple TV",
		"clientProfile": "apple-tv",
	})
	pairingID, _ := create["id"].(string)
	if pairingID == "" || create["status"] != "pending" || len(create["code"].(string)) != 6 {
		t.Fatalf("expected pending pairing request, got %#v", create)
	}

	status := getJSON(t, router, "/api/pairing/requests/"+pairingID)
	if status["status"] != "pending" || status["code"] == "" {
		t.Fatalf("expected public pending status with code, got %#v", status)
	}

	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")
	approved := client.requestJSON(t, router, http.MethodPost, "/api/pairing/requests/"+pairingID+"/approve", map[string]any{})
	if approved.status != http.StatusOK {
		t.Fatalf("expected approval 200, got %d: %s", approved.status, approved.body)
	}
	if approved.payload["status"] != "approved" || approved.payload["deviceId"] == "" || approved.payload["code"] != nil {
		t.Fatalf("expected approved pairing without code, got %#v", approved.payload)
	}

	polled := getJSON(t, router, "/api/pairing/requests/"+pairingID)
	if polled["status"] != "approved" || polled["deviceId"] == "" || polled["code"] != nil {
		t.Fatalf("expected approved polling result without code, got %#v", polled)
	}
}

func TestPairingAdminRoutesRequireAdmin(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	if _, err := deps.Auth.CreateUser(context.Background(), "viewer", "viewer-password-123!", "Viewer", "standard"); err != nil {
		t.Fatalf("create viewer: %v", err)
	}
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "viewer", "viewer-password-123!")

	list := client.requestJSON(t, router, http.MethodGet, "/api/pairing/requests", nil)
	if list.status != http.StatusForbidden {
		t.Fatalf("expected standard user to get 403 on pairing list, got %d: %s", list.status, list.body)
	}
}

func TestReadinessReflectsDegradedRuntimePath(t *testing.T) {
	deps := testDeps(t, time.Now())
	badPath := filepath.Join(t.TempDir(), "not-a-dir")
	if err := os.WriteFile(badPath, []byte("file"), 0o600); err != nil {
		t.Fatalf("write bad path: %v", err)
	}
	deps.Config.TranscodeDir = badPath
	router := NewRouter(deps)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, httptest.NewRequest(http.MethodGet, "/api/ready", nil))

	if response.Code != http.StatusServiceUnavailable {
		t.Fatalf("expected degraded readiness 503, got %d: %s", response.Code, response.Body.String())
	}
	payload := decodeBody(t, response.Body.Bytes())
	if payload["status"] != "degraded" {
		t.Fatalf("expected degraded status, got %#v", payload)
	}
}

func TestOperationalEventCarriesCorrelationID(t *testing.T) {
	deps := testDeps(t, time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	eventsCh, stop := deps.Events.Subscribe(ctx)
	defer stop()
	router := NewRouter(deps)
	request := httptest.NewRequest(http.MethodPost, "/api/sessions", bytes.NewReader(mustJSON(t, map[string]any{
		"mediaSourceId": "source-1",
		"title":         "Example",
	})))
	request.Header.Set("Content-Type", "application/json")
	request.Header.Set("X-Request-ID", "trace-session-1")
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusAccepted {
		t.Fatalf("expected session start 202, got %d: %s", response.Code, response.Body.String())
	}
	deadline := time.After(time.Second)
	for {
		select {
		case event := <-eventsCh:
			if event.Type != "api.session.accepted" {
				continue
			}
			payload := event.Data.(map[string]any)
			if payload["correlationId"] != "trace-session-1" {
				t.Fatalf("expected correlated event, got %#v", payload)
			}
			return
		case <-deadline:
			t.Fatal("timed out waiting for correlated session event")
		}
	}
}

func TestSystemStatusAndRuntimeFolders(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	status := getJSON(t, router, "/api/system/status")

	cpu := status["cpu"].(map[string]any)
	if cpu["cores"].(float64) < 1 {
		t.Fatalf("expected at least one cpu core, got %#v", cpu)
	}
	disks := status["disks"].([]any)
	if len(disks) < 5 {
		t.Fatalf("expected runtime folder disk entries, got %#v", status)
	}
	settings := getJSON(t, router, "/api/settings")
	paths := settings["runtimePaths"].(map[string]any)
	if paths["transcode"] == "" || paths["metadata"] == "" {
		t.Fatalf("expected configurable runtime paths, got %#v", paths)
	}
}

func TestClientHomeReturnsTVRows(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Movies", "Arrival (2016)", "Arrival.2016.1080p.mkv"))
	writeTestFile(t, filepath.Join(root, "TV", "The Bear", "Season 01", "The.Bear.S01E01.1080p.mkv"))

	router := NewRouter(testDeps(t, time.Now()))
	movieScan := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{
		"path":        filepath.Join(root, "Movies"),
		"sampleLimit": 10,
	})
	waitForScan(t, router, movieScan["id"].(string))
	tvScan := postJSON(t, router, "/api/libraries/tv/scan", map[string]any{
		"path":        filepath.Join(root, "TV"),
		"sampleLimit": 10,
	})
	waitForScan(t, router, tvScan["id"].(string))

	home := getJSON(t, router, "/api/client/home?clientProfile=apple-tv")
	if home["profile"] != "apple-tv" {
		t.Fatalf("expected apple-tv profile, got %#v", home)
	}
	hero := home["hero"].(map[string]any)
	if hero["title"] == "" || hero["kind"] == "" {
		t.Fatalf("expected hero item, got %#v", hero)
	}
	rows := home["rows"].([]any)
	if len(rows) != 4 {
		t.Fatalf("expected four TV rows, got %#v", rows)
	}
	rowItems := map[string]int{}
	for _, raw := range rows {
		row := raw.(map[string]any)
		items := row["items"].([]any)
		rowItems[row["id"].(string)] = len(items)
	}
	if rowItems["movies"] != 1 || rowItems["tv"] != 1 || rowItems["recently-added"] < 2 {
		t.Fatalf("expected movie, tv, and recently added rows, got %#v", rowItems)
	}
}

func TestClientDetailEndpointsReturnNativePayloads(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Movies", "Arrival (2016)", "Arrival.2016.1080p.mkv"))
	writeTestFile(t, filepath.Join(root, "Movies", "Arrival (2016)", "Arrival.2016.1080p.en.ass"))
	writeTestFile(t, filepath.Join(root, "TV", "The Bear", "Season 01", "The.Bear.S01E01.1080p.mkv"))

	router := NewRouter(testDeps(t, time.Now()))
	movieScan := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{
		"path":        filepath.Join(root, "Movies"),
		"sampleLimit": 10,
	})
	waitForScan(t, router, movieScan["id"].(string))
	tvScan := postJSON(t, router, "/api/libraries/tv/scan", map[string]any{
		"path":        filepath.Join(root, "TV"),
		"sampleLimit": 10,
	})
	waitForScan(t, router, tvScan["id"].(string))

	movies := getJSON(t, router, "/api/movies")
	movieID := movies["movies"].([]any)[0].(map[string]any)["id"].(string)
	movieDetail := getJSON(t, router, "/api/client/movies/"+movieID+"?clientProfile=apple-tv")
	if movieDetail["defaultMediaSourceId"] == "" {
		t.Fatalf("expected default movie media source, got %#v", movieDetail)
	}
	movieVersions := movieDetail["versions"].([]any)
	if len(movieVersions) != 1 {
		t.Fatalf("expected one movie version, got %#v", movieDetail)
	}
	movieVersion := movieVersions[0].(map[string]any)
	sidecars := movieVersion["sidecars"].([]any)
	if len(sidecars) != 1 {
		t.Fatalf("expected one sidecar subtitle payload, got %#v", movieVersion)
	}
	conversion := sidecars[0].(map[string]any)["conversion"].(map[string]any)
	if conversion["outputFormat"] != "webvtt" || conversion["reasonCode"] != "subtitle_text_conversion_available" {
		t.Fatalf("expected text subtitle conversion plan, got %#v", conversion)
	}

	series := getJSON(t, router, "/api/series")
	seriesID := series["series"].([]any)[0].(map[string]any)["id"].(string)
	seriesDetail := getJSON(t, router, "/api/client/series/"+seriesID+"?clientProfile=apple-tv")
	if seriesDetail["defaultMediaSourceId"] == "" {
		t.Fatalf("expected default series media source, got %#v", seriesDetail)
	}
	seasons := seriesDetail["seasons"].([]any)
	if len(seasons) != 1 {
		t.Fatalf("expected one season payload, got %#v", seriesDetail)
	}
	episodes := seasons[0].(map[string]any)["episodes"].([]any)
	if len(episodes) != 1 {
		t.Fatalf("expected one episode payload, got %#v", seasons[0])
	}
	episodeVersions := episodes[0].(map[string]any)["versions"].([]any)
	if len(episodeVersions) != 1 {
		t.Fatalf("expected one episode version payload, got %#v", episodes[0])
	}
}

func TestClientPlaybackStartHeartbeatAndStop(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Arrival (2016)", "Arrival.2016.1080p.mkv"))

	router := NewRouter(testDeps(t, time.Now()))
	payload := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{"path": root})
	waitForScan(t, router, payload["id"].(string))
	sources := getJSON(t, router, "/api/media-sources")
	sourceID := sources["mediaSources"].([]any)[0].(map[string]any)["id"].(string)

	started := postJSON(t, router, "/api/client/playback/start", map[string]any{
		"mediaSourceId": sourceID,
		"deviceId":      "apple-tv-living-room",
		"clientProfile": "apple-tv",
		"routeType":     "lan",
	})
	if started["sessionId"] == "" || started["heartbeatUrl"] == "" || started["stopUrl"] == "" {
		t.Fatalf("expected client playback session payload, got %#v", started)
	}
	route := started["route"].(map[string]any)
	if route["route"] != "direct" || route["url"] == "" {
		t.Fatalf("expected direct route payload, got %#v", route)
	}
	sessionID := started["sessionId"].(string)

	heartbeat := requestJSON(t, router, http.MethodPatch, "/api/client/playback/"+sessionID, map[string]any{
		"progressSeconds": 48,
		"durationSeconds": 120,
		"status":          "playing",
	})
	session := heartbeat["session"].(map[string]any)
	if session["progressSeconds"] != float64(48) {
		t.Fatalf("expected heartbeat progress update, got %#v", session)
	}
	state := getJSON(t, router, "/api/playback/state/"+sourceID)
	if state["progressSeconds"] != float64(48) {
		t.Fatalf("expected playback state to follow heartbeat, got %#v", state)
	}

	stopped := postJSON(t, router, "/api/client/playback/"+sessionID+"/stop", map[string]any{
		"progressSeconds": 120,
		"durationSeconds": 120,
		"status":          "stopped",
	})
	stoppedSession := stopped["session"].(map[string]any)
	if stoppedSession["status"] != "stopped" {
		t.Fatalf("expected stopped session, got %#v", stoppedSession)
	}
	active := getJSON(t, router, "/api/sessions")
	if len(active["sessions"].([]any)) != 0 {
		t.Fatalf("expected no active sessions after stop, got %#v", active)
	}
}

func TestClientPlaybackStartRequiresPersistentDeviceAuthWhenProtected(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Arrival (2016)", "Arrival.2016.1080p.mkv"))

	deps := testDepsWithAuth(t, time.Now())
	deps.Config.PlaybackPolicy = "full"
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")

	scan := client.requestJSON(t, router, http.MethodPost, "/api/libraries/movies/scan", map[string]any{"path": root})
	if scan.status != http.StatusAccepted {
		t.Fatalf("expected authenticated scan start, got %d: %s", scan.status, scan.body)
	}
	waitForScan(t, router, scan.payload["id"].(string))
	sources := client.requestJSON(t, router, http.MethodGet, "/api/media-sources", nil)
	sourceID := sources.payload["mediaSources"].([]any)[0].(map[string]any)["id"].(string)
	if err := deps.Catalog.SaveProbe(context.Background(), sourceID, catalog.ProbeResult{
		Container:       "mkv",
		DurationSeconds: 120,
		Bitrate:         24_000_000,
		VideoCodec:      "h264",
		Width:           1920,
		Height:          1080,
		AudioStreams:    1,
		SubtitleStreams: 0,
		RawJSON:         "{}",
	}); err != nil {
		t.Fatalf("save probe: %v", err)
	}

	started := client.requestJSON(t, router, http.MethodPost, "/api/client/playback/start", map[string]any{
		"mediaSourceId":     sourceID,
		"deviceId":          "apple-tv-living-room",
		"clientProfile":     "apple-tv",
		"routeType":         "remote",
		"maxNetworkBitrate": 8_000_000,
	})
	if started.status != http.StatusConflict {
		t.Fatalf("expected protected native playback to block until device auth exists, got %d: %#v", started.status, started.payload)
	}
	if !strings.Contains(started.body, "persistent device authentication") {
		t.Fatalf("expected clear device auth blocker, got %s", started.body)
	}
}

func TestMigrationDryRunImportAndRollback(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Movies", "Heat (1995)", "Heat.1995.1080p.BluRay.mkv"))
	writeTestFile(t, filepath.Join(root, "TV", "The Bear", "Season 01", "The.Bear.S01E01.1080p.WEB-DL.mkv"))

	deps := testDepsWithAuth(t, time.Now())
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")

	movieScan := client.requestJSON(t, router, http.MethodPost, "/api/libraries/movies/scan", map[string]any{
		"path": filepath.Join(root, "Movies"),
	})
	if movieScan.status != http.StatusAccepted {
		t.Fatalf("movie scan start: %d %s", movieScan.status, movieScan.body)
	}
	waitForScan(t, router, movieScan.payload["id"].(string))
	tvScan := client.requestJSON(t, router, http.MethodPost, "/api/libraries/tv/scan", map[string]any{
		"path": filepath.Join(root, "TV"),
	})
	if tvScan.status != http.StatusAccepted {
		t.Fatalf("tv scan start: %d %s", tvScan.status, tvScan.body)
	}
	waitForScan(t, router, tvScan.payload["id"].(string))

	movies := client.requestJSON(t, router, http.MethodGet, "/api/movies", nil)
	series := client.requestJSON(t, router, http.MethodGet, "/api/series", nil)
	movieID := movies.payload["movies"].([]any)[0].(map[string]any)["id"].(string)
	seriesID := series.payload["series"].([]any)[0].(map[string]any)["id"].(string)
	if err := deps.Catalog.UpsertExternalID(context.Background(), catalog.ExternalID{Kind: "movie", ItemID: movieID, Provider: "tmdb", ExternalID: "949"}); err != nil {
		t.Fatalf("seed movie external id: %v", err)
	}
	if err := deps.Catalog.UpsertExternalID(context.Background(), catalog.ExternalID{Kind: "series", ItemID: seriesID, Provider: "tvdb", ExternalID: "411211"}); err != nil {
		t.Fatalf("seed series external id: %v", err)
	}

	fixture, err := os.ReadFile(filepath.Join("..", "migration", "testdata", "plex-watch-history.json"))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}

	formats := client.requestJSON(t, router, http.MethodGet, "/api/migrations/formats", nil)
	if formats.status != http.StatusOK || len(formats.payload["formats"].([]any)) == 0 {
		t.Fatalf("expected migration formats, got %d %#v", formats.status, formats.payload)
	}

	dryRun := client.requestJSON(t, router, http.MethodPost, "/api/migrations/dry-run", map[string]any{
		"payload": string(fixture),
		"scopes":  []string{migration.ScopePlayback, migration.ScopeMetadata},
		"userId":  "local",
	})
	if dryRun.status != http.StatusOK {
		t.Fatalf("dry run failed: %d %s", dryRun.status, dryRun.body)
	}
	summary := dryRun.payload["summary"].(map[string]any)
	if summary["importable"] != float64(2) || summary["conflicted"] != float64(1) {
		t.Fatalf("expected dry-run conflict classification, got %#v", dryRun.payload)
	}

	imported := client.requestJSON(t, router, http.MethodPost, "/api/migrations/import", map[string]any{
		"payload":            string(fixture),
		"scopes":             []string{migration.ScopePlayback, migration.ScopeMetadata},
		"userId":             "local",
		"selectedImportKeys": []string{"plex-heat", "plex-bear-s1e1"},
	})
	if imported.status != http.StatusOK || imported.payload["runId"] == "" {
		t.Fatalf("expected successful migration import, got %d %#v", imported.status, imported.payload)
	}
	runID := imported.payload["runId"].(string)
	run := client.requestJSON(t, router, http.MethodGet, "/api/migrations/runs/"+runID, nil)
	if run.status != http.StatusOK || run.payload["status"] != "completed" {
		t.Fatalf("expected stored migration run detail, got %d %#v", run.status, run.payload)
	}

	runs := client.requestJSON(t, router, http.MethodGet, "/api/migrations/runs", nil)
	if runs.status != http.StatusOK || len(runs.payload["runs"].([]any)) == 0 {
		t.Fatalf("expected migration runs list, got %d %#v", runs.status, runs.payload)
	}

	rollback := client.requestJSON(t, router, http.MethodPost, "/api/migrations/runs/"+runID+"/rollback", map[string]any{})
	if rollback.status != http.StatusOK || rollback.payload["status"] != "rolled_back" {
		t.Fatalf("expected rollback report, got %d %#v", rollback.status, rollback.payload)
	}
	recent := getJSON(t, router, "/api/playback/recent")
	if len(recent["recent"].([]any)) != 0 {
		t.Fatalf("expected rollback to clear imported playback state, got %#v", recent)
	}
}

func TestPlaybackDecisionEndpointIsExplicitlyDeferred(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	payload := getJSON(t, router, "/api/playback/decision?clientProfile=android-tv")

	if payload["mode"] != string(playback.DecisionDeferred) {
		t.Fatalf("expected deferred decision mode, got %#v", payload["mode"])
	}
	if payload["clientProfile"] != "android-tv" {
		t.Fatalf("expected client profile to round-trip, got %#v", payload["clientProfile"])
	}
	if payload["reasonCode"] != "source_required" || payload["decisionTraceId"] == "" {
		t.Fatalf("expected v2 decision fields, got %#v", payload)
	}
}

func TestSubtitleConversionEndpointReportsOutputBehavior(t *testing.T) {
	root := t.TempDir()
	mediaPath := filepath.Join(root, "Arrival (2016)", "Arrival.2016.1080p.mkv")
	writeTestFile(t, mediaPath)
	writeTestFile(t, filepath.Join(root, "Arrival (2016)", "Arrival.2016.1080p.en.ass"))

	router := NewRouter(testDeps(t, time.Now()))
	payload := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{
		"path":        root,
		"sampleLimit": 10,
	})
	waitForScan(t, router, payload["id"].(string))
	sources := getJSON(t, router, "/api/media-sources")
	sourceList := sources["mediaSources"].([]any)
	sourceID := sourceList[0].(map[string]any)["id"].(string)

	response := postJSON(t, router, "/api/media-sources/"+sourceID+"/subtitles/0/convert?clientProfile=web", map[string]any{})
	conversion := response["conversion"].(map[string]any)

	if conversion["status"] != "available" || conversion["outputFormat"] != "webvtt" {
		t.Fatalf("expected available WebVTT conversion, got %#v", response)
	}
	if conversion["outputBehavior"] == "" || conversion["serverImpact"] != "low" {
		t.Fatalf("expected clear conversion behavior, got %#v", conversion)
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

	movies := getJSON(t, router, "/api/movies")
	movieList := movies["movies"].([]any)
	if len(movieList) != 1 {
		t.Fatalf("expected one movie in browse API, got %#v", movieList)
	}
	movieID := movieList[0].(map[string]any)["id"].(string)
	movieDetail := getJSON(t, router, "/api/movies/"+movieID)
	if movieDetail["title"] != "Heat" {
		t.Fatalf("expected Heat movie detail, got %#v", movieDetail["title"])
	}
	metadata := getJSON(t, router, "/api/metadata/movie/"+movieID)
	records := metadata["records"].([]any)
	if len(records) != 1 || records[0].(map[string]any)["provider"] != "filename" {
		t.Fatalf("expected filename metadata record, got %#v", metadata)
	}
	match := requestJSON(t, router, http.MethodPut, "/api/metadata/match", map[string]any{
		"kind":     "movie",
		"id":       movieID,
		"title":    "Heat",
		"year":     1995,
		"provider": "manual",
		"overview": "A professional thief weighs one last score.",
	})
	if len(match["records"].([]any)) != 2 {
		t.Fatalf("expected manual and filename metadata records, got %#v", match)
	}

	series := getJSON(t, router, "/api/series")
	seriesList := series["series"].([]any)
	if len(seriesList) != 1 {
		t.Fatalf("expected one series in browse API, got %#v", seriesList)
	}
	seriesID := seriesList[0].(map[string]any)["id"].(string)
	seriesDetail := getJSON(t, router, "/api/series/"+seriesID)
	if seriesDetail["title"] != "The Bear" {
		t.Fatalf("expected The Bear series detail, got %#v", seriesDetail["title"])
	}

	sources := getJSON(t, router, "/api/media-sources?unprobed=true")
	sourceList := sources["mediaSources"].([]any)
	if len(sourceList) != 2 {
		t.Fatalf("expected two unprobed media sources, got %#v", sourceList)
	}
	sourceID := sourceList[0].(map[string]any)["id"].(string)
	sourceDetail := getJSON(t, router, "/api/media-sources/"+sourceID)
	if sourceDetail["id"] != sourceID {
		t.Fatalf("expected media source detail, got %#v", sourceDetail)
	}
	streamRequest := httptest.NewRequest(http.MethodGet, "/api/media-sources/"+sourceID+"/stream", nil)
	streamResponse := httptest.NewRecorder()
	router.ServeHTTP(streamResponse, streamRequest)
	if streamResponse.Code != http.StatusOK {
		t.Fatalf("expected stream 200, got %d: %s", streamResponse.Code, streamResponse.Body.String())
	}
	decision := getJSON(t, router, "/api/playback/decision?mediaSourceId="+sourceID+"&clientProfile=web")
	if decision["mode"] != string(playback.DecisionDeferred) {
		t.Fatalf("expected unprobed source to defer playback decision, got %#v", decision["mode"])
	}
	if decision["reasonCode"] != "probe_required" || decision["decisionTraceId"] == "" {
		t.Fatalf("expected v2 probe decision fields, got %#v", decision)
	}
}

func TestPlaybackStateAndSessions(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Heat (1995)", "Heat.1995.1080p.BluRay.mkv"))

	router := NewRouter(testDeps(t, time.Now()))
	payload := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{"path": root})
	waitForScan(t, router, payload["id"].(string))
	sources := getJSON(t, router, "/api/media-sources")
	sourceID := sources["mediaSources"].([]any)[0].(map[string]any)["id"].(string)

	state := requestJSON(t, router, http.MethodPut, "/api/playback/state/"+sourceID, map[string]any{
		"progressSeconds": 91,
		"durationSeconds": 100,
	})
	if state["watched"] != true {
		t.Fatalf("expected 90 percent progress to mark watched, got %#v", state)
	}
	recent := getJSON(t, router, "/api/playback/recent")
	if len(recent["recent"].([]any)) != 1 {
		t.Fatalf("expected recent playback item, got %#v", recent)
	}

	session := postJSON(t, router, "/api/sessions", map[string]any{"mediaSourceId": sourceID, "deviceId": "web"})
	sessionID := session["id"].(string)
	updated := requestJSON(t, router, http.MethodPatch, "/api/sessions/"+sessionID, map[string]any{
		"progressSeconds": 12,
		"durationSeconds": 100,
	})
	if updated["progressSeconds"] != float64(12) {
		t.Fatalf("expected session progress update, got %#v", updated)
	}
	active := getJSON(t, router, "/api/sessions")
	if len(active["sessions"].([]any)) != 1 {
		t.Fatalf("expected one active session, got %#v", active)
	}

	download := postJSON(t, router, "/api/downloads", map[string]any{"mediaSourceId": sourceID, "targetProfile": "original"})
	if download["status"] != string(downloads.StatusCompleted) {
		t.Fatalf("expected original download to complete immediately, got %#v", download)
	}
	downloadsPayload := getJSON(t, router, "/api/downloads")
	if len(downloadsPayload["downloads"].([]any)) != 1 {
		t.Fatalf("expected one download job, got %#v", downloadsPayload)
	}
}

func TestSessionInspectorEndpointExposesPlaybackFields(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	started := postJSON(t, router, "/api/sessions", map[string]any{
		"mediaSourceId":  "source-1",
		"deviceId":       "web",
		"clientProfile":  "web",
		"route":          "direct",
		"mode":           "Direct Play",
		"reasonCode":     "direct_play_supported",
		"reasonText":     "Direct playback is available.",
		"serverImpact":   "Low impact",
		"selectedTracks": map[string]string{"audio": "Default", "subtitles": "Off"},
	})
	id := started["id"].(string)
	updated := requestJSON(t, router, http.MethodPatch, "/api/sessions/"+id, map[string]any{
		"route":          "transcode",
		"mode":           "Video Transcode",
		"reasonCode":     "video_conversion_required",
		"reasonText":     "Video conversion required.",
		"serverImpact":   "High server load",
		"selectedTracks": map[string]string{"audio": "Default", "subtitles": "English PGS"},
	})
	if updated["route"] != "transcode" {
		t.Fatalf("expected route update, got %#v", updated)
	}

	inspector := getJSON(t, router, "/api/sessions/"+id+"/inspector")
	if inspector["route"] != "transcode" || inspector["reasonCode"] != "video_conversion_required" {
		t.Fatalf("expected inspector route/reason, got %#v", inspector)
	}
	tracks := inspector["selectedTracks"].(map[string]any)
	if tracks["subtitles"] != "English PGS" {
		t.Fatalf("expected selected subtitle track, got %#v", tracks)
	}
	history := inspector["routeHistory"].([]any)
	if len(history) != 1 {
		t.Fatalf("expected route history, got %#v", inspector)
	}
}

func TestMetadataMatchResolvesReview(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Unclear.File.Name.mkv"))

	router := NewRouter(testDeps(t, time.Now()))
	payload := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{"path": root})
	waitForScan(t, router, payload["id"].(string))
	review := getJSON(t, router, "/api/review")
	items := review["items"].([]any)
	if len(items) == 0 {
		t.Fatalf("expected review item")
	}
	item := items[0].(map[string]any)
	requestJSON(t, router, http.MethodPut, "/api/metadata/match", map[string]any{
		"kind":  item["kind"],
		"id":    item["id"],
		"title": "Fixed Title",
		"year":  2026,
	})

	review = getJSON(t, router, "/api/review")
	if len(review["items"].([]any)) != 0 {
		t.Fatalf("expected review queue to be resolved, got %#v", review)
	}
}

func TestLibraryManagementRemoteAndArtworkEndpoints(t *testing.T) {
	root := t.TempDir()
	router := NewRouter(testDeps(t, time.Now()))

	library := postJSON(t, router, "/api/libraries", map[string]any{
		"kind": "movies",
		"name": "Archive Movies",
		"path": root,
	})
	if library["name"] != "Archive Movies" {
		t.Fatalf("expected saved library, got %#v", library)
	}
	libraryID := library["id"].(string)
	librariesPayload := getJSON(t, router, "/api/libraries")
	if len(librariesPayload["libraries"].([]any)) != 1 {
		t.Fatalf("expected one saved library, got %#v", librariesPayload)
	}
	scan := postJSON(t, router, "/api/libraries/"+libraryID+"/scan", map[string]any{})
	waitForScan(t, router, scan["id"].(string))

	remote := getJSON(t, router, "/api/remote/access")
	if remote["wanLookup"] != "available_on_request" {
		t.Fatalf("expected explicit wan lookup flag, got %#v", remote)
	}
	if remote["diagnostics"] != "available" {
		t.Fatalf("expected diagnostics to be advertised, got %#v", remote)
	}
	request := httptest.NewRequest(http.MethodGet, "/api/artwork/movie/example", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("expected svg artwork response, got %d %q", response.Code, response.Header().Get("Content-Type"))
	}
}

func TestRemoteDiagnosticsEndpointRejectsSensitiveURLParts(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	result := postJSON(t, router, "/api/remote/diagnostics", map[string]any{
		"publicUrl": "https://media.example.com/watch?token=secret",
	})
	if result["failureClass"] != "invalid_input" {
		t.Fatalf("expected invalid input without leaking url details, got %#v", result)
	}
	target := result["target"].(map[string]any)
	if len(target) != 0 {
		t.Fatalf("expected empty sanitized target for invalid input, got %#v", target)
	}
}

func TestAdaptiveStreamingPlanManifestAndTelemetry(t *testing.T) {
	root := t.TempDir()
	writeTestFile(t, filepath.Join(root, "Big Movie (2026)", "Big.Movie.2026.2160p.mkv"))

	deps := testDeps(t, time.Now())
	deps.Config.PlaybackPolicy = "full"
	router := NewRouter(deps)
	payload := postJSON(t, router, "/api/libraries/movies/scan", map[string]any{"path": root})
	waitForScan(t, router, payload["id"].(string))
	sources := getJSON(t, router, "/api/media-sources")
	sourceID := sources["mediaSources"].([]any)[0].(map[string]any)["id"].(string)
	if err := deps.Catalog.SaveProbe(context.Background(), sourceID, catalog.ProbeResult{
		Container:       "matroska",
		DurationSeconds: 7200,
		Bitrate:         61_000_000,
		VideoCodec:      "hevc",
		Width:           3840,
		Height:          2160,
		AudioStreams:    1,
		RawJSON:         `{"streams":[]}`,
	}); err != nil {
		t.Fatalf("save probe: %v", err)
	}

	decision := getJSON(t, router, "/api/playback/decision?mediaSourceId="+sourceID+"&clientProfile=web&routeType=remote&maxNetworkBitrate=10000000")
	if decision["mode"] != string(playback.AdaptiveStream) {
		t.Fatalf("expected adaptive decision, got %#v", decision)
	}
	route := getJSON(t, router, "/api/playback/route?mediaSourceId="+sourceID+"&clientProfile=web&routeType=remote&maxNetworkBitrate=10000000")
	if route["route"] != "adaptive" || route["manifestUrl"] == "" {
		t.Fatalf("expected adaptive route, got %#v", route)
	}
	session := postJSON(t, router, "/api/media-sources/"+sourceID+"/adaptive/session", map[string]any{
		"clientProfile":     "web",
		"routeType":         "remote",
		"maxNetworkBitrate": 10_000_000,
	})
	plan := session["plan"].(map[string]any)
	if plan["enabled"] != true {
		t.Fatalf("expected enabled adaptive plan, got %#v", session)
	}

	request := httptest.NewRequest(http.MethodGet, "/api/media-sources/"+sourceID+"/adaptive/master.m3u8?clientProfile=web&routeType=remote&maxNetworkBitrate=10000000", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "#EXT-X-STREAM-INF") {
		t.Fatalf("expected hls master playlist, got %d: %s", response.Code, response.Body.String())
	}
	request = httptest.NewRequest(http.MethodGet, "/api/media-sources/"+sourceID+"/adaptive/variant-720p.m3u8?clientProfile=web&routeType=remote&maxNetworkBitrate=10000000", nil)
	response = httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || !strings.Contains(response.Body.String(), "#EXT-X-TARGETDURATION") {
		t.Fatalf("expected hls variant playlist, got %d: %s", response.Code, response.Body.String())
	}

	telemetry := requestJSON(t, router, http.MethodPost, "/api/adaptive/telemetry", map[string]any{
		"sessionId":       "sess_1",
		"mediaSourceId":   sourceID,
		"clientProfile":   "web",
		"event":           "stall",
		"variantId":       "720p",
		"bufferSeconds":   0.4,
		"stallMs":         1200,
		"observedBitrate": 3_800_000,
	})
	if telemetry["event"] != "adaptive.stall" || telemetry["correlationId"] == "" {
		t.Fatalf("expected correlated adaptive telemetry, got %#v", telemetry)
	}
}

func TestAuthProtectedRouteRequiresSession(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	request := httptest.NewRequest(http.MethodGet, "/api/sessions", nil)
	response := httptest.NewRecorder()

	router.ServeHTTP(response, request)

	if response.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d: %s", response.Code, response.Body.String())
	}
}

func TestAuthLoginAndProtectedRouteAccess(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	client := newAuthTestClient(t)

	login := client.requestJSON(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "admin",
		"password": "test-password-123!",
	})
	if login.status != http.StatusOK {
		t.Fatalf("expected login 200, got %d: %s", login.status, login.body)
	}
	if login.payload["user"].(map[string]any)["username"] != "admin" {
		t.Fatalf("expected admin user, got %#v", login.payload)
	}
	if client.csrfToken == "" {
		t.Fatalf("expected csrf token to be captured")
	}

	protected := client.requestJSON(t, router, http.MethodGet, "/api/sessions", nil)
	if protected.status != http.StatusOK {
		t.Fatalf("expected protected route 200, got %d: %s", protected.status, protected.body)
	}

	session := client.requestJSON(t, router, http.MethodGet, "/api/auth/session", nil)
	if session.status != http.StatusOK {
		t.Fatalf("expected session route 200, got %d: %s", session.status, session.body)
	}
	if session.payload["csrfToken"] == "" {
		t.Fatalf("expected csrf token in session payload, got %#v", session.payload)
	}
}

func TestAuthRevokedSessionDenied(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	client := newAuthTestClient(t)

	login := client.requestJSON(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "admin",
		"password": "test-password-123!",
	})
	if login.status != http.StatusOK {
		t.Fatalf("expected login 200, got %d: %s", login.status, login.body)
	}
	logout := client.requestJSON(t, router, http.MethodPost, "/api/auth/logout", map[string]any{})
	if logout.status != http.StatusOK {
		t.Fatalf("expected logout 200, got %d: %s", logout.status, logout.body)
	}

	protected := client.requestJSON(t, router, http.MethodGet, "/api/sessions", nil)
	if protected.status != http.StatusUnauthorized {
		t.Fatalf("expected revoked session to get 401, got %d: %s", protected.status, protected.body)
	}
}

func TestAuthInvalidLoginsTriggerLockout(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	client := newAuthTestClient(t)

	for attempt := 0; attempt < 4; attempt++ {
		response := client.requestJSON(t, router, http.MethodPost, "/api/auth/login", map[string]any{
			"username": "admin",
			"password": "wrong-password",
		})
		if response.status != http.StatusUnauthorized {
			t.Fatalf("attempt %d expected 401, got %d: %s", attempt+1, response.status, response.body)
		}
	}

	locked := client.requestJSON(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "admin",
		"password": "wrong-password",
	})
	if locked.status != http.StatusTooManyRequests {
		t.Fatalf("expected 429 on lockout, got %d: %s", locked.status, locked.body)
	}
	if locked.payload["lockedUntil"] == "" {
		t.Fatalf("expected lockedUntil in response, got %#v", locked.payload)
	}
}

func TestAuthMutationRejectsMissingCSRF(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	client := newAuthTestClient(t)

	login := client.requestJSON(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"username": "admin",
		"password": "test-password-123!",
	})
	if login.status != http.StatusOK {
		t.Fatalf("expected login 200, got %d: %s", login.status, login.body)
	}

	request := httptest.NewRequest(http.MethodPut, "http://vyrden.test/api/settings", bytes.NewReader(mustJSON(t, map[string]any{
		"httpAddr": "127.0.0.1:8097",
	})))
	request.Header.Set("Content-Type", "application/json")
	client.apply(request, false)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)

	if response.Code != http.StatusForbidden {
		t.Fatalf("expected 403 without csrf header, got %d: %s", response.Code, response.Body.String())
	}
}

func TestAuthzStandardCannotCallAdminSettingsRoutes(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	if _, err := deps.Auth.CreateUser(context.Background(), "viewer", "viewer-password-123!", "Viewer", "standard"); err != nil {
		t.Fatalf("create standard user: %v", err)
	}
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "viewer", "viewer-password-123!")

	response := client.requestJSON(t, router, http.MethodPut, "/api/settings", map[string]any{
		"httpAddr": "127.0.0.1:8097",
	})

	if response.status != http.StatusForbidden {
		t.Fatalf("expected standard user to get 403 on settings, got %d: %s", response.status, response.body)
	}

	browse := client.requestJSON(t, router, http.MethodGet, "/api/settings/folders/browse?path="+url.QueryEscape(t.TempDir()), nil)
	if browse.status != http.StatusForbidden {
		t.Fatalf("expected standard user to get 403 on folder browse, got %d: %s", browse.status, browse.body)
	}
}

func TestAuthzAdminCanCallAdminSettingsRoutes(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")

	response := client.requestJSON(t, router, http.MethodPut, "/api/settings", map[string]any{
		"httpAddr": "127.0.0.1:8097",
	})

	if response.status != http.StatusOK {
		t.Fatalf("expected admin settings update 200, got %d: %s", response.status, response.body)
	}
}

func TestSettingsFolderBrowseListsDirectories(t *testing.T) {
	root := t.TempDir()
	if err := os.Mkdir(filepath.Join(root, "Movies"), 0o755); err != nil {
		t.Fatalf("mkdir movies: %v", err)
	}
	if err := os.WriteFile(filepath.Join(root, "readme.txt"), []byte("not a folder"), 0o644); err != nil {
		t.Fatalf("write file: %v", err)
	}
	router := NewRouter(testDeps(t, time.Now()))

	response := requestJSON(t, router, http.MethodGet, "/api/settings/folders/browse?path="+url.QueryEscape(root), nil)
	if response["path"] == "" || response["parent"] == "" {
		t.Fatalf("expected current path and parent, got %#v", response)
	}
	entries := response["entries"].([]any)
	if len(entries) != 1 || entries[0].(map[string]any)["name"] != "Movies" {
		t.Fatalf("expected only child folders, got %#v", response)
	}
	if response["writable"] != true {
		t.Fatalf("expected writable temp folder, got %#v", response)
	}
}

func TestSettingsRuntimePathsReflectSavedValuesBeforeRestart(t *testing.T) {
	deps := testDeps(t, time.Now())
	router := NewRouter(deps)
	nextData := t.TempDir()
	nextTranscode := filepath.Join(nextData, "transcode")

	update := requestJSON(t, router, http.MethodPut, "/api/settings", map[string]any{
		"dataDir":      nextData,
		"transcodeDir": nextTranscode,
	})
	updatedPaths := update["runtimePaths"].(map[string]any)
	if updatedPaths["data"] != nextData || updatedPaths["transcode"] != nextTranscode {
		t.Fatalf("expected update response to include saved paths, got %#v", update)
	}

	reloaded := getJSON(t, router, "/api/settings")
	reloadedPaths := reloaded["runtimePaths"].(map[string]any)
	if reloadedPaths["data"] != nextData || reloadedPaths["transcode"] != nextTranscode {
		t.Fatalf("expected settings reload to keep saved paths before restart, got %#v", reloaded)
	}
	status := getJSON(t, router, "/api/system/status")
	disks := status["disks"].([]any)
	found := false
	for _, item := range disks {
		disk := item.(map[string]any)
		if disk["name"] == "data" && disk["path"] == nextData {
			found = true
		}
	}
	if !found {
		t.Fatalf("expected system status to use saved data dir before restart, got %#v", status)
	}
}

func TestAuthzProtectedMediaAndDownloadRoutes(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	if _, err := deps.Auth.CreateUser(context.Background(), "viewer", "viewer-password-123!", "Viewer", "standard"); err != nil {
		t.Fatalf("create standard user: %v", err)
	}
	router := NewRouter(deps)

	unauth := httptest.NewRequest(http.MethodGet, "/api/media-sources/missing/stream", nil)
	unauthResponse := httptest.NewRecorder()
	router.ServeHTTP(unauthResponse, unauth)
	if unauthResponse.Code != http.StatusUnauthorized {
		t.Fatalf("expected unauthenticated stream 401, got %d: %s", unauthResponse.Code, unauthResponse.Body.String())
	}

	viewer := newAuthTestClient(t)
	loginAs(t, viewer, router, "viewer", "viewer-password-123!")
	sessionID := startPlaybackSession(t, viewer, router, "missing", "web")
	token := viewer.requestJSON(t, router, http.MethodPost, "/api/media-sources/missing/stream-token", map[string]any{
		"sessionId": sessionID,
		"deviceId":  "web",
	})
	if token.status != http.StatusOK {
		t.Fatalf("expected standard stream token 200, got %d: %s", token.status, token.body)
	}
	streamURL, _ := token.payload["streamUrl"].(string)
	stream := viewer.requestJSON(t, router, http.MethodGet, streamURL, nil)
	if stream.status != http.StatusNotFound {
		t.Fatalf("expected standard stream request to reach handler and return 404, got %d: %s", stream.status, stream.body)
	}
	download := viewer.requestJSON(t, router, http.MethodPost, "/api/downloads", map[string]any{"mediaSourceId": "missing", "targetProfile": "original"})
	if download.status != http.StatusForbidden {
		t.Fatalf("expected standard download creation 403, got %d: %s", download.status, download.body)
	}
}

func TestAuthzAuditEventsIncludeActorAndAction(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	eventsCh, stop := deps.Events.Subscribe(ctx)
	t.Cleanup(func() {
		stop()
		cancel()
	})
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")

	response := client.requestJSON(t, router, http.MethodPut, "/api/settings", map[string]any{
		"httpAddr": "127.0.0.1:8097",
	})
	if response.status != http.StatusOK {
		t.Fatalf("expected settings update 200, got %d: %s", response.status, response.body)
	}

	deadline := time.After(2 * time.Second)
	for {
		select {
		case event := <-eventsCh:
			if event.Type != "audit.route" {
				continue
			}
			data := event.Data.(map[string]any)
			if data["userId"] == "" || data["action"] != "settings.update" || data["result"] != "allowed" {
				t.Fatalf("unexpected audit data: %#v", data)
			}
			return
		case <-deadline:
			t.Fatalf("timed out waiting for audit route event")
		}
	}
}

func TestSensitiveActionsEmitDomainAuditEvents(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	ctx, cancel := context.WithCancel(context.Background())
	eventsCh, stop := deps.Events.Subscribe(ctx)
	t.Cleanup(func() {
		stop()
		cancel()
	})
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")

	settingsResponse := client.requestJSON(t, router, http.MethodPut, "/api/settings", map[string]any{
		"httpAddr": "127.0.0.1:8097",
	})
	if settingsResponse.status != http.StatusOK {
		t.Fatalf("expected settings update 200, got %d: %s", settingsResponse.status, settingsResponse.body)
	}
	libraryResponse := client.requestJSON(t, router, http.MethodPost, "/api/libraries", map[string]any{
		"name":        "Movies",
		"kind":        "movies",
		"path":        t.TempDir(),
		"storageType": "local",
	})
	if libraryResponse.status != http.StatusOK {
		t.Fatalf("expected library save 200, got %d: %s", libraryResponse.status, libraryResponse.body)
	}

	seen := map[string]bool{}
	deadline := time.After(2 * time.Second)
	for len(seen) < 3 {
		select {
		case event := <-eventsCh:
			data, _ := event.Data.(map[string]any)
			switch event.Type {
			case "audit.auth":
				if data["action"] == "login" && data["result"] == "allowed" && data["userId"] != "" {
					seen[event.Type] = true
				}
			case "audit.settings":
				if data["action"] == "settings.update" && data["result"] == "allowed" && data["userId"] != "" {
					seen[event.Type] = true
				}
			case "audit.library":
				if data["action"] == "library.save" && data["result"] == "allowed" && data["libraryId"] != "" {
					seen[event.Type] = true
				}
			}
		case <-deadline:
			t.Fatalf("timed out waiting for domain audit events, saw %#v", seen)
		}
	}
}

func TestSignedStreamWithoutTokenIsDenied(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")
	sessionID := startPlaybackSession(t, client, router, "source_1", "web")

	response := client.requestJSON(t, router, http.MethodGet, "/api/media-sources/source_1/stream?sessionId="+sessionID+"&deviceId=web", nil)

	if response.status != http.StatusUnauthorized {
		t.Fatalf("expected missing stream token 401, got %d: %s", response.status, response.body)
	}
}

func TestSignedStreamExpiredTokenIsDenied(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")
	sessionID := startPlaybackSession(t, client, router, "source_1", "web")
	token, _, err := deps.Streaming.Issue(streaming.Expected{MediaSourceID: "source_1", SessionID: sessionID, UserID: "admin", DeviceID: "web"}, -time.Minute)
	if err != nil {
		t.Fatalf("issue expired token: %v", err)
	}

	response := client.requestJSON(t, router, http.MethodGet, "/api/media-sources/source_1/stream?sessionId="+sessionID+"&deviceId=web&token="+url.QueryEscape(token), nil)

	if response.status != http.StatusForbidden || !strings.Contains(response.body, "expired") {
		t.Fatalf("expected expired stream token 403, got %d: %s", response.status, response.body)
	}
}

func TestSignedStreamForgedTokenIsDenied(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")
	sessionID := startPlaybackSession(t, client, router, "source_1", "web")
	token, _, err := deps.Streaming.Issue(streaming.Expected{MediaSourceID: "source_1", SessionID: sessionID, UserID: "admin", DeviceID: "web"}, 0)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	response := client.requestJSON(t, router, http.MethodGet, "/api/media-sources/source_1/stream?sessionId="+sessionID+"&deviceId=web&token="+url.QueryEscape(token+"x"), nil)

	if response.status != http.StatusForbidden || !strings.Contains(response.body, "invalid") {
		t.Fatalf("expected forged stream token 403, got %d: %s", response.status, response.body)
	}
}

func TestSignedStreamTokenCannotMoveAcrossSessions(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")
	firstSessionID := startPlaybackSession(t, client, router, "source_1", "web")
	secondSessionID := startPlaybackSession(t, client, router, "source_1", "web")
	token, _, err := deps.Streaming.Issue(streaming.Expected{MediaSourceID: "source_1", SessionID: firstSessionID, UserID: "admin", DeviceID: "web"}, 0)
	if err != nil {
		t.Fatalf("issue token: %v", err)
	}

	response := client.requestJSON(t, router, http.MethodGet, "/api/media-sources/source_1/stream?sessionId="+secondSessionID+"&deviceId=web&token="+url.QueryEscape(token), nil)

	if response.status != http.StatusForbidden || !strings.Contains(response.body, "session") {
		t.Fatalf("expected cross-session token 403, got %d: %s", response.status, response.body)
	}
}

func TestSignedStreamLimitBlocksExcessStreams(t *testing.T) {
	deps := testDepsWithAuth(t, time.Now())
	deps.Streaming.SetLimits(1, 1)
	router := NewRouter(deps)
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")
	sessionID := startPlaybackSession(t, client, router, "source_1", "web")
	expected := streaming.Expected{MediaSourceID: "source_1", SessionID: sessionID, UserID: "admin", DeviceID: "web"}
	firstToken, _, err := deps.Streaming.Issue(expected, 0)
	if err != nil {
		t.Fatalf("issue first token: %v", err)
	}
	_, release, err := deps.Streaming.Validate(firstToken, expected)
	if err != nil {
		t.Fatalf("hold first stream: %v", err)
	}
	defer release()
	secondToken, _, err := deps.Streaming.Issue(expected, 0)
	if err != nil {
		t.Fatalf("issue second token: %v", err)
	}

	response := client.requestJSON(t, router, http.MethodGet, "/api/media-sources/source_1/stream?sessionId="+sessionID+"&deviceId=web&token="+url.QueryEscape(secondToken), nil)

	if response.status != http.StatusTooManyRequests {
		t.Fatalf("expected stream limit 429, got %d: %s", response.status, response.body)
	}
}

func TestSignedStreamTokenEndpointReturnsBoundURL(t *testing.T) {
	router := NewRouter(testDepsWithAuth(t, time.Now()))
	client := newAuthTestClient(t)
	loginAs(t, client, router, "admin", "test-password-123!")
	sessionID := startPlaybackSession(t, client, router, "source_1", "web")

	token := client.requestJSON(t, router, http.MethodPost, "/api/media-sources/source_1/stream-token", map[string]any{
		"sessionId": sessionID,
		"deviceId":  "web",
	})

	if token.status != http.StatusOK || token.payload["streamUrl"] == "" || token.payload["token"] == "" {
		t.Fatalf("expected bound stream token response, got %d: %#v", token.status, token.payload)
	}
}

func testDeps(t *testing.T, startedAt time.Time) Deps {
	t.Helper()

	dataDir := t.TempDir()
	cfg := config.Config{
		HTTPAddr:         "127.0.0.1:8097",
		DataDir:          dataDir,
		AuthDisabled:     true,
		TranscodeDir:     filepath.Join(dataDir, "transcode"),
		DownloadsDir:     filepath.Join(dataDir, "downloads"),
		MetadataDir:      filepath.Join(dataDir, "metadata"),
		CacheDir:         filepath.Join(dataDir, "cache"),
		TempDir:          filepath.Join(dataDir, "temp"),
		EventBuffer:      64,
		ScanWorkers:      1,
		ProbeWorkers:     2,
		TranscodeWorkers: 1,
		GPUWorkers:       1,
	}
	for _, dir := range []string{cfg.TranscodeDir, cfg.DownloadsDir, cfg.MetadataDir, cfg.CacheDir, cfg.TempDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("create runtime dir: %v", err)
		}
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
	observe := observability.NewService()
	observe.Subscribe(ctx, eventBus)
	return Deps{
		Config:    cfg,
		StartedAt: startedAt,
		Events:    eventBus,
		Observe:   observe,
		Resources: manager,
		Jobs:      registry,
		Libraries: libraryService,
		Scanner:   scannerService,
		Scans:     scans.NewService(cfg, eventBus, registry.Scan, libraryService, scannerService, catalogService, metaprovider.NewService(cfg, catalogService, eventBus), movieService, tvService),
		Catalog:   catalogService,
		Media:     media.NewService(),
		Metadata:  metaprovider.NewService(cfg, catalogService, eventBus),
		Movies:    movieService,
		TV:        tvService,
		Probe:     probe.NewService("ffprobe"),
		Probes:    probes.NewService(eventBus, registry.Probe, catalogService, probe.NewService("ffprobe")),
		Playback:  playback.NewService(),
		PlayState: playstate.NewService(db, eventBus),
		Streaming: streaming.NewServiceWithKey("test", []byte("01234567890123456789012345678901")),
		Transcode: transcode.NewService(eventBus, registry.Transcode, "ffmpeg", filepath.Join(t.TempDir(), "transcode")),
		Downloads: downloads.NewService(eventBus, registry.Transcode, "ffmpeg", filepath.Join(t.TempDir(), "downloads")),
		Devices:   devices.NewService(),
		Sessions:  sessions.NewService(eventBus),
		Subtitles: subtitles.NewService(),
		Pairing:   pairing.NewService(),
		Migration: migration.NewService(db, eventBus),
	}
}

func testDepsWithAuth(t *testing.T, startedAt time.Time) Deps {
	t.Helper()

	deps := testDeps(t, startedAt)
	deps.Config.AuthDisabled = false
	db, err := database.Open(context.Background(), t.TempDir())
	if err != nil {
		t.Fatalf("open auth database: %v", err)
	}
	t.Cleanup(func() {
		if err := db.Close(); err != nil {
			t.Fatalf("close auth database: %v", err)
		}
	})
	deps.Catalog = catalog.NewService(db)
	deps.Metadata = metaprovider.NewService(deps.Config, deps.Catalog, deps.Events)
	deps.Scans = scans.NewService(deps.Config, deps.Events, deps.Jobs.Scan, deps.Libraries, deps.Scanner, deps.Catalog, deps.Metadata, deps.Movies, deps.TV)
	deps.Probes = probes.NewService(deps.Events, deps.Jobs.Probe, deps.Catalog, probe.NewService("ffprobe"))
	deps.PlayState = playstate.NewService(db, deps.Events)
	deps.Migration = migration.NewService(db, deps.Events)
	deps.Auth = auth.NewService(db, false)
	if err := deps.Auth.Bootstrap(context.Background(), auth.BootstrapOptions{
		Username: "admin",
		Password: "test-password-123!",
	}); err != nil {
		t.Fatalf("bootstrap auth: %v", err)
	}
	return deps
}

type authTestClient struct {
	jar       http.CookieJar
	csrfToken string
}

type authJSONResponse struct {
	status  int
	payload map[string]any
	body    string
}

func newAuthTestClient(t *testing.T) *authTestClient {
	t.Helper()
	jar, err := cookiejar.New(nil)
	if err != nil {
		t.Fatalf("create cookie jar: %v", err)
	}
	return &authTestClient{jar: jar}
}

func (c *authTestClient) requestJSON(t *testing.T, router http.Handler, method string, path string, body any) authJSONResponse {
	t.Helper()

	var reader io.Reader
	if body != nil {
		reader = bytes.NewReader(mustJSON(t, body))
	}
	request := httptest.NewRequest(method, "http://vyrden.test"+path, reader)
	if body != nil {
		request.Header.Set("Content-Type", "application/json")
	}
	c.apply(request, method != http.MethodGet && method != http.MethodHead)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	c.store(request, response)

	result := authJSONResponse{status: response.Code, body: response.Body.String()}
	if response.Body.Len() > 0 {
		_ = json.Unmarshal(response.Body.Bytes(), &result.payload)
	}
	return result
}

func (c *authTestClient) apply(request *http.Request, includeCSRF bool) {
	for _, cookie := range c.jar.Cookies(request.URL) {
		request.AddCookie(cookie)
	}
	if includeCSRF && c.csrfToken != "" {
		request.Header.Set("X-CSRF-Token", c.csrfToken)
	}
}

func (c *authTestClient) store(request *http.Request, response *httptest.ResponseRecorder) {
	c.jar.SetCookies(request.URL, response.Result().Cookies())
	for _, cookie := range response.Result().Cookies() {
		if cookie.Name == auth.CSRFCookieName {
			c.csrfToken = cookie.Value
		}
	}
}

func loginAs(t *testing.T, client *authTestClient, router http.Handler, username string, password string) {
	t.Helper()
	response := client.requestJSON(t, router, http.MethodPost, "/api/auth/login", map[string]any{
		"username": username,
		"password": password,
	})
	if response.status != http.StatusOK {
		t.Fatalf("login %s failed with %d: %s", username, response.status, response.body)
	}
}

func startPlaybackSession(t *testing.T, client *authTestClient, router http.Handler, mediaSourceID string, deviceID string) string {
	t.Helper()
	response := client.requestJSON(t, router, http.MethodPost, "/api/sessions", map[string]any{
		"mediaSourceId": mediaSourceID,
		"deviceId":      deviceID,
		"clientProfile": "web",
		"mode":          "direct",
		"route":         "direct",
	})
	if response.status != http.StatusAccepted {
		t.Fatalf("start session failed with %d: %s", response.status, response.body)
	}
	sessionID, _ := response.payload["id"].(string)
	if sessionID == "" {
		t.Fatalf("expected session id, got %#v", response.payload)
	}
	return sessionID
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

func decodeBody(t *testing.T, body []byte) map[string]any {
	t.Helper()
	var payload map[string]any
	if err := json.Unmarshal(body, &payload); err != nil {
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
	return requestJSON(t, router, http.MethodPost, path, body)
}

func requestJSON(t *testing.T, router http.Handler, method string, path string, body any) map[string]any {
	t.Helper()

	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	request := httptest.NewRequest(method, path, bytes.NewReader(raw))
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

func mustJSON(t *testing.T, value any) []byte {
	t.Helper()
	raw, err := json.Marshal(value)
	if err != nil {
		t.Fatalf("marshal json: %v", err)
	}
	return raw
}
