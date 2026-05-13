package api

import (
	"net/http"
	"strings"
	"time"

	"github.com/jampat000/Lorivo/server/internal/auth"
)

const (
	roleAdmin    = "admin"
	roleStandard = "standard"
)

type routePolicy struct {
	Pattern string
	Group   string
	Action  string
	Roles   []string
}

func route(pattern string, group string, action string, roles ...string) routePolicy {
	return routePolicy{Pattern: pattern, Group: group, Action: action, Roles: roles}
}

var routePolicies = map[string]routePolicy{
	"GET /api/auth/session":         route("GET /api/auth/session", "auth", "session.read", roleAdmin, roleStandard),
	"POST /api/auth/logout":         route("POST /api/auth/logout", "auth", "session.logout", roleAdmin, roleStandard),
	"GET /api/users":                route("GET /api/users", "auth", "users.list", roleAdmin),
	"POST /api/users":               route("POST /api/users", "auth", "user.create", roleAdmin),
	"DELETE /api/users/{id}":        route("DELETE /api/users/{id}", "auth", "user.delete", roleAdmin),
	"POST /api/users/{id}/password": route("POST /api/users/{id}/password", "auth", "user.password.update", roleAdmin),

	"POST /api/libraries":             route("POST /api/libraries", "libraries", "library.save", roleAdmin),
	"DELETE /api/libraries/{id}":      route("DELETE /api/libraries/{id}", "libraries", "library.delete", roleAdmin),
	"POST /api/libraries/{id}/scan":   route("POST /api/libraries/{id}/scan", "libraries", "library.scan", roleAdmin),
	"POST /api/libraries/movies/scan": route("POST /api/libraries/movies/scan", "libraries", "library.scan.movies", roleAdmin),
	"POST /api/libraries/tv/scan":     route("POST /api/libraries/tv/scan", "libraries", "library.scan.tv", roleAdmin),
	"POST /api/libraries/scan":        route("POST /api/libraries/scan", "libraries", "library.scan.all", roleAdmin),

	"PUT /api/metadata/match":                 route("PUT /api/metadata/match", "metadata", "metadata.match", roleAdmin),
	"POST /api/metadata/refresh":              route("POST /api/metadata/refresh", "metadata", "metadata.refresh", roleAdmin),
	"POST /api/metadata/refresh-batch":        route("POST /api/metadata/refresh-batch", "metadata", "metadata.refresh.batch", roleAdmin),
	"GET /api/migrations/formats":             route("GET /api/migrations/formats", "migration", "migration.formats", roleAdmin),
	"GET /api/migrations/runs":                route("GET /api/migrations/runs", "migration", "migration.runs.list", roleAdmin),
	"GET /api/migrations/runs/{id}":           route("GET /api/migrations/runs/{id}", "migration", "migration.run.read", roleAdmin),
	"POST /api/migrations/dry-run":            route("POST /api/migrations/dry-run", "migration", "migration.dry_run", roleAdmin),
	"POST /api/migrations/import":             route("POST /api/migrations/import", "migration", "migration.import", roleAdmin),
	"POST /api/migrations/runs/{id}/rollback": route("POST /api/migrations/runs/{id}/rollback", "migration", "migration.rollback", roleAdmin),

	"GET /api/client/home":                route("GET /api/client/home", "client", "client.home", roleAdmin, roleStandard),
	"GET /api/client/movies/{id}":         route("GET /api/client/movies/{id}", "client", "client.movie.detail", roleAdmin, roleStandard),
	"GET /api/client/series/{id}":         route("GET /api/client/series/{id}", "client", "client.series.detail", roleAdmin, roleStandard),
	"POST /api/client/playback/start":     route("POST /api/client/playback/start", "client", "client.playback.start", roleAdmin, roleStandard),
	"PATCH /api/client/playback/{id}":     route("PATCH /api/client/playback/{id}", "client", "client.playback.heartbeat", roleAdmin, roleStandard),
	"POST /api/client/playback/{id}/stop": route("POST /api/client/playback/{id}/stop", "client", "client.playback.stop", roleAdmin, roleStandard),
	"PUT /api/settings":                   route("PUT /api/settings", "settings", "settings.update", roleAdmin),
	"PUT /api/settings/metadata-sources":  route("PUT /api/settings/metadata-sources", "settings", "settings.metadata_sources.update", roleAdmin),
	"GET /api/settings/folders/browse":    route("GET /api/settings/folders/browse", "settings", "settings.folders.browse", roleAdmin),
	"POST /api/settings/hardware/test":    route("POST /api/settings/hardware/test", "settings", "settings.hardware.test", roleAdmin),
	"POST /api/remote/diagnostics":        route("POST /api/remote/diagnostics", "remote", "remote.diagnostics.run", roleAdmin),
	"POST /api/remote/wan":                route("POST /api/remote/wan", "remote", "remote.wan.lookup", roleAdmin),

	"GET /api/media-sources/{id}/stream":                     route("GET /api/media-sources/{id}/stream", "media", "media.stream", roleAdmin, roleStandard),
	"GET /api/media-sources/{id}/adaptive/master.m3u8":       route("GET /api/media-sources/{id}/adaptive/master.m3u8", "media", "media.adaptive.master", roleAdmin, roleStandard),
	"GET /api/media-sources/{id}/adaptive/{variant}":         route("GET /api/media-sources/{id}/adaptive/{variant}", "media", "media.adaptive.variant", roleAdmin, roleStandard),
	"POST /api/media-sources/{id}/adaptive/session":          route("POST /api/media-sources/{id}/adaptive/session", "media", "media.adaptive.session", roleAdmin, roleStandard),
	"POST /api/adaptive/telemetry":                           route("POST /api/adaptive/telemetry", "media", "media.adaptive.telemetry", roleAdmin, roleStandard),
	"POST /api/media-sources/{id}/stream-token":              route("POST /api/media-sources/{id}/stream-token", "media", "media.stream.token", roleAdmin, roleStandard),
	"GET /api/media-sources/{id}/subtitles/{index}":          route("GET /api/media-sources/{id}/subtitles/{index}", "media", "media.subtitle.stream", roleAdmin, roleStandard),
	"POST /api/media-sources/{id}/subtitles/{index}/convert": route("POST /api/media-sources/{id}/subtitles/{index}/convert", "media", "media.subtitle.convert", roleAdmin, roleStandard),
	"POST /api/media-sources/{id}/probe":                     route("POST /api/media-sources/{id}/probe", "media", "media.probe", roleAdmin),
	"POST /api/probes":                                       route("POST /api/probes", "media", "probe.start", roleAdmin),
	"POST /api/work":                                         route("POST /api/work", "work", "work.start", roleAdmin),
	"DELETE /api/work/{id}":                                  route("DELETE /api/work/{id}", "work", "work.cancel", roleAdmin),
	"GET /api/work/{id}/file":                                route("GET /api/work/{id}/file", "work", "work.file", roleAdmin, roleStandard),
	"POST /api/downloads":                                    route("POST /api/downloads", "downloads", "download.start", roleAdmin),
	"GET /api/downloads/{id}/file":                           route("GET /api/downloads/{id}/file", "downloads", "download.file", roleAdmin, roleStandard),
	"GET /api/pairing/requests":                              route("GET /api/pairing/requests", "pairing", "pairing.list", roleAdmin),
	"POST /api/pairing/requests/{id}/approve":                route("POST /api/pairing/requests/{id}/approve", "pairing", "pairing.approve", roleAdmin),
	"POST /api/pairing/requests/{id}/deny":                   route("POST /api/pairing/requests/{id}/deny", "pairing", "pairing.deny", roleAdmin),
	"GET /api/devices":                                       route("GET /api/devices", "devices", "devices.list", roleAdmin),
	"POST /api/devices/{id}/revoke":                          route("POST /api/devices/{id}/revoke", "devices", "devices.revoke", roleAdmin),
	"GET /api/sessions":                                      route("GET /api/sessions", "sessions", "sessions.list", roleAdmin, roleStandard),
	"GET /api/sessions/{id}/inspector":                       route("GET /api/sessions/{id}/inspector", "sessions", "sessions.inspector", roleAdmin, roleStandard),
	"POST /api/sessions":                                     route("POST /api/sessions", "sessions", "session.start", roleAdmin, roleStandard),
	"PATCH /api/sessions/{id}":                               route("PATCH /api/sessions/{id}", "sessions", "session.update", roleAdmin, roleStandard),
	"DELETE /api/sessions/{id}":                              route("DELETE /api/sessions/{id}", "sessions", "session.stop", roleAdmin, roleStandard),
	"PUT /api/playback/state/{id}":                           route("PUT /api/playback/state/{id}", "playback", "playback.state.update", roleAdmin, roleStandard),
	"GET /play/{id}":                                         route("GET /play/{id}", "playback", "player.open", roleAdmin, roleStandard),
}

