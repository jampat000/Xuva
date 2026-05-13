package api

import (
	"net/http"
	"strings"

	"github.com/jampat000/Lorivo/server/internal/auth"
	"github.com/jampat000/Lorivo/server/internal/config"
)

const (
	devAuthBypassSessionID = "dev-auth-bypass"
	devAuthBypassUserID    = "dev-owner"
)

func devAuthBypassEnabled(deps Deps) bool {
	return config.DevAuthBypassActive(deps.Config)
}

func devAuthBypassAllowsRequest(r *http.Request) bool {
	if !requestHostIsLoopback(r) {
		return false
	}
	path := strings.TrimSpace(r.URL.Path)
	switch {
	case r.Method == http.MethodGet && path == "/api/auth/session":
		return true
	case r.Method == http.MethodPost && path == "/api/auth/logout":
		return true
	case r.Method == http.MethodPut && path == "/api/settings":
		return true
	case r.Method == http.MethodPut && path == "/api/settings/metadata-sources":
		return true
	case r.Method == http.MethodGet && path == "/api/settings/folders/browse":
		return true
	case r.Method == http.MethodPost && path == "/api/libraries":
		return true
	case r.Method == http.MethodDelete && strings.HasPrefix(path, "/api/libraries/"):
		return true
	case r.Method == http.MethodPost && (path == "/api/libraries/movies/scan" || path == "/api/libraries/tv/scan" || path == "/api/libraries/scan"):
		return true
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/api/libraries/") && strings.HasSuffix(path, "/scan"):
		return true
	case r.Method == http.MethodPut && path == "/api/metadata/match":
		return true
	case r.Method == http.MethodPost && (path == "/api/metadata/refresh" || path == "/api/metadata/refresh-batch"):
		return true
	case r.Method == http.MethodGet && path == "/api/pairing/requests":
		return true
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/api/pairing/requests/") && (strings.HasSuffix(path, "/approve") || strings.HasSuffix(path, "/deny")):
		return true
	case r.Method == http.MethodGet && path == "/api/devices":
		return true
	case r.Method == http.MethodPost && strings.HasPrefix(path, "/api/devices/") && strings.HasSuffix(path, "/revoke"):
		return true
	case r.Method == http.MethodGet && (path == "/api/sessions" || strings.HasPrefix(path, "/api/sessions/")):
		return true
	default:
		return false
	}
}

func devAuthBypassSession() auth.ResolvedSession {
	return auth.ResolvedSession{
		Principal: auth.Principal{
			ID:          devAuthBypassUserID,
			Username:    "development-owner",
			DisplayName: "Development Owner",
			Role:        "admin",
		},
		Session: auth.Session{
			ID:     devAuthBypassSessionID,
			UserID: devAuthBypassUserID,
		},
		DevBypass: true,
	}
}
