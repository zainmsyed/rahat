package tokens

import (
	"testing"
	"time"
)

func TestIssueAndVerify(t *testing.T) {
	mgr := NewManager("0123456789abcdef")
	now := time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC)
	mgr.now = func() time.Time { return now }

	token, err := mgr.Issue("user-123", time.Hour)
	if err != nil {
		t.Fatalf("Issue() error = %v", err)
	}
	claims, err := mgr.Verify(token)
	if err != nil {
		t.Fatalf("Verify() error = %v", err)
	}
	if claims.UserID != "user-123" {
		t.Fatalf("UserID = %q, want user-123", claims.UserID)
	}
}

func TestVerifyRejectsTamperedAndExpiredTokens(t *testing.T) {
	mgr := NewManager("0123456789abcdef")
	now := time.Date(2026, 7, 24, 9, 0, 0, 0, time.UTC)
	mgr.now = func() time.Time { return now }
	token, err := mgr.Issue("user-123", time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := mgr.Verify(token + "tampered"); err != ErrInvalid {
		t.Fatalf("tampered Verify() error = %v, want ErrInvalid", err)
	}
	mgr.now = func() time.Time { return now.Add(2 * time.Hour) }
	if _, err := mgr.Verify(token); err != ErrExpired {
		t.Fatalf("expired Verify() error = %v, want ErrExpired", err)
	}
}

func TestUnavailableWhenSecretTooShort(t *testing.T) {
	mgr := NewManager("short")
	if mgr.Available() {
		t.Fatal("expected manager unavailable")
	}
	if _, err := mgr.Issue("user-123", time.Hour); err != ErrUnavailable {
		t.Fatalf("Issue() error = %v, want ErrUnavailable", err)
	}
	if _, err := mgr.Verify("anything"); err != ErrUnavailable {
		t.Fatalf("Verify() error = %v, want ErrUnavailable", err)
	}
}
