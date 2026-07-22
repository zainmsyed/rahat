package events

import (
	"context"
	"database/sql"
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

func (r *Repository) Create(ctx context.Context, event EventLog) (EventLog, error) {
	if event.ID == "" {
		event.ID = store.NewID()
	}
	if event.OccurredAt.IsZero() {
		event.OccurredAt = time.Now().UTC()
	}
	if event.PayloadJSON == "" {
		event.PayloadJSON = "{}"
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO event_logs (id, user_id, occurrence_id, channel, event_type, message_type, payload_json, occurred_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, event.ID, event.UserID, nullIfEmpty(event.OccurrenceID), event.Channel, event.EventType, event.MessageType, event.PayloadJSON, store.FormatTime(event.OccurredAt)); err != nil {
		return EventLog{}, fmt.Errorf("create event log: %w", err)
	}

	return event, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]EventLog, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, occurrence_id, channel, event_type, message_type, payload_json, occurred_at
		FROM event_logs
		WHERE user_id = ?
		ORDER BY occurred_at DESC, id DESC
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list event logs for user %s: %w", userID, err)
	}
	defer rows.Close()

	var events []EventLog
	for rows.Next() {
		var event EventLog
		var occurrenceID sql.NullString
		var occurredAt string
		if err := rows.Scan(&event.ID, &event.UserID, &occurrenceID, &event.Channel, &event.EventType, &event.MessageType, &event.PayloadJSON, &occurredAt); err != nil {
			return nil, fmt.Errorf("scan event log row: %w", err)
		}
		event.OccurrenceID = occurrenceID.String
		if event.OccurredAt, err = store.ParseTime(occurredAt); err != nil {
			return nil, fmt.Errorf("parse event occurred_at: %w", err)
		}
		events = append(events, event)
	}

	return events, rows.Err()
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
