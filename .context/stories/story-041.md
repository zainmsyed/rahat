# Story 041: Align migration tests with production bootstrap

**Status:** complete  
**Type:** test-hardening  
**Created:** 2026-08-03  
**Last accessed:** 2026-08-04  
**Completed:** 2026-08-04

---

## Goal
Make migration integration tests exercise the real production migration/bootstrap path instead of narrow fixtures that mark migrations as applied without creating their tables. New migrations must be covered by tests that apply through the same sequence used by application startup.

## Verification
The full Go test suite passes after new migrations are added. In particular, the Telegram identity migration integration test can apply the complete migration chain, including `013_story_037_day_preference.sql`, without `no such table: tasks` errors.

## Scope — files this story may touch
- internal/store/migrations.go
- internal/store/story002_integration_test.go
- internal/store/story014_integration_test.go
- internal/store/story037_integration_test.go (new, if needed)
- internal/store/onboarding_confirmation_test.go
- internal/db/sqlite_test.go
- cmd/server/main_test.go
- README.md

## Out of scope — do not touch
- Scheduler behavior
- Telegram runtime behavior
- Frontend code
- Migration content already applied to production, unless a compatibility issue is identified

## Dependencies
- Story 002
- Story 014
- Story 037

---

## Checklist
- [x] Reproduce the Story 014 migration fixture failure against the current migration chain.
- [x] Refactor migration tests to start from the real bootstrap schema/migration sequence instead of recording later migrations as pre-applied without their tables.
- [x] Add a migration-chain test that validates every migration can apply in order on a clean database.
- [x] Keep the duplicate Telegram identity conflict coverage intact.
- [x] Run the full Go test suite through the available Go container and document results.

---

## Issues

- Backend verification on 2026-08-03 failed in `TestMigrationResolvesDuplicateTelegramChatIDs` because the test marked migrations 001–010 as applied but only manually created `users`, causing migration 013 to fail with `no such table: tasks`.
- **Resolved:** `TestMigrationResolvesDuplicateTelegramChatIDs` now applies migrations 001–010 through the real bootstrap path, seeds duplicate users, and then applies the remaining migrations. The full Go test suite passes in the `golang:1.25` container.

---

## Completion Summary

- Added `ApplyMigrationsUpTo` to `internal/store/migrations.go` so tests can stop the migration chain at a specific migration and recreate historical schema states.
- Consolidated shared test helpers and migration-chain tests into the allowed `internal/store/story014_integration_test.go` file. The helpers mirror the production SQLite bootstrap path (WAL, foreign keys, `schema_migrations` table, then `ApplyMigrations`).
- Rewrote `internal/store/story014_integration_test.go` to apply migrations 001–010 for real, seed two users with the same Telegram chat ID, and then apply the remaining migrations, preserving the duplicate-conflict assertions. It also contains:
  - `TestMigrationsApplyInOrderOnCleanDatabase` — applies each migration in its own fresh database and verifies it is recorded.
  - `TestMigrationsAreIdempotent` — applies the full chain twice and confirms all migrations remain recorded.
- Simplified `internal/store/onboarding_confirmation_test.go` to use the shared `openTestDB` helper, removing its hand-rolled fixture.
- Ran `gofmt`, `go vet ./internal/store/...`, and the full `go test ./...` suite inside the `golang:1.25` container. All packages pass.
