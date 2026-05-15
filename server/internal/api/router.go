package api

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"html"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"log/slog"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strconv"
	"strings"
	"time"

	_ "golang.org/x/image/webp"

	"github.com/jampat000/Xuva/server/internal/adaptive"
	"github.com/jampat000/Xuva/server/internal/auth"
	"github.com/jampat000/Xuva/server/internal/catalog"
	"github.com/jampat000/Xuva/server/internal/config"
	"github.com/jampat000/Xuva/server/internal/devices"
	"github.com/jampat000/Xuva/server/internal/discovery"
	"github.com/jampat000/Xuva/server/internal/downloads"
	"github.com/jampat000/Xuva/server/internal/events"
	"github.com/jampat000/Xuva/server/internal/jobs"
	"github.com/jampat000/Xuva/server/internal/libraries"
	"github.com/jampat000/Xuva/server/internal/media"
	metaprovider "github.com/jampat000/Xuva/server/internal/metadata"
	"github.com/jampat000/Xuva/server/internal/metasources"
	"github.com/jampat000/Xuva/server/internal/migration"
	"github.com/jampat000/Xuva/server/internal/movies"
	"github.com/jampat000/Xuva/server/internal/observability"
	"github.com/jampat000/Xuva/server/internal/pairing"
	"github.com/jampat000/Xuva/server/internal/playback"
	"github.com/jampat000/Xuva/server/internal/playstate"
	"github.com/jampat000/Xuva/server/internal/probe"
	"github.com/jampat000/Xuva/server/internal/probes"
	"github.com/jampat000/Xuva/server/internal/remote"
	"github.com/jampat000/Xuva/server/internal/resources"
	"github.com/jampat000/Xuva/server/internal/scanner"
	"github.com/jampat000/Xuva/server/internal/scans"
	"github.com/jampat000/Xuva/server/internal/sessions"
	"github.com/jampat000/Xuva/server/internal/streaming"
	"github.com/jampat000/Xuva/server/internal/subtitles"
	"github.com/jampat000/Xuva/server/internal/systemstats"
	"github.com/jampat000/Xuva/server/internal/transcode"
	"github.com/jampat000/Xuva/server/internal/tv"
	"github.com/jampat000/Xuva/server/internal/webapp"
)

type Deps struct {
	Config    config.Config
	StartedAt time.Time
	Auth      *auth.Service
	Events    *events.Bus
	Observe   *observability.Service
	Resources *resources.Manager
	Jobs      *jobs.Registry
	Discovery *discovery.Service
	Libraries *libraries.Service
	Scanner   *scanner.Service
	Scans     *scans.Service
	Catalog   *catalog.Service
	Media     *media.Service
	Metadata  *metaprovider.Service
	Movies    *movies.Service
	TV        *tv.Service
	Probe     *probe.Service
	Probes    *probes.Service
	Playback  *playback.Service
	PlayState *playstate.Service
	Streaming *streaming.Service
	Transcode *transcode.Service
	Downloads *downloads.Service
	Devices   *devices.Service
	Sessions  *sessions.Service
	Subtitles *subtitles.Service
	Pairing   *pairing.Service
	Migration *migration.Service
}

func NewRouter(deps Deps) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/health", healthHandler(deps))
	mux.HandleFunc("GET /api/ready", readinessHandler(deps))
	mux.HandleFunc("GET /api/metrics", metricsHandler(deps))
	mux.HandleFunc("GET /api/client/bootstrap", clientBootstrapHandler(deps))
	mux.HandleFunc("GET /api/discovery/status", discoveryStatusHandler(deps))
	mux.HandleFunc("POST /api/pairing/requests", pairingCreateHandler(deps))
	mux.HandleFunc("GET /api/pairing/requests/{id}", pairingStatusHandler(deps))
	mux.HandleFunc("POST /api/auth/bootstrap", authBootstrapHandler(deps))
	mux.HandleFunc("POST /api/auth/login", authLoginHandler(deps))
	handleProtected(mux, deps, "GET /api/auth/session", authSessionHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/auth/logout", authLogoutHandler(deps))
	handleProtected(mux, deps, "GET /api/users", usersListHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/users", usersCreateHandler(deps))
	handleProtectedCSRF(mux, deps, "DELETE /api/users/{id}", usersDeleteHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/users/{id}/password", usersPasswordHandler(deps))
	mux.HandleFunc("GET /api/events", eventsHandler(deps))
	mux.HandleFunc("GET /api/architecture", architectureHandler(deps))
	mux.HandleFunc("GET /api/libraries", librariesHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/libraries", librarySaveHandler(deps))
	handleProtectedCSRF(mux, deps, "DELETE /api/libraries/{id}", libraryDeleteHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/libraries/{id}/scan", libraryScanByIDHandler(deps))
	mux.HandleFunc("GET /api/catalog/summary", catalogSummaryHandler(deps))
	mux.HandleFunc("GET /api/catalog/health", catalogHealthHandler(deps))
	handleProtected(mux, deps, "GET /api/migrations/formats", migrationFormatsHandler(deps))
	handleProtected(mux, deps, "GET /api/migrations/runs", migrationRunsHandler(deps))
	handleProtected(mux, deps, "GET /api/migrations/runs/{id}", migrationRunDetailHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/migrations/dry-run", migrationDryRunHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/migrations/import", migrationImportHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/migrations/runs/{id}/rollback", migrationRollbackHandler(deps))
	handleProtected(mux, deps, "GET /api/client/home", clientHomeHandler(deps))
	handleProtected(mux, deps, "GET /api/client/movies/{id}", clientMovieDetailHandler(deps))
	handleProtected(mux, deps, "GET /api/client/series/{id}", clientSeriesDetailHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/client/playback/start", clientPlaybackStartHandler(deps))
	handleProtectedCSRF(mux, deps, "PATCH /api/client/playback/{id}", clientPlaybackHeartbeatHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/client/playback/{id}/stop", clientPlaybackStopHandler(deps))
	mux.HandleFunc("GET /api/movies", moviesHandler(deps))
	mux.HandleFunc("GET /api/movies/{id}", movieDetailHandler(deps))
	mux.HandleFunc("GET /api/series", seriesHandler(deps))
	mux.HandleFunc("GET /api/series/{id}", seriesDetailHandler(deps))
	mux.HandleFunc("GET /api/review", reviewHandler(deps))
	mux.HandleFunc("GET /api/metadata/providers", metadataProvidersHandler(deps))
	mux.HandleFunc("GET /api/metadata/suggestions", metadataSuggestionsHandler(deps))
	mux.HandleFunc("GET /api/metadata/{kind}/{id}", metadataRecordsHandler(deps))
	handleProtectedCSRF(mux, deps, "PUT /api/metadata/match", metadataMatchHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/metadata/refresh", metadataRefreshHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/metadata/refresh-batch", metadataRefreshBatchHandler(deps))
	mux.HandleFunc("GET /api/artwork/{kind}/{id}", artworkHandler(deps))
	mux.HandleFunc("GET /api/versions", versionsHandler(deps))
	mux.HandleFunc("GET /api/settings/performance", performanceSettingsHandler(deps))
	mux.HandleFunc("GET /api/settings", settingsHandler(deps))
	handleProtected(mux, deps, "GET /api/settings/folders/browse", settingsFolderBrowseHandler(deps))
	handleProtectedCSRF(mux, deps, "PUT /api/settings", settingsUpdateHandler(deps))
	handleProtectedCSRF(mux, deps, "PUT /api/settings/metadata-sources", metadataSourceSettingsUpdateHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/settings/hardware/test", hardwareTestHandler(deps))
	mux.HandleFunc("GET /api/system/status", systemStatusHandler(deps))
	mux.HandleFunc("GET /api/remote/access", remoteAccessHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/remote/diagnostics", remoteDiagnosticsHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/remote/wan", wanAddressHandler(deps))
	mux.HandleFunc("GET /api/media-sources", mediaSourcesHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}", mediaSourceDetailHandler(deps))
	handleProtected(mux, deps, "GET /api/media-sources/{id}/adaptive/master.m3u8", adaptiveMasterHandler(deps))
	handleProtected(mux, deps, "GET /api/media-sources/{id}/adaptive/{variant}", adaptiveVariantHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/media-sources/{id}/adaptive/session", adaptiveSessionHandler(deps))
	handleProtected(mux, deps, "GET /api/media-sources/{id}/stream", mediaSourceStreamHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/media-sources/{id}/stream-token", mediaSourceStreamTokenHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}/tracks", mediaSourceTracksHandler(deps))
	mux.HandleFunc("GET /api/media-sources/{id}/subtitles", mediaSourceSubtitlesHandler(deps))
	handleProtected(mux, deps, "GET /api/media-sources/{id}/subtitles/{index}", mediaSourceSubtitleStreamHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/media-sources/{id}/subtitles/{index}/convert", mediaSourceSubtitleConvertHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/media-sources/{id}/probe", mediaSourceProbeHandler(deps))
	mux.HandleFunc("GET /api/probes", probesHandler(deps))
	mux.HandleFunc("GET /api/probes/{id}", probeJobHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/probes", probeStartHandler(deps))
	mux.HandleFunc("GET /api/work", workHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/work", workStartHandler(deps))
	handleProtectedCSRF(mux, deps, "DELETE /api/work/{id}", workCancelHandler(deps))
	handleProtected(mux, deps, "GET /api/work/{id}/file", workFileHandler(deps))
	mux.HandleFunc("GET /api/downloads", downloadsHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/downloads", downloadStartHandler(deps))
	mux.HandleFunc("GET /api/downloads/{id}", downloadJobHandler(deps))
	handleProtected(mux, deps, "GET /api/downloads/{id}/file", downloadFileHandler(deps))
	mux.HandleFunc("GET /api/devices/profiles", deviceProfilesHandler(deps))
	handleProtected(mux, deps, "GET /api/devices", approvedDevicesHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/devices/{id}/revoke", approvedDeviceRevokeHandler(deps))
	handleProtected(mux, deps, "GET /api/pairing/requests", pairingListHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/pairing/requests/{id}/approve", pairingApproveHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/pairing/requests/{id}/deny", pairingDenyHandler(deps))
	handleProtected(mux, deps, "GET /api/sessions", sessionsHandler(deps))
	handleProtected(mux, deps, "GET /api/sessions/{id}/inspector", sessionInspectorHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/sessions", sessionStartHandler(deps))
	handleProtectedCSRF(mux, deps, "PATCH /api/sessions/{id}", sessionUpdateHandler(deps))
	handleProtectedCSRF(mux, deps, "DELETE /api/sessions/{id}", sessionStopHandler(deps))
	mux.HandleFunc("GET /api/playback/recent", playbackRecentHandler(deps))
	mux.HandleFunc("GET /api/playback/state/{id}", playbackStateGetHandler(deps))
	handleProtectedCSRF(mux, deps, "PUT /api/playback/state/{id}", playbackStateSetHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/adaptive/telemetry", adaptiveTelemetryHandler(deps))
	mux.HandleFunc("GET /api/scans", scansHandler(deps))
	mux.HandleFunc("GET /api/scans/{id}", scanJobHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/libraries/movies/scan", movieScanHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/libraries/tv/scan", tvScanHandler(deps))
	handleProtectedCSRF(mux, deps, "POST /api/libraries/scan", allLibrariesScanHandler(deps))
	mux.HandleFunc("GET /api/playback/decision", playbackDecisionHandler(deps))
	mux.HandleFunc("GET /api/playback/route", playbackRouteHandler(deps))
	handleProtected(mux, deps, "GET /play/{id}", playerHandler(deps))
	mux.Handle("GET /", webRootHandler(deps))
	return withObservability(deps, withSecurity(deps, withResolvedSession(deps, mux)))
}

func webRootHandler(deps Deps) http.Handler {
	root := webapp.RootHandler()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() || !isHistoryRoutePath(r.URL.Path) {
			root.ServeHTTP(w, r)
			return
		}

		normalizedPath := normalizeHistoryPath(r.URL.Path)
		if normalizedPath == "" {
			normalizedPath = "/"
		}
		resolved, hasSession := auth.ResolvedSessionFromContext(r.Context())
		isSignedIn := hasSession && strings.TrimSpace(resolved.Principal.ID) != ""
		role := strings.ToLower(strings.TrimSpace(resolved.Principal.Role))
		isAdmin := role == "admin"

		if !isSignedIn {
			if normalizedPath == "/signin" {
				root.ServeHTTP(w, r)
				return
			}
			if normalizedPath == "/setup-wizard" {
				needsBootstrap, err := deps.Auth.RequiresBootstrap(r.Context())
				if err == nil && needsBootstrap {
					root.ServeHTTP(w, r)
					return
				}
			}
			http.Redirect(w, r, "/signin", http.StatusTemporaryRedirect)
			return
		}

		if normalizedPath == "/signin" || normalizedPath == "/setup-wizard" {
			http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
			return
		}
		if normalizedPath == "/settings" || strings.HasPrefix(normalizedPath, "/settings/") {
			if !isAdmin {
				http.Redirect(w, r, "/", http.StatusTemporaryRedirect)
				return
			}
		}

		root.ServeHTTP(w, r)
	})
}

func isHistoryRoutePath(requestPath string) bool {
	if requestPath == "" || requestPath == "/" {
		return true
	}
	if strings.HasPrefix(requestPath, "/api/") ||
		strings.HasPrefix(requestPath, "/_app/") ||
		strings.HasPrefix(requestPath, "/legacy/") ||
		strings.HasPrefix(requestPath, "/next/") ||
		requestPath == "/admin" ||
		strings.HasPrefix(requestPath, "/admin/") {
		return false
	}
	lastSegment := requestPath
	if slash := strings.LastIndex(lastSegment, "/"); slash >= 0 {
		lastSegment = lastSegment[slash+1:]
	}
	return !strings.Contains(lastSegment, ".")
}

func normalizeHistoryPath(requestPath string) string {
	trimmed := strings.TrimSpace(requestPath)
	if trimmed == "" {
		return "/"
	}
	if trimmed == "/" {
		return "/"
	}
	normalized := "/" + strings.Trim(trimmed, "/")
	return normalized
}

type statusRecorder struct {
	http.ResponseWriter
	status int
}

func (r *statusRecorder) WriteHeader(status int) {
	r.status = status
	r.ResponseWriter.WriteHeader(status)
}

func (r *statusRecorder) Flush() {
	if flusher, ok := r.ResponseWriter.(http.Flusher); ok {
		flusher.Flush()
	}
}

func (r *statusRecorder) Unwrap() http.ResponseWriter {
	return r.ResponseWriter
}

func withObservability(deps Deps, next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		correlationID := observability.NormalizeCorrelationID(r.Header.Get("X-Request-ID"))
		w.Header().Set("X-Request-ID", correlationID)
		recorder := &statusRecorder{ResponseWriter: w}
		next.ServeHTTP(recorder, r.WithContext(observability.WithCorrelationID(r.Context(), correlationID)))
		status := recorder.status
		if status == 0 {
			status = http.StatusOK
		}
		duration := time.Since(started)
		if deps.Observe != nil {
			deps.Observe.ObserveRequest(r.Method, r.URL.Path, status, duration, correlationID)
		}
		slog.Info("api request",
			"correlation_id", correlationID,
			"method", r.Method,
			"path", r.URL.Path,
			"status", status,
			"duration_ms", float64(duration.Microseconds())/1000,
			"remote_addr", requestRemoteAddr(r),
		)
	})
}

func withResolvedSession(deps Deps, next http.Handler) http.Handler {
	if deps.Auth == nil || deps.Auth.Disabled() {
		return next
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookieToken := ""
		if cookie, err := r.Cookie(auth.SessionCookieName); err == nil {
			cookieToken = strings.TrimSpace(cookie.Value)
		}
		headerToken := strings.TrimSpace(r.Header.Get("X-Auth-Token"))
		if headerToken == "" {
			authz := strings.TrimSpace(r.Header.Get("Authorization"))
			if authz != "" {
				parts := strings.SplitN(authz, " ", 2)
				if len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "Bearer") {
					headerToken = strings.TrimSpace(parts[1])
				}
			}
		}
		if cookieToken == "" && headerToken == "" {
			next.ServeHTTP(w, r)
			return
		}
		resolve := func(token string) (auth.ResolvedSession, error) {
			return deps.Auth.Resolve(r.Context(), token, requestRemoteAddr(r), r.UserAgent())
		}
		resolved := auth.ResolvedSession{}
		var err error
		if cookieToken != "" {
			resolved, err = resolve(cookieToken)
			if err == nil {
				if resolved.Rotated {
					writeAuthCookies(w, r, resolved)
				}
				next.ServeHTTP(w, r.WithContext(auth.ContextWithResolvedSession(r.Context(), resolved)))
				return
			}
		}
		if headerToken != "" {
			resolved, err = resolve(headerToken)
			if err == nil {
				if resolved.Rotated {
					writeAuthCookies(w, r, resolved)
				}
				w.Header().Set("X-Auth-Token", resolved.Token)
				next.ServeHTTP(w, r.WithContext(auth.ContextWithResolvedSession(r.Context(), resolved)))
				return
			}
		}
		// Both candidate tokens failed. Continue unauthenticated and let protected
		// handlers decide response; avoid clearing cookies for transient token races.
		if err != nil {
			next.ServeHTTP(w, r)
			return
		}
		next.ServeHTTP(w, r)
	})
}

func requireAuth(deps Deps, next http.HandlerFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			next(w, r)
			return
		}
		if _, ok := auth.ResolvedSessionFromContext(r.Context()); !ok {
			// Avoid clearing cookies when a header token is present. That path can be
			// hit by transient resolve failures (for example a momentary DB busy write)
			// and should not force a full sign-out.
			if !hasHeaderAuthToken(r) {
				clearAuthCookies(w)
			}
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		next(w, r)
	}
}

func requireAuthCSRF(deps Deps, next http.HandlerFunc) http.HandlerFunc {
	return requireAuth(deps, func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			next(w, r)
			return
		}
		if hasHeaderAuthToken(r) {
			next(w, r)
			return
		}
		resolved, _ := auth.ResolvedSessionFromContext(r.Context())
		csrfCookie, _ := r.Cookie(auth.CSRFCookieName)
		csrfCookieValue := ""
		if csrfCookie != nil {
			csrfCookieValue = csrfCookie.Value
		}
		if err := deps.Auth.ValidateCSRF(resolved, csrfCookieValue, r.Header.Get("X-CSRF-Token")); err != nil {
			writeError(w, http.StatusForbidden, "valid csrf token required")
			return
		}
		next(w, r)
	})
}

func authLoginHandler(deps Deps) http.HandlerFunc {
	type request struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			writeJSON(w, http.StatusOK, map[string]any{"authDisabled": true})
			return
		}
		var payload request
		if !decodeJSON(w, r, &payload) {
			return
		}
		principal, session, token, err := deps.Auth.Authenticate(r.Context(), payload.Username, payload.Password, requestRemoteAddr(r), r.UserAgent())
		if err != nil {
			publishAuthAudit(deps, r, payload.Username, "", "", "login", "denied")
			if until, ok := auth.LockoutUntil(err); ok {
				writeJSON(w, http.StatusTooManyRequests, map[string]any{
					"error":       "too many invalid login attempts",
					"lockedUntil": until.Format(time.RFC3339),
				})
				return
			}
			if errors.Is(err, auth.ErrInvalidCredentials) {
				writeError(w, http.StatusUnauthorized, "invalid username or password")
				return
			}
			writeError(w, http.StatusInternalServerError, "login failed")
			return
		}
		publishAuthAudit(deps, r, principal.Username, principal.ID, principal.Role, "login", "allowed")
		writeAuthCookies(w, r, auth.ResolvedSession{Principal: principal, Session: session, Token: token})
		writeJSON(w, http.StatusOK, map[string]any{
			"user": map[string]any{
				"id":          principal.ID,
				"username":    principal.Username,
				"displayName": principal.DisplayName,
				"role":        principal.Role,
			},
			"session": map[string]any{
				"id":        session.ID,
				"expiresAt": session.ExpiresAt.Format(time.RFC3339),
			},
			"sessionToken": token,
		})
	}
}

func authBootstrapHandler(deps Deps) http.HandlerFunc {
	type request struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			writeJSON(w, http.StatusOK, map[string]any{"authDisabled": true})
			return
		}
		var payload request
		if !decodeJSON(w, r, &payload) {
			return
		}
		principal, err := deps.Auth.BootstrapUser(r.Context(), auth.BootstrapOptions{
			Username:    payload.Username,
			Password:    payload.Password,
			DisplayName: payload.DisplayName,
		})
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrBootstrapComplete):
				writeError(w, http.StatusConflict, "an account already exists; sign in")
			case strings.Contains(strings.ToLower(err.Error()), "password"), strings.Contains(strings.ToLower(err.Error()), "required"):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "bootstrap failed")
			}
			return
		}
		createdPrincipal, session, token, err := deps.Auth.Authenticate(r.Context(), principal.Username, payload.Password, requestRemoteAddr(r), r.UserAgent())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "bootstrap sign-in failed")
			return
		}
		publishAuthAudit(deps, r, createdPrincipal.Username, createdPrincipal.ID, createdPrincipal.Role, "bootstrap", "allowed")
		writeAuthCookies(w, r, auth.ResolvedSession{Principal: createdPrincipal, Session: session, Token: token})
		writeJSON(w, http.StatusCreated, map[string]any{
			"user": map[string]any{
				"id":          createdPrincipal.ID,
				"username":    createdPrincipal.Username,
				"displayName": createdPrincipal.DisplayName,
				"role":        createdPrincipal.Role,
			},
			"session": map[string]any{
				"id":        session.ID,
				"expiresAt": session.ExpiresAt.Format(time.RFC3339),
			},
			"sessionToken": token,
		})
	}
}

