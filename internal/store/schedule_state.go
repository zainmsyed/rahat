package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type ScheduleCheckpoint struct {
	ID                       string
	UserID                   string
	ScheduleDate             string
	NextCheckpointAt         *time.Time
	ScheduledOccurrenceCount int
	GeneratedAt              time.Time
	CreatedAt                time.Time
	UpdatedAt                time.Time
}

type ScheduleCheckpointRepository struct {
	db *sql.DB
}

func NewScheduleCheckpointRepository(db *sql.DB) *ScheduleCheckpointRepository {
	return &ScheduleCheckpointRepository{db: db}
}

func (r *ScheduleCheckpointRepository) Upsert(ctx context.Context, checkpoint ScheduleCheckpoint) (ScheduleCheckpoint, error) {
	now := time.Now().UTC()
	if checkpoint.ID == "" {
		checkpoint.ID = NewID()
	}
	if checkpoint.CreatedAt.IsZero() {
		checkpoint.CreatedAt = now
	}
	if checkpoint.GeneratedAt.IsZero() {
		checkpoint.GeneratedAt = now
	}
	checkpoint.UpdatedAt = now

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO schedule_checkpoints (id, user_id, schedule_date, next_checkpoint_at, scheduled_occurrence_count, generated_at, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, schedule_date) DO UPDATE SET
			next_checkpoint_at = excluded.next_checkpoint_at,
			scheduled_occurrence_count = excluded.scheduled_occurrence_count,
			generated_at = excluded.generated_at,
			updated_at = excluded.updated_at
	`, checkpoint.ID, checkpoint.UserID, checkpoint.ScheduleDate, nullableTime(checkpoint.NextCheckpointAt), checkpoint.ScheduledOccurrenceCount, FormatTime(checkpoint.GeneratedAt), FormatTime(checkpoint.CreatedAt), FormatTime(checkpoint.UpdatedAt)); err != nil {
		return ScheduleCheckpoint{}, fmt.Errorf("upsert schedule checkpoint: %w", err)
	}

	return r.GetByDate(ctx, checkpoint.UserID, checkpoint.ScheduleDate)
}

func (r *ScheduleCheckpointRepository) GetByDate(ctx context.Context, userID, scheduleDate string) (ScheduleCheckpoint, error) {
	var checkpoint ScheduleCheckpoint
	var nextCheckpointAt sql.NullString
	var generatedAt string
	var createdAt string
	var updatedAt string

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, schedule_date, next_checkpoint_at, scheduled_occurrence_count, generated_at, created_at, updated_at
		FROM schedule_checkpoints
		WHERE user_id = ? AND schedule_date = ?
	`, userID, scheduleDate).Scan(&checkpoint.ID, &checkpoint.UserID, &checkpoint.ScheduleDate, &nextCheckpointAt, &checkpoint.ScheduledOccurrenceCount, &generatedAt, &createdAt, &updatedAt); err != nil {
		return ScheduleCheckpoint{}, fmt.Errorf("get schedule checkpoint %s/%s: %w", userID, scheduleDate, err)
	}

	var err error
	checkpoint.NextCheckpointAt, err = ParseNullableTime(nextCheckpointAt)
	if err != nil {
		return ScheduleCheckpoint{}, fmt.Errorf("parse schedule checkpoint next_checkpoint_at: %w", err)
	}
	checkpoint.GeneratedAt, err = ParseTime(generatedAt)
	if err != nil {
		return ScheduleCheckpoint{}, fmt.Errorf("parse schedule checkpoint generated_at: %w", err)
	}
	checkpoint.CreatedAt, err = ParseTime(createdAt)
	if err != nil {
		return ScheduleCheckpoint{}, fmt.Errorf("parse schedule checkpoint created_at: %w", err)
	}
	checkpoint.UpdatedAt, err = ParseTime(updatedAt)
	if err != nil {
		return ScheduleCheckpoint{}, fmt.Errorf("parse schedule checkpoint updated_at: %w", err)
	}
	return checkpoint, nil
}

func nullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return FormatTime(*value)
}
