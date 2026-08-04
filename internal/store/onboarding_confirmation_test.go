package store

import (
	"context"
	"testing"
)

func TestOnboardingConfirmationRepository(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	repo := NewOnboardingConfirmationRepository(sqlDB)

	userID := "user-abc-123"

	_, found, err := repo.Get(ctx, userID)
	if err != nil {
		t.Fatalf("Get() before record error = %v", err)
	}
	if found {
		t.Fatal("expected no confirmation before record")
	}

	if err := repo.Record(ctx, userID, true, ""); err != nil {
		t.Fatalf("Record() error = %v", err)
	}

	record, found, err := repo.Get(ctx, userID)
	if err != nil {
		t.Fatalf("Get() after record error = %v", err)
	}
	if !found {
		t.Fatal("expected confirmation to be found after record")
	}
	if !record.Delivered {
		t.Fatalf("expected Delivered=true, got %v", record.Delivered)
	}
	if record.FailedReason != "" {
		t.Fatalf("expected empty FailedReason, got %q", record.FailedReason)
	}
	if record.SentAt.IsZero() {
		t.Fatal("expected non-zero SentAt")
	}

	// Duplicate records should be ignored and the original state preserved.
	if err := repo.Record(ctx, userID, false, "should be ignored"); err != nil {
		t.Fatalf("duplicate Record() error = %v", err)
	}
	record2, _, _ := repo.Get(ctx, userID)
	if !record2.Delivered {
		t.Fatalf("duplicate record should not overwrite Delivered=true")
	}
}

func TestOnboardingConfirmationRepositoryRecordRequiresUserID(t *testing.T) {
	ctx := context.Background()
	sqlDB := openTestDB(t)
	repo := NewOnboardingConfirmationRepository(sqlDB)
	if err := repo.Record(ctx, "", true, ""); err == nil {
		t.Fatal("expected error when user id is empty")
	}
}