func handleProtected(mux *http.ServeMux, deps Deps, pattern string, next http.HandlerFunc) {
	mux.HandleFunc(pattern, requireRoutePolicy(deps, routePolicies[pattern], false, next))
}

func handleProtectedCSRF(mux *http.ServeMux, deps Deps, pattern string, next http.HandlerFunc) {
	mux.HandleFunc(pattern, requireRoutePolicy(deps, routePolicies[pattern], true, next))
}

func requireRoutePolicy(deps Deps, policy routePolicy, requireCSRF bool, next http.HandlerFunc) http.HandlerFunc {
	return requireAuth(deps, func(w http.ResponseWriter, r *http.Request) {
		if deps.Auth == nil || deps.Auth.Disabled() {
			next(w, r)
			return
		}
		resolved, _ := auth.ResolvedSessionFromContext(r.Context())
		if !policyAllows(policy, resolved.Principal.Role) {
			publishAudit(deps, r, policy, resolved, "denied", "role_not_allowed")
			writeError(w, http.StatusForbidden, "role is not allowed for this action")
			return
		}
		if requireCSRF {
			if resolved.DevBypass {
				publishAudit(deps, r, policy, resolved, "allowed", "dev_auth_bypass")
				next(w, r)
				return
			}
			if hasHeaderAuthToken(r) {
				publishAudit(deps, r, policy, resolved, "allowed", "header_token")
				next(w, r)
				return
			}
			csrfCookie, _ := r.Cookie(auth.CSRFCookieName)
			csrfCookieValue := ""
			if csrfCookie != nil {
				csrfCookieValue = csrfCookie.Value
			}
			if err := deps.Auth.ValidateCSRF(resolved, csrfCookieValue, r.Header.Get("X-CSRF-Token")); err != nil {
				publishAudit(deps, r, policy, resolved, "denied", "csrf_invalid")
				writeError(w, http.StatusForbidden, "valid csrf token required")
				return
			}
		}
		publishAudit(deps, r, policy, resolved, "allowed", "")
		next(w, r)
	})
}

func policyAllows(policy routePolicy, role string) bool {
	role = strings.ToLower(strings.TrimSpace(role))
	for _, allowed := range policy.Roles {
		if role == allowed {
			return true
		}
	}
	return false
}

func publishAudit(deps Deps, r *http.Request, policy routePolicy, resolved auth.ResolvedSession, result string, reason string) {
	if deps.Events == nil {
		return
	}
	deps.Events.Publish("audit.route", map[string]any{
		"userId":    resolved.Principal.ID,
		"username":  resolved.Principal.Username,
		"role":      resolved.Principal.Role,
		"method":    r.Method,
		"path":      r.URL.Path,
		"pattern":   policy.Pattern,
		"group":     policy.Group,
		"action":    policy.Action,
		"result":    result,
		"reason":    reason,
		"createdAt": time.Now().UTC().Format(time.RFC3339Nano),
	})
}
