package auth

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"time"

	"github.com/rahat/rahat/internal/store"
)

var ErrNotFound = errors.New("auth record not found")

type Repository struct {
	db *sql.DB
}

func NewRepository(db *sql.DB) *Repository {
	return &Repository{db: db}
}

func (r *Repository) CreateAccessGrant(ctx context.Context, tx *sql.Tx, grant AccessGrant) (AccessGrant, error) {
	exec := txExec(r.db, tx)
	if grant.ID == "" {
		grant.ID = store.NewID()
	}
	if grant.CreatedAt.IsZero() {
		grant.CreatedAt = time.Now().UTC()
	}
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO beta_access_grants (id, user_id, selector, token_hash, expires_at, used_at, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?)
	`, grant.ID, grant.UserID, grant.Selector, grant.TokenHash, store.FormatTime(grant.ExpiresAt), formatNullableTime(grant.UsedAt), store.FormatTime(grant.CreatedAt)); err != nil {
		return AccessGrant{}, fmt.Errorf("create access grant: %w", err)
	}
	return grant, nil
}

func (r *Repository) GetAccessGrantBySelector(ctx context.Context, tx *sql.Tx, selector string) (AccessGrant, error) {
	exec := txQuery(r.db, tx)
	var grant AccessGrant
	var usedAt sql.NullString
	var expiresAt string
	var createdAt string
	if err := exec.QueryRowContext(ctx, `
			SELECT id, user_id, selector, token_hash, expires_at, used_at, created_at
			FROM beta_access_grants
			WHERE selector = ?
		`, selector).Scan(&grant.ID, &grant.UserID, &grant.Selector, &grant.TokenHash, &expiresAt, &usedAt, &createdAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return AccessGrant{}, ErrNotFound
		}
		return AccessGrant{}, fmt.Errorf("get access grant by selector: %w", err)
	}
	var err error
	grant.ExpiresAt, err = store.ParseTime(expiresAt)
	if err != nil {
		return AccessGrant{}, fmt.Errorf("parse access grant expires_at: %w", err)
	}
	grant.UsedAt, err = store.ParseNullableTime(usedAt)
	if err != nil {
		return AccessGrant{}, fmt.Errorf("parse access grant used_at: %w", err)
	}
	grant.CreatedAt, err = store.ParseTime(createdAt)
	if err != nil {
		return AccessGrant{}, fmt.Errorf("parse access grant created_at: %w", err)
	}
	return grant, nil
}

func (r *Repository) MarkAccessGrantUsed(ctx context.Context, tx *sql.Tx, id string, usedAt time.Time) error {
	exec := txExec(r.db, tx)
	if _, err := exec.ExecContext(ctx, `UPDATE beta_access_grants SET used_at = ? WHERE id = ?`, store.FormatTime(usedAt), id); err != nil {
		return fmt.Errorf("mark access grant used: %w", err)
	}
	return nil
}

func (r *Repository) CreateWebSession(ctx context.Context, tx *sql.Tx, session WebSession) (WebSession, error) {
	exec := txExec(r.db, tx)
	if session.ID == "" {
		session.ID = store.NewID()
	}
	now := time.Now().UTC()
	if session.CreatedAt.IsZero() {
		session.CreatedAt = now
	}
	if session.LastSeenAt.IsZero() {
		session.LastSeenAt = now
	}
	if _, err := exec.ExecContext(ctx, `
		INSERT INTO web_sessions (id, user_id, selector, token_hash, expires_at, revoked_at, created_at, last_seen_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?)
	`, session.ID, session.UserID, session.Selector, session.TokenHash, store.FormatTime(session.ExpiresAt), formatNullableTime(session.RevokedAt), store.FormatTime(session.CreatedAt), store.FormatTime(session.LastSeenAt)); err != nil {
		return WebSession{}, fmt.Errorf("create web session: %w", err)
	}
	return session, nil
}

func (r *Repository) GetWebSessionBySelector(ctx context.Context, selector string) (WebSession, error) {
	var session WebSession
	var revokedAt sql.NullString
	var expiresAt string
	var createdAt string
	var lastSeenAt string
	if err := r.db.QueryRowContext(ctx, `
		SELECT id, user_id, selector, token_hash, expires_at, revoked_at, created_at, last_seen_at
		FROM web_sessions
		WHERE selector = ?
	`, selector).Scan(&session.ID, &session.UserID, &session.Selector, &session.TokenHash, &expiresAt, &revokedAt, &createdAt, &lastSeenAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return WebSession{}, ErrNotFound
		}
		return WebSession{}, fmt.Errorf("get web session by selector: %w", err)
	}
	var err error
	session.ExpiresAt, err = store.ParseTime(expiresAt)
	if err != nil {
		return WebSession{}, fmt.Errorf("parse web session expires_at: %w", err)
	}
	session.RevokedAt, err = store.ParseNullableTime(revokedAt)
	if err != nil {
		return WebSession{}, fmt.Errorf("parse web session revoked_at: %w", err)
	}
	session.CreatedAt, err = store.ParseTime(createdAt)
	if err != nil {
		return WebSession{}, fmt.Errorf("parse web session created_at: %w", err)
	}
	session.LastSeenAt, err = store.ParseTime(lastSeenAt)
	if err != nil {
		return WebSession{}, fmt.Errorf("parse web session last_seen_at: %w", err)
	}
	return session, nil
}

func (r *Repository) RevokeWebSession(ctx context.Context, id string, revokedAt time.Time) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE web_sessions SET revoked_at = ? WHERE id = ?`, store.FormatTime(revokedAt), id); err != nil {
		return fmt.Errorf("revoke web session: %w", err)
	}
	return nil
}

func (r *Repository) TouchWebSession(ctx context.Context, id string, lastSeenAt time.Time) error {
	if _, err := r.db.ExecContext(ctx, `UPDATE web_sessions SET last_seen_at = ? WHERE id = ?`, store.FormatTime(lastSeenAt), id); err != nil {
		return fmt.Errorf("touch web session: %w", err)
	}
	return nil
}

type execer interface {
	ExecContext(context.Context, string, ...any) (sql.Result, error)
}

type queryer interface {
	QueryRowContext(context.Context, string, ...any) *sql.Row
}

func txExec(db *sql.DB, tx *sql.Tx) execer {
	if tx != nil {
		return tx
	}
	return db
}

func txQuery(db *sql.DB, tx *sql.Tx) queryer {
	if tx != nil {
		return tx
	}
	return db
}

func formatNullableTime(value *time.Time) any {
	if value == nil {
		return nil
	}
	return store.FormatTime(value.UTC())
}
