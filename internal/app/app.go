package app

import (
	"database/sql"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/rahat/rahat/internal/config"
)

type Server struct {
	logger *slog.Logger
	config config.Config
	db     *sql.DB
	mux    *http.ServeMux
}

func NewServer(logger *slog.Logger, cfg config.Config, db *sql.DB) http.Handler {
	s := &Server{
		logger: logger,
		config: cfg,
		db:     db,
		mux:    http.NewServeMux(),
	}

	s.routes()

	return s.withLogging(s.mux)
}

func (s *Server) routes() {
	s.mux.HandleFunc("GET /healthz", s.handleHealth)
	s.mux.HandleFunc("GET /readyz", s.handleReady)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, map[string]any{
		"status":    "ok",
		"service":   "rahat-api",
		"timestamp": time.Now().UTC().Format(time.RFC3339),
	})
}

func (s *Server) handleReady(w http.ResponseWriter, r *http.Request) {
	if err := s.db.PingContext(r.Context()); err != nil {
		writeJSON(w, http.StatusServiceUnavailable, map[string]any{
			"status": "not-ready",
			"error":  err.Error(),
		})
		return
	}

	writeJSON(w, http.StatusOK, map[string]any{
		"status":        "ready",
		"database":      "ok",
		"environment":   s.config.AppEnv,
		"database_path": s.config.DatabasePath,
	})
}

func (s *Server) withLogging(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		started := time.Now()
		rw := &responseWriter{ResponseWriter: w, status: http.StatusOK}
		next.ServeHTTP(rw, r)

		s.logger.Info("request completed",
			"method", r.Method,
			"path", r.URL.Path,
			"status", rw.status,
			"duration_ms", time.Since(started).Milliseconds(),
			"remote_addr", r.RemoteAddr,
		)
	})
}

type responseWriter struct {
	http.ResponseWriter
	status int
}

func (rw *responseWriter) WriteHeader(status int) {
	rw.status = status
	rw.ResponseWriter.WriteHeader(status)
}

func writeJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(payload)
}
