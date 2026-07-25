package users

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
	var user User
	var telegramChatID sql.NullString
	var email sql.NullString
	var createdAt string
	var updatedAt string

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at
		FROM users
		WHERE id = ?
	`, id).Scan(&user.ID, &user.DisplayName, &user.Timezone, &user.DailyTimeBudgetMinutes, &telegramChatID, &email, &createdAt, &updatedAt); err != nil {
		return User{}, fmt.Errorf("get user %s: %w", id, err)
	}

	user.TelegramChatID = telegramChatID.String
	user.Email = email.String
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

func (r *Repository) GetByEmail(ctx context.Context, email string) (User, error) {
	var user User
	var telegramChatID sql.NullString
	var nullableEmail sql.NullString
	var createdAt string
	var updatedAt string

	if err := r.db.QueryRowContext(ctx, `
		SELECT id, display_name, timezone, daily_time_budget_minutes, telegram_chat_id, email, created_at, updated_at
		FROM users
		WHERE email = ?
		ORDER BY created_at, id
		LIMIT 1
	`, email).Scan(&user.ID, &user.DisplayName, &user.Timezone, &user.DailyTimeBudgetMinutes, &telegramChatID, &nullableEmail, &createdAt, &updatedAt); err != nil {
		return User{}, fmt.Errorf("get user by email %s: %w", email, err)
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

func (r *Repository) Update(ctx context.Context, user User) (User, error) {
	user.UpdatedAt = time.Now().UTC()
	if _, err := r.db.ExecContext(ctx, `
		UPDATE users
		SET display_name = ?, timezone = ?, daily_time_budget_minutes = ?, telegram_chat_id = ?, email = ?, updated_at = ?
		WHERE id = ?
	`, user.DisplayName, user.Timezone, user.DailyTimeBudgetMinutes, nullIfEmpty(user.TelegramChatID), nullIfEmpty(user.Email), store.FormatTime(user.UpdatedAt), user.ID); err != nil {
		return User{}, fmt.Errorf("update user %s: %w", user.ID, err)
	}

	return r.GetByID(ctx, user.ID)
}

func nullIfEmpty(value string) any {
	if value == "" {
		return nil
	}
	return value
}