func publishAuthAudit(deps Deps, r *http.Request, username string, userID string, role string, action string, result string) {
	if deps.Events == nil {
		return
	}
	deps.Events.Publish("audit.auth", map[string]any{
		"correlationId": observability.CorrelationID(r.Context()),
		"userId":        userID,
		"username":      strings.ToLower(strings.TrimSpace(username)),
		"role":          role,
		"method":        r.Method,
		"path":          r.URL.Path,
		"action":        action,
		"result":        result,
		"createdAt":     time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func publishDomainAudit(deps Deps, r *http.Request, eventType string, action string, result string, fields map[string]any) {
	if deps.Events == nil {
		return
	}
	payload := map[string]any{
		"correlationId": observability.CorrelationID(r.Context()),
		"method":        r.Method,
		"path":          r.URL.Path,
		"action":        action,
		"result":        result,
		"createdAt":     time.Now().UTC().Format(time.RFC3339Nano),
	}
	if resolved, ok := auth.ResolvedSessionFromContext(r.Context()); ok {
		payload["userId"] = resolved.Principal.ID
		payload["username"] = resolved.Principal.Username
		payload["role"] = resolved.Principal.Role
	}
	for key, value := range fields {
		payload[key] = value
	}
	deps.Events.Publish(eventType, payload)
}

func authSessionHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			writeJSON(w, http.StatusOK, map[string]any{"authDisabled": true})
			return
		}
		resolved, ok := auth.ResolvedSessionFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		payload := map[string]any{
			"user": map[string]any{
				"id":          resolved.Principal.ID,
				"username":    resolved.Principal.Username,
				"displayName": resolved.Principal.DisplayName,
				"role":        resolved.Principal.Role,
			},
		}
		payload["session"] = map[string]any{
			"id":        resolved.Session.ID,
			"expiresAt": resolved.Session.ExpiresAt.Format(time.RFC3339),
		}
		payload["csrfToken"] = resolved.Session.CSRFToken
		writeJSON(w, http.StatusOK, payload)
	}
}

func authLogoutHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth != nil && !deps.Auth.Disabled() {
			if cookie, err := r.Cookie(auth.SessionCookieName); err == nil && cookie.Value != "" {
				_ = deps.Auth.Revoke(r.Context(), cookie.Value)
			}
		}
		clearAuthCookies(w)
		writeJSON(w, http.StatusOK, map[string]any{"status": "logged_out"})
	}
}

func usersListHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			writeError(w, http.StatusServiceUnavailable, "user accounts are not available")
			return
		}
		users, err := deps.Auth.ListUsers(r.Context())
		if err != nil {
			if errors.Is(err, auth.ErrUnauthorized) {
				writeError(w, http.StatusUnauthorized, "authentication required")
				return
			}
			writeError(w, http.StatusInternalServerError, "user list failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"users": users})
	}
}

func usersCreateHandler(deps Deps) http.HandlerFunc {
	type request struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		DisplayName string `json:"displayName"`
		Role        string `json:"role"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			writeError(w, http.StatusServiceUnavailable, "user accounts are not available")
			return
		}
		var payload request
		if !decodeJSON(w, r, &payload) {
			return
		}
		principal, err := deps.Auth.CreateUser(r.Context(), payload.Username, payload.Password, payload.DisplayName, payload.Role)
		if err != nil {
			switch {
			case errors.Is(err, auth.ErrUnauthorized):
				writeError(w, http.StatusUnauthorized, "authentication required")
			case strings.Contains(strings.ToLower(err.Error()), "users.username"):
				writeError(w, http.StatusConflict, "username already exists")
			case strings.Contains(strings.ToLower(err.Error()), "required"), strings.Contains(strings.ToLower(err.Error()), "password"):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusBadRequest, "user create failed")
			}
			return
		}
		publishDomainAudit(deps, r, "audit.auth", "user.create", "allowed", map[string]any{
			"targetUserId":   principal.ID,
			"targetUsername": principal.Username,
			"targetRole":     principal.Role,
		})
		writeJSON(w, http.StatusCreated, map[string]any{
			"user": map[string]any{
				"id":          principal.ID,
				"username":    principal.Username,
				"displayName": principal.DisplayName,
				"role":        principal.Role,
			},
		})
	}
}

func usersDeleteHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			writeError(w, http.StatusServiceUnavailable, "user accounts are not available")
			return
		}
		userID := strings.TrimSpace(r.PathValue("id"))
		resolved, _ := auth.ResolvedSessionFromContext(r.Context())
		if userID == "" {
			writeError(w, http.StatusBadRequest, "user id is required")
			return
		}
		if userID == resolved.Principal.ID {
			writeError(w, http.StatusBadRequest, "cannot delete the signed-in account")
			return
		}
		if err := deps.Auth.DeleteUser(r.Context(), userID); err != nil {
			switch {
			case errors.Is(err, auth.ErrUnauthorized):
				writeError(w, http.StatusUnauthorized, "authentication required")
			case errors.Is(err, auth.ErrUserNotFound):
				writeError(w, http.StatusNotFound, "user not found")
			case errors.Is(err, auth.ErrLastAdmin):
				writeError(w, http.StatusConflict, "cannot delete the last admin account")
			default:
				writeError(w, http.StatusInternalServerError, "user delete failed")
			}
			return
		}
		publishDomainAudit(deps, r, "audit.auth", "user.delete", "allowed", map[string]any{
			"targetUserId": userID,
		})
		writeJSON(w, http.StatusOK, map[string]any{"status": "deleted"})
	}
}

func usersPasswordHandler(deps Deps) http.HandlerFunc {
	type request struct {
		Password string `json:"password"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			writeError(w, http.StatusServiceUnavailable, "user accounts are not available")
			return
		}
		userID := strings.TrimSpace(r.PathValue("id"))
		if userID == "" {
			writeError(w, http.StatusBadRequest, "user id is required")
			return
		}
		var payload request
		if !decodeJSON(w, r, &payload) {
			return
		}
		if err := deps.Auth.SetUserPassword(r.Context(), userID, payload.Password); err != nil {
			switch {
			case errors.Is(err, auth.ErrUnauthorized):
				writeError(w, http.StatusUnauthorized, "authentication required")
			case errors.Is(err, auth.ErrUserNotFound):
				writeError(w, http.StatusNotFound, "user not found")
			case strings.Contains(strings.ToLower(err.Error()), "password"):
				writeError(w, http.StatusBadRequest, err.Error())
			default:
				writeError(w, http.StatusInternalServerError, "password update failed")
			}
			return
		}
		publishDomainAudit(deps, r, "audit.auth", "user.password.update", "allowed", map[string]any{
			"targetUserId": userID,
		})
		responsePayload := map[string]any{"status": "updated"}
		if resolved, ok := auth.ResolvedSessionFromContext(r.Context()); ok && resolved.Principal.ID == userID {
			principal, session, token, err := deps.Auth.Authenticate(
				r.Context(),
				resolved.Principal.Username,
				payload.Password,
				requestRemoteAddr(r),
				r.UserAgent(),
			)
			if err != nil {
				writeError(w, http.StatusInternalServerError, "password updated but session refresh failed")
				return
			}
			refreshed := auth.ResolvedSession{Principal: principal, Session: session, Token: token}
			writeAuthCookies(w, r, refreshed)
			w.Header().Set("X-Auth-Token", token)
			responsePayload["user"] = map[string]any{
				"id":          principal.ID,
				"username":    principal.Username,
				"displayName": principal.DisplayName,
				"role":        principal.Role,
			}
			responsePayload["session"] = map[string]any{
				"id":        session.ID,
				"expiresAt": session.ExpiresAt.Format(time.RFC3339),
			}
			responsePayload["sessionToken"] = token
		}
		writeJSON(w, http.StatusOK, responsePayload)
	}
}

func writeAuthCookies(w http.ResponseWriter, r *http.Request, resolved auth.ResolvedSession) {
	secure := requestSecure(r)
	maxAge := int(time.Until(resolved.Session.ExpiresAt).Seconds())
	if maxAge < 0 {
		maxAge = 0
	}
	http.SetCookie(w, &http.Cookie{
		Name:     auth.SessionCookieName,
		Value:    resolved.Token,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
	http.SetCookie(w, &http.Cookie{
		Name:     auth.CSRFCookieName,
		Value:    resolved.Session.CSRFToken,
		Path:     "/",
		MaxAge:   maxAge,
		HttpOnly: false,
		SameSite: http.SameSiteLaxMode,
		Secure:   secure,
	})
}

func clearAuthCookies(w http.ResponseWriter) {
	for _, name := range []string{auth.SessionCookieName, auth.CSRFCookieName} {
		http.SetCookie(w, &http.Cookie{
			Name:     name,
			Value:    "",
			Path:     "/",
			MaxAge:   -1,
			HttpOnly: name == auth.SessionCookieName,
			SameSite: http.SameSiteLaxMode,
		})
	}
}

func requestSecure(r *http.Request) bool {
	// Do not infer secure-cookie behavior from forwarded headers. In desktop and
	// dev topologies those headers can be present even when the browser origin is
	// plain HTTP localhost, which causes cookies to be dropped.
	return r.TLS != nil
}

func hasHeaderAuthToken(r *http.Request) bool {
	if strings.TrimSpace(r.Header.Get("X-Auth-Token")) != "" {
		return true
	}
	authz := strings.TrimSpace(r.Header.Get("Authorization"))
	if authz == "" {
		return false
	}
	parts := strings.SplitN(authz, " ", 2)
	return len(parts) == 2 && strings.EqualFold(strings.TrimSpace(parts[0]), "Bearer") && strings.TrimSpace(parts[1]) != ""
}

func requestHostIsLoopback(r *http.Request) bool {
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" && r.URL != nil {
		host = strings.TrimSpace(r.URL.Host)
	}
	return isLoopbackHost(host)
}

func isLoopbackHost(hostport string) bool {
	host := strings.TrimSpace(hostport)
	if host == "" {
		return false
	}
	parsedHost := host
	if value, _, err := net.SplitHostPort(host); err == nil {
		parsedHost = value
	}
	parsedHost = strings.TrimSpace(strings.Trim(parsedHost, "[]"))
	if parsedHost == "" {
		return false
	}
	lower := strings.ToLower(parsedHost)
	if lower == "localhost" || strings.HasSuffix(lower, ".localhost") {
		return true
	}
	ip := net.ParseIP(parsedHost)
	return ip != nil && ip.IsLoopback()
}

func requestRemoteAddr(r *http.Request) string {
	if forwarded := strings.TrimSpace(r.Header.Get("X-Forwarded-For")); forwarded != "" {
		return strings.Split(forwarded, ",")[0]
	}
	host, _, err := net.SplitHostPort(r.RemoteAddr)
	if err == nil {
		return host
	}
	return r.RemoteAddr
}

func healthHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, _ := healthSnapshot(deps)
		writeJSON(w, http.StatusOK, payload)
	}
}

func readinessHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		payload, degraded := healthSnapshot(deps)
		status := http.StatusOK
		if degraded {
			status = http.StatusServiceUnavailable
		}
		writeJSON(w, status, payload)
	}
}

func metricsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		requests := []observability.RequestMetric{}
		events := []observability.EventMetric{}
		if deps.Observe != nil {
			requests = deps.Observe.Requests()
			events = deps.Observe.Events()
		}
		queues := []map[string]any{}
		timeline := []observability.TimelineEntry{}
		if deps.Jobs != nil {
			queues = deps.Jobs.Snapshot()
		}
		if deps.Observe != nil {
			timeline = deps.Observe.Recent(60)
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"requests": requests,
			"queues":   queues,
			"events":   events,
			"timeline": timeline,
			"outcomes": map[string]any{
				"sessions":  sessionOutcomeCounts(deps),
				"transcode": transcodeOutcomeCounts(deps),
				"downloads": downloadOutcomeCounts(deps),
				"probes":    probeOutcomeCounts(deps),
			},
			"alerts": observability.EvaluateAlerts(queues, requests),
		})
	}
}

func clientBootstrapHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := currentConfig(deps)
		profileID := firstNonEmpty(r.URL.Query().Get("clientProfile"), "apple-tv")
		profile, ok := deps.Devices.GetProfile(profileID)
		if !ok {
			profile, _ = deps.Devices.GetProfile("apple-tv")
		}
		startedAt := deps.StartedAt
		if startedAt.IsZero() {
			startedAt = time.Now().UTC()
		}
		authRequired := deps.Auth != nil && !deps.Auth.Disabled()
		bootstrapAllowed := false
		defaultAdminUsername := strings.TrimSpace(cfg.AdminUsername)
		if defaultAdminUsername == "" {
			defaultAdminUsername = "admin"
		}
		if authRequired {
			if required, err := deps.Auth.RequiresBootstrap(r.Context()); err == nil {
				bootstrapAllowed = required
			}
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"server": map[string]any{
				"product":      "xuva",
				"name":         configDisplayName(cfg.ServerName),
				"baseUrl":      requestBaseURL(r, cfg.HTTPAddr),
				"httpAddr":     cfg.HTTPAddr,
				"lanAddresses": lanAddresses(cfg.HTTPAddr),
				"startedAt":    startedAt.UTC().Format(time.RFC3339),
			},
			"auth": map[string]any{
				"required":          authRequired,
				"bootstrapAllowed":  bootstrapAllowed,
				"defaultUsername":   defaultAdminUsername,
				"bootstrapEndpoint": "/api/auth/bootstrap",
				"methods":           []string{"session_cookie", "local_pairing_code"},
			},
			"client": map[string]any{
				"requestedProfile": profileID,
				"profile":          profile,
			},
			"profiles": deps.Devices.Profiles(),
			"features": map[string]any{
				"directPlay":        true,
				"hlsAdaptive":       true,
				"resume":            true,
				"watchedState":      true,
				"trackSelection":    true,
				"devicePairing":     "local_code",
				"remoteDiagnostics": true,
				"vendorRelay":       false,
			},
			"endpoints": map[string]string{
				"health":           "/api/health",
				"authSession":      "/api/auth/session",
				"login":            "/api/auth/login",
				"pairingCreate":    "/api/pairing/requests",
				"pairingStatus":    "/api/pairing/requests/{id}",
				"clientHome":       "/api/client/home",
				"deviceProfiles":   "/api/devices/profiles",
				"movies":           "/api/movies",
				"series":           "/api/series",
				"playbackDecision": "/api/playback/decision",
				"playbackRoute":    "/api/playback/route",
				"sessions":         "/api/sessions",
				"playbackState":    "/api/playback/state/{mediaSourceId}",
				"streamToken":      "/api/media-sources/{id}/stream-token",
				"directStream":     "/api/media-sources/{id}/stream",
				"adaptiveMaster":   "/api/media-sources/{id}/adaptive/master.m3u8",
				"adaptiveSession":  "/api/media-sources/{id}/adaptive/session",
				"tracks":           "/api/media-sources/{id}/tracks",
				"subtitles":        "/api/media-sources/{id}/subtitles",
				"remoteAccess":     "/api/remote/access",
			},
		})
	}
}

func discoveryStatusHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := currentConfig(deps)
		status := discovery.Status{
			Enabled:     cfg.DiscoveryEnabled,
			ServiceName: configDisplayName(cfg.ServerName),
			ServiceType: "_xuva._tcp.local.",
			Note:        "Local discovery is not running.",
		}
		if deps.Discovery != nil {
			status = deps.Discovery.Status()
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"enabled":     status.Enabled,
			"running":     status.Running,
			"serviceName": status.ServiceName,
			"serviceType": status.ServiceType,
			"port":        status.Port,
			"txtRecords":  status.TXTRecords,
			"lastError":   status.LastError,
			"note":        status.Note,
		})
	}
}

func healthSnapshot(deps Deps) (map[string]any, bool) {
	startedAt := deps.StartedAt
	if startedAt.IsZero() {
		startedAt = time.Now().UTC()
	}
	checks := map[string]any{}
	degraded := false
	for name, path := range runtimePaths(deps.Config) {
		if path == "" {
			continue
		}
		ok, message := pathReady(path)
		checks["path."+name] = map[string]any{"ok": ok, "message": message}
		if !ok {
			degraded = true
		}
	}
	if deps.Jobs != nil {
		for _, queue := range deps.Jobs.Snapshot() {
			name, _ := queue["name"].(string)
			workers := intFromAny(queue["workers"])
			active := intFromAny(queue["active"])
			queued := intFromAny(queue["queued"])
			ok := workers <= 0 || active < workers || queued == 0
			checks["queue."+name] = map[string]any{
				"ok":      ok,
				"workers": workers,
				"active":  active,
				"queued":  queued,
			}
			if !ok {
				degraded = true
			}
		}
	}
	status := "ok"
	if degraded {
		status = "degraded"
	}
	return map[string]any{
		"status":    status,
		"service":   "xuva-server",
		"startedAt": startedAt.UTC().Format(time.RFC3339),
		"httpAddr":  deps.Config.HTTPAddr,
		"checks":    checks,
		"libraries": map[string]string{
			"movies": deps.Config.MovieLibraryPath,
			"tv":     deps.Config.TVLibraryPath,
		},
	}, degraded
}

func pathReady(path string) (bool, string) {
	info, err := os.Stat(path)
	if err != nil {
		return false, err.Error()
	}
	if !info.IsDir() {
		return false, "path is not a directory"
	}
	testPath, ok := safeChildPath(path, ".xuva-healthcheck")
	if !ok {
		return false, "path cannot be checked safely"
	}
	file, err := os.OpenFile(testPath, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0o600)
	if err != nil {
		return false, err.Error()
	}
	_ = file.Close()
	_ = os.Remove(testPath)
	return true, "ready"
}

func sessionOutcomeCounts(deps Deps) map[string]int {
	output := map[string]int{}
	if deps.Sessions == nil {
		return output
	}
	for _, session := range deps.Sessions.List() {
		key := strings.TrimSpace(session.Status)
		if key == "" {
			key = "unknown"
		}
		output[key]++
	}
	return output
}

func transcodeOutcomeCounts(deps Deps) map[string]int {
	output := map[string]int{}
	if deps.Transcode == nil {
		return output
	}
	for _, job := range deps.Transcode.List() {
		output[string(job.Status)]++
	}
	return output
}

func downloadOutcomeCounts(deps Deps) map[string]int {
	output := map[string]int{}
	if deps.Downloads == nil {
		return output
	}
	for _, job := range deps.Downloads.List() {
		output[string(job.Status)]++
	}
	return output
}

func probeOutcomeCounts(deps Deps) map[string]int {
	output := map[string]int{}
	if deps.Probes == nil {
		return output
	}
	for _, job := range deps.Probes.List() {
		output[string(job.Status)]++
	}
	return output
}

func intFromAny(value any) int {
	switch typed := value.(type) {
	case int:
		return typed
	case int64:
		return int(typed)
	case float64:
		return int(typed)
	default:
		return 0
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
		cfg := currentConfig(deps)
		writeJSON(w, http.StatusOK, map[string]any{
			"libraries":       deps.Libraries.List(),
			"metadataSources": metadataSourceCatalogPayload(r.Context(), cfg, deps.Metadata),
		})
	}
}

func librarySaveHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request libraries.Library
		if !decodeJSON(w, r, &request) {
			return
		}
		if len(request.MetadataSources) == 0 {
			request.MetadataSources = defaultMetadataSourcePreferenceForLibrary(currentConfig(deps), request.Kind)
		}
		request.MetadataSources = metasources.NormalizeRequestedSourceOrder(string(request.Kind), request.MetadataSources)
		library, err := deps.Catalog.SaveLibrary(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		deps.Libraries.Set(library)
		deps.Events.Publish("library.updated", library)
		publishDomainAudit(deps, r, "audit.library", "library.save", "allowed", map[string]any{
			"libraryId":   library.ID,
			"libraryKind": string(library.Kind),
			"storageType": string(library.StorageType),
		})
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
		publishDomainAudit(deps, r, "audit.library", "library.delete", "allowed", map[string]any{
			"libraryId": id,
		})
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
		job, err := deps.Scans.Start(r.Context(), scans.Request{Kind: kind, LibraryID: library.ID, Path: library.Path})
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		publishDomainAudit(deps, r, "audit.library", "library.scan", "allowed", map[string]any{
			"libraryId":   library.ID,
			"libraryKind": string(library.Kind),
			"jobId":       job.ID,
		})
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

func migrationFormatsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"formats": migration.Formats()})
	}
}

func migrationRunsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Migration == nil {
			writeError(w, http.StatusServiceUnavailable, "migration tooling is not available")
			return
		}
		runs, err := deps.Migration.ListRuns(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "migration runs lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"runs": runs})
	}
}

func migrationRunDetailHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Migration == nil {
			writeError(w, http.StatusServiceUnavailable, "migration tooling is not available")
			return
		}
		report, ok, err := deps.Migration.GetRun(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "migration run lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "migration run not found")
			return
		}
		writeJSON(w, http.StatusOK, report)
	}
}

func migrationDryRunHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Migration == nil {
			writeError(w, http.StatusServiceUnavailable, "migration tooling is not available")
			return
		}
		var request migration.Request
		if !decodeJSON(w, r, &request) {
			return
		}
		report, err := deps.Migration.DryRun(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, report)
	}
}

func migrationImportHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Migration == nil {
			writeError(w, http.StatusServiceUnavailable, "migration tooling is not available")
			return
		}
		var request migration.Request
		if !decodeJSON(w, r, &request) {
			return
		}
		report, err := deps.Migration.Import(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, report)
	}
}

func migrationRollbackHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Migration == nil {
			writeError(w, http.StatusServiceUnavailable, "migration tooling is not available")
			return
		}
		report, err := deps.Migration.Rollback(r.Context(), r.PathValue("id"))
		if err != nil {
			switch err {
			case migration.ErrRunNotFound:
				writeError(w, http.StatusNotFound, "migration run not found")
			case migration.ErrRunRollback:
				writeError(w, http.StatusConflict, "migration run cannot be rolled back")
			default:
				writeError(w, http.StatusBadRequest, err.Error())
			}
			return
		}
		writeJSON(w, http.StatusOK, report)
	}
}

func clientHomeHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		limit := queryInt(r, "limit", 24)
		recent, err := deps.PlayState.Recent(r.Context(), requestUserID(r), limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "recent playback lookup failed")
			return
		}
		movieItems, err := deps.Catalog.ListMovies(r.Context(), limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "movie list failed")
			return
		}
		seriesItems, err := deps.Catalog.ListSeries(r.Context(), limit)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "series list failed")
			return
		}
		rows := []map[string]any{
			{"id": "continue", "title": "Continue Watching", "items": tvRecentItems(recent)},
			{"id": "movies", "title": "Movies", "items": tvMovieItems(movieItems)},
			{"id": "tv", "title": "TV Shows", "items": tvSeriesItems(seriesItems)},
			{"id": "recently-added", "title": "Recently Added", "items": tvRecentlyAddedItems(movieItems, seriesItems, limit)},
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"profile": firstNonEmpty(r.URL.Query().Get("clientProfile"), "apple-tv"),
			"hero":    firstTVHomeItem(rows),
			"rows":    rows,
			"actions": map[string]string{
				"movieDetail":  "/api/movies/{id}",
				"seriesDetail": "/api/series/{id}",
				"playback":     "/api/playback/route",
			},
		})
	}
}

func clientMovieDetailHandler(deps Deps) http.HandlerFunc {
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
		writeJSON(w, http.StatusOK, clientMovieDetailPayload(r.Context(), deps, r, detail))
	}
}

func clientSeriesDetailHandler(deps Deps) http.HandlerFunc {
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
		writeJSON(w, http.StatusOK, clientSeriesDetailPayload(r.Context(), deps, r, detail))
	}
}

func clientPlaybackStartHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload clientPlaybackStartRequest
		if !decodeJSON(w, r, &payload) {
			return
		}
		if payload.MediaSourceID == "" {
			writeError(w, http.StatusBadRequest, "media source id is required")
			return
		}
		source, ok, err := deps.Catalog.GetMediaSource(r.Context(), payload.MediaSourceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		decision := clientPlaybackDecision(r.Context(), deps, source, clientPlaybackOptions{
			ClientProfile:      payload.ClientProfile,
			RouteType:          payload.RouteType,
			MaxNetworkBitrate:  payload.MaxNetworkBitrate,
			AudioTrackIndex:    payload.AudioTrackIndex,
			SubtitleTrackIndex: payload.SubtitleTrackIndex,
			SubtitleMode:       payload.SubtitleMode,
			SubtitleActive:     payload.SubtitleActive,
		})
		routePayload, status, err := clientPlaybackRoutePayload(deps, r, source, decision, payload)
		if err != nil {
			writeError(w, status, err.Error())
			return
		}
		start := sessions.StartRequest{
			UserID:          requestUserID(r),
			DeviceID:        firstNonEmpty(payload.DeviceID, "apple-tv"),
			MediaSourceID:   source.ID,
			ClientProfile:   firstNonEmpty(payload.ClientProfile, "apple-tv"),
			Route:           clientRouteLabel(decision),
			Mode:            string(decision.Mode),
			ReasonCode:      decision.ReasonCode,
			ReasonText:      decision.ReasonText,
			SelectedTracks:  clientSelectedTracks(payload.AudioTrackIndex, payload.SubtitleTrackIndex),
			ProgressSeconds: 0,
		}
		if err := enrichSessionStartRequest(deps, r.Context(), &start); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		session, err := deps.Sessions.Start(start)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"sessionId":           session.ID,
			"deviceId":            session.DeviceID,
			"mediaSourceId":       session.MediaSourceID,
			"playbackStateUrl":    "/api/playback/state/" + session.MediaSourceID,
			"heartbeatUrl":        "/api/client/playback/" + session.ID,
			"stopUrl":             "/api/client/playback/" + session.ID + "/stop",
			"heartbeatIntervalMs": 2000,
			"decision":            decision,
			"route":               routePayload,
		})
	}
}

func clientPlaybackHeartbeatHandler(deps Deps) http.HandlerFunc {
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
		writeJSON(w, http.StatusOK, map[string]any{"session": session})
	}
}

func clientPlaybackStopHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request sessions.UpdateRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		if request.Status == "" {
			request.Status = "stopped"
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
		stopped, _ := deps.Sessions.Stop(r.PathValue("id"))
		writeJSON(w, http.StatusOK, map[string]any{"session": stopped})
	}
}

type clientPlaybackOptions struct {
	ClientProfile      string
	RouteType          string
	MaxNetworkBitrate  int64
	AudioTrackIndex    int
	SubtitleTrackIndex int
	SubtitleMode       string
	SubtitleActive     bool
}

type clientPlaybackStartRequest struct {
	MediaSourceID      string `json:"mediaSourceId"`
	DeviceID           string `json:"deviceId"`
	ClientProfile      string `json:"clientProfile"`
	RouteType          string `json:"routeType"`
	MaxNetworkBitrate  int64  `json:"maxNetworkBitrate"`
	AudioTrackIndex    int    `json:"audioTrackIndex"`
	SubtitleTrackIndex int    `json:"subtitleTrackIndex"`
	SubtitleMode       string `json:"subtitleMode"`
	SubtitleActive     bool   `json:"subtitleActive"`
}

func clientMovieDetailPayload(ctx context.Context, deps Deps, r *http.Request, detail catalog.MovieDetail) map[string]any {
	versions := make([]map[string]any, 0, len(detail.Versions))
	defaultMediaSourceID := ""
	for _, version := range detail.Versions {
		if defaultMediaSourceID == "" {
			defaultMediaSourceID = version.MediaSourceID
		}
		if payload, ok := clientVersionPayload(ctx, deps, r, version.MediaSourceID); ok {
			versions = append(versions, payload)
		}
	}
	return map[string]any{
		"item": map[string]any{
			"id":           detail.ID,
			"kind":         "movie",
			"title":        detail.Title,
			"year":         detail.Year,
			"overview":     metadataOverview(detail.Metadata),
			"posterUrl":    metadataPoster(detail.Metadata),
			"backdropUrl":  metadataBackdrop(detail.Metadata),
			"versionCount": detail.VersionCount,
		},
		"defaultMediaSourceId": defaultMediaSourceID,
		"versions":             versions,
	}
}

func clientSeriesDetailPayload(ctx context.Context, deps Deps, r *http.Request, detail catalog.SeriesDetail) map[string]any {
	seasons := make([]map[string]any, 0, len(detail.Seasons))
	defaultMediaSourceID := ""
	for _, season := range detail.Seasons {
		episodes := make([]map[string]any, 0, len(season.Episodes))
		for _, episode := range season.Episodes {
			versionPayloads := make([]map[string]any, 0, len(episode.Versions))
			for _, version := range episode.Versions {
				if defaultMediaSourceID == "" {
					defaultMediaSourceID = version.MediaSourceID
				}
				if payload, ok := clientVersionPayload(ctx, deps, r, version.MediaSourceID); ok {
					versionPayloads = append(versionPayloads, payload)
				}
			}
			episodes = append(episodes, map[string]any{
				"id":            episode.ID,
				"title":         firstNonEmpty(episode.Title, episodeEpisodeLabel(episode)),
				"seasonNumber":  episode.SeasonNumber,
				"episodeNumber": episode.EpisodeNumber,
				"versionCount":  episode.VersionCount,
				"versions":      versionPayloads,
			})
		}
		seasons = append(seasons, map[string]any{
			"id":           season.ID,
			"seasonNumber": season.SeasonNumber,
			"episodes":     episodes,
		})
	}
	return map[string]any{
		"item": map[string]any{
			"id":           detail.ID,
			"kind":         "series",
			"title":        detail.Title,
			"overview":     metadataOverview(detail.Metadata),
			"posterUrl":    metadataPoster(detail.Metadata),
			"backdropUrl":  metadataBackdrop(detail.Metadata),
			"seasonCount":  detail.SeasonCount,
			"episodeCount": detail.EpisodeCount,
		},
		"defaultMediaSourceId": defaultMediaSourceID,
		"seasons":              seasons,
	}
}

func clientVersionPayload(ctx context.Context, deps Deps, r *http.Request, mediaSourceID string) (map[string]any, bool) {
	source, ok, err := deps.Catalog.GetMediaSource(ctx, mediaSourceID)
	if err != nil || !ok {
		return nil, false
	}
	tracks, _, _ := deps.Catalog.GetMediaSourceTracks(ctx, mediaSourceID)
	sidecars := subtitles.DiscoverSidecars(source.Path)
	decision := clientPlaybackDecision(ctx, deps, source, clientPlaybackOptions{
		ClientProfile:     firstNonEmpty(r.URL.Query().Get("clientProfile"), "apple-tv"),
		RouteType:         firstNonEmpty(r.URL.Query().Get("routeType"), "lan"),
		MaxNetworkBitrate: queryInt64(r, "maxNetworkBitrate", 0),
	})
	return map[string]any{
		"mediaSourceId":  source.ID,
		"path":           source.RelPath,
		"qualityLabel":   sessionQualityLabel(source),
		"source":         source,
		"audioTracks":    tracks.AudioTracks,
		"subtitleTracks": tracks.SubtitleTracks,
		"sidecars":       clientSidecarPayloads(deps.Subtitles, sidecars, firstNonEmpty(r.URL.Query().Get("clientProfile"), "apple-tv")),
		"decision":       decision,
	}, true
}

func clientPlaybackDecision(ctx context.Context, deps Deps, source catalog.MediaSourceItem, options clientPlaybackOptions) playback.Decision {
	request := playback.Request{
		MediaSourceID:       source.ID,
		ClientProfile:       firstNonEmpty(options.ClientProfile, "apple-tv"),
		RouteType:           firstNonEmpty(options.RouteType, "lan"),
		MaxNetworkBitrate:   options.MaxNetworkBitrate,
		AudioTrackIndex:     options.AudioTrackIndex,
		SubtitleTrackIndex:  options.SubtitleTrackIndex,
		SubtitleMode:        options.SubtitleMode,
		SubtitleTrackActive: options.SubtitleActive,
	}
	request = applyClientProfile(deps, request)
	return deps.Playback.DecideSource(ctx, request, playbackSourceFacts(ctx, deps, request, source))
}

func clientPlaybackRoutePayload(deps Deps, r *http.Request, source catalog.MediaSourceItem, decision playback.Decision, payload clientPlaybackStartRequest) (map[string]any, int, error) {
	if decision.Mode == playback.DirectPlay || decision.Mode == playback.DecisionDeferred {
		return map[string]any{
			"route":    "direct",
			"status":   "ready",
			"url":      "/api/media-sources/" + source.ID + "/stream",
			"decision": decision,
		}, http.StatusOK, nil
	}
	if !playbackPolicyAllows(deps.Config.PlaybackPolicy, decision) {
		return map[string]any{
			"route":           "blocked",
			"status":          "blocked_by_policy",
			"policy":          playbackPolicyStatus(deps.Config.PlaybackPolicy),
			"decision":        decision,
			"fallbackOptions": playbackPolicyFallbacks(deps.Config.PlaybackPolicy, decision),
		}, http.StatusOK, nil
	}
	if deps.Auth != nil && !deps.Auth.Disabled() {
		return nil, http.StatusConflict, fmt.Errorf("native client playback requires persistent device authentication before protected stream URLs can be issued")
	}
	if decision.Mode == playback.AdaptiveStream {
		plan := adaptivePlanForSource(deps, source, payload.ClientProfile, firstNonEmpty(payload.RouteType, "remote"), payload.MaxNetworkBitrate)
		if plan.Enabled {
			return map[string]any{
				"route":       "adaptive",
				"status":      "ready",
				"protocol":    plan.Protocol,
				"manifestUrl": "/api/media-sources/" + source.ID + "/adaptive/master.m3u8?clientProfile=" + url.QueryEscape(firstNonEmpty(payload.ClientProfile, "apple-tv")) + "&routeType=" + url.QueryEscape(firstNonEmpty(payload.RouteType, "remote")) + "&maxNetworkBitrate=" + fmt.Sprintf("%d", payload.MaxNetworkBitrate),
				"plan":        plan,
				"decision":    decision,
			}, http.StatusOK, nil
		}
	}
	mode := transcode.ModeTranscode
	if decision.Mode == playback.Remux {
		mode = transcode.ModeRemux
	}
	if job, ok := deps.Transcode.FindCompleted(source.ID, mode); ok {
		return map[string]any{
			"route":    string(mode),
			"status":   "ready",
			"url":      "/api/work/" + job.ID + "/file",
			"job":      job,
			"decision": decision,
		}, http.StatusOK, nil
	}
	if job, ok := deps.Transcode.FindActive(source.ID, mode); ok {
		return map[string]any{
			"route":    string(mode),
			"status":   string(job.Status),
			"job":      job,
			"decision": decision,
		}, http.StatusAccepted, nil
	}
	request := transcode.Request{MediaSourceID: source.ID, Mode: mode, SourcePath: source.Path}
	if mode == transcode.ModeTranscode {
		if encoder, ok := selectedHardwareEncoder(r.Context(), deps.Config); ok {
			request.Acceleration = "hardware"
			request.VideoEncoder = encoder
		}
	}
	job, err := deps.Transcode.Start(r.Context(), request)
	if err != nil {
		return nil, http.StatusBadRequest, err
	}
	return map[string]any{
		"route":    string(mode),
		"status":   string(job.Status),
		"job":      job,
		"decision": decision,
	}, http.StatusAccepted, nil
}

func clientRouteLabel(decision playback.Decision) string {
	switch decision.Mode {
	case playback.DirectPlay:
		return "direct"
	case playback.AdaptiveStream:
		return "adaptive"
	case playback.Remux:
		return "remux"
	case playback.AudioTranscode:
		return "audio-transcode"
	case playback.SubtitleBurn:
		return "subtitle-burn"
	case playback.VideoTranscode:
		return "transcode"
	default:
		return "pending"
	}
}

func clientSelectedTracks(audioTrackIndex int, subtitleTrackIndex int) map[string]string {
	selected := map[string]string{}
	if audioTrackIndex > 0 {
		selected["audioTrack"] = fmt.Sprintf("%d", audioTrackIndex)
	}
	if subtitleTrackIndex > 0 {
		selected["subtitleTrack"] = fmt.Sprintf("%d", subtitleTrackIndex)
	}
	if len(selected) == 0 {
		return nil
	}
	return selected
}

func clientSidecarPayloads(service *subtitles.Service, sidecars []subtitles.Sidecar, clientProfile string) []map[string]any {
	if len(sidecars) == 0 {
		return nil
	}
	output := make([]map[string]any, 0, len(sidecars))
	for _, sidecar := range sidecars {
		output = append(output, map[string]any{
			"path":              sidecar.Path,
			"relPath":           sidecar.RelPath,
			"format":            sidecar.Format,
			"language":          sidecar.Language,
			"forced":            sidecar.Forced,
			"hearingImpaired":   sidecar.HearingImpaired,
			"requiresVideoBurn": sidecar.RequiresVideoBurn,
			"conversion":        service.PlanConversion(sidecar, clientProfile),
		})
	}
	return output
}

func metadataOverview(record *catalog.MetadataRecord) string {
	if record == nil {
		return ""
	}
	return record.Overview
}

func episodeEpisodeLabel(episode catalog.EpisodeBrief) string {
	return fmt.Sprintf("S%02d E%02d", episode.SeasonNumber, episode.EpisodeNumber)
}

func tvRecentItems(items []playstate.RecentItem) []map[string]any {
	output := make([]map[string]any, 0, len(items))
	for _, item := range items {
		output = append(output, map[string]any{
			"id":              item.MediaSourceID,
			"kind":            item.Kind,
			"title":           firstNonEmpty(item.Name, item.RelPath, "Resume playback"),
			"subtitle":        "Resume from " + formatProgress(item.Percent),
			"mediaSourceId":   item.MediaSourceID,
			"progressPercent": item.Percent,
			"lastPlayedAt":    item.LastPlayedAt,
			"route":           "Resume",
		})
	}
	return output
}

func tvMovieItems(items []catalog.MovieListItem) []map[string]any {
	output := make([]map[string]any, 0, len(items))
	for _, item := range items {
		subtitle := "Movie"
		if item.Year > 0 {
			subtitle = fmt.Sprintf("%d", item.Year)
		}
		output = append(output, map[string]any{
			"id":           item.ID,
			"kind":         "movie",
			"title":        item.Title,
			"subtitle":     subtitle,
			"posterUrl":    metadataPoster(item.Metadata),
			"backdropUrl":  metadataBackdrop(item.Metadata),
			"versionCount": item.VersionCount,
			"route":        "Ready",
		})
	}
	return output
}

func tvSeriesItems(items []catalog.SeriesListItem) []map[string]any {
	output := make([]map[string]any, 0, len(items))
	for _, item := range items {
		output = append(output, map[string]any{
			"id":           item.ID,
			"kind":         "series",
			"title":        item.Title,
			"subtitle":     fmt.Sprintf("%d season%s / %d episode%s", item.SeasonCount, plural(item.SeasonCount), item.EpisodeCount, plural(item.EpisodeCount)),
			"posterUrl":    metadataPoster(item.Metadata),
			"backdropUrl":  metadataBackdrop(item.Metadata),
			"seasonCount":  item.SeasonCount,
			"episodeCount": item.EpisodeCount,
			"route":        "Ready",
		})
	}
	return output
}

func tvRecentlyAddedItems(movies []catalog.MovieListItem, series []catalog.SeriesListItem, limit int) []map[string]any {
	output := []map[string]any{}
	output = append(output, tvMovieItems(movies)...)
	output = append(output, tvSeriesItems(series)...)
	if len(output) > limit {
		return output[:limit]
	}
	return output
}

