package web

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestNewStaticHandler(t *testing.T) {
	dir := t.TempDir()
	_ = os.WriteFile(filepath.Join(dir, "index.html"), []byte("<html>app</html>"), 0o644)
	_ = os.WriteFile(filepath.Join(dir, "asset.txt"), []byte("static asset"), 0o644)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
		_, _ = w.Write([]byte("next"))
	})

	handler := NewStaticHandler(dir, next)

	t.Run("serves existing static files", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/asset.txt", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		if got := rr.Body.String(); got != "static asset" {
			t.Fatalf("expected static asset body, got %q", got)
		}
	})

	t.Run("falls back to index.html for unknown non-API paths", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, "/login", nil)
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("expected status 200, got %d", rr.Code)
		}
		if got := rr.Body.String(); got != "<html>app</html>" {
			t.Fatalf("expected index.html body, got %q", got)
		}
	})

	t.Run("passes API paths to next handler", func(t *testing.T) {
		for _, path := range []string{"/healthz", "/readyz", "/auth/session", "/tasks", "/lookahead/plan"} {
			req := httptest.NewRequest(http.MethodGet, path, nil)
			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)
			if rr.Code != http.StatusTeapot {
				t.Fatalf("path %s: expected status 418, got %d", path, rr.Code)
			}
		}
	})

	t.Run("passes non-GET methods to next handler", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, "/asset.txt", strings.NewReader(""))
		rr := httptest.NewRecorder()
		handler.ServeHTTP(rr, req)
		if rr.Code != http.StatusTeapot {
			t.Fatalf("expected status 418, got %d", rr.Code)
		}
	})
}

func TestNewStaticHandlerDisabled(t *testing.T) {
	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusTeapot)
	})
	handler := NewStaticHandler("", next)

	req := httptest.NewRequest(http.MethodGet, "/", nil)
	rr := httptest.NewRecorder()
	handler.ServeHTTP(rr, req)
	if rr.Code != http.StatusTeapot {
		t.Fatalf("expected next handler when static dir is empty, got %d", rr.Code)
	}
}
