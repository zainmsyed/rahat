package store

import (
	"context"
	"database/sql"
	"fmt"
	"time"
)

type OAuthState struct {
	ID          string
	UserID      string
	Provider    string
	StateToken  string
	RedirectURI string
	ExpiresAt   time.Time
	CreatedAt   time.Time
	ConsumedAt  *time.Time
}

type OAuthStateRepository struct{ db *sql.DB }

func NewOAuthStateRepository(db *sql.DB) *OAuthStateRepository { return &OAuthStateRepository{db: db} }

func (r *OAuthStateRepository) Create(ctx context.Context, state OAuthState) (OAuthState, error) {
	now := time.Now().UTC()
	if state.ID == "" {
		state.ID = NewID()
	}
	if state.StateToken == "" {
		state.StateToken = NewID()
	}
	if state.CreatedAt.IsZero() {
		state.CreatedAt = now
	}
	if state.ExpiresAt.IsZero() {
		state.ExpiresAt = now.Add(15 * time.Minute)
	}
	if _, err := r.db.ExecContext(ctx, `
		INSERT INTO oauth_states (id, user_id, provider, state_token, redirect_uri, expires_at, created_at, consumed_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, state.ID, state.UserID, state.Provider, state.StateToken, state.RedirectURI, FormatTime(state.ExpiresAt), FormatTime(state.CreatedAt), nullableTime(state.ConsumedAt)); err != nil {
		return OAuthState{}, fmt.Errorf("create oauth state: %w", err)
	}
	return state, nil
}

func (r *OAuthStateRepository) GetPendingByUserAndProvider(ctx context.Context, userID, provider string, now time.Time) (OAuthState, error) {
	var state OAuthState
	var expiresAt, createdAt string
	var consumedAt sql.NullString
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, provider, state_token, redirect_uri, expires_at, created_at, consumed_at
		FROM oauth_states
		WHERE user_id = ? AND provider = ? AND consumed_at IS NULL AND expires_at > ?
		ORDER BY created_at DESC
		LIMIT 1
	`, userID, provider, FormatTime(now)).Scan(&state.ID, &state.UserID, &state.Provider, &state.StateToken, &state.RedirectURI, &expiresAt, &createdAt, &consumedAt); err != nil {
		return OAuthState{}, fmt.Errorf("get pending oauth state %s/%s: %w", userID, provider, err)
	}
	var err error
	state.ExpiresAt, err = ParseTime(expiresAt)
	if err != nil {
		return OAuthState{}, fmt.Errorf("parse oauth state expires_at: %w", err)
	}
	state.CreatedAt, err = ParseTime(createdAt)
	if err != nil {
		return OAuthState{}, fmt.Errorf("parse oauth state created_at: %w", err)
	}
	state.ConsumedAt, err = ParseNullableTime(consumedAt)
	if err != nil {
		return OAuthState{}, fmt.Errorf("parse oauth state consumed_at: %w", err)
	}
	return state, nil
}

func (r *OAuthStateRepository) Consume(ctx context.Context, provider, token string, now time.Time) (OAuthState, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return OAuthState{}, fmt.Errorf("begin consume oauth state: %w", err)
	}
	defer func() {
		if tx != nil {
			_ = tx.Rollback()
		}
	}()

	var state OAuthState
	var expiresAt, createdAt string
	var consumedAt sql.NullString
	if err := tx.QueryRowContext(ctx, `
		SELECT id, user_id, provider, state_token, redirect_uri, expires_at, created_at, consumed_at
		FROM oauth_states WHERE provider = ? AND state_token = ?
	`, provider, token).Scan(&state.ID, &state.UserID, &state.Provider, &state.StateToken, &state.RedirectURI, &expiresAt, &createdAt, &consumedAt); err != nil {
		return OAuthState{}, fmt.Errorf("get oauth state %s/%s: %w", provider, token, err)
	}
	state.ExpiresAt, err = ParseTime(expiresAt)
	if err != nil {
		return OAuthState{}, fmt.Errorf("parse oauth state expires_at: %w", err)
	}
	state.CreatedAt, err = ParseTime(createdAt)
	if err != nil {
		return OAuthState{}, fmt.Errorf("parse oauth state created_at: %w", err)
	}
	state.ConsumedAt, err = ParseNullableTime(consumedAt)
	if err != nil {
		return OAuthState{}, fmt.Errorf("parse oauth state consumed_at: %w", err)
	}
	if state.ConsumedAt != nil {
		return OAuthState{}, fmt.Errorf("oauth state already consumed")
	}
	if now.UTC().After(state.ExpiresAt) {
		return OAuthState{}, fmt.Errorf("oauth state expired")
	}
	consumed := now.UTC()
	if _, err := tx.ExecContext(ctx, `UPDATE oauth_states SET consumed_at = ? WHERE id = ?`, FormatTime(consumed), state.ID); err != nil {
		return OAuthState{}, fmt.Errorf("mark oauth state consumed: %w", err)
	}
	state.ConsumedAt = &consumed
	if err := tx.Commit(); err != nil {
		return OAuthState{}, fmt.Errorf("commit consume oauth state: %w", err)
	}
	tx = nil
	return state, nil
}