func firstTVHomeItem(rows []map[string]any) map[string]any {
	for _, row := range rows {
		items, _ := row["items"].([]map[string]any)
		if len(items) > 0 {
			return items[0]
		}
	}
	return map[string]any{
		"id":       "empty",
		"kind":     "empty",
		"title":    "Add your first library",
		"subtitle": "Open Xuva Settings to add Movies or TV Shows.",
		"route":    "Setup",
	}
}

func metadataPoster(record *catalog.MetadataRecord) string {
	if record == nil {
		return ""
	}
	return normalizeArtworkSourceURL(record.PosterURL, "poster")
}

func metadataBackdrop(record *catalog.MetadataRecord) string {
	if record == nil {
		return ""
	}
	return normalizeArtworkSourceURL(record.BackdropURL, "backdrop")
}

func normalizeArtworkSourceURL(raw string, artType string) string {
	_ = artType
	source := strings.TrimSpace(raw)
	if source == "" {
		return ""
	}
	lower := strings.ToLower(source)
	if !strings.Contains(lower, "image.tmdb.org") {
		return source
	}
	parsed, err := url.Parse(source)
	if err != nil {
		return source
	}
	const prefix = "/t/p/"
	if !strings.HasPrefix(parsed.Path, prefix) {
		return source
	}
	parts := strings.SplitN(strings.TrimPrefix(parsed.Path, prefix), "/", 2)
	if len(parts) != 2 || strings.TrimSpace(parts[0]) == "" {
		return source
	}
	current := parts[0]
	if strings.EqualFold(current, "original") {
		return source
	}
	target := "original"
	if strings.EqualFold(current, target) {
		return source
	}
	parsed.Path = prefix + target + "/" + parts[1]
	return parsed.String()
}

func formatProgress(value float64) string {
	if value <= 0 {
		return "the beginning"
	}
	if value > 1 {
		value = 1
	}
	return fmt.Sprintf("%.0f%%", value*100)
}

func plural(value int) string {
	if value == 1 {
		return ""
	}
	return "s"
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
			"providers":   metadataProviders(r.Context(), deps),
			"strategy":    "Xuva runs strict managed metadata mode: local-first signals and account-free online sources stay active, managed providers auto-run when server credentials are provisioned, and fallback paths continue when limits or provider outages occur.",
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
			"providers": metadataProviders(r.Context(), deps),
		})
	}
}

func metadataProvidersHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{
			"managedMode": "strict",
			"providers":   metadataProviders(r.Context(), deps),
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

func metadataRefreshHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Metadata == nil {
			writeError(w, http.StatusServiceUnavailable, "metadata providers are not available")
			return
		}
		var request metaprovider.RefreshRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		result, err := deps.Metadata.Refresh(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func metadataRefreshBatchHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Metadata == nil {
			writeError(w, http.StatusServiceUnavailable, "metadata providers are not available")
			return
		}
		var request struct {
			Kind  string `json:"kind"`
			Limit int    `json:"limit"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}
		result, err := deps.Metadata.RefreshBatch(r.Context(), request.Kind, request.Limit)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, result)
	}
}

func metadataProviders(ctx context.Context, deps Deps) []map[string]any {
	cfg := currentConfig(deps)
	health := map[string]metaprovider.ProviderHealth{}
	if deps.Metadata != nil {
		health = deps.Metadata.ProviderHealth(ctx)
	}
	providers := []map[string]any{
		{"id": "filename", "name": "Filename and folders", "status": "active", "local": true},
		{"id": "manual", "name": "Manual correction", "status": "active", "local": true},
		{"id": "nfo", "name": "Local NFO", "status": "active", "local": true},
		{"id": "artwork", "name": "Poster and fanart sidecars", "status": "active", "local": true},
		{"id": "tvmaze", "name": "TVMaze", "status": "automatic", "local": false},
		{"id": "wikipedia", "name": "Wikipedia", "status": "automatic", "local": false},
		{"id": "wikidata", "name": "Wikidata", "status": "automatic", "local": false},
		{"id": "tmdb", "name": "TMDB", "status": "managed-ready", "local": false},
		{"id": "tvdb", "name": "TheTVDB", "status": "managed-ready", "local": false},
		{"id": "omdb", "name": "OMDb", "status": "managed-ready", "local": false},
	}
	for _, provider := range providers {
		id, _ := provider["id"].(string)
		if id == "" {
			continue
		}
		if id != "tmdb" && id != "tvdb" && id != "omdb" {
			continue
		}
		state, ok := health[id]
		if !ok {
			state = metaprovider.ProviderHealth{
				ID:         id,
				Managed:    true,
				Configured: managedProviderConfiguredForConfig(id, cfg),
				Healthy:    managedProviderConfiguredForConfig(id, cfg),
				Status:     "ready",
			}
			if !state.Configured {
				state.Status = "not_provisioned"
				state.Healthy = false
			}
		}
		provider["managedMode"] = "strict"
		provider["configured"] = state.Configured
		provider["healthy"] = state.Healthy
		provider["health"] = state
		provider["status"] = state.Status
	}
	return providers
}

func artworkHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		kind := r.PathValue("kind")
		id := r.PathValue("id")
		title := id
		artType := r.URL.Query().Get("type")
		if artType == "" {
			artType = "poster"
		}
		if !safePathSegment(kind) || !safePathSegment(id) || !safePathSegment(artType) {
			writeError(w, http.StatusBadRequest, "invalid artwork path")
			return
		}
		if artType != "poster" && artType != "backdrop" {
			writeError(w, http.StatusBadRequest, "invalid artwork type")
			return
		}
		if deps.Catalog != nil {
			if record, ok, err := deps.Catalog.GetBestMetadata(r.Context(), kind, id); err == nil && ok {
				if record.Title != "" {
					title = record.Title
				}
			}
			if records, err := deps.Catalog.ListMetadataRecords(r.Context(), kind, id); err == nil {
				candidates := metadataArtworkCandidates(records, artType)
				for _, candidate := range candidates {
					if serveArtwork(w, r, deps.Config.MetadataDir, kind, id, artType, candidate, true) {
						return
					}
				}
				for _, candidate := range candidates {
					if serveArtwork(w, r, deps.Config.MetadataDir, kind, id, artType, candidate, false) {
						return
					}
				}
			}
		}
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
		_, _ = io.WriteString(w, fallbackArtworkSVG(title, artType))
	}
}

func metadataArtworkCandidates(records []catalog.MetadataRecord, artType string) []string {
	output := []string{}
	seen := map[string]struct{}{}
	push := func(value string) {
		value = strings.TrimSpace(value)
		if value == "" {
			return
		}
		key := strings.ToLower(value)
		if _, exists := seen[key]; exists {
			return
		}
		seen[key] = struct{}{}
		output = append(output, value)
	}
	if strings.EqualFold(artType, "backdrop") {
		for _, record := range records {
			push(metadataBackdrop(&record))
		}
		for _, record := range records {
			push(metadataPoster(&record))
		}
		return output
	}
	for _, record := range records {
		push(metadataPoster(&record))
	}
	for _, record := range records {
		push(metadataBackdrop(&record))
	}
	return output
}

func fallbackArtworkSVG(title string, artType string) string {
	safeTitle := html.EscapeString(truncate(strings.TrimSpace(title), 20))
	if safeTitle == "" {
		safeTitle = "Xuva"
	}
	if strings.EqualFold(artType, "backdrop") {
		return `<svg xmlns="http://www.w3.org/2000/svg" width="1280" height="720" viewBox="0 0 1280 720">
<defs>
  <linearGradient id="bg" x1="0" x2="1" y1="0" y2="1">
    <stop stop-color="#0a1522"/>
    <stop offset="0.52" stop-color="#11233a"/>
    <stop offset="1" stop-color="#0d1a2c"/>
  </linearGradient>
  <radialGradient id="glow" cx="0.82" cy="0.14" r="0.6">
    <stop stop-color="#5bc2d6" stop-opacity="0.2"/>
    <stop offset="1" stop-color="#5bc2d6" stop-opacity="0"/>
  </radialGradient>
</defs>
<rect width="1280" height="720" fill="url(#bg)"/>
<rect width="1280" height="720" fill="url(#glow)"/>
</svg>`
	}
	return fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" width="600" height="900" viewBox="0 0 600 900">
<defs>
  <linearGradient id="g" x1="0" x2="1" y1="0" y2="1">
    <stop stop-color="#101f33"/>
    <stop offset="1" stop-color="#0a1522"/>
  </linearGradient>
  <radialGradient id="glow" cx="0.78" cy="0.16" r="0.55">
    <stop stop-color="#61d0e2" stop-opacity="0.24"/>
    <stop offset="1" stop-color="#61d0e2" stop-opacity="0"/>
  </radialGradient>
</defs>
<rect width="600" height="900" fill="url(#g)"/>
<rect width="600" height="900" fill="url(#glow)"/>
<rect x="26" y="26" width="548" height="848" rx="22" fill="none" stroke="#d7ebf8" stroke-opacity="0.12" stroke-width="2"/>
<text x="72" y="702" fill="#ecf5fc" fill-opacity="0.9" font-family="Inter,Segoe UI,sans-serif" font-size="40" font-weight="760">%s</text>
</svg>`, safeTitle)
}

func serveArtwork(w http.ResponseWriter, r *http.Request, metadataDir string, kind string, id string, artType string, source string, strict bool) bool {
	if serveLocalArtwork(w, r, source, artType, strict) {
		return true
	}
	return serveCachedArtwork(w, r, metadataDir, kind, id, artType, source, strict)
}

func serveLocalArtwork(w http.ResponseWriter, r *http.Request, path string, artType string, strict bool) bool {
	path = strings.TrimSpace(path)
	if path == "" {
		return false
	}
	lower := strings.ToLower(path)
	if strings.HasPrefix(lower, "http://") || strings.HasPrefix(lower, "https://") {
		return false
	}
	info, err := os.Stat(path)
	if err != nil || info.IsDir() {
		return false
	}
	switch strings.ToLower(filepath.Ext(path)) {
	case ".jpg", ".jpeg", ".png", ".webp":
		if !artworkPassesQualityGatePath(path, artType, strict) {
			return false
		}
		http.ServeFile(w, r, path)
		return true
	default:
		return false
	}
}

func serveCachedArtwork(w http.ResponseWriter, r *http.Request, metadataDir string, kind string, id string, artType string, sourceURL string, strict bool) bool {
	if !strings.HasPrefix(strings.ToLower(sourceURL), "https://") && !strings.HasPrefix(strings.ToLower(sourceURL), "http://") {
		return false
	}
	if metadataDir == "" {
		return false
	}
	dir, ok := safeChildPath(metadataDir, "artwork", kind, id)
	if !ok {
		return false
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return false
	}
	normalizedSource := strings.TrimSpace(sourceURL)
	sourceMarkerPath, ok := safeChildPath(dir, artType+".source")
	if !ok {
		return false
	}
	cachedSource := readCachedArtworkSource(sourceMarkerPath)
	refreshBecauseSourceChanged := cachedSource != "" && !strings.EqualFold(cachedSource, normalizedSource)
	refreshTMDBCacheWithoutMarker := cachedSource == "" && strings.Contains(strings.ToLower(normalizedSource), "image.tmdb.org/t/p/")
	for _, ext := range []string{".jpg", ".png", ".webp"} {
		path, ok := safeChildPath(dir, artType+ext)
		if !ok {
			return false
		}
		if _, err := os.Stat(path); err == nil {
			if refreshBecauseSourceChanged || refreshTMDBCacheWithoutMarker {
				_ = os.Remove(path)
				continue
			}
			if !artworkPassesQualityGatePath(path, artType, strict) {
				return false
			}
			http.ServeFile(w, r, path)
			return true
		}
	}
	ctx, cancel := context.WithTimeout(r.Context(), 15*time.Second)
	defer cancel()
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, sourceURL, nil)
	if err != nil {
		return false
	}
	request.Header.Set("User-Agent", "Xuva/0.1 (+https://github.com/xuvahq/xuva)")
	request.Header.Set("Accept", "image/avif,image/webp,image/apng,image/*,*/*;q=0.8")
	response, err := http.DefaultClient.Do(request)
	if err != nil {
		return false
	}
	defer response.Body.Close()
	if response.StatusCode < 200 || response.StatusCode > 299 {
		return false
	}
	ext := artworkExtension(response.Header.Get("Content-Type"), sourceURL)
	path, ok := safeChildPath(dir, artType+ext)
	if !ok {
		return false
	}
	payload, err := io.ReadAll(io.LimitReader(response.Body, 32*1024*1024))
	if err != nil {
		return false
	}
	if !artworkPassesQualityGateBytes(payload, artType, sourceURL, strict) {
		return false
	}
	file, err := os.Create(path)
	if err != nil {
		return false
	}
	if _, err := io.Copy(file, bytes.NewReader(payload)); err != nil {
		_ = file.Close()
		_ = os.Remove(path)
		return false
	}
	if err := file.Close(); err != nil {
		_ = os.Remove(path)
		return false
	}
	writeCachedArtworkSource(sourceMarkerPath, normalizedSource)
	http.ServeFile(w, r, path)
	return true
}

func readCachedArtworkSource(path string) string {
	raw, err := os.ReadFile(path)
	if err != nil {
		return ""
	}
	return strings.TrimSpace(string(raw))
}

func writeCachedArtworkSource(path string, source string) {
	_ = os.WriteFile(path, []byte(strings.TrimSpace(source)), 0o644)
}

func artworkExtension(contentType string, sourceURL string) string {
	contentType = strings.ToLower(contentType)
	switch {
	case strings.Contains(contentType, "png"):
		return ".png"
	case strings.Contains(contentType, "webp"):
		return ".webp"
	case strings.Contains(contentType, "jpeg"), strings.Contains(contentType, "jpg"):
		return ".jpg"
	}
	ext := strings.ToLower(filepath.Ext(strings.Split(sourceURL, "?")[0]))
	if ext == ".png" || ext == ".webp" || ext == ".jpg" || ext == ".jpeg" {
		if ext == ".jpeg" {
			return ".jpg"
		}
		return ext
	}
	return ".jpg"
}

func artworkPassesQualityGatePath(path string, artType string, strict bool) bool {
	data, err := os.ReadFile(path)
	if err != nil {
		return false
	}
	return artworkPassesQualityGateBytes(data, artType, path, strict)
}

func artworkPassesQualityGateBytes(payload []byte, artType string, sourceHint string, strict bool) bool {
	minWidth, minHeight := artworkQualityThresholds(artType, strict)
	config, _, err := image.DecodeConfig(bytes.NewReader(payload))
	if err == nil {
		return config.Width >= minWidth && config.Height >= minHeight
	}
	widthHint, heightHint := artworkDimensionHints(sourceHint, artType)
	if widthHint <= 0 || heightHint <= 0 {
		return false
	}
	return widthHint >= minWidth && heightHint >= minHeight
}

func artworkQualityThresholds(artType string, strict bool) (minWidth int, minHeight int) {
	if !strict {
		if strings.EqualFold(artType, "backdrop") {
			return 480, 270
		}
		return 180, 260
	}
	if strings.EqualFold(artType, "backdrop") {
		return 1280, 720
	}
	return 600, 900
}

func artworkDimensionHints(source string, artType string) (int, int) {
	value := strings.ToLower(strings.TrimSpace(source))
	if value == "" {
		return 0, 0
	}
	if strings.Contains(value, "image.tmdb.org/t/p/") {
		return tmdbArtworkHint(value, artType)
	}
	if strings.Contains(value, "tvmaze.com") {
		if strings.Contains(value, "/original") || strings.Contains(value, "original_") {
			return 1280, 720
		}
		if strings.Contains(value, "/medium") || strings.Contains(value, "medium_") {
			if strings.EqualFold(artType, "backdrop") {
				return 0, 0
			}
			return 210, 295
		}
	}
	if strings.Contains(value, "/thumb/") {
		if width := wikipediaThumbWidth(value); width > 0 {
			if strings.EqualFold(artType, "backdrop") {
				return width, int(float64(width) * 9.0 / 16.0)
			}
			return width, int(float64(width) * 1.5)
		}
	}
	return 0, 0
}

func tmdbArtworkHint(source string, artType string) (int, int) {
	const prefix = "/t/p/"
	parsed, err := url.Parse(source)
	if err != nil {
		return 0, 0
	}
	if !strings.HasPrefix(parsed.Path, prefix) {
		return 0, 0
	}
	parts := strings.SplitN(strings.TrimPrefix(parsed.Path, prefix), "/", 2)
	if len(parts) != 2 {
		return 0, 0
	}
	size := strings.TrimSpace(parts[0])
	switch strings.ToLower(size) {
	case "original":
		if strings.EqualFold(artType, "backdrop") {
			return 1920, 1080
		}
		return 1000, 1500
	}
	if !strings.HasPrefix(strings.ToLower(size), "w") {
		return 0, 0
	}
	width, err := strconv.Atoi(strings.TrimPrefix(strings.ToLower(size), "w"))
	if err != nil || width <= 0 {
		return 0, 0
	}
	if strings.EqualFold(artType, "backdrop") {
		return width, int(float64(width) * 9.0 / 16.0)
	}
	return width, int(float64(width) * 1.5)
}

func wikipediaThumbWidth(source string) int {
	const marker = "px-"
	index := strings.Index(source, marker)
	if index <= 0 {
		return 0
	}
	start := index - 1
	for start >= 0 {
		char := source[start]
		if char < '0' || char > '9' {
			break
		}
		start--
	}
	width, err := strconv.Atoi(source[start+1 : index])
	if err != nil || width <= 0 {
		return 0
	}
	return width
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
			"profile":              profile,
			"playbackPolicy":       playbackPolicyStatus(deps.Config.PlaybackPolicy),
			"limits":               deps.Resources.Limits(),
			"queues":               deps.Jobs.Snapshot(),
			"libraries":            libraryList,
			"storageDefaults":      storageDefaults(libraryList),
			"hardwareAcceleration": hardwareAccelerationStatus(deps.Config),
			"recommendations": []string{
				"Keep scan concurrency low for network/removable storage",
				"Probe jobs are isolated from scan and transcode queues",
				"Playback-critical work is separated from background jobs",
			},
		})
	}
}

func storageDefaults(libraryList []libraries.Library) map[string]libraries.WorkerDefaults {
	output := map[string]libraries.WorkerDefaults{}
	for _, library := range libraryList {
		output[library.ID] = libraries.DefaultsForStorage(library.StorageType)
	}
	return output
}

func metadataSourceCatalogPayload(ctx context.Context, cfg config.Config, service *metaprovider.Service) map[string][]map[string]any {
	catalog := metasources.SourceCatalogByKind(cfg)
	health := map[string]metaprovider.ProviderHealth{}
	if service != nil {
		health = service.ProviderHealth(ctx)
	}
	output := map[string][]map[string]any{
		"movie":  {},
		"series": {},
	}
	for kind, definitions := range catalog {
		rows := make([]map[string]any, 0, len(definitions))
		for _, definition := range definitions {
			item := map[string]any{
				"id":             definition.ID,
				"name":           definition.Name,
				"description":    definition.Description,
				"coverage":       definition.Coverage,
				"note":           definition.Note,
				"kinds":          definition.Kinds,
				"local":          definition.Local,
				"managed":        definition.Managed,
				"requiresConfig": definition.RequiresConfig,
				"available":      definition.Available,
			}
			if definition.Managed {
				state, ok := health[definition.ID]
				if !ok {
					state = metaprovider.ProviderHealth{
						ID:         definition.ID,
						Managed:    true,
						Configured: managedProviderConfiguredForConfig(definition.ID, cfg),
						Healthy:    definition.Available,
						Status:     "ready",
					}
					if !state.Configured {
						state.Status = "not_provisioned"
						state.Healthy = false
					}
				}
				item["available"] = state.Configured
				item["providerHealth"] = state
				item["runtimeReady"] = state.Healthy
				item["status"] = state.Status
			}
			rows = append(rows, item)
		}
		output[kind] = rows
	}
	return output
}

func metadataSourcePreferencesPayload(cfg config.Config) map[string][]string {
	return map[string][]string{
		"movie":  configuredMetadataSourceOrder(cfg, "movie"),
		"series": configuredMetadataSourceOrder(cfg, "series"),
	}
}

