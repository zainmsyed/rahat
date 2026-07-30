package web

import (
	"net/http"
	"path"
	"strings"
)

// prefersHTML reports whether the request is best treated as a browser page
// navigation. Only requests that explicitly include "text/html" in their Accept
// header are considered page navigations; requests with no Accept header or
// with JSON/API Accept headers are forwarded to the API mux.
func prefersHTML(r *http.Request) bool {
	accept := r.Header.Get("Accept")
	if accept == "" {
		return false
	}
	return strings.Contains(accept, "text/html")
}

// NewStaticHandler returns a handler that serves files from staticDir.
// Existing files are served directly. When a file does not exist and the
// request prefers HTML, index.html is served so the client-side router can
// take over. All other requests are forwarded to next.
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

		file, err := root.Open(requestPath)
		if err == nil {
			defer file.Close()
			stat, err := file.Stat()
			if err == nil && !stat.IsDir() {
				fs.ServeHTTP(w, r)
				return
			}
		}

		if prefersHTML(r) {
			http.ServeFile(w, r, path.Join(staticDir, "index.html"))
			return
		}

		next.ServeHTTP(w, r)
	})
}
