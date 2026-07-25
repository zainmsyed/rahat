package users

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rahat/rahat/internal/store"
)

var ErrNotFound = errors.New("user not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) Create(ctx context.Context, user User) (User, error) {
	now := time.Now().UTC()
	if user.ID == "" {
		user.ID = store.NewID()
	}
	if user.CreatedAt.IsZero() {
		user.CreatedAt = now
	}
	if user.UpdatedAt.IsZero() {
		user.UpdatedAt = now
	}

	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO users (id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, user.ID, user.DisplayName, user.Timezone, user.DailyTimeBudgetMinutes, nullIfEmpty(user.TelegramChatID), nullIfEmpty(user.Email), store.FormatTime(user.CreatedAt), store.FormatTime(user.UpdatedAt)); err != nil {
		return User{}, fmt.Errorf("create user: %w", err)
	}

	return user, nil
}

func (r *Repository) GetByID(ctx context.Context, id string) (User, error) {
	return r.getByID(ctx, r.db, id)
}

func (r *Repository) getByID(ctx context.Context, q queryer, id string) (User, error) {
	user, err := scanUser(q.QueryRowContext(ctx, `
		SELECT id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user %s: %w", id, err)
	}
	return user, nil
}

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	user, err := scanUser(r.db.QueryRowContext(ctx, `
		SELECT id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at
		FROM users
		WHERE email = ?
		ORDER BY created_at, id
		LIMIT 1
	`, email))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by email %s: %w", email, err)
	}
	return user, nil
}

func (r *Repository) GetByTelegramChatID(ctx context.Context, chatID string) (User, error) {
	return r.getByTelegramChatID(ctx, r.db, chatID)
}

func (r *Repository) getByTelegramChatID(ctx context.Context, q queryer, chatID string) (User, error) {
	user, err := scanUser(q.QueryRowContext(ctx, `
		SELECT id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at
		FROM users
		WHERE telegram_chat_id = ?
		LIMIT 1
	`, chatID))
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return User{}, ErrNotFound
		}
		return User{}, fmt.Errorf("get user by telegram chat id: %w", err)
	}
	return user, nil
}

func (r *Repository) Update(ctx context.Context, user User) (User, error) {
	if _, err := r.update(ctx, r.db, user); err != nil {
		return User{}, err
	}
	return r.GetByID(ctx, user.ID)
}

func (r *Repository) update(ctx context.Context, e execer, user User) (sql.Result, error) {
	user.UpdatedAt = time.Now().UTC()
	return e.ExecContext(ctx, `
		UPDATE users
		SET display_name = ?, timezone = ?, daily_time_budget_minutes = ?, telegram_chat_id = ?, email = ?, updated_at = ?
		WHERE id = ?
	`, user.DisplayName, user.Timezone, user.DailyTimeBudgetMinutes, nullIfEmpty(user.TelegramChatID), nullIfEmpty(user.Email), store.FormatTime(user.UpdatedAt), user.ID)
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func scanUser(row *sql.Row) (User, error) {
	var user User
	var telegramChatID sql.NullString
	var nullableEmail sql.NullString
	var createdAt string
	var updatedAt string
	if err := row.Scan(&user.ID, &user.DisplayName, &user.Timezone, &user.DailyTimeBudgetMinutes, &telegramChatID, &nullableEmail, &createdAt, &updatedAt); err != nil {
		return User{}, err
	}
	user.TelegramChatID = telegramChatID.String
	user.Email = nullableEmail.String
	var err error
	user.CreatedAt, err = store.ParseTime(createdAt)
	if err != nil {
		return User{}, fmt.Errorf("parse user created_at: %w", err)
	}
	user.UpdatedAt, err = store.ParseTime(updatedAt)
	if err != nil {
		return User{}, fmt.Errorf("parse user updated_at: %w", err)
	}
	return user, nil
}
