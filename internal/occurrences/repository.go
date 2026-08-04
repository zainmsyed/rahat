package occurrences

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rahat/rahat/internal/store"
)

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, occurrence Occurrence) (Occurrence, error) {
	now := time.Now().UTC()
	if occurrence.ID == "" {
		occurrence.ID = store.NewID()
	}
	if occurrence.CreatedAt.IsZero() {
		occurrence.CreatedAt = now
	}
	if occurrence.UpdatedAt.IsZero() {
		occurrence.UpdatedAt = now
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO occurrences (id, user_id, task_id, subtask_id, status, scheduled_for_date, original_scheduled_for_date, scheduled_time_of_day, rollover_count, consecutive_no_count, snoozed_until_at, ready_at, completed_at, skipped_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
	`, occurrence.ID, occurrence.UserID, occurrence.TaskID, nullIfEmpty(occurrence.SubtaskID), occurrence.Status, occurrence.ScheduledForDate, occurrence.OriginalScheduledForDate, occurrence.ScheduledTimeOfDay, occurrence.RolloverCount, occurrence.ConsecutiveNoCount, nullableTime(occurrence.SnoozedUntilAt), nullableTime(occurrence.ReadyAt), nullableTime(occurrence.CompletedAt), nullableTime(occurrence.SkippedAt), store.FormatTime(occurrence.CreatedAt), store.FormatTime(occurrence.UpdatedAt)); err != nil {
		return Occurrence{}, fmt.Errorf("create occurrence: %w", err)
	}

	return occurrence, nil
}

func (r *Repository) GetOpenByIdentity(ctx context.Context, identity OpenOccurrenceIdentity) (Occurrence, error) {
	var id string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id
		FROM occurrences
		WHERE user_id = ?
		  AND task_id = ?
		  AND COALESCE(subtask_id, '') = ?
		  AND original_scheduled_for_date = ?
		  AND status IN (?, ?)
		ORDER BY CASE WHEN status = ? THEN 0 ELSE 1 END, updated_at DESC, id
		LIMIT 1
	`, identity.UserID, identity.TaskID, identity.SubtaskID, identity.OriginalScheduledForDate, StatusPending, StatusScheduled, StatusScheduled).Scan(&id); err != nil {
		return Occurrence{}, fmt.Errorf("get open occurrence identity: %w", err)
	}
	return r.GetByID(ctx, id)
}

func (r *Repository) GetByID(ctx context.Context, id string) (Occurrence, error) {
	var occurrence Occurrence
	var subtaskID sql.NullString
	var snoozedUntilAt sql.NullString
	var readyAt sql.NullString
	var completedAt sql.NullString
	var skippedAt sql.NullString
	var createdAt string
	var updatedAt string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, task_id, subtask_id, status, scheduled_for_date, original_scheduled_for_date, scheduled_time_of_day, rollover_count, consecutive_no_count, snoozed_until_at, ready_at, completed_at, skipped_at, created_at, updated_at
		FROM occurrences
		WHERE id = ?
	`, id).Scan(&occurrence.ID, &occurrence.UserID, &occurrence.TaskID, &subtaskID, &occurrence.Status, &occurrence.ScheduledForDate, &occurrence.OriginalScheduledForDate, &occurrence.ScheduledTimeOfDay, &occurrence.RolloverCount, &occurrence.ConsecutiveNoCount, &snoozedUntilAt, &readyAt, &completedAt, &skippedAt, &createdAt, &updatedAt); err != nil {
		return Occurrence{}, fmt.Errorf("get occurrence %s: %w", id, err)
	}
	occurrence.SubtaskID = subtaskID.String

	var err error
	occurrence.SnoozedUntilAt, err = store.ParseNullableTime(snoozedUntilAt)
	if err != nil {
		return Occurrence{}, fmt.Errorf("parse occurrence snoozed_until_at: %w", err)
	}
	occurrence.ReadyAt, err = store.ParseNullableTime(readyAt)
	if err != nil {
		return Occurrence{}, fmt.Errorf("parse occurrence ready_at: %w", err)
	}
	occurrence.CompletedAt, err = store.ParseNullableTime(completedAt)
	if err != nil {
		return Occurrence{}, fmt.Errorf("parse occurrence completed_at: %w", err)
	}
	occurrence.SkippedAt, err = store.ParseNullableTime(skippedAt)
	if err != nil {
		return Occurrence{}, fmt.Errorf("parse occurrence skipped_at: %w", err)
	}
	occurrence.CreatedAt, err = store.ParseTime(createdAt)
	if err != nil {
		return Occurrence{}, fmt.Errorf("parse occurrence created_at: %w", err)
	}
	occurrence.UpdatedAt, err = store.ParseTime(updatedAt)
	if err != nil {
		return Occurrence{}, fmt.Errorf("parse occurrence updated_at: %w", err)
	}

	return occurrence, nil
}

func (r *Repository) SaveOpen(ctx context.Context, occurrence Occurrence) (Occurrence, error) {
	existing, err := r.GetOpenByIdentity(ctx, occurrence.OpenIdentity())
	if err == nil {
		occurrence.ID = existing.ID
		occurrence.CreatedAt = existing.CreatedAt
		return r.Update(ctx, occurrence)
	}
	if !IsNotFound(err) {
		return Occurrence{}, err
	}
	created, err := r.Create(ctx, occurrence)
	if err == nil {
		return created, nil
	}

	// A concurrent planner may have inserted the same logical occurrence after
	// the lookup. Re-read the unique identity and update that row instead.
	existing, lookupErr := r.GetOpenByIdentity(ctx, occurrence.OpenIdentity())
	if lookupErr != nil {
		return Occurrence{}, err
	}
	occurrence.ID = existing.ID
	occurrence.CreatedAt = existing.CreatedAt
	return r.Update(ctx, occurrence)
}

func (r *Repository) Update(ctx context.Context, occurrence Occurrence) (Occurrence, error) {
	occurrence.UpdatedAt = time.Now().UTC()
	if _, err := r.db.ExecContext(ctx, `
		UPDATE occurrences
		SET status = ?, scheduled_for_date = ?, original_scheduled_for_date = ?, scheduled_time_of_day = ?, rollover_count = ?, consecutive_no_count = ?, snoozed_until_at = ?, ready_at = ?, completed_at = ?, skipped_at = ?, updated_at = ?
		WHERE id = ?
	`, occurrence.Status, occurrence.ScheduledForDate, occurrence.OriginalScheduledForDate, occurrence.ScheduledTimeOfDay, occurrence.RolloverCount, occurrence.ConsecutiveNoCount, nullableTime(occurrence.SnoozedUntilAt), nullableTime(occurrence.ReadyAt), nullableTime(occurrence.CompletedAt), nullableTime(occurrence.SkippedAt), store.FormatTime(occurrence.UpdatedAt), occurrence.ID); err != nil {
		return Occurrence{}, fmt.Errorf("update occurrence %s: %w", occurrence.ID, err)
	}
	return r.GetByID(ctx, occurrence.ID)
}

func (r *Repository) Delete(ctx context.Context, id string) error {
	if _, err := r.db.ExecContext(ctx, `DELETE FROM occurrences WHERE id = ?`, id); err != nil {
		return fmt.Errorf("delete occurrence %s: %w", id, err)
	}
	return nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Occurrence, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id
		FROM occurrences
		WHERE user_id = ?
		ORDER BY scheduled_for_date, id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list occurrences for user %s: %w", userID, err)
	}
	defer rows.Close()

	var occurrences []Occurrence
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan occurrence id: %w", err)
		}
		occurrence, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		occurrences = append(occurrences, occurrence)
	}

	return occurrences, rows.Err()
}

func (r *Repository) ListByTask(ctx context.Context, taskID string) ([]Occurrence, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id
		FROM occurrences
		WHERE task_id = ?
		ORDER BY scheduled_for_date, id
	`, taskID)
	if err != nil {
		return nil, fmt.Errorf("list occurrences for task %s: %w", taskID, err)
	}
	defer rows.Close()

	var occurrences []Occurrence
	for rows.Next() {
		var id string
		if err := rows.Scan(&id); err != nil {
			return nil, fmt.Errorf("scan occurrence id: %w", err)
		}
		occurrence, err := r.GetByID(ctx, id)
		if err != nil {
			return nil, err
		}
		occurrences = append(occurrences, occurrence)
	}

	return occurrences, rows.Err()
}

func IsNotFound(err error) bool {
	return err != nil && errors.Is(err, sql.ErrNoRows)
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return store.FormatTime(*value)
}
