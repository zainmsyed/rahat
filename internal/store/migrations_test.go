package store

import (
	"context"
	"testing"
)

func TestMigrationsApplyInOrderOnCleanDatabase(t *testing.T) {
	ctx := context.Background()

	dir, err := migrationDir()
	if err != nil {
		t.Fatalf("migrationDir() error = %v", err)
	}
	names, err := migrationNames(dir)
	if err != nil {
		t.Fatalf("migrationNames() error = %v", err)
	}
	if len(names) == 0 {
		t.Fatal("no migrations found")
	}

	for _, name := range names {
		t.Run(name, func(t *testing.T) {
			db := openCleanTestDB(t)
			if err := ApplyMigrationsUpTo(ctx, db, name); err != nil {
				t.Fatalf("ApplyMigrationsUpTo(%q) error = %v", name, err)
			}
			applied, err := migrationApplied(ctx, db, name)
			if err != nil {
				t.Fatalf("migrationApplied(%q) error = %v", name, err)
			}
			if !applied {
				t.Fatalf("migration %q was not recorded as applied", name)
			}
		})
	}
}

func TestMigrationsAreIdempotent(t *testing.T) {
	ctx := context.Background()
	db := openTestDB(t)

	// The database was already migrated by openTestDB; a second pass must be a no-op.
	if err := ApplyMigrations(ctx, db); err != nil {
		t.Fatalf("second ApplyMigrations() error = %v", err)
	}

	dir, err := migrationDir()
	if err != nil {
		t.Fatalf("migrationDir() error = %v", err)
	}
	names, err := migrationNames(dir)
	if err != nil {
		t.Fatalf("migrationNames() error = %v", err)
	}
	for _, name := range names {
		applied, err := migrationApplied(ctx, db, name)
		if err != nil {
			t.Fatalf("migrationApplied(%q) error = %v", name, err)
		}
		if !applied {
			t.Fatalf("migration %q not recorded after idempotent re-run", name)
		}
	}
}
