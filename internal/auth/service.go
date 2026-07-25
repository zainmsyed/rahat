package auth

import (
	"context"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrUnavailable  = errors.New("web sessions are not configured")
	ErrInvalidToken = errors.New("access token is invalid")
	ErrExpired      = errors.New("access token has expired")
	ErrUsed         = errors.New("access token has already been used")
	ErrRevoked      = errors.New("session has been revoked")
)

type Service struct {
	repo       *Repository
	db         *sql.DB
	secret     []byte
	sessionTTL time.Duration
	now        func() time.Time
}

func NewService(db *sql.DB, repo *Repository, secret string, sessionTTL time.Duration) *Service {
	return &Service{db: db, repo: repo, secret: []byte(strings.TrimSpace(secret)), sessionTTL: sessionTTL, now: func() time.Time { return time.Now().UTC() }}
}

func (s *Service) Available() bool {
	return len(s.secret) > 0
}

func (s *Service) IssueAccessGrant(ctx context.Context, userID string, ttl time.Duration) (AccessGrant, string, error) {
	if !s.Available() {
		return AccessGrant{}, "", ErrUnavailable
	}
	selector, secret, err := randomOpaqueParts()
	if err != nil {
		return AccessGrant{}, "", err
	}
	grant := AccessGrant{UserID: userID, Selector: selector, TokenHash: s.hashSecret(selector, secret), ExpiresAt: s.now().Add(ttl)}
	created, err := s.repo.CreateAccessGrant(ctx, nil, grant)
	if err != nil {
		return AccessGrant{}, "", err
	}
	return created, selector + "." + secret, nil
}

func (s *Service) CreateSessionForUser(ctx context.Context, userID string) (WebSession, string, error) {
	if !s.Available() {
		return WebSession{}, "", ErrUnavailable
	}
	selector, secret, err := randomOpaqueParts()
	if err != nil {
		return WebSession{}, "", err
	}
	session := WebSession{UserID: userID, Selector: selector, TokenHash: s.hashSecret(selector, secret), ExpiresAt: s.now().Add(s.sessionTTL)}
	created, err := s.repo.CreateWebSession(ctx, nil, session)
	if err != nil {
		return WebSession{}, "", err
	}
	return created, selector + "." + secret, nil
}

func (s *Service) ExchangeAccessGrant(ctx context.Context, rawToken string) (WebSession, string, error) {
	if !s.Available() {
		return WebSession{}, "", ErrUnavailable
	}
	selector, secret, err := parseOpaqueToken(rawToken)
	if err != nil {
		return WebSession{}, "", ErrInvalidToken
	}
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return WebSession{}, "", fmt.Errorf("begin access grant exchange: %w", err)
	}
	defer tx.Rollback()
	grant, err := s.repo.GetAccessGrantBySelector(ctx, tx, selector)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return WebSession{}, "", ErrInvalidToken
		}
		return WebSession{}, "", err
	}
	now := s.now()
	if !s.verifySecret(grant.Selector, secret, grant.TokenHash) {
		return WebSession{}, "", ErrInvalidToken
	}
	if grant.UsedAt != nil {
		return WebSession{}, "", ErrUsed
	}
	if !grant.ExpiresAt.After(now) {
		return WebSession{}, "", ErrExpired
	}
	selector, sessionSecret, err := randomOpaqueParts()
	if err != nil {
		return WebSession{}, "", err
	}
	session := WebSession{UserID: grant.UserID, Selector: selector, TokenHash: s.hashSecret(selector, sessionSecret), ExpiresAt: now.Add(s.sessionTTL), CreatedAt: now, LastSeenAt: now}
	created, err := s.repo.CreateWebSession(ctx, tx, session)
	if err != nil {
		return WebSession{}, "", err
	}
	if err := s.repo.MarkAccessGrantUsed(ctx, tx, grant.ID, now); err != nil {
		return WebSession{}, "", err
	}
	if err := tx.Commit(); err != nil {
		return WebSession{}, "", fmt.Errorf("commit access grant exchange: %w", err)
	}
	return created, created.Selector + "." + sessionSecret, nil
}

func (s *Service) VerifySession(ctx context.Context, rawToken string) (WebSession, error) {
	if !s.Available() {
		return WebSession{}, ErrUnavailable
	}
	selector, secret, err := parseOpaqueToken(rawToken)
	if err != nil {
		return WebSession{}, ErrInvalidToken
	}
	session, err := s.repo.GetWebSessionBySelector(ctx, selector)
	if err != nil {
		if errors.Is(err, ErrNotFound) {
			return WebSession{}, ErrInvalidToken
		}
		return WebSession{}, err
	}
	now := s.now()
	if !s.verifySecret(session.Selector, secret, session.TokenHash) {
		return WebSession{}, ErrInvalidToken
	}
	if session.RevokedAt != nil {
		return WebSession{}, ErrRevoked
	}
	if !session.ExpiresAt.After(now) {
		return WebSession{}, ErrExpired
	}
	_ = s.repo.TouchWebSession(ctx, session.ID, now)
	return session, nil
}

func (s *Service) RevokeSession(ctx context.Context, rawToken string) error {
	session, err := s.VerifySession(ctx, rawToken)
	if err != nil {
		if errors.Is(err, ErrRevoked) || errors.Is(err, ErrExpired) || errors.Is(err, ErrInvalidToken) {
			return nil
		}
		return err
	}
	return s.repo.RevokeWebSession(ctx, session.ID, s.now())
}

func (s *Service) hashSecret(selector, secret string) string {
	mac := hmac.New(sha256.New, s.secret)
	_, _ = mac.Write([]byte(selector))
	_, _ = mac.Write([]byte{0})
	_, _ = mac.Write([]byte(secret))
	return hex.EncodeToString(mac.Sum(nil))
}

func (s *Service) verifySecret(selector, secret, want string) bool {
	got := s.hashSecret(selector, secret)
	return hmac.Equal([]byte(got), []byte(want))
}

func parseOpaqueToken(raw string) (string, string, error) {
	parts := strings.SplitN(strings.TrimSpace(raw), ".", 2)
	if len(parts) != 2 || parts[0] == "" || parts[1] == "" {
		return "", "", ErrInvalidToken
	}
	return parts[0], parts[1], nil
}

func randomOpaqueParts() (string, string, error) {
	selector, err := randomHex(16)
	if err != nil {
		return "", "", err
	}
	secret, err := randomHex(32)
	if err != nil {
		return "", "", err
	}
	return selector, secret, nil
}

func randomHex(n int) (string, error) {
	buf := make([]byte, n)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("random bytes: %w", err)
	}
	return hex.EncodeToString(buf), nil
}
