package auth

import (
	"context"
	"errors"
	"path/filepath"
	"testing"
	"time"

	"github.com/rahat/rahat/internal/db"
)

func newTestService(t *testing.T) *Service {
	t.Helper()
	sqlDB, err := db.OpenSQLite(context.Background(), filepath.Join(t.TempDir(), "rahat.sqlite3"))
	if err != nil {
		t.Fatalf("OpenSQLite() error = %v", err)
	}
	t.Cleanup(func() { _ = sqlDB.Close() })
	return NewService(sqlDB, NewRepository(sqlDB), "test-secret", 14*24*time.Hour)
}

func TestIssueExchangeVerifyAndRevoke(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	if _, err := svc.db.ExecContext(ctx, `INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, created_at, updated_at) VALUES ('u1', 'Tester', 'UTC', 30, '2026-07-25T12:00:00Z', '2026-07-25T12:00:00Z')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	_, accessToken, err := svc.IssueAccessGrant(ctx, "u1", time.Hour)
	if err != nil {
		t.Fatalf("IssueAccessGrant() error = %v", err)
	}
	session, sessionToken, err := svc.ExchangeAccessGrant(ctx, accessToken)
	if err != nil {
		t.Fatalf("ExchangeAccessGrant() error = %v", err)
	}
	if session.UserID != "u1" || sessionToken == "" {
		t.Fatalf("unexpected session exchange: %+v token=%q", session, sessionToken)
	}
	if _, _, err := svc.ExchangeAccessGrant(ctx, accessToken); !errors.Is(err, ErrUsed) {
		t.Fatalf("expected ErrUsed, got %v", err)
	}
	verified, err := svc.VerifySession(ctx, sessionToken)
	if err != nil {
		t.Fatalf("VerifySession() error = %v", err)
	}
	if verified.UserID != "u1" {
		t.Fatalf("verified user = %q, want u1", verified.UserID)
	}
	if err := svc.RevokeSession(ctx, sessionToken); err != nil {
		t.Fatalf("RevokeSession() error = %v", err)
	}
	if _, err := svc.VerifySession(ctx, sessionToken); !errors.Is(err, ErrRevoked) {
		t.Fatalf("expected ErrRevoked, got %v", err)
	}
}

func TestExpiredAccessGrantRejected(t *testing.T) {
	svc := newTestService(t)
	ctx := context.Background()
	if _, err := svc.db.ExecContext(ctx, `INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, created_at, updated_at) VALUES ('u1', 'Tester', 'UTC', 30, '2026-07-25T12:00:00Z', '2026-07-25T12:00:00Z')`); err != nil {
		t.Fatalf("seed user: %v", err)
	}
	now := time.Date(2026, 7, 25, 12, 0, 0, 0, time.UTC)
	svc.now = func() time.Time { return now }
	_, token, err := svc.IssueAccessGrant(ctx, "u1", time.Minute)
	if err != nil {
		t.Fatalf("IssueAccessGrant() error = %v", err)
	}
	svc.now = func() time.Time { return now.Add(2 * time.Minute) }
	if _, _, err := svc.ExchangeAccessGrant(ctx, token); !errors.Is(err, ErrExpired) {
		t.Fatalf("expected ErrExpired, got %v", err)
	}
}
