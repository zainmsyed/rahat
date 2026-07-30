package web

import (
	"net/http"
	"path"
	"strings"
)

var apiPrefixes = []string{
	"/healthz",
	"/readyz",
	"/auth",
	"/onboarding",
	"/tasks",
	"/calendar",
	"/schedule",
	"/lookahead",
	"/telegram",
	"/webhooks",
}

func isAPIPath(requestPath string) bool {
	for _, prefix := range apiPrefixes {
		if requestPath == prefix || strings.HasPrefix(requestPath, prefix+"/") {
			return true
		}
	}
	return false
}

// NewStaticHandler returns a handler that serves files from staticDir.
// Requests matching known API paths are forwarded to next. If the requested
// file does not exist, it falls back to index.html so a client-side router
// can take over.
func NewStaticHandler(staticDir string, next http.Handler) http.Handler {
	if staticDir == "" {
		return next
	}

	fs := http.FileServer(http.Dir(staticDir))
	root := http.Dir(staticDir)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			next.ServeHTTP(w, r)
			return
		}

		requestPath := path.Clean(r.URL.Path)
		if requestPath == "." {
			requestPath = "/"
		}

		if isAPIPath(requestPath) {
			next.ServeHTTP(w, r)
			return
		}

		file, err := root.Open(requestPath)
		if err == nil {
			defer file.Close()
			stat, err := file.Stat()
			if err == nil && !stat.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}

		http.ServeFile(w, r, path.Join(staticDir, "index.html"))
	})
}
