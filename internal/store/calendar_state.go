package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type CalendarConnection struct {
	ID           string
	UserID       string
	Provider     string
	CalendarID   string
	AccessToken  string
	RefreshToken string
	TokenType    string
	Expiry       *time.Time
	Scope        string
	Timezone     string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

type CalendarBlock struct {
	ID              string
	UserID          string
	Provider        string
	ExternalEventID string
	LocalDate       string
	Timezone        string
	Title           string
	Detail          string
	StartAt         *time.Time
	EndAt           *time.Time
	IsAllDay        bool
	Classification  string
	Window          string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

type CalendarConnectionRepository struct{ db *sql.DB }

type CalendarBlockRepository struct{ db *sql.DB }

func NewCalendarConnectionRepository(db *sql.DB) *CalendarConnectionRepository {
	return &CalendarConnectionRepository{db: db}
}

func NewCalendarBlockRepository(db *sql.DB) *CalendarBlockRepository {
	return &CalendarBlockRepository{db: db}
}

func (r *CalendarConnectionRepository) Upsert(ctx context.Context, conn CalendarConnection) (CalendarConnection, error) {
	now := time.Now().UTC()
	if conn.ID == "" {
		conn.ID = NewID()
	}
	if conn.CalendarID == "" {
		conn.CalendarID = "primary"
	}
	if conn.TokenType == "" {
		conn.TokenType = "Bearer"
	}
	if conn.Timezone == "" {
		conn.Timezone = "UTC"
	}
	if conn.CreatedAt.IsZero() {
		conn.CreatedAt = now
	}
	conn.UpdatedAt = now

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO calendar_connections (id, user_id, provider, calendar_id, access_token, refresh_token, token_type, expiry, scope, timezone, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		ON CONFLICT(user_id, provider) DO UPDATE SET
			calendar_id = excluded.calendar_id,
			access_token = excluded.access_token,
			refresh_token = excluded.refresh_token,
			token_type = excluded.token_type,
			expiry = excluded.expiry,
			scope = excluded.scope,
			timezone = excluded.timezone,
			updated_at = excluded.updated_at
	`, conn.ID, conn.UserID, conn.Provider, conn.CalendarID, conn.AccessToken, nullableString(conn.RefreshToken), conn.TokenType, nullableTime(conn.Expiry), conn.Scope, conn.Timezone, FormatTime(conn.CreatedAt), FormatTime(conn.UpdatedAt)); err != nil {
		return CalendarConnection{}, fmt.Errorf("upsert calendar connection: %w", err)
	}
	return r.GetByUserAndProvider(ctx, conn.UserID, conn.Provider)
}

func (r *CalendarConnectionRepository) GetByUserAndProvider(ctx context.Context, userID, provider string) (CalendarConnection, error) {
	var conn CalendarConnection
	var refreshToken sql.NullString
	var expiry sql.NullString
	var createdAt, updatedAt string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, provider, calendar_id, access_token, refresh_token, token_type, expiry, scope, timezone, created_at, updated_at
		FROM calendar_connections WHERE user_id = ? AND provider = ?
	`, userID, provider).Scan(&conn.ID, &conn.UserID, &conn.Provider, &conn.CalendarID, &conn.AccessToken, &refreshToken, &conn.TokenType, &expiry, &conn.Scope, &conn.Timezone, &createdAt, &updatedAt); err != nil {
		return CalendarConnection{}, fmt.Errorf("get calendar connection %s/%s: %w", userID, provider, err)
	}
	conn.RefreshToken = refreshToken.String
	var err error
	conn.Expiry, err = ParseNullableTime(expiry)
	if err != nil {
		return CalendarConnection{}, fmt.Errorf("parse calendar connection expiry: %w", err)
	}
	conn.CreatedAt, err = ParseTime(createdAt)
	if err != nil {
		return CalendarConnection{}, fmt.Errorf("parse calendar connection created_at: %w", err)
	}
	conn.UpdatedAt, err = ParseTime(updatedAt)
	if err != nil {
		return CalendarConnection{}, fmt.Errorf("parse calendar connection updated_at: %w", err)
	}
	return conn, nil
}

func (r *CalendarBlockRepository) ReplaceDay(ctx context.Context, userID, provider, localDate string, blocks []CalendarBlock) error {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin replace calendar blocks: %w", err)
	}
	if _, err := tx.ExecContext(ctx, `DELETE FROM calendar_blocks WHERE user_id = ? AND provider = ? AND local_date = ?`, userID, provider, localDate); err != nil {
		_ = tx.Rollback()
		return fmt.Errorf("delete calendar blocks: %w", err)
	}
	for _, block := range blocks {
		now := time.Now().UTC()
		if block.ID == "" {
			block.ID = NewID()
		}
		if block.CreatedAt.IsZero() {
			block.CreatedAt = now
		}
		block.UpdatedAt = now
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO calendar_blocks (id, user_id, provider, external_event_id, local_date, timezone, title, detail, start_at, end_at, is_all_day, classification, window, created_at, updated_at)
			VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?, ?)
		`, block.ID, block.UserID, block.Provider, block.ExternalEventID, block.LocalDate, block.Timezone, block.Title, block.Detail, nullableTime(block.StartAt), nullableTime(block.EndAt), boolInt(block.IsAllDay), block.Classification, block.Window, FormatTime(block.CreatedAt), FormatTime(block.UpdatedAt)); err != nil {
			_ = tx.Rollback()
			return fmt.Errorf("insert calendar block: %w", err)
		}
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit replace calendar blocks: %w", err)
	}
	return nil
}

func (r *CalendarBlockRepository) ListByUserAndDate(ctx context.Context, userID, localDate string) ([]CalendarBlock, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, user_id, provider, external_event_id, local_date, timezone, title, detail, start_at, end_at, is_all_day, classification, window, created_at, updated_at
		FROM calendar_blocks WHERE user_id = ? AND local_date = ? ORDER BY start_at, title
	`, userID, localDate)
	if err != nil {
		return nil, fmt.Errorf("list calendar blocks: %w", err)
	}
	defer rows.Close()
	var blocks []CalendarBlock
	for rows.Next() {
		var block CalendarBlock
		var startAt, endAt sql.NullString
		var isAllDay int
		var createdAt, updatedAt string
		if err := rows.Scan(&block.ID, &block.UserID, &block.Provider, &block.ExternalEventID, &block.LocalDate, &block.Timezone, &block.Title, &block.Detail, &startAt, &endAt, &isAllDay, &block.Classification, &block.Window, &createdAt, &updatedAt); err != nil {
			return nil, fmt.Errorf("scan calendar block: %w", err)
		}
		block.IsAllDay = isAllDay == 1
		var err error
		block.StartAt, err = ParseNullableTime(startAt)
		if err != nil {
			return nil, fmt.Errorf("parse calendar block start_at: %w", err)
		}
		block.EndAt, err = ParseNullableTime(endAt)
		if err != nil {
			return nil, fmt.Errorf("parse calendar block end_at: %w", err)
		}
		block.CreatedAt, err = ParseTime(createdAt)
		if err != nil {
			return nil, fmt.Errorf("parse calendar block created_at: %w", err)
		}
		block.UpdatedAt, err = ParseTime(updatedAt)
		if err != nil {
			return nil, fmt.Errorf("parse calendar block updated_at: %w", err)
		}
		blocks = append(blocks, block)
	}
	return blocks, rows.Err()
}

func nullableString(value string) any {
	if value == "" {
		return nil
	}
	return value
}

func boolInt(value bool) int {
	if value {
		return 1
	}
	return 0
}
