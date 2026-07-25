package main

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestWithCORSAllowsWebOriginAndDevOrigins(t *testing.T) {
	handler := withCORS(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		"http://localhost:5200",
		[]string{"http://127.0.0.1:5200", "http://192.168.1.50:5200"},
		"development",
	)

	for _, origin := range []string{"http://localhost:5200", "http://127.0.0.1:5200", "http://192.168.1.50:5200"} {
		req := httptest.NewRequest(http.MethodOptions, "/auth/access-link/exchange", http.NoBody)
		req.Header.Set("Origin", origin)
		rec := httptest.NewRecorder()
		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNoContent {
			t.Fatalf("origin %s: expected status %d, got %d", origin, http.StatusNoContent, rec.Code)
		}
		allowed := rec.Header().Get("Access-Control-Allow-Origin")
		if allowed != origin {
			t.Fatalf("origin %s: expected Allow-Origin %q, got %q", origin, origin, allowed)
		}
		if !strings.Contains(rec.Header().Get("Vary"), "Origin") {
			t.Fatalf("origin %s: expected Vary to include Origin, got %q", origin, rec.Header().Get("Vary"))
		}
	}
}

func TestWithCORSRejectsUnknownOrigin(t *testing.T) {
	handler := withCORS(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { w.WriteHeader(http.StatusOK) }),
		"http://localhost:5200",
		[]string{"http://127.0.0.1:5200"},
		"development",
	)

	req := httptest.NewRequest(http.MethodOptions, "/auth/access-link/exchange", http.NoBody)
	req.Header.Set("Origin", "http://evil.example.com")
	rec := httptest.NewRecorder()
	handler.ServeHTTP(rec, req)

	if rec.Header().Get("Access-Control-Allow-Origin") != "" {
		t.Fatalf("expected no Allow-Origin for untrusted origin, got %q", rec.Header().Get("Access-Control-Allow-Origin"))
	}
}
