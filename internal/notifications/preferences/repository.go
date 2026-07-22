package preferences

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

func (r *Repository) Upsert(ctx context.Context, pref Preference) (Preference, error) {
	now := time.Now().UTC()
	if pref.ID == "" {
		pref.ID = store.NewID()
	}
	if pref.CreatedAt.IsZero() {
		pref.CreatedAt = now
	}
	pref.UpdatedAt = now

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO channel_preferences (id, user_id, channel, enabled, is_primary, supports_interactive, recap_enabled, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, channel) DO UPDATE SET
			enabled = excluded.enabled,
			is_primary = excluded.is_primary,
			supports_interactive = excluded.supports_interactive,
			recap_enabled = excluded.recap_enabled,
			updated_at = excluded.updated_at
	`, pref.ID, pref.UserID, pref.Channel, pref.Enabled, pref.IsPrimary, pref.SupportsInteractive, pref.RecapEnabled, store.FormatTime(pref.CreatedAt), store.FormatTime(pref.UpdatedAt)); err != nil {
		return Preference{}, fmt.Errorf("upsert channel preference: %w", err)
	}

	return r.GetByChannel(ctx, pref.UserID, pref.Channel)
}

func (r *Repository) GetByChannel(ctx context.Context, userID string, channel Channel) (Preference, error) {
	var pref Preference
	var createdAt string
	var updatedAt string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, channel, enabled, is_primary, supports_interactive, recap_enabled, created_at, updated_at
		FROM channel_preferences
		WHERE user_id = ? AND channel = ?
	`, userID, channel).Scan(&pref.ID, &pref.UserID, &pref.Channel, &pref.Enabled, &pref.IsPrimary, &pref.SupportsInteractive, &pref.RecapEnabled, &createdAt, &updatedAt); err != nil {
		return Preference{}, fmt.Errorf("get channel preference %s/%s: %w", userID, channel, err)
	}
	var err error
	pref.CreatedAt, err = store.ParseTime(createdAt)
	if err != nil {
		return Preference{}, fmt.Errorf("parse preference created_at: %w", err)
	}
	pref.UpdatedAt, err = store.ParseTime(updatedAt)
	if err != nil {
		return Preference{}, fmt.Errorf("parse preference updated_at: %w", err)
	}
	return pref, nil
}

func (r *Repository) ListByUser(ctx context.Context, userID string) ([]Preference, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, channel, enabled, is_primary, supports_interactive, recap_enabled, created_at, updated_at
		FROM channel_preferences
		WHERE user_id = ?
		ORDER BY channel
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list channel preferences for user %s: %w", userID, err)
	}
	defer rows.Close()

	var prefs []Preference
	for rows.Next() {
		var pref Preference
		var createdAt string
		var updatedAt string
		if err := rows.Scan(&pref.ID, &pref.UserID, &pref.Channel, &pref.Enabled, &pref.IsPrimary, &pref.SupportsInteractive, &pref.RecapEnabled, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan preference row: %w", err)
		}
		if pref.CreatedAt, err = store.ParseTime(createdAt); err != nil {
			return nil, fmt.Errorf("parse preference created_at: %w", err)
		}
		if pref.UpdatedAt, err = store.ParseTime(updatedAt); err != nil {
			return nil, fmt.Errorf("parse preference updated_at: %w", err)
		}
		prefs = append(prefs, pref)
	}

	return prefs, rows.Err()
}

func (r *Repository) CreatePause(ctx context.Context, pause Pause) (Pause, error) {
	if pause.ID == "" {
		pause.ID = store.NewID()
	}
	if pause.CreatedAt.IsZero() {
		pause.CreatedAt = time.Now().UTC()
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO pauses (id, user_id, task_id, scope, reason, starts_at, ends_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, pause.ID, pause.UserID, nullIfEmpty(pause.TaskID), pause.Scope, pause.Reason, store.FormatTime(pause.StartsAt), store.FormatTime(pause.EndsAt), store.FormatTime(pause.CreatedAt)); err != nil {
		return Pause{}, fmt.Errorf("create pause: %w", err)
	}

	return pause, nil
}

func (r *Repository) ListPausesByUser(ctx context.Context, userID string) ([]Pause, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, task_id, scope, reason, starts_at, ends_at, created_at
		FROM pauses
		WHERE user_id = ?
		ORDER BY starts_at, id
	`, userID)
	if err != nil {
		return nil, fmt.Errorf("list pauses for user %s: %w", userID, err)
	}
	defer rows.Close()

	var pauses []Pause
	for rows.Next() {
		var pause Pause
		var taskID sql.NullString
		var startsAt string
		var endsAt string
		var createdAt string
		if err := rows.Scan(&pause.ID, &pause.UserID, &taskID, &pause.Scope, &pause.Reason, &startsAt, &endsAt, &createdAt); err != nil {
			return nil, fmt.Errorf("scan pause row: %w", err)
		}
		pause.TaskID = taskID.String
		if pause.StartsAt, err = store.ParseTime(startsAt); err != nil {
			return nil, fmt.Errorf("parse pause starts_at: %w", err)
		}
		if pause.EndsAt, err = store.ParseTime(endsAt); err != nil {
			return nil, fmt.Errorf("parse pause ends_at: %w", err)
		}
		if pause.CreatedAt, err = store.ParseTime(createdAt); err != nil {
			return nil, fmt.Errorf("parse pause created_at: %w", err)
		}
		pauses = append(pauses, pause)
	}

	return pauses, rows.Err()
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
