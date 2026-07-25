package events

import (
	"context"
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/rahat/rahat/internal/store"
)

type ReportFilter struct {
	From *time.Time
	To   *time.Time
}

type SummaryRow struct {
	Channel     string
	EventType   string
	MessageType string
	Count       int
}

func (r *Repository) Summary(ctx context.Context, filter ReportFilter) ([]SummaryRow, error) {
	query := `
		SELECT channel, event_type, message_type, COUNT(*)
		FROM event_logs
	`
	where, args := buildReportWhere(filter)
	if where != "" {
		query += " WHERE " + where
	}
	query += " GROUP BY channel, event_type, message_type ORDER BY channel, event_type, message_type"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("summarize event logs: %w", err)
	}
	defer rows.Close()
	result := []SummaryRow{}
	for rows.Next() {
		var row SummaryRow
		if err := rows.Scan(&row.Channel, &row.EventType, &row.MessageType, &row.Count); err != nil {
			return nil, fmt.Errorf("scan event summary: %w", err)
		}
		result = append(result, row)
	}
	return result, rows.Err()
}

func (r *Repository) ListFiltered(ctx context.Context, filter ReportFilter) ([]EventLog, error) {
	query := `
		SELECT id, user_id, occurrence_id, channel, event_type, message_type, payload_json, occurred_at
		FROM event_logs
	`
	where, args := buildReportWhere(filter)
	if where != "" {
		query += " WHERE " + where
	}
	query += " ORDER BY occurred_at DESC, id DESC"
	rows, err := r.db.QueryContext(ctx, query, args...)
	if err != nil {
		return nil, fmt.Errorf("list filtered event logs: %w", err)
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

func buildReportWhere(filter ReportFilter) (string, []any) {
	clauses := []string{}
	args := []any{}
	if filter.From != nil {
		clauses = append(clauses, "occurred_at >= ?")
		args = append(args, store.FormatTime(filter.From.UTC()))
	}
	if filter.To != nil {
		clauses = append(clauses, "occurred_at <= ?")
		args = append(args, store.FormatTime(filter.To.UTC()))
	}
	return strings.Join(clauses, " AND "), args
}
