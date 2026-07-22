package app

import (
	"context"
	"encoding/json"
	"io"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/rahat/rahat/internal/config"
	"github.com/rahat/rahat/internal/db"
)

func TestHealthEndpoint(t *testing.T) {
	t.Parallel()

	handler := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/healthz", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["status"] != "ok" {
		t.Fatalf("status payload = %v, want ok", payload["status"])
	}
	if payload["service"] != "rahat-api" {
		t.Fatalf("service payload = %v, want rahat-api", payload["service"])
	}
	if _, ok := payload["timestamp"].(string); !ok {
		t.Fatalf("timestamp payload missing or not a string: %v", payload["timestamp"])
	}
}

func TestReadyEndpoint(t *testing.T) {
	t.Parallel()

	handler := newTestServer(t)
	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)

	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusOK {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusOK)
	}

	var payload map[string]any
	if err := json.Unmarshal(recorder.Body.Bytes(), &payload); err != nil {
		t.Fatalf("decode response: %v", err)
	}

	if payload["status"] != "ready" {
		t.Fatalf("status payload = %v, want ready", payload["status"])
	}
	if payload["database"] != "ok" {
		t.Fatalf("database payload = %v, want ok", payload["database"])
	}
	if payload["environment"] != "test" {
		t.Fatalf("environment payload = %v, want test", payload["environment"])
	}
}

func TestReadyEndpointReturnsServiceUnavailableWhenDatabaseIsClosed(t *testing.T) {
	t.Parallel()

	cfg := config.Config{
		AppEnv:       "test",
		DatabasePath: filepath.Join(t.TempDir(), "rahat.sqlite3"),
	}

	sqlDB, err := db.OpenSQLite(context.Background(), cfg.DatabasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	handler := NewServer(logger, cfg, sqlDB)

	if err := sqlDB.Close(); err != nil {
		t.Fatalf("Close() error = %v", err)
	}

	recorder := httptest.NewRecorder()
	request := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	handler.ServeHTTP(recorder, request)

	if recorder.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want %d", recorder.Code, http.StatusServiceUnavailable)
	}
}

func newTestServer(t *testing.T) http.Handler {
	t.Helper()

	cfg := config.Config{
		AppEnv:       "test",
		DatabasePath: filepath.Join(t.TempDir(), "rahat.sqlite3"),
	}

	sqlDB, err := db.OpenSQLite(context.Background(), cfg.DatabasePath)
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() {
		_ = sqlDB.Close()
	})

	logger := slog.New(slog.NewTextHandler(io.Discard, nil))
	return NewServer(logger, cfg, sqlDB)
}