func configuredMetadataSourceOrder(cfg config.Config, kind string) []string {
	switch metasources.NormalizeKind(kind) {
	case "series":
		return metasources.NormalizeRequestedSourceOrder("series", cfg.SeriesMetadataSources)
	default:
		return metasources.NormalizeRequestedSourceOrder("movie", cfg.MovieMetadataSources)
	}
}

func defaultMetadataSourcePreferenceForLibrary(cfg config.Config, kind libraries.Kind) []string {
	switch kind {
	case libraries.KindTV:
		return configuredMetadataSourceOrder(cfg, "series")
	default:
		return configuredMetadataSourceOrder(cfg, "movie")
	}
}

func managedProviderConfiguredForConfig(provider string, cfg config.Config) bool {
	return metaprovider.ManagedProviderConfigured(provider, cfg)
}

func settingsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		cfg := currentConfig(deps)
		health := map[string]metaprovider.ProviderHealth{}
		if deps.Metadata != nil {
			health = deps.Metadata.ProviderHealth(r.Context())
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"config":                    settingsPayload(cfg),
			"runtimePaths":              runtimePaths(cfg),
			"libraries":                 deps.Libraries.List(),
			"metadataSources":           metadataSourceCatalogPayload(r.Context(), cfg, deps.Metadata),
			"metadataSourcePreferences": metadataSourcePreferencesPayload(cfg),
			"metadataHealth":            health,
			"managedMode":               "strict",
		})
	}
}

func settingsFolderBrowseHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		path := strings.TrimSpace(r.URL.Query().Get("path"))
		if path == "" {
			path = currentConfig(deps).DataDir
		}
		payload, status := browseFolderPayload(path)
		writeJSON(w, status, payload)
	}
}

func settingsUpdateHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request config.Config
		raw, err := io.ReadAll(r.Body)
		if err != nil {
			writeError(w, http.StatusBadRequest, "invalid json body")
			return
		}
		if len(strings.TrimSpace(string(raw))) > 0 {
			if err := json.Unmarshal(raw, &request); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json body")
				return
			}
		}
		fields := map[string]json.RawMessage{}
		if len(strings.TrimSpace(string(raw))) > 0 {
			if err := json.Unmarshal(raw, &fields); err != nil {
				writeError(w, http.StatusBadRequest, "invalid json body")
				return
			}
		}
		updated := currentConfig(deps)
		if value, ok := fields["serverName"]; ok {
			var serverName string
			if err := json.Unmarshal(value, &serverName); err != nil {
				writeError(w, http.StatusBadRequest, "server name must be text")
				return
			}
			normalized, err := config.NormalizeServerName(serverName)
			if err != nil {
				writeError(w, http.StatusBadRequest, err.Error())
				return
			}
			updated.ServerName = normalized
		}
		mergeString(&updated.HTTPAddr, request.HTTPAddr)
		mergeString(&updated.DataDir, request.DataDir)
		mergeString(&updated.TranscodeDir, request.TranscodeDir)
		mergeString(&updated.DownloadsDir, request.DownloadsDir)
		mergeString(&updated.MetadataDir, request.MetadataDir)
		mergeString(&updated.CacheDir, request.CacheDir)
		mergeString(&updated.TempDir, request.TempDir)
		mergeString(&updated.FFmpegPath, request.FFmpegPath)
		mergeString(&updated.FFprobePath, request.FFprobePath)
		mergeInt(&updated.EventBuffer, request.EventBuffer)
		mergeInt(&updated.ScanWorkers, request.ScanWorkers)
		mergeInt(&updated.ProbeWorkers, request.ProbeWorkers)
		mergeInt(&updated.TranscodeWorkers, request.TranscodeWorkers)
		mergeInt(&updated.GPUWorkers, request.GPUWorkers)
		if value, ok := fields["hardwareUnlocked"]; ok {
			var hardwareUnlocked bool
			if err := json.Unmarshal(value, &hardwareUnlocked); err != nil {
				writeError(w, http.StatusBadRequest, "hardware unlocked must be true or false")
				return
			}
			updated.HardwareUnlocked = hardwareUnlocked
		}
		mergeString(&updated.LibrarySyncMode, request.LibrarySyncMode)
		mergeString(&updated.PlaybackPolicy, request.PlaybackPolicy)
		updated.PlaybackPolicy = normalizedPlaybackPolicy(updated.PlaybackPolicy)
		mergeInt(&updated.SyncIntervalMins, request.SyncIntervalMins)
		mergeInt(&updated.WatchDebounceSecs, request.WatchDebounceSecs)
		mergeInt(&updated.ProbeBatchLimit, request.ProbeBatchLimit)
		if request.AllowedOrigins != nil {
			updated.AllowedOrigins = request.AllowedOrigins
		}
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
			EventBuffer:      updated.EventBuffer,
			ScanWorkers:      updated.ScanWorkers,
			ProbeWorkers:     updated.ProbeWorkers,
			TranscodeWorkers: updated.TranscodeWorkers,
			GPUWorkers:       updated.GPUWorkers,
		}); err != nil {
			writeError(w, http.StatusInternalServerError, "settings save failed")
			return
		}
		deps.Events.Publish("settings.updated", settingsPayload(updated))
		publishDomainAudit(deps, r, "audit.settings", "settings.update", "allowed", map[string]any{
			"restartRequired": true,
			"playbackPolicy":  updated.PlaybackPolicy,
			"syncInterval":    updated.SyncIntervalMins,
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"config":          settingsPayload(updated),
			"runtimePaths":    runtimePaths(updated),
			"restartRequired": true,
		})
	}
}

func metadataSourceSettingsUpdateHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request struct {
			Movie  []string `json:"movie"`
			Series []string `json:"series"`
		}
		if !decodeJSON(w, r, &request) {
			return
		}

		movie := metasources.NormalizeRequestedSourceOrder("movie", request.Movie)
		series := metasources.NormalizeRequestedSourceOrder("series", request.Series)
		if len(movie) == 0 {
			writeError(w, http.StatusBadRequest, "choose at least one movie metadata source")
			return
		}
		if len(series) == 0 {
			writeError(w, http.StatusBadRequest, "choose at least one TV metadata source")
			return
		}

		updated := currentConfig(deps)
		updated.MovieMetadataSources = append([]string(nil), movie...)
		updated.SeriesMetadataSources = append([]string(nil), series...)
		if err := config.SaveFile(deps.Config.DataDir, updated); err != nil {
			writeError(w, http.StatusInternalServerError, "metadata source settings save failed")
			return
		}

		librariesList, err := deps.Catalog.ListLibraries(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "metadata source settings save failed")
			return
		}
		for _, library := range librariesList {
			switch library.Kind {
			case libraries.KindMovies:
				library.MetadataSources = append([]string(nil), movie...)
			case libraries.KindTV:
				library.MetadataSources = append([]string(nil), series...)
			default:
				continue
			}
			saved, saveErr := deps.Catalog.SaveLibrary(r.Context(), library)
			if saveErr != nil {
				writeError(w, http.StatusInternalServerError, "metadata source settings save failed")
				return
			}
			deps.Libraries.Set(saved)
			deps.Events.Publish("library.updated", saved)
		}

		payload := metadataSourcePreferencesPayload(updated)
		deps.Events.Publish("settings.updated", map[string]any{"metadataSourcePreferences": payload})
		publishDomainAudit(deps, r, "audit.settings", "settings.metadata_sources.update", "allowed", map[string]any{
			"movieSources":  payload["movie"],
			"seriesSources": payload["series"],
		})
		writeJSON(w, http.StatusOK, map[string]any{
			"metadataSourcePreferences": payload,
			"metadataSources":           metadataSourceCatalogPayload(r.Context(), updated, deps.Metadata),
			"restartRequired":           false,
		})
	}
}

func browseFolderPayload(path string) (map[string]any, int) {
	resolved, err := resolveBrowsePath(path)
	if err != nil {
		return map[string]any{
			"path":  path,
			"roots": browseRoots(),
			"error": err.Error(),
		}, http.StatusBadRequest
	}
	info, err := os.Stat(resolved)
	if err != nil {
		return map[string]any{
			"path":   resolved,
			"parent": parentFolder(resolved),
			"roots":  browseRoots(),
			"error":  err.Error(),
		}, http.StatusOK
	}
	if !info.IsDir() {
		return map[string]any{
			"path":   resolved,
			"parent": parentFolder(resolved),
			"roots":  browseRoots(),
			"error":  "path is not a folder",
		}, http.StatusOK
	}
	entries, err := os.ReadDir(resolved)
	if err != nil {
		return map[string]any{
			"path":     resolved,
			"parent":   parentFolder(resolved),
			"roots":    browseRoots(),
			"writable": false,
			"error":    err.Error(),
		}, http.StatusOK
	}
	folders := make([]map[string]any, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		name := entry.Name()
		folders = append(folders, map[string]any{
			"name": name,
			"path": filepath.Join(resolved, name),
		})
	}
	sort.Slice(folders, func(i, j int) bool {
		return strings.ToLower(folders[i]["name"].(string)) < strings.ToLower(folders[j]["name"].(string))
	})
	writable, message := pathReady(resolved)
	return map[string]any{
		"path":     resolved,
		"parent":   parentFolder(resolved),
		"roots":    browseRoots(),
		"entries":  folders,
		"writable": writable,
		"message":  message,
	}, http.StatusOK
}

func resolveBrowsePath(path string) (string, error) {
	path = strings.TrimSpace(path)
	if path == "" {
		return filepath.Abs(".")
	}
	cleaned := filepath.Clean(os.ExpandEnv(path))
	if filepath.IsAbs(cleaned) {
		return cleaned, nil
	}
	return filepath.Abs(cleaned)
}

func parentFolder(path string) string {
	parent := filepath.Dir(path)
	if parent == path || parent == "." {
		return ""
	}
	return parent
}

