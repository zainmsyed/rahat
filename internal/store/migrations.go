package store

import (
	"context"
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
)

func ApplyMigrations(ctx context.Context, db *sql.DB) error {
	return applyMigrations(ctx, db, "")
}

// ApplyMigrationsUpTo applies all unapplied migrations up to and including the
// named migration. It is intended for tests that need to recreate a historical
// schema state before applying later migrations.
func ApplyMigrationsUpTo(ctx context.Context, db *sql.DB, name string) error {
	return applyMigrations(ctx, db, name)
}

func applyMigrations(ctx context.Context, db *sql.DB, stopAt string) error {
	migrationDir, err := migrationDir()
	if err != nil {
		return err
	}

	names, err := migrationNames(migrationDir)
	if err != nil {
		return err
	}

	foundStop := stopAt == ""
	for _, name := range names {
		if stopAt != "" && name > stopAt {
			break
		}
		if name == stopAt {
			foundStop = true
		}

		applied, err := migrationApplied(ctx, db, name)
		if err != nil {
			return err
		}
		if applied {
			continue
		}

		sqlBytes, err := os.ReadFile(filepath.Join(migrationDir, name))
		if err != nil {
			return fmt.Errorf("read migration %s: %w", name, err)
		}

		tx, err := db.BeginTx(ctx, nil)
		if err != nil {
			return fmt.Errorf("begin migration %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, string(sqlBytes)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("apply migration %s: %w", name, err)
		}

		if _, err := tx.ExecContext(ctx, `INSERT INTO schema_migrations(name) VALUES (?)`, name); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("record migration %s: %w", name, err)
		}

		if err := tx.Commit(); err != nil {
			return fmt.Errorf("commit migration %s: %w", name, err)
		}
	}

	if stopAt != "" && !foundStop {
		return fmt.Errorf("stop migration %q not found", stopAt)
	}

	return nil
}

func migrationNames(migrationDir string) ([]string, error) {
	entries, err := os.ReadDir(migrationDir)
	if err != nil {
		return nil, fmt.Errorf("read migrations: %w", err)
	}

	names := make([]string, 0, len(entries))
	for _, entry := range entries {
		if entry.IsDir() || !strings.HasSuffix(entry.Name(), ".sql") {
			continue
		}
		names = append(names, entry.Name())
	}
	sort.Strings(names)
	return names, nil
}

func migrationApplied(ctx context.Context, db *sql.DB, name string) (bool, error) {
	var count int
	if err := db.QueryRowContext(ctx, `SELECT COUNT(*) FROM schema_migrations WHERE name = ?`, name).Scan(&count); err != nil {
		return false, fmt.Errorf("check migration %s: %w", name, err)
	}
	return count > 0, nil
}

func migrationDir() (string, error) {
	_, file, _, ok := runtime.Caller(0)
	if !ok {
		return "", fmt.Errorf("locate migrations: runtime caller unavailable")
	}
	return filepath.Clean(filepath.Join(filepath.Dir(file), "..", "..", "db", "migrations")), nil
}
