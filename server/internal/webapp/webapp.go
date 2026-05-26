package webapp

import (
	"embed"
	"io/fs"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"path"
	"strings"
)

//go:embed all:static-next/_app
//go:embed all:static-next/avatars
//go:embed static-next/index.html
//go:embed static-next/robots.txt
//go:embed static-next/favicon.ico static-next/favicon.svg
//go:embed static-next/favicon-16.png static-next/favicon-32.png
//go:embed static-next/favicon-48.png static-next/favicon-64.png
//go:embed static-next/apple-touch-icon.png
var assets embed.FS

func RootHandler() http.Handler {
	if proxy := devProxyHandler(); proxy != nil {
		return proxy
	}

	staticFS, err := fs.Sub(assets, "static-next")
	if err != nil {
		panic(err)
	}
	fileServer := http.FileServer(http.FS(staticFS))
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/legacy/") ||
			r.URL.Path == "/legacy" ||
			strings.HasPrefix(r.URL.Path, "/next/") ||
			r.URL.Path == "/next" {
			http.NotFound(w, r)
			return
		}

		relativePath, ok := sanitizeRootRelativePath(r.URL.Path)
		if !ok {
			http.NotFound(w, r)
			return
		}

		if relativePath == "" {
			serveSPAIndex(w, staticFS, "Xuva")
			return
		}

		info, err := fs.Stat(staticFS, relativePath)
		if err == nil && !info.IsDir() {
			if relativePath == "build-info.json" || strings.HasSuffix(relativePath, ".html") {
				setNoStoreHeaders(w)
			} else if shouldDisableAssetCaching() {
				setNoStoreHeaders(w)
			} else {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			}
			fileServer.ServeHTTP(w, r)
			return
		}

		if path.Ext(relativePath) != "" {
			http.NotFound(w, r)
			return
		}

		serveSPAIndex(w, staticFS, "Xuva")
	})
}

func sanitizeRootRelativePath(urlPath string) (string, bool) {
	if urlPath == "/" {
		return "", true
	}
	trimmed := strings.TrimPrefix(urlPath, "/")
	if trimmed == "" {
		return "", true
	}
	cleaned := path.Clean("/" + trimmed)
	if strings.HasPrefix(cleaned, "/..") {
		return "", false
	}
	relative := strings.TrimPrefix(cleaned, "/")
	if relative == "." {
		return "", true
	}
	return relative, true
}

func serveSPAIndex(w http.ResponseWriter, staticFS fs.FS, title string) {
	index, err := fs.ReadFile(staticFS, "index.html")
	if err != nil {
		setNoStoreHeaders(w)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte(`<!doctype html><html lang="en"><head><meta charset="utf-8"><title>` + title + `</title></head><body><main><h1>` + title + `</h1><p>Publish the Svelte build with: <code>npm --prefix apps/web/svelte run publish:go-static</code></p></main></body></html>`))
		return
	}
	setNoStoreHeaders(w)
	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write(index)
}

func setNoStoreHeaders(w http.ResponseWriter) {
	w.Header().Set("Cache-Control", "no-store, max-age=0")
	w.Header().Set("Pragma", "no-cache")
	w.Header().Set("Expires", "0")
}

func shouldDisableAssetCaching() bool {
	value := strings.TrimSpace(strings.ToLower(os.Getenv("XUVA_WEB_DISABLE_ASSET_CACHE")))
	return value == "1" || value == "true" || value == "yes"
}

func devProxyHandler() http.Handler {
	origin := strings.TrimSpace(os.Getenv("XUVA_WEB_DEV_ORIGIN"))
	if origin == "" {
		return nil
	}
	target, err := url.Parse(origin)
	if err != nil || target.Scheme == "" || target.Host == "" {
		return nil
	}
	proxy := httputil.NewSingleHostReverseProxy(target)
	originalDirector := proxy.Director
	proxy.Director = func(req *http.Request) {
		originalDirector(req)
		req.Host = target.Host
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/api/") ||
			strings.HasPrefix(r.URL.Path, "/legacy/") ||
			r.URL.Path == "/legacy" ||
			strings.HasPrefix(r.URL.Path, "/next/") ||
			r.URL.Path == "/next" {
			http.NotFound(w, r)
			return
		}
		proxy.ServeHTTP(w, r)
	})
}