func browseRoots() []map[string]string {
	roots := []map[string]string{}
	if runtime.GOOS == "windows" {
		for drive := 'A'; drive <= 'Z'; drive++ {
			root := string(drive) + `:\`
			if info, err := os.Stat(root); err == nil && info.IsDir() {
				roots = append(roots, map[string]string{"name": strings.TrimSuffix(root, `\`), "path": root})
			}
		}
		return roots
	}
	roots = append(roots, map[string]string{"name": "Root", "path": string(filepath.Separator)})
	if home, err := os.UserHomeDir(); err == nil && home != "" {
		roots = append(roots, map[string]string{"name": "Home", "path": home})
	}
	return roots
}

func hardwareAccelerationStatus(cfg config.Config) map[string]any {
	encoders, err := detectHardwareEncoders(cfg.FFmpegPath)
	available := len(encoders) > 0
	unlockState := "locked"
	if cfg.HardwareUnlocked {
		unlockState = "unlocked"
	}
	status := "not_detected"
	if err != nil {
		status = "unknown"
	}
	if available {
		status = "available"
	}
	result := map[string]any{
		"status":         status,
		"available":      available,
		"unlockState":    unlockState,
		"gpuWorkers":     cfg.GPUWorkers,
		"encoders":       encoders,
		"recommendation": hardwareAccelerationRecommendation(available, cfg.GPUWorkers),
	}
	if err != nil {
		result["error"] = err.Error()
	}
	return result
}

func hardwareAccelerationRecommendation(available bool, gpuWorkers int) string {
	if !available {
		return "No FFmpeg hardware encoder support was detected. Video conversion will fall back to CPU until a compatible GPU driver and FFmpeg build are available."
	}
	if gpuWorkers <= 0 {
		return "FFmpeg exposes hardware encoder support, but GPU worker slots are disabled. Enable one or more slots to reserve GPU conversion capacity."
	}
	return "FFmpeg exposes hardware encoder support. Once licensed and runtime-tested, Xuva can use GPU conversion for heavy video routes and subtitle burn-in."
}

func playbackPolicyStatus(policy string) map[string]any {
	policy = normalizedPlaybackPolicy(policy)
	labels := map[string]string{
		"original_only": "Original Only",
		"light":         "Light Compatibility",
		"full":          "Full Compatibility",
		"cinema":        "Cinema Server",
	}
	descriptions := map[string]string{
		"original_only": "Xuva plays the original file only. If this device cannot play it as-is, Xuva shows fallback options instead of converting automatically.",
		"light":         "Xuva may repackage while playing or convert audio. Video stays untouched, so quality is preserved.",
		"full":          "Xuva may convert video while playing when a device needs it. Work is temporary unless the user creates an optimized version.",
		"cinema":        "Xuva allows heavier live conversion and future automated optimization controls for power users.",
	}
	return map[string]any{
		"id":          policy,
		"label":       labels[policy],
		"description": descriptions[policy],
	}
}

func normalizedPlaybackPolicy(policy string) string {
	switch policy {
	case "light", "full", "cinema":
		return policy
	default:
		return "original_only"
	}
}

func playbackPolicyAllows(policy string, decision playback.Decision) bool {
	switch decision.Mode {
	case playback.DirectPlay, playback.DecisionDeferred:
		return true
	}
	switch normalizedPlaybackPolicy(policy) {
	case "light":
		return decision.Mode == playback.Remux || decision.Mode == playback.AudioTranscode
	case "full", "cinema":
		return decision.Mode == playback.Remux || decision.Mode == playback.AudioTranscode || decision.Mode == playback.AdaptiveStream || decision.Mode == playback.VideoTranscode || decision.Mode == playback.SubtitleBurn
	default:
		return false
	}
}

func playbackPolicyFallbacks(policy string, decision playback.Decision) []map[string]string {
	mode := string(decision.Mode)
	fallbacks := []map[string]string{
		{"label": "Play on a compatible device", "detail": "Use a player that supports this file as-is so Xuva does not need to convert anything."},
		{"label": "Allow this session to adapt", "detail": "Switch to a compatibility policy that permits the required playback work: " + mode + "."},
	}
	if decision.Mode == playback.Remux {
		fallbacks = append(fallbacks, map[string]string{"label": "Allow live repackage", "detail": "Repackage while playing. No video quality loss and no permanent copy."})
	}
	if decision.Mode == playback.AudioTranscode {
		fallbacks = append(fallbacks, map[string]string{"label": "Allow audio conversion", "detail": "Convert audio while playing. Video remains untouched and temporary work is discarded."})
	}
	if decision.Mode == playback.VideoTranscode || decision.Mode == playback.SubtitleBurn {
		fallbacks = append(fallbacks, map[string]string{"label": "Allow live video conversion", "detail": "Convert video only while playing. This may use high CPU or GPU if hardware acceleration is unlocked and working."})
	}
	fallbacks = append(fallbacks, map[string]string{"label": "Create optimized version", "detail": "Optional stored version for easier future playback. Xuva should show size, quality, and storage impact first."})
	return fallbacks
}

func hardwareTestHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		encoders, err := detectHardwareEncoders(deps.Config.FFmpegPath)
		if err != nil {
			writeJSON(w, http.StatusOK, map[string]any{
				"status": "failed",
				"error":  err.Error(),
				"tests":  []map[string]any{},
			})
			return
		}
		tests := make([]map[string]any, 0, len(encoders))
		working := 0
		for _, encoder := range encoders {
			id := encoder["id"]
			result := testHardwareEncoder(r.Context(), deps.Config.FFmpegPath, id)
			if ok, _ := result["ok"].(bool); ok {
				working++
			}
			for key, value := range encoder {
				result[key] = value
			}
			tests = append(tests, result)
		}
		status := "failed"
		if working > 0 {
			status = "passed"
		}
		writeJSON(w, http.StatusOK, map[string]any{
			"status":  status,
			"working": working,
			"tested":  len(tests),
			"tests":   tests,
		})
	}
}

func testHardwareEncoder(parent context.Context, ffmpegPath string, encoder string) map[string]any {
	if strings.TrimSpace(ffmpegPath) == "" {
		ffmpegPath = "ffmpeg"
	}
	started := time.Now()
	ctx, cancel := context.WithTimeout(parent, 10*time.Second)
	defer cancel()
	args := []string{
		"-hide_banner", "-loglevel", "error",
		"-f", "lavfi", "-i", "testsrc2=size=128x72:rate=1",
		"-frames:v", "1", "-an",
		"-c:v", encoder,
		"-f", "null", os.DevNull,
	}
	output, err := exec.CommandContext(ctx, ffmpegPath, args...).CombinedOutput()
	result := map[string]any{
		"ok":         err == nil && ctx.Err() == nil,
		"durationMs": time.Since(started).Milliseconds(),
	}
	if ctx.Err() != nil {
		result["error"] = ctx.Err().Error()
	} else if err != nil {
		message := strings.TrimSpace(string(output))
		if message == "" {
			message = err.Error()
		}
		if len(message) > 500 {
			message = message[:500]
		}
		result["error"] = message
	}
	return result
}

func detectHardwareEncoders(ffmpegPath string) ([]map[string]string, error) {
	if strings.TrimSpace(ffmpegPath) == "" {
		ffmpegPath = "ffmpeg"
	}
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
	defer cancel()
	output, err := exec.CommandContext(ctx, ffmpegPath, "-hide_banner", "-encoders").CombinedOutput()
	if ctx.Err() != nil {
		return nil, ctx.Err()
	}
	if err != nil {
		return nil, fmt.Errorf("ffmpeg encoder scan failed: %w", err)
	}
	text := strings.ToLower(string(output))
	candidates := []struct {
		id     string
		label  string
		vendor string
		codec  string
	}{
		{"h264_qsv", "H.264 Intel Quick Sync", "Intel Quick Sync", "H.264"},
		{"hevc_qsv", "HEVC Intel Quick Sync", "Intel Quick Sync", "HEVC"},
		{"av1_qsv", "AV1 Intel Quick Sync", "Intel Quick Sync", "AV1"},
		{"h264_nvenc", "H.264 NVIDIA NVENC", "NVIDIA NVENC", "H.264"},
		{"hevc_nvenc", "HEVC NVIDIA NVENC", "NVIDIA NVENC", "HEVC"},
		{"av1_nvenc", "AV1 NVIDIA NVENC", "NVIDIA NVENC", "AV1"},
		{"h264_amf", "H.264 AMD AMF", "AMD AMF", "H.264"},
		{"hevc_amf", "HEVC AMD AMF", "AMD AMF", "HEVC"},
		{"av1_amf", "AV1 AMD AMF", "AMD AMF", "AV1"},
		{"h264_vaapi", "H.264 VAAPI", "VAAPI", "H.264"},
		{"hevc_vaapi", "HEVC VAAPI", "VAAPI", "HEVC"},
		{"av1_vaapi", "AV1 VAAPI", "VAAPI", "AV1"},
		{"h264_videotoolbox", "H.264 VideoToolbox", "Apple VideoToolbox", "H.264"},
		{"hevc_videotoolbox", "HEVC VideoToolbox", "Apple VideoToolbox", "HEVC"},
	}
	encoders := make([]map[string]string, 0)
	for _, candidate := range candidates {
		if strings.Contains(text, candidate.id) {
			encoders = append(encoders, map[string]string{
				"id":     candidate.id,
				"label":  candidate.label,
				"vendor": candidate.vendor,
				"codec":  candidate.codec,
			})
		}
	}
	return encoders, nil
}

func selectedHardwareEncoder(ctx context.Context, cfg config.Config) (string, bool) {
	if !cfg.HardwareUnlocked || cfg.GPUWorkers <= 0 {
		return "", false
	}
	encoders, err := detectHardwareEncoders(cfg.FFmpegPath)
	if err != nil {
		return "", false
	}
	preferred := []string{"h264_qsv", "h264_nvenc", "h264_amf", "h264_vaapi", "h264_videotoolbox"}
	for _, id := range preferred {
		for _, encoder := range encoders {
			if encoder["id"] == id {
				result := testHardwareEncoder(ctx, cfg.FFmpegPath, id)
				if ok, _ := result["ok"].(bool); ok {
					return id, true
				}
			}
		}
	}
	return "", false
}

func systemStatusHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, systemstats.Collect(runtimePaths(currentConfig(deps))))
	}
}

func currentConfig(deps Deps) config.Config {
	if saved, err := config.LoadFile(deps.Config.DataDir); err == nil {
		return config.Merge(deps.Config, saved)
	}
	return deps.Config
}

func settingsPayload(cfg config.Config) map[string]any {
	return map[string]any{
		"serverName":        configDisplayName(cfg.ServerName),
		"httpAddr":          cfg.HTTPAddr,
		"dataDir":           cfg.DataDir,
		"transcodeDir":      cfg.TranscodeDir,
		"downloadsDir":      cfg.DownloadsDir,
		"metadataDir":       cfg.MetadataDir,
		"cacheDir":          cfg.CacheDir,
		"tempDir":           cfg.TempDir,
		"ffmpegPath":        cfg.FFmpegPath,
		"ffprobePath":       cfg.FFprobePath,
		"eventBuffer":       cfg.EventBuffer,
		"scanWorkers":       cfg.ScanWorkers,
		"probeWorkers":      cfg.ProbeWorkers,
		"transcodeWorkers":  cfg.TranscodeWorkers,
		"gpuWorkers":        cfg.GPUWorkers,
		"hardwareUnlocked":  cfg.HardwareUnlocked,
		"playbackPolicy":    cfg.PlaybackPolicy,
		"librarySyncMode":   cfg.LibrarySyncMode,
		"syncIntervalMins":  cfg.SyncIntervalMins,
		"watchDebounceSecs": cfg.WatchDebounceSecs,
		"probeBatchLimit":   cfg.ProbeBatchLimit,
		"allowedOrigins":    cfg.AllowedOrigins,
		"metadataProviders": map[string]any{
			"automatic": []map[string]any{
				{"id": "filename", "name": "Filename and folders", "coverage": "All libraries", "note": "Always on"},
				{"id": "nfo", "name": "Local NFO", "coverage": "Movies and TV with sidecars", "note": "Always on"},
				{"id": "artwork", "name": "Poster and fanart sidecars", "coverage": "Movies and TV with local images", "note": "Always on"},
				{"id": "tvmaze", "name": "TVMaze", "coverage": "Series metadata and TV ratings", "note": "No user account required"},
				{"id": "wikipedia", "name": "Wikipedia", "coverage": "Movie and series summaries and art where available", "note": "No user account required"},
				{"id": "wikidata", "name": "Wikidata", "coverage": "Movie and series labels, external IDs, and Wikimedia artwork", "note": "No user account required"},
			},
			"managedOverrides": []map[string]any{
				{"id": "omdb", "name": "OMDb", "configured": managedProviderConfiguredForConfig("omdb", cfg), "coverage": "IMDb, Rotten Tomatoes, Metacritic"},
				{"id": "tmdb", "name": "TMDB", "configured": managedProviderConfiguredForConfig("tmdb", cfg), "coverage": "TMDB community ratings and IDs"},
				{"id": "tvdb", "name": "TheTVDB", "configured": managedProviderConfiguredForConfig("tvdb", cfg), "coverage": "TV and movie metadata, IDs, and ratings"},
			},
		},
	}
}

func configDisplayName(value string) string {
	if strings.TrimSpace(value) == "My Server" {
		return "Xuva"
	}
	normalized, err := config.NormalizeServerName(value)
	if err != nil {
		return "Xuva"
	}
	return normalized
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
			"diagnostics":    "available",
			"failureClasses": []string{remote.ClassNotConfigured, remote.ClassPrivateRoute, remote.ClassDNS, remote.ClassNATFirewall, remote.ClassCertificate, remote.ClassThroughput},
			"recommendation": "Use your own VPN, reverse proxy, or port-forwarding setup. Xuva does not require hosted relay servers.",
		})
	}
}

func remoteDiagnosticsHandler(deps Deps) http.HandlerFunc {
	checker := remote.NewChecker()
	return func(w http.ResponseWriter, r *http.Request) {
		var payload remote.Request
		if !decodeJSON(w, r, &payload) {
			return
		}
		result := checker.Diagnose(r.Context(), payload, lanAddresses(deps.Config.HTTPAddr))
		writeJSON(w, http.StatusOK, result)
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
		publishOperationalEvent(deps, r, "api.probe.accepted", map[string]any{
			"jobId":         job.ID,
			"mediaSourceId": job.MediaSourceID,
			"limit":         job.Limit,
		})
		writeJSON(w, http.StatusAccepted, job)
	}
}

func mediaSourceStreamHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		release, ok := authorizeStreamRequest(deps, w, r, r.PathValue("id"))
		if !ok {
			return
		}
		defer release()
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

func mediaSourceStreamTokenHandler(deps Deps) http.HandlerFunc {
	type request struct {
		SessionID string `json:"sessionId"`
		DeviceID  string `json:"deviceId"`
	}
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Streaming == nil {
			writeError(w, http.StatusServiceUnavailable, "stream signing is not available")
			return
		}
		resolved, ok := auth.ResolvedSessionFromContext(r.Context())
		if !ok {
			writeError(w, http.StatusUnauthorized, "authentication required")
			return
		}
		var payload request
		if !decodeJSON(w, r, &payload) {
			return
		}
		session, ok := deps.Sessions.Get(payload.SessionID)
		if !ok {
			writeError(w, http.StatusForbidden, "playback session is not active")
			return
		}
		if session.MediaSourceID != r.PathValue("id") || session.UserID != resolved.Principal.ID {
			writeError(w, http.StatusForbidden, "playback session does not match this media source")
			return
		}
		deviceID := firstNonEmpty(payload.DeviceID, session.DeviceID)
		if deviceID == "" || deviceID != session.DeviceID {
			writeError(w, http.StatusForbidden, "playback device does not match this session")
			return
		}
		token, claims, err := deps.Streaming.Issue(streaming.Expected{
			MediaSourceID: session.MediaSourceID,
			SessionID:     session.ID,
			UserID:        session.UserID,
			DeviceID:      session.DeviceID,
		}, 0)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "stream token issue failed")
			return
		}
		query := "?sessionId=" + url.QueryEscape(session.ID) + "&deviceId=" + url.QueryEscape(session.DeviceID) + "&token=" + url.QueryEscape(token)
		writeJSON(w, http.StatusOK, map[string]any{
			"token":           token,
			"expiresAt":       time.Unix(claims.ExpiresAt, 0).UTC().Format(time.RFC3339),
			"streamUrl":       "/api/media-sources/" + session.MediaSourceID + "/stream" + query,
			"subtitleBaseUrl": "/api/media-sources/" + session.MediaSourceID + "/subtitles/",
			"query":           query,
		})
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
		title := item.Name
		if display, ok, err := deps.Catalog.GetMediaSourceDisplay(r.Context(), id); err != nil {
			writeError(w, http.StatusInternalServerError, "media source display lookup failed")
			return
		} else if ok && display.Title != "" {
			title = display.Title
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprintf(w, `<!doctype html>
<html lang="en">
<head>
  <meta charset="utf-8">
  <meta name="viewport" content="width=device-width, initial-scale=1">
  <title>%s - Xuva</title>
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
      left:50%%;
      width:min(96vw, 1760px);
      bottom:clamp(18px, 2vw, 34px);
      transform:translateX(-50%%);
      display:block;
      pointer-events:none;
      transition:opacity .18s ease, transform .18s ease;
    }
    body.is-idle .overlay, body.is-idle .topbar {
      opacity:0;
      pointer-events:none;
    }
    body.is-idle .overlay { transform:translate(-50%%, 10px); }
    body.is-idle .topbar { transform:translateY(10px); }
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
    .player-dock {
      display:grid;
      grid-template-columns:minmax(0, 1fr) minmax(330px, 420px);
      gap:clamp(18px, 1.8vw, 30px);
      align-items:stretch;
      max-width:100%%;
      overflow:visible;
      min-height:clamp(170px, 15vh, 220px);
      pointer-events:auto;
    }
    .player-primary {
      display:grid;
      grid-template-rows:auto auto auto;
      align-content:end;
      min-width:0;
    }
    .eyebrow {
      color:var(--champagne);
      font-size:12px;
      font-weight:900;
      text-transform:uppercase;
    }
    h1 {
      margin:5px 0 0;
      max-width:100%%;
      font-size:clamp(30px, 2.15vw, 46px);
      line-height:1.02;
      font-weight:900;
      white-space:normal;
      overflow:visible;
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
      padding-top:12px;
    }
    button, a.button {
      pointer-events:auto;
      min-height:42px;
      display:inline-flex;
      align-items:center;
      justify-content:center;
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
      vertical-align:middle;
    }
    button.primary { border-color:transparent; background:var(--champagne); color:#11100d; }
    .icon-button {
      width:42px;
      padding:0;
      display:inline-grid;
      place-items:center;
    }
    .seek-row {
      display:grid;
      grid-template-columns:max-content minmax(0, 1fr) max-content;
      gap:12px;
      align-items:center;
      margin-top:14px;
    }
    .seek-row span {
      color:var(--soft);
      font-size:13px;
      font-weight:850;
      min-width:4.5em;
      text-align:center;
    }
    input[type="range"] {
      width:100%%;
      accent-color:var(--champagne);
      cursor:pointer;
    }
    .control-field,
    .menu-control {
      position:relative;
      display:inline-flex;
      align-items:center;
      gap:8px;
      min-height:42px;
      padding:0 12px;
      border:1px solid var(--line);
      border-radius:8px;
      background:rgba(245,240,231,.055);
      color:var(--soft);
      font-size:13px;
      font-weight:850;
      white-space:nowrap;
    }
    .menu-trigger {
      min-height:auto;
      padding:0;
      border:0;
      background:transparent;
      color:var(--text);
      box-shadow:none;
      font-size:13px;
    }
    .menu-trigger::after {
      content:"";
      width:8px;
      height:8px;
      margin-left:6px;
      border-right:2px solid var(--soft);
      border-bottom:2px solid var(--soft);
      transform:rotate(45deg) translateY(-2px);
    }
    .player-menu {
      position:absolute;
      left:0;
      bottom:calc(100%% + 8px);
      z-index:8;
      display:none;
      min-width:calc(150px * var(--density-scale));
      padding:6px;
      border:1px solid var(--line);
      border-radius:10px;
      background:rgba(18,18,17,.96);
      backdrop-filter:blur(18px);
      box-shadow:0 20px 70px rgba(0,0,0,.48);
    }
    .menu-control.open .player-menu { display:grid; gap:4px; }
    .player-menu button {
      justify-content:flex-start;
      min-height:34px;
      width:100%%;
      border-color:transparent;
      background:transparent;
      color:var(--soft);
      font-size:13px;
    }
    .player-menu button:hover,
    .player-menu button.selected {
      color:var(--text);
      background:rgba(245,240,231,.08);
      border-color:rgba(245,240,231,.09);
    }
    .volume {
      width:110px;
    }
    .forecast {
      display:grid;
      gap:10px;
      align-content:start;
      padding-left:clamp(14px, 1.2vw, 22px);
      border-left:1px solid rgba(245,240,231,.1);
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
      .overlay { grid-template-columns:1fr; left:12px; right:12px; width:auto; bottom:12px; transform:none; }
      body.is-idle .overlay { transform:translateY(10px); }
      .player-dock { grid-template-columns:1fr; }
      .control-stack { justify-content:flex-start; }
      .forecast { display:none; }
      h1 { font-size:clamp(24px, 8vw, 38px); white-space:normal; text-wrap:balance; }
    }
  </style>
</head>
<body>
  <div class="player-shell">
    <div class="topbar">
      <div class="brand"><b>V</b> Xuva Player</div>
      <div class="status-pill" id="sessionState">Starting</div>
    </div>
    <button class="hud-toggle" id="hudToggle" type="button">Controls</button>
    <video id="player" autoplay></video>
  </div>
  <div class="overlay">
    <section class="panel player-dock">
      <div class="player-primary">
        <div>
        <div class="eyebrow">Now playing</div>
        <h1 id="playerTitle">%s</h1>
        <div class="meta">
          <span class="chip route" id="decisionMode">Preparing</span>
          <span class="chip" id="progressChip">0:00</span>
          <span class="chip" id="subtitleChip">Subtitles checking</span>
        </div>
        <div class="hint">Space toggles playback. Arrow keys seek. Progress saves automatically.</div>
      </div>
        <div class="seek-row">
          <span id="currentTime">0:00</span>
          <input id="seekBar" type="range" min="0" max="1000" value="0" aria-label="Seek">
          <span id="durationTime">--:--</span>
        </div>
        <div class="control-stack">
          <button class="primary" id="playToggle" type="button">Pause</button>
          <button class="icon-button" id="backButton" type="button" aria-label="Back 10 seconds">-10</button>
          <button class="icon-button" id="forwardButton" type="button" aria-label="Forward 10 seconds">+10</button>
          <button id="restartButton" type="button">Restart</button>
          <button id="markButton" type="button">Mark Watched</button>
          <button id="fullscreenButton" type="button">Fullscreen</button>
          <div class="menu-control" id="speedControl"><span>Speed</span><button class="menu-trigger" id="speedButton" type="button">1x</button><div class="player-menu" id="speedMenu"></div></div>
          <div class="menu-control" id="subtitleControl"><span>Subs</span><button class="menu-trigger" id="subtitleButton" type="button">Off</button><div class="player-menu" id="subtitleMenu"></div></div>
          <label class="control-field">Volume <input class="volume" id="volumeSlider" type="range" min="0" max="1" step="0.01" value="1"></label>
          <a class="button" href="/">Dashboard</a>
        </div>
      </div>
      <aside class="forecast">
        <h2>Playback Forecast</h2>
        <div class="decision"><strong id="forecastMode">Checking</strong><span id="forecastReason">Inspecting selected source and client profile.</span></div>
        <div class="kv">
          <div><span>Client</span><span>Web</span></div>
          <div><span>Route</span><span id="forecastRoute">LAN direct</span></div>
          <div><span>Server</span><span id="forecastServer">Low impact</span></div>
        </div>
      </aside>
    </section>
  </div>
  <script>
    const mediaSourceId = %q;
    const resumeEnabled = new URLSearchParams(location.search).get("start") !== "0";
    let sessionId = "";
    let saveInFlight = false;
    let queuedSaveStatus = "";
    let isSeeking = false;
    let lastMouseMove = Date.now();
    let signedStream = null;
    let currentDecision = {};
    let currentRoute = "";
    const player = document.getElementById("player");
    const sessionState = document.getElementById("sessionState");
    const playToggle = document.getElementById("playToggle");
    const backButton = document.getElementById("backButton");
    const forwardButton = document.getElementById("forwardButton");
    const restartButton = document.getElementById("restartButton");
    const markButton = document.getElementById("markButton");
    const fullscreenButton = document.getElementById("fullscreenButton");
    const hudToggle = document.getElementById("hudToggle");
    const decisionMode = document.getElementById("decisionMode");
    const progressChip = document.getElementById("progressChip");
    const subtitleChip = document.getElementById("subtitleChip");
    const seekBar = document.getElementById("seekBar");
    const currentTimeLabel = document.getElementById("currentTime");
    const durationTimeLabel = document.getElementById("durationTime");
    const volumeSlider = document.getElementById("volumeSlider");
    const speedControl = document.getElementById("speedControl");
    const speedButton = document.getElementById("speedButton");
    const speedMenu = document.getElementById("speedMenu");
    const subtitleControl = document.getElementById("subtitleControl");
    const subtitleButton = document.getElementById("subtitleButton");
    const subtitleMenu = document.getElementById("subtitleMenu");
    const playerTitle = document.getElementById("playerTitle");
    const forecastMode = document.getElementById("forecastMode");
    const forecastReason = document.getElementById("forecastReason");
    const forecastServer = document.getElementById("forecastServer");
    const forecastRoute = document.getElementById("forecastRoute");
    let idleTimer = 0;
    const speedOptions = ["0.75", "1", "1.25", "1.5", "2"];
    let selectedSubtitleTrack = "-1";

    function csrfToken() {
      return document.cookie.split(";").map(item => item.trim()).find(item => item.startsWith("xuva_csrf="))?.split("=").slice(1).join("=") || "";
    }
    async function send(path, body, method = "POST", keepalive = false) {
      const token = csrfToken();
      const headers = { "Content-Type": "application/json" };
      if (token) headers["X-CSRF-Token"] = decodeURIComponent(token);
      const response = await fetch(path, { method, keepalive, headers, body: JSON.stringify(body || {}) });
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
    function fitPlayerTitle() {
      if (!playerTitle || window.innerWidth <= 900) return;
      playerTitle.style.fontSize = "";
      const styles = getComputedStyle(playerTitle);
      const maxSize = parseFloat(styles.fontSize) || 54;
      const minSize = 24;
      let size = maxSize;
      while (playerTitle.scrollWidth > playerTitle.clientWidth && size > minSize) {
        size -= 1;
        playerTitle.style.fontSize = size + "px";
      }
    }
    function progressBody(status) {
      return { progressSeconds: player.currentTime || 0, durationSeconds: Number.isFinite(player.duration) ? player.duration : 0, status: status || (player.paused ? "paused" : "playing") };
    }
    async function loadForecast() {
      const decision = await getJSON("/api/playback/decision?mediaSourceId=" + mediaSourceId + "&clientProfile=web");
      currentDecision = decision || {};
      const mode = decision.mode || "direct";
      const label = mode.replaceAll("_", " ");
      decisionMode.textContent = label;
      decisionMode.className = "chip " + (mode === "direct" ? "good" : mode === "remux" ? "route" : "warn");
      forecastMode.textContent = label;
      forecastReason.textContent = decision.reason || "Direct file stream is available for this client.";
      forecastServer.textContent = mode === "direct" ? "Low impact" : "Server work required";
      forecastRoute.textContent = label;
      return mode;
    }
    function supportsNativeHLS() {
      if (!player || typeof player.canPlayType !== "function") return false;
      return Boolean(player.canPlayType("application/vnd.apple.mpegurl") || player.canPlayType("application/x-mpegURL"));
    }
    async function playbackURLForRoute(route) {
      if (route.protocol === "hls" && route.manifestUrl) {
        if (supportsNativeHLS()) return route.manifestUrl;
        return "";
      }
      if (route.url) {
        if (route.route === "direct") {
          const signed = await authorizeStream();
          return signed.streamUrl || route.url;
        }
        return route.url;
      }
      return "";
    }
    async function resolvePlaybackRoute() {
      sessionState.textContent = "Selecting route";
      const route = await getJSON("/api/playback/route?mediaSourceId=" + mediaSourceId + "&clientProfile=web", {});
      currentRoute = route.route || currentRoute;
      if (route.status === "blocked_by_policy") {
        const policy = route.policy || {};
        const decision = route.decision || {};
        const fallbacks = route.fallbackOptions || [];
        player.removeAttribute("src");
        sessionState.textContent = "Playback blocked by policy";
        forecastMode.textContent = "Fallback needed";
        forecastReason.textContent = (policy.label || "Current policy") + " will not automatically " + (decision.mode || "adapt this file") + ". Choose a fallback from the dashboard or change Playback Policy in Settings.";
        forecastServer.textContent = "No work started";
        forecastRoute.textContent = policy.label || "Blocked";
        if (fallbacks.length) {
          forecastReason.textContent += " Options: " + fallbacks.map(item => item.label).join(", ") + ".";
        }
        return "blocked";
      }
      if (route.status === "ready") {
        const playbackURL = await playbackURLForRoute(route);
        if (playbackURL) {
          player.src = playbackURL;
          sessionState.textContent = route.route === "direct" ? "Direct stream" : route.route === "adaptive" ? "Adaptive stream" : "Prepared stream";
          await updateInspectorRoute(route.route || "direct", route.decision || currentDecision);
          await player.play().catch(() => {});
          return route.route || "direct";
        }
        if (route.protocol === "hls" && route.manifestUrl && !supportsNativeHLS()) {
          sessionState.textContent = "Adaptive stream unavailable";
          forecastMode.textContent = "Browser fallback needed";
          forecastReason.textContent = "This browser does not support native HLS playback for adaptive streams. Use a native player profile or switch to a direct-compatible file.";
          forecastServer.textContent = "Route blocked";
          forecastRoute.textContent = "Adaptive HLS";
          return "blocked";
        }
      }
      if (route.job && route.job.id) {
        sessionState.textContent = "Preparing " + (route.route || "playback");
        await updateInspectorRoute(route.route || "preparing", route.decision || currentDecision);
        setTimeout(resolvePlaybackRoute, 1800);
        return route.route || "preparing";
      }
      const signed = await authorizeStream();
      if (signed.streamUrl) {
        player.src = signed.streamUrl;
        sessionState.textContent = "Direct fallback";
        await updateInspectorRoute("direct", currentDecision);
        return "direct";
      }
      sessionState.textContent = "Playback unavailable";
      forecastReason.textContent = "Xuva could not resolve a playable route for this browser. Check source compatibility and playback policy.";
      await updateInspectorRoute("blocked", currentDecision);
      return "blocked";
    }
    async function updateInspectorRoute(route, decision) {
      if (!sessionId) return;
      await send("/api/sessions/" + sessionId, {
        route,
        mode: decision.mode || route,
        reasonCode: decision.reasonCode || "",
        reasonText: decision.reasonText || decision.reason || "",
        serverImpact: decision.estimatedCpuCost === "high" ? "High server load" : decision.estimatedCpuCost === "medium" ? "Moderate server load" : "Low impact",
        selectedTracks: {
          audio: "Default",
          subtitles: subtitleButton.textContent || "Off"
        }
      }, "PATCH").catch(() => {});
    }
    async function authorizeStream() {
      if (signedStream && signedStream.streamUrl) return signedStream;
      if (!sessionId) return {};
      signedStream = await send("/api/media-sources/" + mediaSourceId + "/stream-token", { sessionId, deviceId: "web" });
      return signedStream || {};
    }
    async function loadSubtitles() {
      const subtitles = await getJSON("/api/media-sources/" + mediaSourceId + "/subtitles", { sidecars: [] });
      const sidecars = subtitles.sidecars || [];
      subtitleChip.textContent = sidecars.length ? sidecars.length + " subtitle file" + (sidecars.length === 1 ? "" : "s") : "No sidecar subtitles";
      let subtitleTrackIndex = 0;
      sidecars.forEach((item, index) => {
        if (item.format !== "vtt") return;
        const track = document.createElement("track");
        track.kind = "subtitles";
        track.label = item.language || item.relPath || "Subtitle";
        track.srclang = item.language || "und";
        const query = signedStream && signedStream.query ? signedStream.query : "";
        track.src = "/api/media-sources/" + mediaSourceId + "/subtitles/" + index + query;
        player.appendChild(track);
        addMenuOption(subtitleMenu, subtitleControl, subtitleButton, track.label, String(subtitleTrackIndex), value => {
          selectedSubtitleTrack = value;
          syncSubtitleSelection();
        });
        subtitleTrackIndex += 1;
      });
      setTimeout(() => syncSubtitleSelection(), 100);
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
      const session = await send("/api/sessions", { mediaSourceId, deviceId: "web", clientProfile: "web", mode, route: mode });
      sessionId = session.id || "";
      sessionState.textContent = sessionId ? "Live session" : "Local playback";
    }
    async function saveProgress(status) {
      if (saveInFlight) {
        queuedSaveStatus = status || queuedSaveStatus || (player.paused ? "paused" : "playing");
        return;
      }
      saveInFlight = true;
      const body = progressBody(status);
      try {
        if (sessionId) await send("/api/sessions/" + sessionId, body, "PATCH");
        await send("/api/playback/state/" + mediaSourceId, body, "PUT");
      } finally {
        saveInFlight = false;
        if (queuedSaveStatus) {
          const queued = queuedSaveStatus;
          queuedSaveStatus = "";
          saveProgress(queued);
        }
      }
    }
    async function stopSession(status = "stopped") {
      const body = progressBody(status);
      if (sessionId) {
        await send("/api/sessions/" + sessionId, body, "PATCH").catch(() => {});
        await send("/api/sessions/" + sessionId, {}, "DELETE", true).catch(() => {});
        sessionId = "";
      }
      await send("/api/playback/state/" + mediaSourceId, body, "PUT", true).catch(() => {});
    }
    function refreshProgress() {
      const current = formatTime(player.currentTime || 0);
      const total = Number.isFinite(player.duration) && player.duration > 0 ? formatTime(player.duration) : "--:--";
      progressChip.textContent = current + " / " + total;
      currentTimeLabel.textContent = current;
      durationTimeLabel.textContent = total;
      if (!isSeeking) {
        seekBar.value = Number.isFinite(player.duration) && player.duration > 0 ? Math.round(((player.currentTime || 0) / player.duration) * 1000) : 0;
      }
      playToggle.textContent = player.paused ? "Play" : "Pause";
      sessionState.textContent = player.paused ? "Paused" : "Playing";
    }
    function seekBy(seconds) {
      player.currentTime = Math.max(0, Math.min(player.duration || 0, (player.currentTime || 0) + seconds));
      refreshProgress();
      saveProgress("playing");
    }
    function syncSubtitleSelection() {
      Array.from(player.textTracks || []).forEach((track, index) => {
        track.mode = String(index) === selectedSubtitleTrack ? "showing" : "disabled";
      });
      updateInspectorRoute(currentRoute || "direct", currentDecision);
    }
    function toggleFullscreen() {
      const target = document.querySelector(".player-shell");
      if (!document.fullscreenElement) {
        target.requestFullscreen?.();
      } else {
        document.exitFullscreen?.();
      }
    }
    function addMenuOption(menu, control, labelButton, label, value, onSelect) {
      const item = document.createElement("button");
      item.type = "button";
      item.dataset.value = value;
      item.textContent = label;
      item.addEventListener("click", () => {
        labelButton.textContent = label;
        menu.querySelectorAll("button").forEach(button => button.classList.toggle("selected", button === item));
        control.classList.remove("open");
        onSelect(value, label);
        showHud(4200);
      });
      menu.appendChild(item);
      return item;
    }
    function toggleMenu(control) {
      document.querySelectorAll(".menu-control.open").forEach(item => {
        if (item !== control) item.classList.remove("open");
      });
      control.classList.toggle("open");
      showHud(4200);
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
    backButton.addEventListener("click", () => seekBy(-10));
    forwardButton.addEventListener("click", () => seekBy(10));
    restartButton.addEventListener("click", () => { player.currentTime = 0; player.play(); saveProgress("playing"); });
    markButton.addEventListener("click", async () => {
      await send("/api/playback/state/" + mediaSourceId, { watched: true, progressSeconds: player.duration || player.currentTime || 0, durationSeconds: player.duration || 0 }, "PUT");
      markButton.textContent = "Marked";
    });
    fullscreenButton.addEventListener("click", toggleFullscreen);
    volumeSlider.addEventListener("input", () => { player.volume = Number(volumeSlider.value || 0); player.muted = player.volume <= 0; });
    speedOptions.forEach(value => {
      const label = value + "x";
      const item = addMenuOption(speedMenu, speedControl, speedButton, label, value, next => { player.playbackRate = Number(next || 1); });
      item.classList.toggle("selected", value === "1");
    });
    addMenuOption(subtitleMenu, subtitleControl, subtitleButton, "Off", "-1", value => {
      selectedSubtitleTrack = value;
      syncSubtitleSelection();
    }).classList.add("selected");
    speedButton.addEventListener("click", () => toggleMenu(speedControl));
    subtitleButton.addEventListener("click", () => toggleMenu(subtitleControl));
    document.addEventListener("click", event => {
      if (!event.target.closest(".menu-control")) {
        document.querySelectorAll(".menu-control.open").forEach(item => item.classList.remove("open"));
      }
    });
    seekBar.addEventListener("input", () => {
      isSeeking = true;
      const duration = Number.isFinite(player.duration) ? player.duration : 0;
      currentTimeLabel.textContent = formatTime((Number(seekBar.value || 0) / 1000) * duration);
    });
    seekBar.addEventListener("change", () => {
      const duration = Number.isFinite(player.duration) ? player.duration : 0;
      if (duration > 0) player.currentTime = (Number(seekBar.value || 0) / 1000) * duration;
      isSeeking = false;
      refreshProgress();
      saveProgress(player.paused ? "paused" : "playing");
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
      if (event.code === "ArrowRight") seekBy(10);
      if (event.code === "ArrowLeft") seekBy(-10);
      if (event.key && event.key.toLowerCase() === "f") toggleFullscreen();
    });
    window.addEventListener("mousemove", () => {
      showHud();
    });
    window.addEventListener("pointerdown", event => {
      if (event.target === player || event.target === document.body) showHud();
    });
    window.addEventListener("resize", fitPlayerTitle);
    setInterval(() => {
      refreshProgress();
      if (!player.paused && !player.ended) saveProgress("playing");
    }, 2000);
    window.addEventListener("beforeunload", () => { stopSession("stopped"); });
    (async function boot() {
      await loadResumeState();
      const mode = await loadForecast();
      await startSession(mode);
      await resolvePlaybackRoute();
      await loadSubtitles();
      fitPlayerTitle();
      refreshProgress();
      showHud(2400);
    })();
  </script>
</body>
</html>`, html.EscapeString(title), html.EscapeString(title), id)
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
		release, ok := authorizeStreamRequest(deps, w, r, r.PathValue("id"))
		if !ok {
			return
		}
		defer release()
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

func mediaSourceSubtitleConvertHandler(deps Deps) http.HandlerFunc {
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
		service := deps.Subtitles
		if service == nil {
			service = subtitles.NewService()
		}
		plan := service.PlanConversion(sidecars[index], firstNonEmpty(r.URL.Query().Get("clientProfile"), "web"))
		writeJSON(w, http.StatusAccepted, map[string]any{
			"mediaSourceId": item.ID,
			"subtitleIndex": index,
			"sidecar":       sidecars[index],
			"conversion":    plan,
		})
	}
}

func authorizeStreamRequest(deps Deps, w http.ResponseWriter, r *http.Request, mediaSourceID string) (func(), bool) {
	if deps.Auth == nil || deps.Auth.Disabled() {
		return func() {}, true
	}
	if deps.Streaming == nil {
		writeError(w, http.StatusServiceUnavailable, "stream signing is not available")
		return nil, false
	}
	resolved, ok := auth.ResolvedSessionFromContext(r.Context())
	if !ok {
		writeError(w, http.StatusUnauthorized, "authentication required")
		return nil, false
	}
	sessionID := r.URL.Query().Get("sessionId")
	deviceID := r.URL.Query().Get("deviceId")
	session, ok := deps.Sessions.Get(sessionID)
	if !ok {
		publishStreamDenied(deps, r, resolved, mediaSourceID, sessionID, "session_inactive")
		writeError(w, http.StatusForbidden, "playback session is not active")
		return nil, false
	}
	if session.MediaSourceID != mediaSourceID || session.UserID != resolved.Principal.ID || session.DeviceID != deviceID {
		publishStreamDenied(deps, r, resolved, mediaSourceID, sessionID, "session_mismatch")
		writeError(w, http.StatusForbidden, "playback session does not match this stream")
		return nil, false
	}
	_, release, err := deps.Streaming.Validate(r.URL.Query().Get("token"), streaming.Expected{
		MediaSourceID: mediaSourceID,
		SessionID:     session.ID,
		UserID:        session.UserID,
		DeviceID:      session.DeviceID,
	})
	if err != nil {
		status, reason := streamTokenError(err)
		publishStreamDenied(deps, r, resolved, mediaSourceID, sessionID, reason)
		writeError(w, status, reason)
		return nil, false
	}
	return release, true
}

func streamTokenError(err error) (int, string) {
	switch {
	case errors.Is(err, streaming.ErrMissingToken):
		return http.StatusUnauthorized, "stream token is required"
	case errors.Is(err, streaming.ErrExpiredToken):
		return http.StatusForbidden, "stream token is expired"
	case errors.Is(err, streaming.ErrInvalidSignature), errors.Is(err, streaming.ErrMalformedToken), errors.Is(err, streaming.ErrSigningKeyMissing):
		return http.StatusForbidden, "stream token is invalid"
	case errors.Is(err, streaming.ErrTokenMismatch):
		return http.StatusForbidden, "stream token does not match playback session"
	case errors.Is(err, streaming.ErrStreamLimit):
		return http.StatusTooManyRequests, "stream limit exceeded"
	default:
		return http.StatusForbidden, "stream token rejected"
	}
}

func publishStreamDenied(deps Deps, r *http.Request, resolved auth.ResolvedSession, mediaSourceID string, sessionID string, reason string) {
	if deps.Events == nil {
		return
	}
	deps.Events.Publish("audit.stream.denied", map[string]any{
		"correlationId": observability.CorrelationID(r.Context()),
		"userId":        resolved.Principal.ID,
		"username":      resolved.Principal.Username,
		"role":          resolved.Principal.Role,
		"method":        r.Method,
		"path":          r.URL.Path,
		"mediaSourceId": mediaSourceID,
		"sessionId":     sessionID,
		"reason":        reason,
		"createdAt":     time.Now().UTC().Format(time.RFC3339Nano),
	})
}

func publishOperationalEvent(deps Deps, r *http.Request, eventType string, fields map[string]any) {
	if deps.Events == nil {
		return
	}
	payload := map[string]any{
		"correlationId": observability.CorrelationID(r.Context()),
		"method":        r.Method,
		"path":          r.URL.Path,
		"createdAt":     time.Now().UTC().Format(time.RFC3339Nano),
	}
	for key, value := range fields {
		payload[key] = value
	}
	deps.Events.Publish(eventType, payload)
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
		if request.Mode == transcode.ModeTranscode && request.VideoEncoder == "" {
			if encoder, ok := selectedHardwareEncoder(r.Context(), deps.Config); ok {
				request.Acceleration = "hardware"
				request.VideoEncoder = encoder
			}
		}
		job, err := deps.Transcode.Start(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		publishOperationalEvent(deps, r, "api.work.accepted", map[string]any{
			"jobId":         job.ID,
			"mediaSourceId": job.MediaSourceID,
			"mode":          string(job.Mode),
		})
		writeJSON(w, http.StatusAccepted, job)
	}
}

func workCancelHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := deps.Transcode.Cancel(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "work job not found")
			return
		}
		writeJSON(w, http.StatusOK, job)
	}
}

func workFileHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		job, ok := deps.Transcode.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "work job not found")
			return
		}
		if job.Status != transcode.StatusCompleted || job.OutputPath == "" {
			writeError(w, http.StatusConflict, "work job is not ready")
			return
		}
		http.ServeFile(w, r, job.OutputPath)
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
		if request.TargetProfile != downloads.ProfileOriginal && request.VideoEncoder == "" {
			if encoder, ok := selectedHardwareEncoder(r.Context(), deps.Config); ok {
				request.Acceleration = "hardware"
				request.VideoEncoder = encoder
			}
		}
		job, err := deps.Downloads.Start(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		publishOperationalEvent(deps, r, "api.download.accepted", map[string]any{
			"jobId":         job.ID,
			"mediaSourceId": job.MediaSourceID,
			"profile":       job.TargetProfile,
		})
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

func approvedDevicesHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Devices == nil {
			writeJSON(w, http.StatusOK, map[string]any{"devices": []map[string]any{}})
			return
		}
		items, err := deps.Devices.ListApproved(r.Context())
		if err != nil {
			writeError(w, http.StatusInternalServerError, "approved devices lookup failed")
			return
		}
		payload := make([]map[string]any, 0, len(items))
		for _, item := range items {
			payload = append(payload, approvedDevicePayload(item))
		}
		writeJSON(w, http.StatusOK, map[string]any{"devices": payload})
	}
}

func approvedDeviceRevokeHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Devices == nil {
			writeError(w, http.StatusServiceUnavailable, "approved device registry is not available")
			return
		}
		item, err := deps.Devices.Revoke(r.Context(), r.PathValue("id"))
		if err != nil {
			switch err {
			case devices.ErrNotFound:
				writeError(w, http.StatusNotFound, "approved device not found")
			case devices.ErrRegistryUnavailable:
				writeError(w, http.StatusServiceUnavailable, "approved device registry is not available")
			default:
				writeError(w, http.StatusInternalServerError, "approved device update failed")
			}
			return
		}
		publishOperationalEvent(deps, r, "device.revoked", map[string]any{
			"deviceId":      item.DeviceID,
			"displayName":   item.DisplayName,
			"clientProfile": item.ClientProfile,
		})
		writeJSON(w, http.StatusOK, approvedDevicePayload(item))
	}
}

func pairingCreateHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Pairing == nil {
			writeError(w, http.StatusServiceUnavailable, "pairing is not available")
			return
		}
		var request pairing.CreateRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		if _, ok := deps.Devices.GetProfile(firstNonEmpty(request.ClientProfile, "apple-tv")); !ok {
			writeError(w, http.StatusBadRequest, "unknown client profile")
			return
		}
		item, err := deps.Pairing.Create(request)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "pairing request failed")
			return
		}
		publishOperationalEvent(deps, r, "pairing.request.created", map[string]any{
			"pairingId":     item.ID,
			"clientProfile": item.ClientProfile,
			"deviceName":    item.DeviceName,
			"expiresAt":     item.ExpiresAt,
		})
		writeJSON(w, http.StatusAccepted, item)
	}
}

func pairingStatusHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Pairing == nil {
			writeError(w, http.StatusServiceUnavailable, "pairing is not available")
			return
		}
		item, ok := deps.Pairing.Get(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "pairing request not found")
			return
		}
		writeJSON(w, http.StatusOK, item)
	}
}

func pairingListHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if deps.Pairing == nil {
			writeError(w, http.StatusServiceUnavailable, "pairing is not available")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"requests": deps.Pairing.List()})
	}
}

func pairingApproveHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		closePairingRequest(w, r, deps, true)
	}
}

func pairingDenyHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		closePairingRequest(w, r, deps, false)
	}
}

func closePairingRequest(w http.ResponseWriter, r *http.Request, deps Deps, approve bool) {
	if deps.Pairing == nil {
		writeError(w, http.StatusServiceUnavailable, "pairing is not available")
		return
	}
	actor := "local-admin"
	if resolved, ok := auth.ResolvedSessionFromContext(r.Context()); ok {
		actor = firstNonEmpty(resolved.Principal.Username, resolved.Principal.ID, actor)
	}
	var (
		item pairing.Request
		err  error
	)
	if approve {
		item, err = deps.Pairing.Approve(r.PathValue("id"), actor)
	} else {
		item, err = deps.Pairing.Deny(r.PathValue("id"), actor)
	}
	if err != nil {
		switch err {
		case pairing.ErrNotFound:
			writeError(w, http.StatusNotFound, "pairing request not found")
		case pairing.ErrExpired:
			writeError(w, http.StatusGone, "pairing request expired")
		case pairing.ErrClosed:
			writeError(w, http.StatusConflict, "pairing request is already closed")
		default:
			writeError(w, http.StatusInternalServerError, "pairing request failed")
		}
		return
	}
	if approve && deps.Devices != nil {
		if _, registerErr := deps.Devices.Approve(r.Context(), devices.ApproveInput{
			DeviceID:      item.DeviceID,
			DeviceName:    item.DeviceName,
			ClientProfile: item.ClientProfile,
			ApprovedBy:    actor,
		}); registerErr != nil {
			writeError(w, http.StatusInternalServerError, "approved device registry update failed")
			return
		}
		publishOperationalEvent(deps, r, "device.approved", map[string]any{
			"deviceId":      item.DeviceID,
			"deviceName":    item.DeviceName,
			"clientProfile": item.ClientProfile,
		})
	}
	publishOperationalEvent(deps, r, "pairing.request."+item.Status, map[string]any{
		"pairingId":     item.ID,
		"clientProfile": item.ClientProfile,
		"deviceName":    item.DeviceName,
		"deviceId":      item.DeviceID,
	})
	writeJSON(w, http.StatusOK, item)
}

func approvedDevicePayload(item devices.ApprovedDevice) map[string]any {
	return map[string]any{
		"id":            item.ID,
		"deviceName":    item.DeviceName,
		"clientProfile": item.ClientProfile,
		"displayName":   item.DisplayName,
		"status":        item.Status,
		"approvedAt":    item.ApprovedAt,
		"approvedBy":    item.ApprovedBy,
		"createdAt":     item.CreatedAt,
		"updatedAt":     item.UpdatedAt,
	}
}

func sessionsHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		writeJSON(w, http.StatusOK, map[string]any{"sessions": deps.Sessions.List()})
	}
}

func sessionInspectorHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		inspector, ok := deps.Sessions.Inspector(r.PathValue("id"))
		if !ok {
			writeError(w, http.StatusNotFound, "session not found")
			return
		}
		writeJSON(w, http.StatusOK, inspector)
	}
}

func sessionStartHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request sessions.StartRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		if request.UserID == "" {
			request.UserID = requestUserID(r)
		}
		if err := enrichSessionStartRequest(deps, r.Context(), &request); err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		session, err := deps.Sessions.Start(request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		publishOperationalEvent(deps, r, "api.session.accepted", map[string]any{
			"sessionId":     session.ID,
			"mediaSourceId": session.MediaSourceID,
			"deviceId":      session.DeviceID,
			"mode":          session.Mode,
		})
		writeJSON(w, http.StatusAccepted, session)
	}
}

func enrichSessionStartRequest(deps Deps, ctx context.Context, request *sessions.StartRequest) error {
	if request == nil || request.MediaSourceID == "" || deps.Catalog == nil {
		return nil
	}
	source, ok, err := deps.Catalog.GetMediaSource(ctx, request.MediaSourceID)
	if err != nil {
		return err
	}
	if !ok {
		return nil
	}
	if request.SourceName == "" {
		request.SourceName = source.Name
	}
	if request.Container == "" {
		request.Container = source.Container
	}
	if request.VideoCodec == "" {
		request.VideoCodec = source.VideoCodec
	}
	if request.Bitrate == 0 {
		request.Bitrate = source.Bitrate
	}
	if request.DurationSeconds == 0 {
		request.DurationSeconds = source.DurationSeconds
	}
	display, displayOK, err := deps.Catalog.GetMediaSourceDisplay(ctx, request.MediaSourceID)
	if err != nil {
		return err
	}
	if displayOK {
		if request.Title == "" {
			request.Title = display.Title
		}
		if request.QualityLabel == "" {
			request.QualityLabel = display.QualityLabel
		}
		if request.ArtworkURL == "" && display.ArtworkKind != "" && display.ArtworkID != "" {
			request.ArtworkURL = "/api/artwork/" + display.ArtworkKind + "/" + display.ArtworkID
		}
	}
	if request.Title == "" {
		request.Title = source.Name
	}
	if request.QualityLabel == "" {
		request.QualityLabel = sessionQualityLabel(source)
	}
	if request.ClientProfile == "" {
		request.ClientProfile = request.DeviceID
	}
	if request.ClientProfile == "" {
		request.ClientProfile = "web"
	}
	if request.Route == "" {
		request.Route = request.Mode
	}
	if request.ServerImpact == "" {
		request.ServerImpact = sessionServerImpact(request.Route, request.Mode)
	}
	if request.ReasonCode == "" || request.ReasonText == "" {
		decision := deps.Playback.DecideSource(ctx, playback.Request{
			MediaSourceID: request.MediaSourceID,
			ClientProfile: request.ClientProfile,
		}, playbackSourceFacts(ctx, deps, playback.Request{MediaSourceID: request.MediaSourceID, ClientProfile: request.ClientProfile}, source))
		if request.ReasonCode == "" {
			request.ReasonCode = decision.ReasonCode
		}
		if request.ReasonText == "" {
			request.ReasonText = decision.ReasonText
		}
	}
	return nil
}

func sessionQualityLabel(source catalog.MediaSourceItem) string {
	parts := []string{}
	if source.Width >= 3800 || source.Height >= 2000 {
		parts = append(parts, "4K")
	} else if source.Width >= 1900 || source.Height >= 1000 {
		parts = append(parts, "1080p")
	} else if source.Height > 0 {
		parts = append(parts, fmt.Sprintf("%dp", source.Height))
	}
	if source.VideoCodec != "" {
		parts = append(parts, strings.ToUpper(source.VideoCodec))
	}
	if source.Container != "" {
		parts = append(parts, strings.ToUpper(source.Container))
	}
	return strings.Join(parts, " ")
}

func sessionServerImpact(route string, mode string) string {
	value := strings.ToLower(route + " " + mode)
	switch {
	case strings.Contains(value, "transcode"):
		return "Converting while playing"
	case strings.Contains(value, "remux"):
		return "Repackaging while playing"
	case strings.Contains(value, "preparing"):
		return "Selecting playback route"
	case strings.Contains(value, "direct"):
		return "Low impact"
	default:
		return "Route pending"
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
		items, err := deps.PlayState.Recent(r.Context(), requestUserID(r), queryInt(r, "limit", 24))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "recent playback lookup failed")
			return
		}
		writeJSON(w, http.StatusOK, map[string]any{"recent": items})
	}
}

func playbackStateGetHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		state, _, err := deps.PlayState.Get(r.Context(), requestUserID(r), r.PathValue("id"))
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
		if request.UserID == "" {
			request.UserID = requestUserID(r)
		}
		state, err := deps.PlayState.Set(r.Context(), r.PathValue("id"), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusOK, state)
	}
}

func requestUserID(r *http.Request) string {
	if resolved, ok := auth.ResolvedSessionFromContext(r.Context()); ok && strings.TrimSpace(resolved.Principal.ID) != "" {
		return resolved.Principal.ID
	}
	return r.URL.Query().Get("userId")
}

func movieScanHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var request libraryScanRequest
		if !decodeJSON(w, r, &request) {
			return
		}
		path := firstNonEmpty(
			request.Path,
			request.MoviesPath,
			deps.Config.MovieLibraryPath,
			firstLibraryPathByKind(deps.Libraries, libraries.KindMovies),
		)
		if path == "" {
			writeError(w, http.StatusBadRequest, "Add a Movies library folder before scanning movies.")
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
		path := firstNonEmpty(
			request.Path,
			request.TVPath,
			deps.Config.TVLibraryPath,
			firstLibraryPathByKind(deps.Libraries, libraries.KindTV),
		)
		if path == "" {
			writeError(w, http.StatusBadRequest, "Add a TV library folder before scanning TV.")
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

		moviesPath := firstNonEmpty(
			request.MoviesPath,
			deps.Config.MovieLibraryPath,
			firstLibraryPathByKind(deps.Libraries, libraries.KindMovies),
		)
		tvPath := firstNonEmpty(
			request.TVPath,
			deps.Config.TVLibraryPath,
			firstLibraryPathByKind(deps.Libraries, libraries.KindTV),
		)
		if moviesPath == "" && tvPath == "" {
			writeError(w, http.StatusBadRequest, "Add a Movies or TV library folder before starting a library scan.")
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

func adaptiveSessionHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		source, ok, err := deps.Catalog.GetMediaSource(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		var payload struct {
			ClientProfile     string `json:"clientProfile"`
			RouteType         string `json:"routeType"`
			MaxNetworkBitrate int64  `json:"maxNetworkBitrate"`
		}
		if !decodeJSON(w, r, &payload) {
			return
		}
		plan := adaptivePlanForSource(deps, source, payload.ClientProfile, payload.RouteType, payload.MaxNetworkBitrate)
		writeJSON(w, http.StatusOK, map[string]any{
			"plan":        plan,
			"manifestUrl": "/api/media-sources/" + source.ID + "/adaptive/master.m3u8?clientProfile=" + url.QueryEscape(firstNonEmpty(payload.ClientProfile, "web")) + "&routeType=" + url.QueryEscape(payload.RouteType) + "&maxNetworkBitrate=" + fmt.Sprintf("%d", payload.MaxNetworkBitrate),
		})
	}
}

func adaptiveMasterHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		source, ok, err := deps.Catalog.GetMediaSource(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		plan := adaptivePlanForSource(deps, source, r.URL.Query().Get("clientProfile"), r.URL.Query().Get("routeType"), queryInt64(r, "maxNetworkBitrate", 0))
		if !plan.Enabled {
			writeError(w, http.StatusPreconditionFailed, plan.Reason)
			return
		}
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte(adaptive.MasterPlaylist(plan)))
	}
}

func adaptiveVariantHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		source, ok, err := deps.Catalog.GetMediaSource(r.Context(), r.PathValue("id"))
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		variantID, validVariant := adaptiveVariantID(r.PathValue("variant"))
		if !validVariant {
			writeError(w, http.StatusNotFound, "adaptive variant not found")
			return
		}
		plan := adaptivePlanForSource(deps, source, r.URL.Query().Get("clientProfile"), r.URL.Query().Get("routeType"), queryInt64(r, "maxNetworkBitrate", 0))
		playlist, ok := adaptive.MediaPlaylist(plan, variantID)
		if !plan.Enabled || !ok {
			writeError(w, http.StatusPreconditionFailed, "adaptive variant is not available")
			return
		}
		w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
		_, _ = w.Write([]byte(playlist))
	}
}

func adaptiveTelemetryHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload adaptive.Telemetry
		if !decodeJSON(w, r, &payload) {
			return
		}
		event := adaptive.NormalizeTelemetry(payload, observability.CorrelationID(r.Context()))
		if deps.Events != nil {
			deps.Events.Publish(event.Event, map[string]any{
				"sessionId":         event.SessionID,
				"mediaSourceId":     event.MediaSourceID,
				"clientProfile":     event.ClientProfile,
				"variantId":         event.VariantID,
				"previousVariantId": event.PreviousVariant,
				"bufferSeconds":     event.BufferSeconds,
				"stallMs":           event.StallMS,
				"observedBitrate":   event.ObservedBitrate,
				"correlationId":     event.CorrelationID,
				"createdAt":         event.CreatedAt,
			})
		}
		writeJSON(w, http.StatusAccepted, event)
	}
}

func playbackDecisionHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		request := playback.Request{
			MediaSourceID:       r.URL.Query().Get("mediaSourceId"),
			ClientProfile:       r.URL.Query().Get("clientProfile"),
			Policy:              r.URL.Query().Get("policy"),
			RouteType:           r.URL.Query().Get("routeType"),
			MaxNetworkBitrate:   queryInt64(r, "maxNetworkBitrate", 0),
			AudioTrackIndex:     queryInt(r, "audioTrackIndex", 0),
			AudioCodec:          r.URL.Query().Get("audioCodec"),
			AudioChannels:       queryInt(r, "audioChannels", 0),
			SubtitleTrackIndex:  queryInt(r, "subtitleTrackIndex", 0),
			SubtitleCodec:       r.URL.Query().Get("subtitleCodec"),
			SubtitleMode:        r.URL.Query().Get("subtitleMode"),
			SubtitleTrackActive: r.URL.Query().Get("subtitleTrackActive") == "true",
			SupportsAdaptive:    r.URL.Query().Get("supportsAdaptive") == "true",
		}
		request = applyClientProfile(deps, request)
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
			tracks, _, _ := deps.Catalog.GetMediaSourceTracks(r.Context(), request.MediaSourceID)
			if request.AudioCodec == "" && len(tracks.AudioTracks) > 0 {
				audio := trackByIndex(tracks.AudioTracks, request.AudioTrackIndex)
				request.AudioTrackIndex = audio.Index
				request.AudioCodec = audio.Codec
				request.AudioChannels = audio.Channels
			}
			if request.SubtitleTrackActive && request.SubtitleCodec == "" && len(tracks.SubtitleTracks) > 0 {
				subtitle := trackByIndex(tracks.SubtitleTracks, request.SubtitleTrackIndex)
				request.SubtitleTrackIndex = subtitle.Index
				request.SubtitleCodec = subtitle.Codec
			}
			decision = deps.Playback.DecideSource(r.Context(), request, playback.SourceFacts{
				MediaSourceID:    source.ID,
				Container:        source.Container,
				VideoCodec:       source.VideoCodec,
				Width:            source.Width,
				Height:           source.Height,
				AudioStreams:     source.AudioStreams,
				SubtitleStreams:  source.SubtitleStreams,
				SidecarSubtitles: len(subtitles.DiscoverSidecars(source.Path)),
				Bitrate:          source.Bitrate,
				Probed:           source.Probed,
				AudioCodec:       request.AudioCodec,
				AudioChannels:    request.AudioChannels,
				SubtitleCodec:    request.SubtitleCodec,
				SubtitleActive:   request.SubtitleTrackActive,
			})
		}
		writeJSON(w, http.StatusOK, decision)
	}
}

func playbackRouteHandler(deps Deps) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		mediaSourceID := r.URL.Query().Get("mediaSourceId")
		if mediaSourceID == "" {
			writeError(w, http.StatusBadRequest, "media source id is required")
			return
		}
		item, ok, err := deps.Catalog.GetMediaSource(r.Context(), mediaSourceID)
		if err != nil {
			writeError(w, http.StatusInternalServerError, "media source lookup failed")
			return
		}
		if !ok {
			writeError(w, http.StatusNotFound, "media source not found")
			return
		}
		decision := playbackDecisionForSource(r.Context(), deps, r, item)
		if decision.Mode == playback.DirectPlay || decision.Mode == playback.DecisionDeferred {
			writeJSON(w, http.StatusOK, map[string]any{
				"route":    "direct",
				"status":   "ready",
				"url":      "/api/media-sources/" + mediaSourceID + "/stream",
				"decision": decision,
			})
			return
		}
		if !playbackPolicyAllows(deps.Config.PlaybackPolicy, decision) {
			writeJSON(w, http.StatusOK, map[string]any{
				"route":           "blocked",
				"status":          "blocked_by_policy",
				"policy":          playbackPolicyStatus(deps.Config.PlaybackPolicy),
				"decision":        decision,
				"fallbackOptions": playbackPolicyFallbacks(deps.Config.PlaybackPolicy, decision),
			})
			return
		}
		if decision.Mode == playback.AdaptiveStream {
			plan := adaptivePlanForSource(deps, item, r.URL.Query().Get("clientProfile"), firstNonEmpty(r.URL.Query().Get("routeType"), "remote"), queryInt64(r, "maxNetworkBitrate", 0))
			if plan.Enabled {
				writeJSON(w, http.StatusOK, map[string]any{
					"route":       "adaptive",
					"status":      "ready",
					"protocol":    plan.Protocol,
					"manifestUrl": "/api/media-sources/" + mediaSourceID + "/adaptive/master.m3u8?clientProfile=" + url.QueryEscape(firstNonEmpty(r.URL.Query().Get("clientProfile"), "web")) + "&routeType=" + url.QueryEscape(firstNonEmpty(r.URL.Query().Get("routeType"), "remote")) + "&maxNetworkBitrate=" + fmt.Sprintf("%d", queryInt64(r, "maxNetworkBitrate", 0)),
					"plan":        plan,
					"decision":    decision,
				})
				return
			}
		}
		mode := transcode.ModeTranscode
		if decision.Mode == playback.Remux {
			mode = transcode.ModeRemux
		}
		if job, ok := deps.Transcode.FindCompleted(mediaSourceID, mode); ok {
			writeJSON(w, http.StatusOK, map[string]any{
				"route":    string(mode),
				"status":   "ready",
				"url":      "/api/work/" + job.ID + "/file",
				"job":      job,
				"decision": decision,
			})
			return
		}
		if job, ok := deps.Transcode.FindActive(mediaSourceID, mode); ok {
			writeJSON(w, http.StatusAccepted, map[string]any{
				"route":    string(mode),
				"status":   string(job.Status),
				"job":      job,
				"decision": decision,
			})
			return
		}
		request := transcode.Request{MediaSourceID: mediaSourceID, Mode: mode, SourcePath: item.Path}
		if mode == transcode.ModeTranscode {
			if encoder, ok := selectedHardwareEncoder(r.Context(), deps.Config); ok {
				request.Acceleration = "hardware"
				request.VideoEncoder = encoder
			}
		}
		job, err := deps.Transcode.Start(r.Context(), request)
		if err != nil {
			writeError(w, http.StatusBadRequest, err.Error())
			return
		}
		writeJSON(w, http.StatusAccepted, map[string]any{
			"route":    string(mode),
			"status":   string(job.Status),
			"job":      job,
			"decision": decision,
		})
	}
}

func playbackDecisionForSource(ctx context.Context, deps Deps, r *http.Request, source catalog.MediaSourceItem) playback.Decision {
	request := playback.Request{
		MediaSourceID:       source.ID,
		ClientProfile:       firstNonEmpty(r.URL.Query().Get("clientProfile"), "web"),
		Policy:              r.URL.Query().Get("policy"),
		RouteType:           r.URL.Query().Get("routeType"),
		MaxNetworkBitrate:   queryInt64(r, "maxNetworkBitrate", 0),
		AudioTrackIndex:     queryInt(r, "audioTrackIndex", 0),
		AudioCodec:          r.URL.Query().Get("audioCodec"),
		AudioChannels:       queryInt(r, "audioChannels", 0),
		SubtitleTrackIndex:  queryInt(r, "subtitleTrackIndex", 0),
		SubtitleCodec:       r.URL.Query().Get("subtitleCodec"),
		SubtitleMode:        r.URL.Query().Get("subtitleMode"),
		SubtitleTrackActive: r.URL.Query().Get("subtitleTrackActive") == "true",
		SupportsAdaptive:    r.URL.Query().Get("supportsAdaptive") == "true",
	}
	request = applyClientProfile(deps, request)
	return deps.Playback.DecideSource(ctx, request, playbackSourceFacts(ctx, deps, request, source))
}

func adaptivePlanForSource(deps Deps, source catalog.MediaSourceItem, clientProfile string, routeType string, maxNetworkBitrate int64) adaptive.Plan {
	profileID := firstNonEmpty(clientProfile, "web")
	profile, ok := deps.Devices.GetProfile(profileID)
	supportsHLS := ok && profile.SupportsHLS
	return adaptive.BuildPlan(adaptive.Request{
		MediaSourceID:     source.ID,
		ClientProfile:     profileID,
		RouteType:         routeType,
		SourceBitrate:     source.Bitrate,
		MaxNetworkBitrate: maxNetworkBitrate,
		Width:             source.Width,
		Height:            source.Height,
		SupportsHLS:       supportsHLS,
	})
}

func adaptiveVariantID(segment string) (string, bool) {
	if !strings.HasPrefix(segment, "variant-") || !strings.HasSuffix(segment, ".m3u8") {
		return "", false
	}
	id := strings.TrimSuffix(strings.TrimPrefix(segment, "variant-"), ".m3u8")
	if id == "" || strings.ContainsAny(id, `/\`) {
		return "", false
	}
	return id, true
}

func applyClientProfile(deps Deps, request playback.Request) playback.Request {
	profile, ok := deps.Devices.GetProfile(request.ClientProfile)
	if !ok {
		return request
	}
	request.ClientProfile = profile.ID
	request.Containers = profile.Containers
	request.VideoCodecs = profile.VideoCodecs
	request.AudioCodecs = profile.AudioCodecs
	request.SubtitleCodecs = profile.SubtitleCodecs
	request.SupportsAdaptive = request.SupportsAdaptive || profile.SupportsHLS
	return request
}

func playbackSourceFacts(ctx context.Context, deps Deps, request playback.Request, source catalog.MediaSourceItem) playback.SourceFacts {
	tracks, _, _ := deps.Catalog.GetMediaSourceTracks(ctx, request.MediaSourceID)
	if request.AudioCodec == "" && len(tracks.AudioTracks) > 0 {
		audio := trackByIndex(tracks.AudioTracks, request.AudioTrackIndex)
		request.AudioTrackIndex = audio.Index
		request.AudioCodec = audio.Codec
		request.AudioChannels = audio.Channels
	}
	if request.SubtitleTrackActive && request.SubtitleCodec == "" && len(tracks.SubtitleTracks) > 0 {
		subtitle := trackByIndex(tracks.SubtitleTracks, request.SubtitleTrackIndex)
		request.SubtitleTrackIndex = subtitle.Index
		request.SubtitleCodec = subtitle.Codec
	}
	return playback.SourceFacts{
		MediaSourceID:    source.ID,
		Container:        source.Container,
		VideoCodec:       source.VideoCodec,
		Width:            source.Width,
		Height:           source.Height,
		AudioStreams:     source.AudioStreams,
		SubtitleStreams:  source.SubtitleStreams,
		SidecarSubtitles: len(subtitles.DiscoverSidecars(source.Path)),
		Bitrate:          source.Bitrate,
		Probed:           source.Probed,
		AudioCodec:       request.AudioCodec,
		AudioChannels:    request.AudioChannels,
		SubtitleCodec:    request.SubtitleCodec,
		SubtitleActive:   request.SubtitleTrackActive,
	}
}

func trackByIndex(tracks []probe.Track, index int) probe.Track {
	for _, track := range tracks {
		if track.Index == index {
			return track
		}
	}
	return tracks[0]
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

func firstLibraryPathByKind(service *libraries.Service, kind libraries.Kind) string {
	if service == nil {
		return ""
	}
	for _, library := range service.List() {
		if library.Kind == kind {
			return strings.TrimSpace(library.Path)
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

func requestBaseURL(r *http.Request, fallbackAddr string) string {
	host := strings.TrimSpace(r.Header.Get("X-Forwarded-Host"))
	if host == "" {
		host = strings.TrimSpace(r.Host)
	}
	if host == "" {
		host = fallbackAddr
	}
	scheme := "http"
	if requestSecure(r) {
		scheme = "https"
	}
	return scheme + "://" + host
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

func queryInt64(r *http.Request, key string, fallback int64) int64 {
	value := r.URL.Query().Get(key)
	if value == "" {
		return fallback
	}
	var output int64
	for _, ch := range value {
		if ch < '0' || ch > '9' {
			return fallback
		}
		output = output*10 + int64(ch-'0')
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
