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
	"github.com/vyrdenhq/vyrden/server/internal/movies"
	"github.com/vyrdenhq/vyrden/server/internal/playback"
	"github.com/vyrdenhq/vyrden/server/internal/playstate"
	"github.com/vyrdenhq/vyrden/server/internal/probe"
	"github.com/vyrdenhq/vyrden/server/internal/probes"
	"github.com/vyrdenhq/vyrden/server/internal/resources"
	"github.com/vyrdenhq/vyrden/server/internal/scanner"
	"github.com/vyrdenhq/vyrden/server/internal/scans"
	"github.com/vyrdenhq/vyrden/server/internal/sessions"
	"github.com/vyrdenhq/vyrden/server/internal/streaming"
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

func TestPlaybackDecisionEndpointIsExplicitlyDeferred(t *testing.T) {
	router := NewRouter(testDeps(t, time.Now()))
	payload := getJSON(t, router, "/api/playback/decision?clientProfile=android-tv")

	if payload["mode"] != string(playback.DecisionDeferred) {
		t.Fatalf("expected deferred decision mode, got %#v", payload["mode"])
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
	request := httptest.NewRequest(http.MethodGet, "/api/artwork/movie/example", nil)
	response := httptest.NewRecorder()
	router.ServeHTTP(response, request)
	if response.Code != http.StatusOK || response.Header().Get("Content-Type") != "image/svg+xml" {
		t.Fatalf("expected svg artwork response, got %d %q", response.Code, response.Header().Get("Content-Type"))
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
