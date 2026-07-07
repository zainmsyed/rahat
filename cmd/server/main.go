package main

import (
	"context"
	"errors"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rahat/rahat/internal/app"
	"github.com/rahat/rahat/internal/config"
	"github.com/rahat/rahat/internal/db"
)

func main() {
	cfg := config.Load()
	logger := slog.New(slog.NewJSONHandler(os.Stdout, &slog.HandlerOptions{Level: cfg.LogLevel}))

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	sqlDB, err := db.OpenSQLite(ctx, cfg.DatabasePath)
	if err != nil {
		logger.Error("failed to initialize database", "error", err)
		os.Exit(1)
	}
	defer sqlDB.Close()

	if len(os.Args) > 1 && os.Args[1] == "db:setup" {
		logger.Info("database setup complete", "database_path", cfg.DatabasePath)
		return
	}

	server := &http.Server{
		Addr:              cfg.HTTPAddr,
		Handler:           app.NewServer(logger, cfg, sqlDB),
		ReadHeaderTimeout: 5 * time.Second,
	}

	go func() {
		<-ctx.Done()

		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		if err := server.Shutdown(shutdownCtx); err != nil {
			logger.Error("server shutdown failed", "error", err)
		}
	}()

	logger.Info("starting rahat api", "addr", cfg.HTTPAddr, "env", cfg.AppEnv, "database_path", cfg.DatabasePath)

	if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		logger.Error("server stopped with error", "error", err)
		os.Exit(1)
	}

	logger.Info("server stopped")
}
