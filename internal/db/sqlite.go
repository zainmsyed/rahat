package db

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"

	_ "modernc.org/sqlite"
)

func OpenSQLite(ctx context.Context, databasePath string) (*sql.DB, error) {
	if err := os.MkdirAll(filepath.Dir(databasePath), 0o755); err != nil {
		return nil, fmt.Errorf("create database directory: %w", err)
	}

	db, err := sql.Open("sqlite", databasePath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite database: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA journal_mode = WAL;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable wal mode: %w", err)
	}

	if _, err := db.ExecContext(ctx, "PRAGMA foreign_keys = ON;"); err != nil {
		db.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	if _, err := db.ExecContext(ctx, `
		CREATE TABLE IF NOT EXISTS schema_migrations (
			id INTEGER PRIMARY KEY AUTOINCREMENT,
			name TEXT NOT NULL UNIQUE,
			applied_at TEXT NOT NULL DEFAULT CURRENT_TIMESTAMP
		);
	`); err != nil {
		db.Close()
		return nil, fmt.Errorf("create bootstrap table: %w", err)
	}

	if err := db.PingContext(ctx); err != nil {
		db.Close()
		return nil, fmt.Errorf("ping sqlite database: %w", err)
	}

	return db, nil
}
