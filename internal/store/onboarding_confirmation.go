package store

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"
)

// OnboardingConfirmation records whether a Telegram onboarding summary has
// already been sent for a user. It is used to keep the finish flow idempotent:
// repeated calls should not deliver duplicate confirmations.
type OnboardingConfirmation struct {
	UserID       string
	Delivered    bool
	FailedReason string
	SentAt       time.Time
}

// OnboardingConfirmationRepository persists onboarding confirmation state.
type OnboardingConfirmationRepository struct {
	db *sql.DB
}

// NewOnboardingConfirmationRepository creates a new repository.
func NewOnboardingConfirmationRepository(db *sql.DB) *OnboardingConfirmationRepository {
	return &OnboardingConfirmationRepository{db: db}
}

// Get returns the stored confirmation for a user, if any.
func (r *OnboardingConfirmationRepository) Get(ctx context.Context, userID string) (OnboardingConfirmation, bool, error) {
	var row OnboardingConfirmation
	var delivered int
	var failedReason sql.NullString
	var sentAt string
	err := r.db.QueryRowContext(ctx, `
		SELECT user_id, delivered, failed_reason, sent_at
		FROM onboarding_telegram_confirmations
		WHERE user_id = ?
	`, userID).Scan(&row.UserID, &delivered, &failedReason, &sentAt)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return OnboardingConfirmation{}, false, nil
		}
		return OnboardingConfirmation{}, false, fmt.Errorf("get onboarding confirmation: %w", err)
	}
	row.Delivered = delivered == 1
	row.FailedReason = failedReason.String
	var parseErr error
	row.SentAt, parseErr = ParseTime(sentAt)
	if parseErr != nil {
		return OnboardingConfirmation{}, false, fmt.Errorf("parse onboarding confirmation sent_at: %w", parseErr)
	}
	return row, true, nil
}

// Record inserts a confirmation record for a user. Because the user_id is the
// primary key, repeated recordings are ignored.
func (r *OnboardingConfirmationRepository) Record(ctx context.Context, userID string, delivered bool, failedReason string) error {
	if userID == "" {
		return fmt.Errorf("user id is required")
	}
	now := time.Now().UTC()
	deliveredInt := 0
	if delivered {
		deliveredInt = 1
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO onboarding_telegram_confirmations (user_id, delivered, failed_reason, sent_at)
		VALUES (?, ?, ?, ?)
		ON CONFLICT(user_id) DO NOTHING
	`, userID, deliveredInt, nullableString(failedReason), FormatTime(now)); err != nil {
		return fmt.Errorf("record onboarding confirmation: %w", err)
	}
	return nil
}
